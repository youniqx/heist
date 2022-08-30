package transit

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type readKeyResponse struct {
	Data readKeyData
}

type readKeyData struct {
	Name       string  `json:"name"`
	Type       KeyType `json:"type"`
	*KeyConfig `json:",inline"`
}

func (t *transitAPI) ReadTransitKey(engine core.MountPathEntity, key KeyNameEntity) (*Key, error) {
	log := t.Core.Log().WithValues("method", "ReadTransitKey")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	keyName, err := key.GetTransitKeyName()
	if err != nil {
		log.Info("failed to get transit key name", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get transit key name").WithCause(err)
	}

	log = log.WithValues("key", keyName)

	configPath := filepath.Join("/v1", path, "keys", keyName)

	response := &readKeyResponse{}
	if err := t.Core.MakeRequest(core.MethodGet, configPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch transit key config", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch transit key config").WithCause(err)
	}

	return &Key{
		Name:   response.Data.Name,
		Type:   response.Data.Type,
		Config: response.Data.KeyConfig,
	}, nil
}
