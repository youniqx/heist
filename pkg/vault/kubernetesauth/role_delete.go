package kubernetesauth

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (k *kubernetesAuthAPI) DeleteKubernetesAuthRole(method core.MountPathEntity, role core.RoleNameEntity) error {
	log := k.Core.Log().WithValues("method", "DeleteKubernetesAuthRole")

	path, err := method.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	roleName, err := role.GetRoleName()
	if err != nil {
		log.Info("failed to get k8s role name", "error", err)
		return core.ErrAPIError.WithDetails("failed to get k8s role name").WithCause(err)
	}

	log = log.WithValues("role", roleName)

	deletePath := filepath.Join("/v1/auth", path, "role", roleName)
	if err := k.Core.MakeRequest(core.MethodDelete, deletePath, nil, nil); err != nil {
		log.Info("failed to delete role", "error", err)
		return core.ErrAPIError.WithDetails("failed to delete k8s auth role").WithCause(err)
	}

	return nil
}
