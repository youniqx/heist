package kvengine

import (
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

type engineAPI struct {
	Core  core.API
	Mount mount.API
}

func NewAPI(coreAPI core.API, mountAPI mount.API) API {
	return &engineAPI{
		Core:  coreAPI,
		Mount: mountAPI,
	}
}

type API interface {
	UpdateKvEngine(engine Entity) error
	ListKvSecrets(engine core.MountPathEntity) ([]core.SecretPath, error)
	ReadKvEngine(engine core.MountPathEntity) (*KvEngine, error)
}

type Config struct {
	MaxVersions        int    `json:"max_versions"`
	CasRequired        bool   `json:"cas_required"`
	DeleteVersionAfter string `json:"delete_version_after"`
}

type Entity interface {
	core.MountPathEntity
	GetKvEngineConfig() (*Config, error)
}

type KvEngine struct {
	Path   string
	Config *Config
}

func (k *KvEngine) GetMountPath() (string, error) {
	return k.Path, nil
}

func (k *KvEngine) GetKvEngineConfig() (*Config, error) {
	return k.Config, nil
}
