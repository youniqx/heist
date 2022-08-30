package auth

import "github.com/youniqx/heist/pkg/vault/core"

type API interface {
	HasAuthMethod(auth core.MountPathEntity) (bool, error)
	ListAuthMethods() ([]*Method, error)
	DeleteAuthMethod(auth core.MountPathEntity) error
	CreateAuthMethod(auth MethodEntity) error
	ReadAuthMethod(auth core.MountPathEntity) (*Method, error)
}

type Type string

const (
	// MethodKubernetes authenticates all requests with a kubernetes service account token.
	MethodKubernetes Type = "kubernetes"
	// MethodToken authenticates all requests with a vault issued token.
	MethodToken Type = "token"
)

type MethodEntity interface {
	core.MountPathEntity
	GetMethod() (Type, error)
}

type Method struct {
	Path string
	Type Type
}

func (a *Method) GetMountPath() (string, error) {
	return a.Path, nil
}

func (a *Method) GetMethod() (Type, error) {
	return a.Type, nil
}

type authAPI struct {
	Core core.API
}

func NewAPI(core core.API) API {
	return &authAPI{Core: core}
}
