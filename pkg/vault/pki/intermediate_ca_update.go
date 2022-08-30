package pki

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) UpdateIntermediateCA(issuer core.MountPathEntity, ca CAEntity) error {
	switch action, err := p.DeterminePKIUpdateAction(ca); {
	case err != nil:
		return core.ErrAPIError.WithDetails("failed to determine action for intermediate ca").WithCause(err)
	case action == UpdateActionCreate:
		_, err := p.CreateIntermediateCA(ModeInternal, issuer, ca)
		if err != nil {
			return core.ErrAPIError.WithDetails("failed to create intermediate ca").WithCause(err)
		}
		return nil
	case action == UpdateActionUpdate:
		if err := p.UpdatePKIEngine(ca); err != nil {
			return core.ErrAPIError.WithDetails("failed to update pki engine of intermediate ca").WithCause(err)
		}
		return nil
	default:
		return core.ErrAPIError.WithDetails(fmt.Sprintf("unknown action: %s", action)).WithCause(err)
	}
}
