package pki

import (
	"path/filepath"
	"strings"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type listCertsResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}

func (p *pkiAPI) ListCerts(engine core.MountPathEntity) ([]string, error) {
	log := p.Core.Log().WithValues("method", "ListCerts")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	response := &listCertsResponse{}
	if err := p.Core.MakeRequest(core.MethodList, filepath.Join("/v1", path, "certs"), nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to list certs in pki engine", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to list certs in pki engine").WithCause(err)
	}

	serialNumbers := make([]string, len(response.Data.Keys))
	for index, key := range response.Data.Keys {
		serialNumbers[index] = strings.ReplaceAll(key, "-", ":")
	}

	return serialNumbers, nil
}
