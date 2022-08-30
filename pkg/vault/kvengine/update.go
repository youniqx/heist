package kvengine

import (
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

func (a *engineAPI) UpdateKvEngine(engine Entity) error {
	log := a.Core.Log().WithValues("method", "UpdateKvEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	exists, err := a.Mount.HasEngine(engine)
	if err != nil {
		log.Info("failed to check if engine exists", "error", err)
		return core.ErrAPIError.WithDetails("failed to check if engine exists").WithCause(err)
	}

	log = log.WithValues("exists", exists)

	if !exists {
		mountRequest := &mount.Mount{
			Path: path,
			Type: mount.TypeKVV2,
			Options: map[string]string{
				"version": "2",
			},
		}

		log.Info("creating new kv engine")

		if err := a.Mount.MountEngine(mountRequest); err != nil {
			log.Info("failed to create kv engine", "error", err)
			return core.ErrAPIError.WithDetails("failed to create kv engine").WithCause(err)
		}
	}

	if err := a.updateKvSecretEngineConfig(engine); err != nil {
		log.Info("failed to update kv engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to update kv engine config").WithCause(err)
	}

	return nil
}
