package auth

import (
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type methodInfo struct {
	Type Type `json:"type"`
}

type listMethodResponse struct {
	Data map[string]*methodInfo `json:"data"`
}

func (a *authAPI) ListAuthMethods() ([]*Method, error) {
	log := a.Core.Log().WithValues("method", "ListAuthMethods")

	response := &listMethodResponse{}
	if err := a.Core.MakeRequest(core.MethodGet, "/v1/sys/auth", nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("unable to list auth methods", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to list auth methods").WithCause(err)
	}

	result := make([]*Method, 0, len(response.Data))
	for path, info := range response.Data {
		result = append(result, &Method{
			Path: strings.TrimSuffix(path, "/"),
			Type: info.Type,
		})
	}

	return result, nil
}
