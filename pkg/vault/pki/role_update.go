package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type updateRoleRequest struct {
	*RoleSettings
	*SubjectSettings
}

func (p *pkiAPI) UpdateCertificateRole(ca core.MountPathEntity, role CertificateRoleEntity) error {
	log := p.Core.Log().WithValues("method", "UpdateRole")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	roleName, err := role.GetRoleName()
	if err != nil {
		log.Info("failed to get certificate role name", "error", err)
		return core.ErrAPIError.WithDetails("failed to get certificate role name").WithCause(err)
	}

	log = log.WithValues("role_name", roleName)

	settings, err := role.GetSettings()
	if err != nil {
		log.Info("failed to get settings from role entity", "error", err)
		return core.ErrAPIError.WithDetails("failed to get settings form role entity").WithCause(err)
	}

	subject, err := role.GetSubject()
	if err != nil {
		log.Info("failed to get subject from role entity", "error", err)
		return core.ErrAPIError.WithDetails("failed to get subject from role entity").WithCause(err)
	}

	request := &updateRoleRequest{
		RoleSettings:    settings,
		SubjectSettings: subject,
	}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "roles", roleName), httpclient.JSON(request), nil); err != nil {
		return core.ErrAPIError.WithDetails("failed to update role in pki engine").WithCause(err)
	}

	return nil
}
