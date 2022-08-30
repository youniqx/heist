package mount

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type mountRequest struct {
	Type    Type              `json:"type"`
	Config  *TuneConfig       `json:"config"`
	Options map[string]string `json:"options"`
}

type mountRequestConfig struct {
	Options map[string]string `json:"options" mapstructure:"options"`
}

func (a *mountAPI) MountEngine(engine Entity) error {
	log := a.Core.Log().WithValues("method", "MountEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	engineType, err := engine.GetMountType()
	if err != nil {
		log.Info("failed to get engine type", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine type").WithCause(err)
	}

	log = log.WithValues("type", engineType)

	mountConfig, err := engine.GetMountConfig()
	if err != nil {
		log.Info("failed to get mount config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get mount config").WithCause(err)
	}

	mountOptions, err := engine.GetMountOptions()
	if err != nil {
		log.Info("failed to get mount options", "error", err)
		return core.ErrAPIError.WithDetails("failed to get mount options").WithCause(err)
	}

	log = log.WithValues("config", mountConfig)

	mountPath := filepath.Join("/v1/sys/mounts", path)
	request := &mountRequest{
		Type:    engineType,
		Config:  mountConfig,
		Options: mountOptions,
	}

	if err := a.Core.MakeRequest(core.MethodPost, mountPath, httpclient.JSON(request), nil); err != nil {
		log.Info("unable to mount engine", "error", err)
		return core.ErrAPIError.WithDetails("failed to mount engine").WithCause(err)
	}

	log.Info("successfully mounted secret engine")

	return nil
}
