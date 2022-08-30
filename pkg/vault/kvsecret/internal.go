package kvsecret

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type setKvSecretOptions struct {
	CAS int `json:"cas"`
}

type setKvSecretRequest struct {
	Options setKvSecretOptions `json:"options"`
	Data    map[string]string  `json:"data"`
}

type data struct {
	Metadata metadata          `json:"metadata"`
	Data     map[string]string `json:"data"`
}

type metadata struct {
	CreatedTime  string `json:"created_time"`
	DeletionTime string `json:"deletion_time"`
	Destroyed    bool   `json:"destroyed"`
	Version      int    `json:"version"`
}

type setKvSecretResponse struct {
	Metadata metadata `json:"data"`
}

type getKvSecretResponse struct {
	RequestID string `json:"request_id"`
	Data      data   `json:"data"`
}

func (a *api) fetchKvSecret(path string) (*getKvSecretResponse, error) {
	log := a.Core.Log().WithValues("method", "fetchKvSecret", "path", path)

	response := &getKvSecretResponse{}
	if err := a.Core.MakeRequest(core.MethodGet, path, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("couldn't fetch secret data", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch kv secret").WithCause(err)
	}

	return response, nil
}

func (a *api) deleteKvSecret(path string) error {
	log := a.Core.Log().WithValues("method", "deleteKvSecret", "path", path)

	if err := a.Core.MakeRequest(core.MethodDelete, path, nil, nil); err != nil {
		log.Info("could not delete secret", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return core.ErrDoesNotExist.WithCause(err)
		}

		return core.ErrAPIError.WithDetails("failed to delete kv secret").WithCause(err)
	}

	return nil
}

func (a *api) writeKvSecret(path string, cas int, fields map[string]string) (*setKvSecretResponse, error) {
	log := a.Core.Log().WithValues("method", "writeKvSecret", "path", path)

	request := &setKvSecretRequest{
		Options: setKvSecretOptions{CAS: cas},
		Data:    fields,
	}
	response := &setKvSecretResponse{}

	if err := a.Core.MakeRequest(core.MethodPost, path, httpclient.JSON(request), httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("couldn't write secret data", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to write kv secret data").WithCause(err)
	}

	return response, nil
}

func getSecretDataPath(engine core.MountPathEntity, secret core.SecretPathEntity) (string, error) {
	enginePath, err := engine.GetMountPath()
	if err != nil {
		return "", core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	secretPath, err := secret.GetSecretPath()
	if err != nil {
		return "", core.ErrAPIError.WithDetails("failed to get secret path").WithCause(err)
	}

	path := filepath.Join("/v1", enginePath, "data", secretPath)

	return path, nil
}

func getSecretMetadataPath(engine core.MountPathEntity, secret core.SecretPathEntity) (string, error) {
	enginePath, err := engine.GetMountPath()
	if err != nil {
		return "", core.ErrAPIError.WithDetails("failed to get engine path").WithCause(err)
	}

	secretPath, err := secret.GetSecretPath()
	if err != nil {
		return "", core.ErrAPIError.WithDetails("failed to get secret path").WithCause(err)
	}

	path := filepath.Join("/v1", enginePath, "metadata", secretPath)

	return path, nil
}
