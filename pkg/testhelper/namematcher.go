//nolint:dupl
package testhelper

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nameMatcher struct {
	Expected string
}

func (p *nameMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(client.Object); ok {
		if reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		return data.GetName() == p.Expected, nil
	}

	return false, fmt.Errorf("can only match against objects which implement the client.Object interface")
}

func (p *nameMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(client.Object); ok {
		if !reflect.ValueOf(data).IsNil() {
			return format.MessageWithDiff(data.GetName(), "to equal", p.Expected)
		}
	}

	return format.Message(actual, "to have name", p.Expected)
}

func (p *nameMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have name", p.Expected)
}
