package kvengine

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *engineAPI) ReadKvEngine(engine core.MountPathEntity) (*KvEngine, error) {
	log := a.Core.Log().WithValues("method", "ReadKvEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	config, err := a.fetchKvSecretEngineConfig(engine)
	if err != nil {
		log.Info("failed to fetch engine config", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to fetch engine config").WithCause(err)
	}

	return &KvEngine{
		Path:   path,
		Config: config,
	}, nil
}
