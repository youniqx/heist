package transit

import (
	"errors"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type createKeyRequest struct {
	Type KeyType `json:"type"`
}

//nolint:cyclop
func (t *transitAPI) UpdateTransitKey(engine core.MountPathEntity, key KeyEntity) error {
	log := t.Core.Log().WithValues("method", "UpdateTransitKey")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	keyName, err := key.GetTransitKeyName()
	if err != nil {
		log.Info("failed to get transit key name", "error", err)
		return core.ErrAPIError.WithDetails("failed to get transit key name").WithCause(err)
	}

	log = log.WithValues("key", keyName)

	keyType, err := key.GetTransitKeyType()
	if err != nil {
		log.Info("failed to get transit key type", "error", err)
		return core.ErrAPIError.WithDetails("failed to get transit key type").WithCause(err)
	}

	log = log.WithValues("type", keyType)

	keyConfig, err := key.GetTransitKeyConfig()
	if err != nil {
		log.Info("failed to get transit key config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get transit key config").WithCause(err)
	}

	log = log.WithValues("config", keyConfig)

	var (
		keyCreationRequired  bool
		configUpdateRequired bool
	)

	switch currentKey, err := t.ReadTransitKey(engine, key); {
	case errors.Is(err, core.ErrDoesNotExist):
		keyCreationRequired = true
		configUpdateRequired = true
	case err == nil:
		if currentKey.Type != keyType {
			return core.ErrAPIError.WithDetails("a key with this name but different type already exists in the engine, key type is immutable after creation")
		}

		keyCreationRequired = false
		configUpdateRequired = !reflect.DeepEqual(currentKey.Config, keyConfig)
	default:
		log.Info("failed to check current state of the transit key", "error", err)
		return core.ErrAPIError.WithDetails("failed to check current state transit key").WithCause(err)
	}

	if keyCreationRequired {
		createPath := filepath.Join("/v1", path, "keys", keyName)
		request := &createKeyRequest{
			Type: keyType,
		}

		if err := t.Core.MakeRequest(core.MethodPost, createPath, httpclient.JSON(request), nil); err != nil {
			log.Info("failed to create new transit key", "error", err)
			return core.ErrAPIError.WithDetails("failed to create new transit key").WithCause(err)
		}
	}

	if configUpdateRequired {
		configPath := filepath.Join("/v1", path, "keys", keyName, "config")
		if err := t.Core.MakeRequest(core.MethodPost, configPath, httpclient.JSON(keyConfig), nil); err != nil {
			log.Info("failed to update transit key configuration", "error", err)

			var responseError *core.VaultHTTPError
			if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
				return core.ErrDoesNotExist.WithCause(err)
			}

			return core.ErrAPIError.WithDetails("failed to update transit key config").WithCause(err)
		}
	}

	return nil
}
