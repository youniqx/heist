package pki

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) DeleteCertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) error {
	log := p.Core.Log().WithValues("method", "DeleteCertificateRole")

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

	if err := p.Core.MakeRequest(core.MethodDelete, filepath.Join("/v1", path, "roles", roleName), nil, nil); err != nil {
		log.Info("failed to delete cert role from pki engine", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil
		}

		return core.ErrAPIError.WithDetails("failed to delete cert role from pki engine").WithCause(err)
	}

	return nil
}
