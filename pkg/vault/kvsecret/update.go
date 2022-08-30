package kvsecret

import (
	"errors"
	"reflect"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *api) UpdateKvSecret(engine core.MountPathEntity, secret Entity) error {
	log := a.Core.Log().WithValues("method", "UpdateKvSecret")

	path, err := getSecretDataPath(engine, secret)
	if err != nil {
		log.Info("failed to get secret data path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get secret data path").WithCause(err)
	}

	log = log.WithValues("path", path)

	expectedFields, err := secret.GetFields()
	if err != nil {
		return core.ErrAPIError.WithDetails("failed to get secret fields").WithCause(err)
	}

	var (
		cas            int
		updateRequired bool
	)

	switch kvSecret, err := a.fetchKvSecret(path); {
	case errors.Is(err, core.ErrDoesNotExist):
		cas = 0
		updateRequired = true
	case err == nil:
		cas = kvSecret.Data.Metadata.Version
		updateRequired = !reflect.DeepEqual(expectedFields, kvSecret.Data.Data)
	default:
		return core.ErrAPIError.WithDetails("failed to check state of secret in Vault").WithCause(err)
	}

	if !updateRequired {
		return nil
	}

	writeResponse, err := a.writeKvSecret(path, cas, expectedFields)
	if err != nil {
		log.Info("failed to write secret data", "error", err)
		return core.ErrAPIError.WithDetails("failed to write secret to Vault").WithCause(err)
	}

	log.Info("secret has been updated", "newVersion", writeResponse.Metadata.Version)

	return nil
}
