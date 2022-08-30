package kubernetesauth

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

type loginRequest struct {
	Role string `json:"role"`
	JWT  string `json:"jwt"`
}

func (k *kubernetesAuthAPI) LoginWithKubernetesAuth(method core.MountPathEntity, role core.RoleNameEntity, jwt string) (*core.AuthResponse, error) {
	log := k.Core.Log().WithValues("method", "LoginWithKubernetesAuth")

	roleName, err := role.GetRoleName()
	if err != nil {
		log.Info("failed to get role name", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get k8s role name").WithCause(err)
	}

	provider := &authProvider{
		Method: method,
		Role:   core.Value(roleName),
		JWT:    core.Value(jwt),
	}

	return provider.Authenticate(k.Core)
}
