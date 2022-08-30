package pki

import (
	"fmt"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type generateRootCARequest struct {
	*Subject
	*CASettings
}

type newCAData struct {
	IssuingCA      string  `json:"issuing_ca"`
	SerialNumber   string  `json:"serial_number"`
	PrivateKey     string  `json:"private_key"`
	PrivateKeyType KeyType `json:"private_key_type"`
}

type generateRootCAResponse struct {
	Data newCAData `json:"data"`
}

//nolint:cyclop
func (p *pkiAPI) CreateRootCA(mode Mode, ca CAEntity) (*CAInfo, error) {
	log := p.Core.Log().WithValues("method", "CreateRootCA")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	switch mode {
	case ModeInternal, ModeExported:
	default:
		log.Info("unknown ca mode setting", "mode", mode)
		return nil, core.ErrAPIError.WithDetails(fmt.Sprintf("unknown ca mode setting: %s", mode))
	}

	log = log.WithValues("mode", mode)

	initialized, err := p.IsPKIEngineInitialized(ca)
	if err != nil {
		log.Info("failed to check if pki engine has been initialized", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to check if pki engine has been initialized").WithCause(err)
	}

	if initialized {
		log.Info("ca already exists, cannot create it again", "error", err)
		return nil, core.ErrAPIError.WithDetails("ca already exists, cannot create it again")
	}

	if err := p.UpdatePKIEngine(ca); err != nil {
		log.Info("failed to update pki engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to update pki engine").WithCause(err)
	}

	importedCert, err := ca.GetImportedCert()
	if err != nil {
		log.Info("failed to get imported cert from ca entity", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get imported cert from ca entity").WithCause(err)
	}

	data, err := p.importOrCreateRootCA(ca, importedCert, mode)
	if err != nil {
		log.Info("failed to import or create cert root ca pki engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to import or create cert root ca pki engine").WithCause(err)
	}

	log = log.WithValues(
		"issuing_ca", data.IssuingCA,
		"serial_number", data.SerialNumber,
		"private_key_exported", data.PrivateKey != "",
	)

	chain, err := p.ReadCACertificateChain(ca)
	if err != nil {
		log.Info("failed to read ca certificate chain", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to read ca certificate chain").WithCause(err)
	}

	log = log.WithValues("cert_chain", chain)

	cert, err := p.ReadCACertificatePEM(ca)
	if err != nil {
		log.Info("failed to read ca certificate pem", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to read ca certificate pem").WithCause(err)
	}

	log = log.WithValues("certificate", cert)

	urls := &CertificateURLs{
		IssuingCertificates:   []string{p.Core.GetVaultAddress("v1", path, "ca")},
		CrlDistributionPoints: []string{p.Core.GetVaultAddress("v1", path, "crl")},
		OcspServers:           nil,
	}

	log = log.WithValues(
		"issuing_certificates", urls.IssuingCertificates,
		"crl_distribution_points", urls.CrlDistributionPoints,
		"ocsp_servers", urls.OcspServers,
	)

	if err := p.SetCertificateURLs(ca, urls); err != nil {
		log.Info("failed to set certificate urls", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to set certificate urls").WithCause(err)
	}

	log.Info("Created new root CA")

	return &CAInfo{
		Path:                        path,
		CertificateChain:            chain,
		IssuingCertificateAuthority: data.IssuingCA,
		SerialNumber:                data.SerialNumber,
		PrivateKey:                  data.PrivateKey,
		PrivateKeyType:              data.PrivateKeyType,
		Certificate:                 cert,
	}, nil
}

func (p *pkiAPI) importOrCreateRootCA(ca CAEntity, importedCert *ImportedCert, mode Mode) (*newCAData, error) {
	log := p.Core.Log().WithValues("method", "importOrCreateRootCA")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	if importedCert != nil {
		parsed, err := parseImportedCert(importedCert)
		if err != nil {
			log.Info("failed to parse cert to import", "error", err)
			return nil, core.ErrAPIError.WithDetails("failed to parse imported cert").WithCause(err)
		}

		if err := p.ImportCert(ca, importedCert); err != nil {
			log.Info("failed to import cert into root pki engine", "error", err)
			return nil, core.ErrAPIError.WithDetails("failed to import root ca in pki engine").WithCause(err)
		}

		var (
			privateKey     string
			privateKeyType KeyType
		)

		if mode == ModeExported {
			privateKey = parsed.PrivateKey
			privateKeyType = parsed.PrivateKeyType
		}

		return &newCAData{
			IssuingCA:      importedCert.Certificate,
			SerialNumber:   formatSerialNumber(parsed.Certificate.SerialNumber),
			PrivateKey:     privateKey,
			PrivateKeyType: privateKeyType,
		}, nil
	}

	settings, err := ca.GetSettings()
	if err != nil {
		log.Info("failed to get settings from ca entity", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get settings form ca entity").WithCause(err)
	}

	subject, err := ca.GetSubject()
	if err != nil {
		log.Info("failed to get subject from ca entity", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get subject from ca entity").WithCause(err)
	}

	request := &generateRootCARequest{
		Subject:    subject,
		CASettings: settings,
	}
	response := &generateRootCAResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "root", "generate", string(mode)), httpclient.JSON(request), httpclient.JSON(response)); err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to generate root ca in pki engine").WithCause(err)
	}

	return &response.Data, nil
}
