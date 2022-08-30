package pki

import "github.com/youniqx/heist/pkg/vault/core"

func (p *pkiAPI) IsPKIEngineInitialized(ca core.MountPathEntity) (bool, error) {
	exists, err := p.Mount.HasEngine(ca)
	if err != nil {
		return false, err
	}
	if exists {
		cert, err := p.ReadCACertificatePEM(ca)
		if err != nil {
			return false, err
		}
		return cert != "", nil
	}
	return false, nil
}

type UpdateAction string

const (
	UpdateActionNone   UpdateAction = ""
	UpdateActionCreate UpdateAction = "create"
	UpdateActionUpdate UpdateAction = "update"
)

func (p *pkiAPI) DeterminePKIUpdateAction(ca core.MountPathEntity) (UpdateAction, error) {
	initialized, err := p.IsPKIEngineInitialized(ca)
	if err != nil {
		return UpdateActionNone, err
	}

	if initialized {
		return UpdateActionUpdate, nil
	}

	return UpdateActionCreate, nil
}
