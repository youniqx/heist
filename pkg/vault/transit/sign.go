package transit

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type signRequest struct {
	Input Base64EncodedBlob `json:"input"`
}

type signResponse struct {
	Data signResponseData `json:"data"`
}

type signResponseData struct {
	Signature string `json:"signature"`
}

func (t *transitAPI) TransitSign(engine core.MountPathEntity, key KeyNameEntity, input []byte) (string, error) {
	log := t.Core.Log().WithValues("method", "TransitSign")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return "", core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	keyName, err := key.GetTransitKeyName()
	if err != nil {
		log.Info("failed to get transit key name", "error", err)
		return "", core.ErrAPIError.WithDetails("failed to get transit key name").WithCause(err)
	}

	log = log.WithValues("key", keyName)

	signPath := filepath.Join("/v1", path, "sign", keyName)
	request := &signRequest{
		Input: Base64EncodedBlob(input),
	}
	response := &signResponse{}

	if err := t.Core.MakeRequest(core.MethodPost, signPath, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to sign data", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return "", core.ErrDoesNotExist.WithCause(err)
		}

		return "", core.ErrAPIError.WithDetails("failed to sign data").WithCause(err)
	}

	return response.Data.Signature, nil
}
