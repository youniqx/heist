package auth

import "github.com/youniqx/heist/pkg/vault/core"

func (a *authAPI) HasAuthMethod(auth core.MountPathEntity) (bool, error) {
	log := a.Core.Log().WithValues("method", "HasAuthMethod")

	path, err := auth.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	methods, err := a.ListAuthMethods()
	if err != nil {
		log.Info("failed to list auth methods", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to list auth methods").WithCause(err)
	}

	for _, method := range methods {
		if method.Path == path {
			return true, nil
		}
	}

	return false, nil
}
