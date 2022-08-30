package pki

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *pkiAPI) UpdateRootCA(ca CAEntity) error {
	switch action, err := p.DeterminePKIUpdateAction(ca); {
	case err != nil:
		return core.ErrAPIError.WithDetails("failed to determine action for root ca").WithCause(err)
	case action == UpdateActionCreate:
		_, err := p.CreateRootCA(ModeInternal, ca)
		if err != nil {
			return core.ErrAPIError.WithDetails("failed to create root ca").WithCause(err)
		}
		return nil
	case action == UpdateActionUpdate:
		if err := p.UpdatePKIEngine(ca); err != nil {
			return core.ErrAPIError.WithDetails("failed to update pki engine of root ca").WithCause(err)
		}
		return nil
	default:
		return core.ErrAPIError.WithDetails(fmt.Sprintf("unknown action: %s", action)).WithCause(err)
	}
}
