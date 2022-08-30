package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
)

type boundServiceAccountNamespaceMatcher struct {
	Expected []string
}

const ErrBoundServiceAccountNamespaceMatcherTypeMismatch TypeMismatchError = "can only match against kubernetesauth.RoleEntity objects"

func (p *boundServiceAccountNamespaceMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kubernetesauth.RoleEntity)
	if !ok {
		return false, ErrBoundServiceAccountNamespaceMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	namespaces, err := data.GetBoundNamespaces()
	if err != nil {
		return false, fmt.Errorf("failed to get namespaces from role entity: %w", err)
	}

	return reflect.DeepEqual(p.Expected, namespaces), nil
}

func (p *boundServiceAccountNamespaceMatcher) FailureMessage(actual interface{}) (message string) {
	data, ok := actual.(kubernetesauth.RoleEntity)
	if ok {
		namespaces, _ := data.GetBoundNamespaces()
		return format.MessageWithDiff(fmt.Sprintf("%v", namespaces), "to equal", fmt.Sprintf("%v", p.Expected))
	}

	return format.Message(actual, "to be bound to namespaces", p.Expected)
}

func (p *boundServiceAccountNamespaceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be bound to namespaces", p.Expected)
}
