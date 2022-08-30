package kubernetesauth

import (
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
)

type RoleEntity interface {
	core.RoleNameEntity
	core.RolePoliciesEntity
	GetBoundNamespaces() ([]string, error)
	GetBoundServiceAccounts() ([]string, error)
}

type MethodEntity interface {
	core.MountPathEntity
	GetMethodConfig() (*Config, error)
}

type API interface {
	UpdateKubernetesAuthMethod(method MethodEntity) error
	UpdateKubernetesAuthRole(method core.MountPathEntity, role RoleEntity) error
	DeleteKubernetesAuthRole(method core.MountPathEntity, role core.RoleNameEntity) error
	ReadKubernetesAuthRole(method core.MountPathEntity, role core.RoleNameEntity) (*Role, error)
	LoginWithKubernetesAuth(method core.MountPathEntity, role core.RoleNameEntity, jwt string) (*core.AuthResponse, error)
}

type Role struct {
	Name                 string
	Policies             []core.PolicyName
	BoundNamespaces      []string
	BoundServiceAccounts []string
}

func (k *Role) GetRoleName() (string, error) {
	return k.Name, nil
}

func (k *Role) GetRolePolicies() ([]core.PolicyName, error) {
	return k.Policies, nil
}

func (k *Role) GetBoundNamespaces() ([]string, error) {
	return k.BoundNamespaces, nil
}

func (k *Role) GetBoundServiceAccounts() ([]string, error) {
	return k.BoundServiceAccounts, nil
}

type Method struct {
	Path   string
	Config *Config
}

func (k *Method) GetMethodConfig() (*Config, error) {
	return k.Config, nil
}

func (k *Method) GetMountPath() (string, error) {
	return k.Path, nil
}

type kubernetesAuthAPI struct {
	Core core.API
	Auth auth.API
}

func NewAPI(coreAPI core.API, authAPI auth.API) API {
	return &kubernetesAuthAPI{
		Core: coreAPI,
		Auth: authAPI,
	}
}
