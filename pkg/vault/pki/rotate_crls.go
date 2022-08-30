package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type rotateCRLsResponse struct {
	Data struct {
		Success bool `json:"success"`
	} `json:"data"`
}

func (p *pkiAPI) RotateCRLs(ca core.MountPathEntity) error {
	log := p.Core.Log().WithValues("method", "RotateCRLs")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	response := &rotateCRLsResponse{}

	if err := p.Core.MakeRequest(core.MethodGet, filepath.Join("/v1", path, "crl", "rotate"), nil, httpclient.JSON(response)); err != nil {
		log.Info("failed to rotate CRLs", "error", err)
		return core.ErrAPIError.WithDetails("failed to rotate CRLs").WithCause(err)
	}

	if !response.Data.Success {
		log.Info("CRL rotation was not successful")
		return core.ErrAPIError.WithDetails("CRL rotation was not successful")
	}

	return nil
}
