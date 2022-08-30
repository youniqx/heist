package kubernetesauth

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func AuthProvider(method core.MountPathEntity, role core.StringSource, jwt core.StringSource) core.AuthProvider {
	return &authProvider{
		Method: method,
		Role:   role,
		JWT:    jwt,
	}
}

type authProvider struct {
	Method core.MountPathEntity
	Role   core.StringSource
	JWT    core.StringSource
}

func (a *authProvider) Authenticate(api core.API) (*core.AuthResponse, error) {
	log := api.Log().WithValues("method", "LoginWithKubernetesAuth")

	path, err := a.Method.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	roleName, err := a.Role.FetchStringValue()
	if err != nil {
		log.Info("failed to get k8s role name", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to fetch k8s role name").WithCause(err)
	}

	log = log.WithValues("role", roleName)

	jwt, err := a.JWT.FetchStringValue()
	if err != nil {
		log.Info("failed to fetch jwt", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to fetch jwt").WithCause(err)
	}

	loginPath := filepath.Join("/v1/auth", path, "login")
	request := &loginRequest{
		Role: roleName,
		JWT:  jwt,
	}
	response := &core.AuthResponse{}

	if err := api.MakeRequest(core.MethodPost, loginPath, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to login using kubernetes authentication", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to login using kubernetes authentication").WithCause(err)
	}

	return response, nil
}
