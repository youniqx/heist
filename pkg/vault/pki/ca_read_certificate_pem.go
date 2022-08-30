package pki

import (
	"bytes"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) ReadCACertificatePEM(ca core.MountPathEntity) (string, error) {
	log := p.Core.Log().WithValues("method", "ReadCACertificatePEM")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return "", core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	var buffer bytes.Buffer
	if err := p.Core.MakeRequest(core.MethodGet, filepath.Join("/v1", path, "ca", "pem"), nil, httpclient.Raw(&buffer)); err != nil {
		log.Info("failed to get ca cert in pem format", "error", err)
		return "", core.ErrAPIError.WithDetails("failed to get ca cert in pem format").WithCause(err)
	}

	return buffer.String(), nil
}
