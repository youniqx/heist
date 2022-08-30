//nolint:dupl
package testhelper

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type namespaceMatcher struct {
	Expected string
}

func (p *namespaceMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(client.Object); ok {
		if reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		return data.GetNamespace() == p.Expected, nil
	}

	return false, fmt.Errorf("can only match against objects which implement the client.Object interface")
}

func (p *namespaceMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(client.Object); ok {
		if !reflect.ValueOf(data).IsNil() {
			return format.MessageWithDiff(data.GetNamespace(), "to equal", p.Expected)
		}
	}

	return format.Message(actual, "to be in namespace", p.Expected)
}

func (p *namespaceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be in namespace", p.Expected)
}
