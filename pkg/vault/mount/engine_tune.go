package mount

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *mountAPI) TuneEngine(engine core.MountPathEntity, config *TuneConfig) error {
	log := a.Core.Log().WithValues("method", "MountEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	if err := a.Core.MakeRequest(core.MethodPost, filepath.Join("/v1/sys/mounts", path, "tune"), httpclient.JSON(config), nil); err != nil {
		log.Info("failed to tune engine", "error", err)
		return core.ErrAPIError.WithDetails("failed to tune engine").WithCause(err)
	}

	return nil
}
