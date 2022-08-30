package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) Tidy(ca core.MountPathEntity, settings *TidySettings) error {
	log := p.Core.Log().WithValues("method", "Tidy")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "tidy"), httpclient.JSON(settings), nil); err != nil {
		log.Info("failed to tidy vault pki storage", "error", err)
		return core.ErrAPIError.WithDetails("failed to tidy vault pki storage").WithCause(err)
	}

	return nil
}
