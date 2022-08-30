package kubernetesauth

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type fetchRoleResponse struct {
	Data fetchRoleResponseData `json:"data"`
}

type fetchRoleResponseData struct {
	BoundServiceAccountNames      []string          `json:"bound_service_account_names"`
	BoundServiceAccountNamespaces []string          `json:"bound_service_account_namespaces"`
	Policies                      []core.PolicyName `json:"policies"`
}

func (k *kubernetesAuthAPI) ReadKubernetesAuthRole(method core.MountPathEntity, role core.RoleNameEntity) (*Role, error) {
	log := k.Core.Log().WithValues("method", "ReadKubernetesAuthRole")

	path, err := method.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	roleName, err := role.GetRoleName()
	if err != nil {
		log.Info("failed to get k8s role name", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get k8s role name").WithCause(err)
	}

	log = log.WithValues("role", roleName)

	fetchPath := filepath.Join("/v1/auth", path, "role", roleName)
	response := &fetchRoleResponse{}

	if err := k.Core.MakeRequest(core.MethodGet, fetchPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch k8s auth role", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch k8s auth role").WithCause(err)
	}

	return &Role{
		Name:                 roleName,
		Policies:             response.Data.Policies,
		BoundNamespaces:      response.Data.BoundServiceAccountNamespaces,
		BoundServiceAccounts: response.Data.BoundServiceAccountNames,
	}, nil
}
