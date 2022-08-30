package random

import (
	"path/filepath"
	"strconv"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type randomResponse struct {
	Data randomResponseData `json:"data"`
}

type randomResponseData struct {
	Base64RandomData string `json:"random_bytes"`
}

func (r *randomAPI) fetchRandomBase64String(length int) (string, error) {
	log := r.Core.Log().WithValues("method", "fetchRandomBase64String", "length", length)
	if length == 0 {
		log.Info("cannot generate random byte sequence with length 0")
		return "", core.ErrAPIError.WithDetails("tried to generate random byte slice with length 0")
	}

	randomPath := filepath.Join("/v1/sys/tools/random", strconv.Itoa(length))
	response := &randomResponse{}

	if err := r.Core.MakeRequest(core.MethodPost, randomPath, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch random byte sequence", "error", err)
		return "", core.ErrAPIError.WithDetails("failed to fetch random byte sequence").WithCause(err)
	}

	return response.Data.Base64RandomData, nil
}
