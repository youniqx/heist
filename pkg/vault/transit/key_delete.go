package transit

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (t *transitAPI) DeleteTransitKey(engine core.MountPathEntity, key KeyNameEntity) error {
	log := t.Core.Log().WithValues("method", "DeleteTransitKey")

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

	deletePath := filepath.Join("/v1", path, "keys", keyName)

	if err := t.Core.MakeRequest(core.MethodDelete, deletePath, nil, nil); err != nil {
		log.Info("failed to delete transit key", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) {
			switch responseError.StatusCode {
			case http.StatusNotFound:
				return nil
			case http.StatusBadRequest:
				for _, vaultError := range responseError.VaultErrors {
					if strings.Contains(vaultError, "deletion is not allowed for this key") {
						return err
					}
				}
				log.Info("bad request during deletion", "errors", responseError.VaultErrors)
				return nil
			}
			return nil
		}

		return core.ErrAPIError.WithDetails("failed to delete transit key").WithCause(err)
	}

	return nil
}
