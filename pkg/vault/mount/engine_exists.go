package mount

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *mountAPI) HasEngine(engine core.MountPathEntity) (bool, error) {
	log := a.Core.Log().WithValues("method", "HasEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	mounts, err := a.ListMounts()
	if err != nil {
		log.Info("error during mount list", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to list mounts").WithCause(err)
	}

	for _, mount := range mounts {
		if mount.Path == path {
			log.Info("secret engine already exists")
			return true, nil
		}
	}

	return false, nil
}
