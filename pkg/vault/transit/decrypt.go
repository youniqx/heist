package transit

import (
	"encoding/base64"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type decryptRequest struct {
	CipherText string `json:"ciphertext"`
}

type decryptResponse struct {
	Data decryptResponseData `json:"data"`
}

type decryptResponseData struct {
	Base64PlainText string `json:"plaintext"`
}

func (t *transitAPI) TransitDecrypt(engine core.MountPathEntity, key KeyNameEntity, cipherText string) ([]byte, error) {
	log := t.Core.Log().WithValues("method", "TransitDecrypt")

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

	decryptPath := filepath.Join("/v1", path, "decrypt", keyName)
	request := &decryptRequest{
		CipherText: cipherText,
	}
	response := &decryptResponse{}

	if err := t.Core.MakeRequest(core.MethodPost, decryptPath, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to decrypt plain text", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to decrypt plain text").WithCause(err)
	}

	plainText, err := base64.StdEncoding.DecodeString(response.Data.Base64PlainText)
	if err != nil {
		log.Info("failed to decode base64 plain text", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to decode base64 plain text").WithCause(err)
	}

	return plainText, nil
}
