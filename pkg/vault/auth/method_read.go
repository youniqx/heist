package auth

import "github.com/youniqx/heist/pkg/vault/core"

func (a *authAPI) ReadAuthMethod(auth core.MountPathEntity) (*Method, error) {
	log := a.Core.Log().WithValues("method", "ReadAuthMethod")

	path, err := auth.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	methods, err := a.ListAuthMethods()
	if err != nil {
		log.Info("failed to fetch auth method", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to list auth methods").WithCause(err)
	}

	for _, method := range methods {
		if method.Path == path {
			return method, nil
		}
	}

	return nil, core.ErrDoesNotExist
}
