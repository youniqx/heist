package transit

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type Base64EncodedBlob []byte

func (b Base64EncodedBlob) MarshalJSON() ([]byte, error) {
	result := base64.StdEncoding.EncodeToString(b)

	data, err := json.Marshal(result)
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to encode base64 blob").WithCause(err)
	}

	return data, nil
}

type encryptRequest struct {
	PlainText Base64EncodedBlob `json:"plaintext"`
}

type encryptResponse struct {
	Data encryptResponseData `json:"data"`
}

type encryptResponseData struct {
	CipherText string `json:"ciphertext"`
}

func (t *transitAPI) TransitEncrypt(engine core.MountPathEntity, key KeyNameEntity, plainText []byte) (string, error) {
	log := t.Core.Log().WithValues("method", "TransitEncrypt")

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

	encryptPath := filepath.Join("/v1", path, "encrypt", keyName)
	request := &encryptRequest{
		PlainText: Base64EncodedBlob(plainText),
	}
	response := &encryptResponse{}

	if err := t.Core.MakeRequest(core.MethodPost, encryptPath, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to encrypt plain text", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return "", core.ErrDoesNotExist.WithCause(err)
		}

		return "", core.ErrAPIError.WithDetails("failed to encrypt plain text").WithCause(err)
	}

	return response.Data.CipherText, nil
}
