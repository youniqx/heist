package transit

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (t *transitAPI) RotateTransitKey(engine core.MountPathEntity, key KeyNameEntity) error {
	log := t.Core.Log().WithValues("method", "RotateTransitKey")

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

	rotatePath := filepath.Join("/v1", path, "keys", keyName, "rotate")
	if err := t.Core.MakeRequest(core.MethodPost, rotatePath, nil, nil); err != nil {
		log.Info("failed to rotate transit key", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return core.ErrDoesNotExist.WithCause(err)
		}

		return core.ErrAPIError.WithDetails("failed to rotate transit key").WithCause(err)
	}

	return nil
}
