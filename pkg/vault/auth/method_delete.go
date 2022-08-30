package auth

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *authAPI) DeleteAuthMethod(auth core.MountPathEntity) error {
	log := a.Core.Log().WithValues("method", "DeleteAuthMethod")

	path, err := auth.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	authPath := filepath.Join("/v1/sys/auth", path)
	if err := a.Core.MakeRequest(core.MethodDelete, authPath, nil, nil); err != nil {
		log.Info("unable to delete auth method", "error", err)
		return core.ErrAPIError.WithDetails("failed to delete auth method").WithCause(err)
	}

	log.Info("successfully deleted auth method")

	return nil
}
