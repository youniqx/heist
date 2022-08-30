package pki

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/core"
)

//nolint:cyclop
func (p *pkiAPI) CreateIntermediateCA(mode Mode, issuer core.MountPathEntity, ca CAEntity) (*CAInfo, error) {
	log := p.Core.Log().WithValues("method", "CreateIntermediateCA")

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

	data, err := p.importOrCreateIntermediateCA(issuer, ca, mode, importedCert)
	if err != nil {
		log.Info("failed to get import or create intermediate ca pki engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get import or create intermediate ca pki engine").WithCause(err)
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

	log.Info("Created new intermediate CA")

	return &CAInfo{
		Path:                        path,
		SerialNumber:                data.SerialNumber,
		PrivateKey:                  data.PrivateKey,
		PrivateKeyType:              data.PrivateKeyType,
		IssuingCertificateAuthority: data.IssuingCA,
		CertificateChain:            chain,
		Certificate:                 cert,
	}, nil
}

func (p *pkiAPI) importOrCreateIntermediateCA(issuer core.MountPathEntity, ca CAEntity, mode Mode, importedCert *ImportedCert) (*newCAData, error) {
	log := p.Core.Log().WithValues("method", "importOrCreateIntermediateCA")
	if importedCert != nil {
		parsed, err := parseImportedCert(importedCert)
		if err != nil {
			return nil, core.ErrAPIError.WithDetails("failed to parse imported cert").WithCause(err)
		}

		issuerPEM, err := p.ReadCACertificatePEM(issuer)
		if err != nil {
			log.Info("failed to read issuer cert PEM", "error", err)
			return nil, core.ErrAPIError.WithDetails("failed to read issuer cert PEM").WithCause(err)
		}

		if err := p.ImportCert(ca, importedCert); err != nil {
			return nil, core.ErrAPIError.WithDetails("failed to import intermediate ca in pki engine").WithCause(err)
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
			IssuingCA:      issuerPEM,
			SerialNumber:   formatSerialNumber(parsed.Certificate.SerialNumber),
			PrivateKey:     privateKey,
			PrivateKeyType: privateKeyType,
		}, nil
	}

	csrInfo, err := p.GenerateIntermediateCSR(mode, ca)
	if err != nil {
		log.Info("failed to generated intermediate csr", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to generate csr for intermediate ca").WithCause(err)
	}

	log = log.WithValues("csr", csrInfo.CSR)

	signResult, err := p.SignIntermediateCSR(issuer, ca, csrInfo.CSR)
	if err != nil {
		log.Info("failed to generated intermediate csr", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to sign csr of intermediate ca").WithCause(err)
	}

	if err := p.SetIntermediateCACert(ca, signResult.Certificate); err != nil {
		log.Info("failed to generated intermediate csr", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to set signed intermediate cert").WithCause(err)
	}

	return &newCAData{
		IssuingCA:      signResult.IssuingCA,
		SerialNumber:   signResult.SerialNumber,
		PrivateKey:     csrInfo.PrivateKey,
		PrivateKeyType: csrInfo.PrivateKeyType,
	}, nil
}
