package mount

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *mountAPI) ReadTuneConfig(engine core.MountPathEntity) (*TuneConfig, error) {
	log := a.Core.Log().WithValues("method", "MountEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	config := &TuneConfig{}

	if err := a.Core.MakeRequest(core.MethodGet, filepath.Join("/v1/sys/mounts", path, "tune"), nil, httpclient.JSON(config)); err != nil {
		log.Info("failed to read tune configuration from engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to read tune configuration from engine").WithCause(err)
	}

	return config, nil
}
