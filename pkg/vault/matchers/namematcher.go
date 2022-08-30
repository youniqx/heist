package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/transit"
)

type nameMatcher struct {
	Expected string
}

const ErrNameMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing core.RoleNameEntity, core.PolicyNameEntity or transit.KeyNameEntity"

func (p *nameMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(core.RoleNameEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		name, err := data.GetRoleName()
		if err != nil {
			return false, fmt.Errorf("failed to get name from role: %w", err)
		}

		return p.Expected == name, nil
	}

	if data, ok := actual.(core.PolicyNameEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		name, err := data.GetPolicyName()
		if err != nil {
			return false, fmt.Errorf("failed to get name from policy: %w", err)
		}

		return p.Expected == name, nil
	}

	if data, ok := actual.(transit.KeyNameEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		name, err := data.GetTransitKeyName()
		if err != nil {
			return false, fmt.Errorf("failed to get name from transit key: %w", err)
		}

		return p.Expected == name, nil
	}

	return false, ErrNameMatcherTypeMismatch
}

func (p *nameMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(core.RoleNameEntity); ok {
		name, _ := data.GetRoleName()
		return format.MessageWithDiff(name, "to equal", p.Expected)
	}

	if data, ok := actual.(core.PolicyNameEntity); ok {
		name, _ := data.GetPolicyName()
		return format.MessageWithDiff(name, "to equal", p.Expected)
	}

	if data, ok := actual.(transit.KeyNameEntity); ok {
		name, _ := data.GetTransitKeyName()
		return format.MessageWithDiff(name, "to equal", p.Expected)
	}

	return format.Message(actual, "to have name", p.Expected)
}

func (p *nameMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have name", p.Expected)
}
