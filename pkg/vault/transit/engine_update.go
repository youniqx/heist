package transit

import (
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

func (t *transitAPI) UpdateTransitEngine(engine EngineEntity) error {
	log := t.Core.Log().WithValues("method", "UpdateTransitEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	exists, err := t.Mount.HasEngine(engine)
	if err != nil {
		log.Info("failed to check if engine exists", "error", err)
		return core.ErrAPIError.WithDetails("failed to check if engine exists").WithCause(err)
	}

	log = log.WithValues("exists", exists)

	pluginName, err := engine.GetPluginName()
	if err != nil {
		log.Info("failed to fetch plugin name", "error", err)
		return core.ErrAPIError.WithDetails("failed to fetch plugin name").WithCause(err)
	}

	if pluginName == "" {
		pluginName = string(mount.TypeTransit)
	}

	if !exists {
		mountRequest := &mount.Mount{
			Path: path,
			Type: mount.Type(pluginName),
		}

		log.Info("creating new transit engine")

		if err := t.Mount.MountEngine(mountRequest); err != nil {
			log.Info("failed to create transit engine", "error", err)
			return core.ErrAPIError.WithDetails("failed to create transit engine").WithCause(err)
		}
	}

	config, err := engine.GetTransitEngineConfig()
	if err != nil {
		log.Info("failed to get desired transit engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to fetch desired transit engine config").WithCause(err)
	}

	if err := t.updateTransitEngineConfig(engine, config); err != nil {
		log.Info("failed to update transit engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to update transit engine config").WithCause(err)
	}

	return nil
}
