package auth

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type createAuthRequest struct {
	Type Type `json:"type"`
}

func (a *authAPI) CreateAuthMethod(auth MethodEntity) error {
	log := a.Core.Log().WithValues("method", "CreateAuthMethod")

	path, err := auth.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	method, err := auth.GetMethod()
	if err != nil {
		log.Info("failed to get auth method type", "error", err)
		return core.ErrAPIError.WithDetails("failed to get auth method type").WithCause(err)
	}

	log = log.WithValues("type", method)

	authPath := filepath.Join("/v1/sys/auth", path)
	request := &createAuthRequest{
		Type: method,
	}

	if err := a.Core.MakeRequest(core.MethodPost, authPath, httpclient.JSON(request), nil); err != nil {
		log.Info("unable to create auth method", "error", err)
		return core.ErrAPIError.WithDetails("failed to create auth method").WithCause(err)
	}

	log.Info("successfully created auth method")

	return nil
}
