package policy

import "github.com/youniqx/heist/pkg/vault/core"

func (p *policyAPI) ReadPolicy(policy core.PolicyNameEntity) (*Policy, error) {
	policyPath, err := getPolicyPath(policy)
	if err != nil {
		return nil, err
	}

	response, err := p.fetchPolicy(policyPath)
	if err != nil {
		return nil, err
	}

	result := response.Data.Policy
	result.Name = response.Data.Name

	return result, nil
}
