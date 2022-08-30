package transit

import "github.com/youniqx/heist/pkg/vault/core"

func (t *transitAPI) ReadTransitEngine(engine core.MountPathEntity) (*Engine, error) {
	log := t.Core.Log().WithValues("method", "ReadTransitEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	config, err := t.fetchTransitEngineConfig(engine)
	if err != nil {
		log.Info("failed to fetch current transit engine config", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to fetch transit engine").WithCause(err)
	}

	return &Engine{
		Path:   path,
		Config: config,
	}, nil
}
