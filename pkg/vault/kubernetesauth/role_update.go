package kubernetesauth

import (
	"errors"
	"path/filepath"
	"reflect"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type createRoleRequest struct {
	BoundServiceAccountNames      []string          `json:"bound_service_account_names"`
	BoundServiceAccountNamespaces []string          `json:"bound_service_account_namespaces"`
	Policies                      []core.PolicyName `json:"policies"`
}

//nolint:cyclop
func (k *kubernetesAuthAPI) UpdateKubernetesAuthRole(method core.MountPathEntity, role RoleEntity) error {
	log := k.Core.Log().WithValues("method", "UpdateKubernetesAuthRole")

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

	rolePolicies, err := role.GetRolePolicies()
	if err != nil {
		log.Info("failed to get k8s role policies", "error", err)
		return core.ErrAPIError.WithDetails("failed to get k8s role policies").WithCause(err)
	}

	log = log.WithValues("policies", rolePolicies)

	boundNamespaces, err := role.GetBoundNamespaces()
	if err != nil {
		log.Info("failed to get k8s role bound namespaces", "error", err)
		return core.ErrAPIError.WithDetails("failed to list bound namespaces").WithCause(err)
	}

	log = log.WithValues("boundNamespaces", boundNamespaces)

	boundServiceAccounts, err := role.GetBoundServiceAccounts()
	if err != nil {
		log.Info("failed to get k8s role bound service accounts", "error", err)
		return core.ErrAPIError.WithDetails("failed to list bound service accounts").WithCause(err)
	}

	log = log.WithValues("boundServiceAccounts", boundServiceAccounts)

	desiredRole := &Role{
		Name:                 roleName,
		Policies:             rolePolicies,
		BoundNamespaces:      boundNamespaces,
		BoundServiceAccounts: boundServiceAccounts,
	}

	var updateRequired bool

	switch currentRole, err := k.ReadKubernetesAuthRole(method, role); {
	case errors.Is(err, core.ErrDoesNotExist):
		updateRequired = true
	case err == nil:
		updateRequired = !reflect.DeepEqual(currentRole, desiredRole)
	default:
		log.Info("failed to check current state of the k8s role", "error", err)
		return core.ErrAPIError.WithDetails("failed to check current state of k8s role").WithCause(err)
	}

	if !updateRequired {
		return nil
	}

	rolePath := filepath.Join("/v1/auth", path, "role", roleName)
	request := &createRoleRequest{
		BoundServiceAccountNames:      boundServiceAccounts,
		BoundServiceAccountNamespaces: boundNamespaces,
		Policies:                      rolePolicies,
	}

	if err := k.Core.MakeRequest(core.MethodPost, rolePath, httpclient.JSON(request), nil); err != nil {
		log.Info("failed to create k8s auth role", "error", err)
		return core.ErrAPIError.WithDetails("failed to create k8s auth role").WithCause(err)
	}

	return nil
}
