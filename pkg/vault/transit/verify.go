package transit

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type verifyRequest struct {
	Input     Base64EncodedBlob `json:"input"`
	Signature string            `json:"signature"`
}

type verifyResponse struct {
	Data verifyResponseData `json:"data"`
}

type verifyResponseData struct {
	Valid bool `json:"valid"`
}

func (t *transitAPI) TransitVerify(engine core.MountPathEntity, key KeyNameEntity, input []byte, signature string) (bool, error) {
	log := t.Core.Log().WithValues("method", "TransitVerify")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	keyName, err := key.GetTransitKeyName()
	if err != nil {
		log.Info("failed to get transit key name", "error", err)
		return false, core.ErrAPIError.WithDetails("failed to get transit key name").WithCause(err)
	}

	log = log.WithValues("key", keyName)

	verifyPath := filepath.Join("/v1", path, "verify", keyName)
	request := &verifyRequest{
		Input:     Base64EncodedBlob(input),
		Signature: signature,
	}
	response := &verifyResponse{}

	if err := t.Core.MakeRequest(core.MethodPost, verifyPath, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to verify signature", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return false, core.ErrDoesNotExist.WithCause(err)
		}

		return false, core.ErrAPIError.WithDetails("failed to verify signature").WithCause(err)
	}

	return response.Data.Valid, nil
}
