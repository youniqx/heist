package policy

import (
	"errors"
	"reflect"

	"github.com/youniqx/heist/pkg/vault/core"
)

func (p *policyAPI) UpdatePolicy(policy Entity) error {
	policyPath, err := getPolicyPath(policy)
	if err != nil {
		return core.ErrAPIError.WithDetails("failed to fetch policy path").WithCause(err)
	}

	expectedRules, err := policy.GetPolicyRules()
	if err != nil {
		return core.ErrAPIError.WithDetails("failed to fetch policy rules").WithCause(err)
	}

	var updateRequired bool

	switch fetchResponse, err := p.fetchPolicy(policyPath); {
	case errors.Is(err, core.ErrDoesNotExist):
		updateRequired = true
	case err == nil:
		updateRequired = !reflect.DeepEqual(expectedRules, fetchResponse.Data.Policy.Rules)
	default:
		return core.ErrAPIError.WithDetails("failed to fetch policy from vault").WithCause(err)
	}

	if !updateRequired {
		return nil
	}

	err = p.writePolicy(policyPath, expectedRules)
	if err != nil {
		return core.ErrAPIError.WithDetails("failed to write policy to vault").WithCause(err)
	}

	return nil
}
