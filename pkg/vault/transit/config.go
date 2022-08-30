package transit

import (
	"errors"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

func (t *transitAPI) updateTransitEngineConfig(engine core.MountPathEntity, config *EngineConfig) error {
	log := t.Core.Log().WithValues("method", "updateTransitEngineConfig")

	currentConfig, err := t.fetchTransitEngineConfig(engine)
	if err != nil {
		log.Info("failed to fetch current transit engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to fetch current transit engine config").WithCause(err)
	}

	if reflect.DeepEqual(config, currentConfig) {
		return nil
	}

	if err := t.writeTransitEngineConfig(engine, config); err != nil {
		log.Info("failed to write transit engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to write transit engine config").WithCause(err)
	}

	if err := t.Mount.ReloadPluginBackends(mount.PluginTransit); err != nil {
		log.Info("failed to reload plugin backends so changes to the transit cache are live immediately", "error", err)
		return core.ErrAPIError.WithDetails("failed to reload plugin backends so changes to the transit cache are live immediately").WithCause(err)
	}

	return nil
}

type fetchCacheConfigResponse struct {
	Data EngineCacheConfig `json:"data"`
}

func (t *transitAPI) fetchTransitEngineConfig(engine core.MountPathEntity) (*EngineConfig, error) {
	log := t.Core.Log().WithValues("method", "fetchTransitEngineConfig")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	configPath := filepath.Join("/v1", path, "cache-config")

	response := &fetchCacheConfigResponse{}
	if err := t.Core.MakeRequest(core.MethodGet, configPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch transit engine config", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch transit engine config").WithCause(err)
	}

	return &EngineConfig{Cache: response.Data}, nil
}

func (t *transitAPI) writeTransitEngineConfig(engine core.MountPathEntity, config *EngineConfig) error {
	log := t.Core.Log().WithValues("method", "writeTransitEngineConfig")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	if config == nil {
		return nil
	}

	configPath := filepath.Join("/v1", path, "cache-config")
	if err := t.Core.MakeRequest(core.MethodPost, configPath, httpclient.JSON(config.Cache), nil); err != nil {
		log.Info("failed to write transit engine config", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return core.ErrDoesNotExist.WithCause(err)
		}

		return core.ErrAPIError.WithDetails("failed to write transit engine config").WithCause(err)
	}

	return nil
}
