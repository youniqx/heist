package transit

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type listKeysResponse struct {
	Data listKeysData `json:"data"`
}

type listKeysData struct {
	Keys []string `json:"keys"`
}

func (t *transitAPI) ListKeys(engine core.MountPathEntity) ([]KeyName, error) {
	log := t.Core.Log().WithValues("method", "ListKeys")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	keyPath := filepath.Join("/v1", path, "keys")

	response := &listKeysResponse{}
	if err := t.Core.MakeRequest(core.MethodList, keyPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to list keys in transit engine", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		return nil, core.ErrAPIError.WithDetails("failed to list keys in transit engine").WithCause(err)
	}

	result := make([]KeyName, len(response.Data.Keys))
	for index, key := range response.Data.Keys {
		result[index] = KeyName(key)
	}

	return result, nil
}
