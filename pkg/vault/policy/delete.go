package policy

import "github.com/youniqx/heist/pkg/vault/core"

func (p *policyAPI) DeletePolicy(policy core.PolicyNameEntity) error {
	policyPath, err := getPolicyPath(policy)
	if err != nil {
		return err
	}

	return p.deletePolicy(policyPath)
}
