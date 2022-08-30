package random

import "github.com/youniqx/heist/pkg/vault/core"

type API interface {
	GenerateRandomBytes(length int) ([]byte, error)
	GenerateRandomString(length int) (string, error)
}

type randomAPI struct {
	Core core.API
}

func NewAPI(core core.API) API {
	return &randomAPI{Core: core}
}
