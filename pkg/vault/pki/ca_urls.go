package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type CertificateURLs struct {
	IssuingCertificates   []string `json:"issuing_certificates"`
	CrlDistributionPoints []string `json:"crl_distribution_points"`
	OcspServers           []string `json:"ocsp_servers"`
}

func (p *pkiAPI) SetCertificateURLs(ca core.MountPathEntity, urls *CertificateURLs) error {
	log := p.Core.Log().WithValues("method", "SetCertificateURLs")

	log = log.WithValues(
		"issuing_certificates", urls.IssuingCertificates,
		"crl_distribution_points", urls.CrlDistributionPoints,
		"ocsp_servers", urls.OcspServers,
	)

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "config", "urls"), httpclient.JSON(urls), nil); err != nil {
		log.Info("failed to set certificate urls", "error", err)
		return core.ErrAPIError.WithDetails("failed to generate root ca in pki engine").WithCause(err)
	}

	return nil
}
