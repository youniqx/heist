package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type setIntermediateCertRequest struct {
	Certificate string `json:"certificate"`
}

func (p *pkiAPI) SetIntermediateCACert(ca core.MountPathEntity, cert string) error {
	log := p.Core.Log().WithValues("method", "SignIntermediateCSR")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	request := &setIntermediateCertRequest{
		Certificate: cert,
	}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "intermediate", "set-signed"), httpclient.JSON(request), nil); err != nil {
		return core.ErrAPIError.WithDetails("failed to sign intermediate csr").WithCause(err)
	}

	return nil
}
