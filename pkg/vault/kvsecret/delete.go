package kvsecret

import (
	"errors"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *api) DeleteKvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) error {
	log := a.Core.Log().WithValues("method", "DeleteKvSecret")

	path, err := getSecretMetadataPath(engine, secret)
	if err != nil {
		log.Info("failed to get secret metadata path", "error", err)
		return err
	}

	log = log.WithValues("path", path)

	switch err := a.deleteKvSecret(path); {
	case errors.Is(err, core.ErrDoesNotExist):
		return nil
	case err == nil:
		return nil
	default:
		log.Info("failed to delete secret", "error", err)
		return err
	}
}
