package pki

import (
	"fmt"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type generateIntermediateCSRRequest struct {
	*Subject
	*CASettings
}

type generateIntermediateCSRResponse struct {
	Data *IntermediateCAInfo `json:"data"`
}

type IntermediateCAInfo struct {
	CSR            string  `json:"csr"`
	PrivateKey     string  `json:"private_key"`
	PrivateKeyType KeyType `json:"private_key_type"`
}

func (p *pkiAPI) GenerateIntermediateCSR(mode Mode, ca CAEntity) (*IntermediateCAInfo, error) {
	log := p.Core.Log().WithValues("method", "GenerateIntermediateCSR")

	switch mode {
	case ModeInternal, ModeExported:
	default:
		log.Info("unknown ca mode setting", "mode", mode)
		return nil, core.ErrAPIError.WithDetails(fmt.Sprintf("unknown ca mode setting: %s", mode))
	}

	log = log.WithValues("mode", mode)

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	settings, err := ca.GetSettings()
	if err != nil {
		log.Info("failed to get settings from ca entity", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get settings form ca entity").WithCause(err)
	}

	subject, err := ca.GetSubject()
	if err != nil {
		log.Info("failed to get subject from ca entity", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get subject from ca entity").WithCause(err)
	}

	request := &generateIntermediateCSRRequest{
		Subject:    subject,
		CASettings: settings,
	}
	response := &generateIntermediateCSRResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "intermediate", "generate", string(mode)), httpclient.JSON(request), httpclient.JSON(response)); err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to generate intermediate csr in pki engine").WithCause(err)
	}

	return response.Data, nil
}
