package random

import (
	"encoding/base64"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (r *randomAPI) GenerateRandomBytes(length int) ([]byte, error) {
	log := r.Core.Log().WithValues("method", "GenerateRandomBytes", "length", length)

	randomBase64String, err := r.fetchRandomBase64String(length)
	if err != nil {
		log.Info("failed to fetch random base64 encoded string", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to fetch random base64 string").WithCause(err)
	}

	randomBytes, err := base64.StdEncoding.DecodeString(randomBase64String)
	if err != nil {
		log.Info("failed to decode base64 encoded random bytes", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to decode random bytes").WithCause(err)
	}

	return randomBytes, nil
}
