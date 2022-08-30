package pki

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type revokeRequest struct {
	SerialNumber string `json:"serial_number"`
}

type revokeResponse struct {
	Data struct {
		RevocationTime int64 `json:"revocation_time"`
	} `json:"data"`
}

func (p *pkiAPI) RevokeCertificate(ca core.MountPathEntity, serial SerialNumberEntity) error {
	log := p.Core.Log().WithValues("method", "RevokeCertificate")

	path, err := ca.GetMountPath()
	if err != nil {
		log.Info("failed to get pki engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	serialNumber, err := serial.GetSerialNumber()
	if err != nil {
		log.Info("failed to get serial number", "error", err)
		return core.ErrAPIError.WithDetails("failed to get serial number").WithCause(err)
	}

	log = log.WithValues("serial_number", serialNumber)

	request := &revokeRequest{
		SerialNumber: serialNumber,
	}
	response := &revokeResponse{}

	if err := p.Core.MakeRequest(core.MethodPost, filepath.Join("/v1", path, "revoke"), httpclient.JSON(request), httpclient.JSON(response)); err != nil {
		log.Info("failed to revoke certificate in role", "error", err)
		return core.ErrAPIError.WithDetails("failed to revoke certificate in role").WithCause(err)
	}

	return nil
}
