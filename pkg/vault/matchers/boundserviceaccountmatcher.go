package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
)

type boundServiceAccountMatcher struct {
	Expected []string
}

const ErrBoundServiceAccountMatcherTypeMismatch TypeMismatchError = "can only match against kubernetesauth.RoleEntity objects"

func (p *boundServiceAccountMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kubernetesauth.RoleEntity)
	if !ok {
		return false, ErrBoundServiceAccountMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	serviceAccounts, err := data.GetBoundServiceAccounts()
	if err != nil {
		return false, fmt.Errorf("failed to get service accounts from role entity: %w", err)
	}

	return reflect.DeepEqual(p.Expected, serviceAccounts), nil
}

func (p *boundServiceAccountMatcher) FailureMessage(actual interface{}) (message string) {
	data, ok := actual.(kubernetesauth.RoleEntity)
	if ok {
		serviceAccounts, _ := data.GetBoundServiceAccounts()
		return format.MessageWithDiff(fmt.Sprintf("%v", serviceAccounts), "to equal", fmt.Sprintf("%v", p.Expected))
	}

	return format.Message(actual, "to be bound to service accounts", p.Expected)
}

func (p *boundServiceAccountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be bound to service accounts", p.Expected)
}
