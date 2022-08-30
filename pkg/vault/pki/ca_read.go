package pki

import (
	"crypto/x509"
	"encoding/pem"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) ReadCA(ca core.MountPathEntity) (*CA, error) {
	log := p.Core.Log().WithValues("method", "ReadCA")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	pemData, err := p.ReadCACertificatePEM(ca)
	if err != nil {
		log.Info("failed to read ca certificate", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to read ca certificate").WithCause(err)
	}

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		log.Info("failed to decode ca certificate PEM", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to decode ca certificate PEM").WithCause(err)
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Info("failed to parse ca certificate", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to parse ca certificate").WithCause(err)
	}

	tuneConfig, err := p.Mount.ReadTuneConfig(ca)
	if err != nil {
		log.Info("failed to read pki tune config", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to pki tune config").WithCause(err)
	}

	return &CA{
		Path:     path,
		Settings: nil,
		Subject: &Subject{
			CommonName:   certificate.Subject.CommonName,
			SerialNumber: certificate.Subject.SerialNumber,
			SubjectSettings: &SubjectSettings{
				Organization:       certificate.Subject.Organization,
				OrganizationalUnit: certificate.Subject.OrganizationalUnit,
				Country:            certificate.Subject.Country,
				Locality:           certificate.Subject.Locality,
				Province:           certificate.Subject.Province,
				StreetAddress:      certificate.Subject.StreetAddress,
				PostalCode:         certificate.Subject.PostalCode,
			},
		},
		Config: tuneConfig,
	}, nil
}
