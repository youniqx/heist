package pki

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type CertificateFormat string

const (
	// CertificateFormatPEM defines that the certificate is PEM encoded.
	CertificateFormatPEM CertificateFormat = "pem"
	// CertificateFormatDER defines that the certificate is DER encoded.
	CertificateFormatDER CertificateFormat = "der"
	// CertificateFormatPEMBundle defines that the certificate is a PEM encoded bundle.
	CertificateFormatPEMBundle CertificateFormat = "pem_bundle"
)

type issueCertRequest struct {
	CommonName        string            `json:"common_name"`
	Format            CertificateFormat `json:"format"`
	DNSSans           string            `json:"alt_names,omitempty"`
	OtherSans         string            `json:"other_sans,omitempty"`
	IPSans            string            `json:"ip_sans,omitempty"`
	URISans           string            `json:"uri_sans,omitempty"`
	TTL               core.VaultTTL     `json:"ttl,omitempty"`
	ExcludeCNFromSans bool              `json:"exclude_cn_from_sans,omitempty"`
}

type issueCertResponse struct {
	Data *Certificate `json:"data"`
}

type IssueCertOptions struct {
	CommonName        string
	DNSSans           []string
	OtherSans         []string
	IPSans            []string
	URISans           []string
	TTL               time.Duration
	ExcludeCNFromSans bool
}

func (p *pkiAPI) IssueCertificate(ca core.MountPathEntity, role core.RoleNameEntity, options *IssueCertOptions) (*Certificate, error) {
	log := p.Core.Log().WithValues("method", "IssueCertificate")
	log = log.WithValues(
		"common_name", options.CommonName,
		"dns_sans", options.DNSSans,
		"other_sans", options.OtherSans,
		"ip_sans", options.IPSans,
		"uri_sans", options.URISans,
		"ttl", options.TTL,
		"exclude_cn_from_sans", options.ExcludeCNFromSans,
	)

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	roleName, err := role.GetRoleName()
	if err != nil {
		log.Info("failed to get certificate role name", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get certificate role name").WithCause(err)
	}

	log = log.WithValues("role_name", roleName)

	request := &issueCertRequest{
		CommonName:        options.CommonName,
		Format:            CertificateFormatPEM,
		DNSSans:           strings.Join(options.DNSSans, ","),
		OtherSans:         strings.Join(options.OtherSans, ","),
		IPSans:            strings.Join(options.IPSans, ","),
		URISans:           strings.Join(options.URISans, ","),
		TTL:               core.VaultTTL{TTL: options.TTL},
		ExcludeCNFromSans: options.ExcludeCNFromSans,
	}

	response := &issueCertResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "issue", roleName), httpclient.JSON(request), httpclient.JSON(response)); err != nil {
		log.Info("failed to issue certificate from role", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to issue certificate from role").WithCause(err)
	}

	return response.Data, nil
}
