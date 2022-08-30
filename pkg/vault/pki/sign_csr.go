package pki

import (
	"path/filepath"
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type signCsrRequest struct {
	CSR              string        `json:"csr,omitempty"`
	CommonName       string        `json:"common_name,omitempty"`
	AlternativeNames string        `json:"alt_names,omitempty"`
	OtherSans        string        `json:"other_sans,omitempty"`
	IPSans           string        `json:"ip_sans,omitempty"`
	URISans          string        `json:"uri_sans,omitempty"`
	TTL              core.VaultTTL `json:"ttl,omitempty"`
}

type signCsrResponse struct {
	Data *Certificate `json:"data"`
}

func (p *pkiAPI) SignCertificateSigningRequest(ca core.MountPathEntity, role core.RoleNameEntity, request *SignCsr) (*Certificate, error) {
	log := p.Core.Log().WithValues("method", "IssueCertificate")
	log = log.WithValues("csr", request.CSR)

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

	csrRequest := &signCsrRequest{
		CSR:              request.CSR,
		CommonName:       request.CommonName,
		AlternativeNames: strings.Join(request.AlternativeNames, ","),
		OtherSans:        strings.Join(request.OtherSans, ","),
		IPSans:           strings.Join(request.IPSans, ","),
		URISans:          strings.Join(request.URISans, ","),
		TTL:              core.VaultTTL{TTL: request.TTL},
	}

	csrResponse := &signCsrResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "sign", roleName), httpclient.JSON(csrRequest), httpclient.JSON(csrResponse)); err != nil {
		log.Info("failed to issue certificate from role", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to issue certificate from role").WithCause(err)
	}

	return csrResponse.Data, nil
}
