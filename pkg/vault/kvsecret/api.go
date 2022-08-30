package kvsecret

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

type api struct {
	Core core.API
}

func NewAPI(core core.API) API {
	return &api{Core: core}
}

type API interface {
	UpdateKvSecret(engine core.MountPathEntity, secret Entity) error
	DeleteKvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) error
	ReadKvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) (*KvSecret, error)
}

type Entity interface {
	core.SecretPathEntity
	GetFields() (map[string]string, error)
}

type KvSecret struct {
	Path   string
	Fields map[string]string
}

func (k *KvSecret) GetSecretPath() (string, error) {
	return k.Path, nil
}

func (k *KvSecret) GetFields() (map[string]string, error) {
	return k.Fields, nil
}
