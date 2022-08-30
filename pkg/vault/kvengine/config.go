package kvengine

import (
	"errors"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *engineAPI) updateKvSecretEngineConfig(engine Entity) error {
	log := a.Core.Log().WithValues("method", "updateKvSecretEngineConfig")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	log.Info("checking if kv engine config needs updates")

	currentConfig, err := a.fetchKvSecretEngineConfig(engine)
	if err != nil {
		log.Info("failed to fetch engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to fetch kv engine config").WithCause(err)
	}

	desiredConfig, err := engine.GetKvEngineConfig()
	if err != nil {
		log.Info("failed to get desired engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get desired engine config").WithCause(err)
	}

	if reflect.DeepEqual(currentConfig, desiredConfig) {
		log.Info("kv engine config is up to date")
		return nil
	}

	log.Info("kv engine needs updates, writing desired config to vault", "currentConfig", currentConfig, "desiredConfig", desiredConfig)

	if err := a.writeKvSecretEngineConfig(engine, desiredConfig); err != nil {
		log.Info("failed to update kv engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get write engine config").WithCause(err)
	}

	return nil
}

type kvEngineConfigResponse struct {
	Data *Config `json:"data"`
}

func (a *engineAPI) fetchKvSecretEngineConfig(engine core.MountPathEntity) (*Config, error) {
	log := a.Core.Log().WithValues("method", "fetchKvSecretEngineConfig")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	configPath := filepath.Join("/v1", path, "config")

	response := &kvEngineConfigResponse{}
	if err := a.Core.MakeRequest(core.MethodGet, configPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch kv engine config", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch kv engine config").WithCause(err)
	}

	return response.Data, nil
}

func (a *engineAPI) writeKvSecretEngineConfig(engine Entity, config *Config) error {
	log := a.Core.Log().WithValues("method", "writeKvSecretEngineConfig")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	configPath := filepath.Join("/v1", path, "config")
	if err := a.Core.MakeRequest(core.MethodPost, configPath, httpclient.JSON(config), nil); err != nil {
		log.Info("couldn't write kv engine config", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return core.ErrDoesNotExist.WithCause(err)
		}

		return core.ErrAPIError.WithDetails("failed to write kv engine config").WithCause(err)
	}

	return nil
}
