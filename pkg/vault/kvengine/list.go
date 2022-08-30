package kvengine

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type listSecretsResponse struct {
	Data listSecretData `json:"data"`
}

type listSecretData struct {
	Keys []string `json:"keys"`
}

func (a *engineAPI) ListKvSecrets(engine core.MountPathEntity) ([]core.SecretPath, error) {
	log := a.Core.Log().WithValues("method", "ListKvSecrets")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	return a.fetchSecretsRecursively(path, "")
}

func (a *engineAPI) fetchSecretsRecursively(mountPath string, relativePath string) ([]core.SecretPath, error) {
	log := a.Core.Log().WithValues("method", "fetchSecretsRecursively", "mountPath", mountPath, "relativePath", relativePath)

	requestPath := filepath.Join("/v1", mountPath, "metadata", relativePath)
	response := &listSecretsResponse{}

	if err := a.Core.MakeRequest(core.MethodList, requestPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to list secrets in kv engine", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		return nil, core.ErrAPIError.WithDetails("failed to list secrets in kv engine").WithCause(err)
	}

	result := make([]core.SecretPath, 0, len(response.Data.Keys))

	for _, key := range response.Data.Keys {
		if strings.HasSuffix(key, "/") {
			nestedSecrets, err := a.fetchSecretsRecursively(mountPath, filepath.Join(relativePath, key))
			if err != nil {
				log.Info("failed to list secrets in kv engine recursively", "error", err)
				return nil, core.ErrAPIError.WithDetails("failed to list secrets in kv engine recursively").WithCause(err)
			}

			result = append(result, nestedSecrets...)

			continue
		}

		result = append(result, core.SecretPath(filepath.Join(relativePath, key)))
	}

	return result, nil
}
