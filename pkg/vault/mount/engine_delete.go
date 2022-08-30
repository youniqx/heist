package mount

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *mountAPI) DeleteEngine(engine core.MountPathEntity) error {
	log := a.Core.Log().WithValues("method", "DeleteEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	log.Info("trying to delete secret engine")

	mountPath := filepath.Join("/v1/sys/mounts", path)
	if err := a.Core.MakeRequest(core.MethodDelete, mountPath, nil, nil); err != nil {
		log.Info("unable to delete engine", "error", err)
		return core.ErrAPIError.WithDetails("failed to delete engine").WithCause(err)
	}

	log.Info("secret engine deleted")

	return nil
}
