package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type readRoleResponse struct {
	Data struct {
		*SubjectSettings
		*RoleSettings
	} `json:"data"`
}

func (p *pkiAPI) ReadCertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) (*CertificateRole, error) {
	log := p.Core.Log().WithValues("method", "ReadCertificateRole")

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

	response := &readRoleResponse{}

	if err := p.Core.MakeRequest(core.MethodGet, filepath.Join("/v1", path, "roles", roleName), nil, httpclient.JSON(response)); err != nil {
		log.Info("failed to read cert role from pki engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to read cert role from pki engine").WithCause(err)
	}

	return &CertificateRole{
		Name:     roleName,
		Settings: response.Data.RoleSettings,
		Subject:  response.Data.SubjectSettings,
	}, nil
}
