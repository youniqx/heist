package kvsecret

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

func (a *api) ReadKvSecret(engine core.MountPathEntity, secret core.SecretPathEntity) (*KvSecret, error) {
	log := a.Core.Log().WithValues("method", "ReadKvSecret")

	path, err := getSecretDataPath(engine, secret)
	if err != nil {
		log.Info("failed to get secret data path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get secret data path").WithCause(err)
	}

	log = log.WithValues("path", path)

	kvSecret, err := a.fetchKvSecret(path)
	if err != nil {
		log.Info("failed to fetch secret data", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get secret data").WithCause(err)
	}

	secretPath, err := secret.GetSecretPath()
	if err != nil {
		log.Info("failed to get secret path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get secret path").WithCause(err)
	}

	return &KvSecret{
		Path:   secretPath,
		Fields: kvSecret.Data.Data,
	}, nil
}
