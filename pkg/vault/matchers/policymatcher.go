package matchers

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/core"
)

type rolePolicyMatcher struct {
	Expected []core.PolicyNameEntity
}

const ErrRolePolicyMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing core.RolePoliciesEntity"

func (p *rolePolicyMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(core.RolePoliciesEntity)
	if !ok {
		return false, ErrRolePolicyMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	policies, err := data.GetRolePolicies()
	if err != nil {
		return false, fmt.Errorf("failed to get policies from role entity: %w", err)
	}

	actualPolicies := make([]string, 0, len(policies))
	for _, policy := range policies {
		actualPolicies = append(actualPolicies, string(policy))
	}

	sort.Strings(actualPolicies)

	expectedPolicies := make([]string, 0, len(p.Expected))

	for _, entity := range p.Expected {
		name, err := entity.GetPolicyName()
		if err != nil {
			return false, fmt.Errorf("failed to get policy name from role entity: %w", err)
		}

		expectedPolicies = append(expectedPolicies, name)
	}

	sort.Strings(expectedPolicies)

	return reflect.DeepEqual(expectedPolicies, actualPolicies), nil
}

func (p *rolePolicyMatcher) FailureMessage(actual interface{}) (message string) {
	data, ok := actual.(core.RolePoliciesEntity)
	if ok {
		if data != nil && !reflect.ValueOf(data).IsNil() {
			policies, _ := data.GetRolePolicies()
			if policies != nil {
				return format.MessageWithDiff(fmt.Sprintf("%v", policies), "to equal", fmt.Sprintf("%v", p.Expected))
			}
		}
	}

	return format.Message(actual, "to have policies", p.Expected)
}

func (p *rolePolicyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have policies", p.Expected)
}
