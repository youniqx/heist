package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type signIntermediateCSRRequest struct {
	CSR    string            `json:"csr"`
	Format CertificateFormat `json:"format"`
	*Subject
	*CASettings
}

type signIntermediateCSRResponse struct {
	Data *SignIntermediateCSRData `json:"data"`
}

type SignIntermediateCSRData struct {
	Certificate  string `json:"certificate"`
	IssuingCA    string `json:"issuing_ca"`
	SerialNumber string `json:"serial_number"`
}

func (p *pkiAPI) SignIntermediateCSR(issuer core.MountPathEntity, ca CAEntity, csr string) (*SignIntermediateCSRData, error) {
	log := p.Core.Log().WithValues("method", "SignIntermediateCSR")

	path, err := issuer.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return nil, core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

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

	request := &signIntermediateCSRRequest{
		CSR:        csr,
		Format:     CertificateFormatPEMBundle,
		Subject:    subject,
		CASettings: settings,
	}
	response := &signIntermediateCSRResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "root", "sign-intermediate"), httpclient.JSON(request), httpclient.JSON(response)); err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to sign intermediate csr").WithCause(err)
	}

	return response.Data, nil
}
