package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
)

type kvSecretFieldMapMatcher struct {
	Expected map[string]string
}

const ErrKvSecretFieldMapMatcherTypeMismatch TypeMismatchError = "can only match against kvsecret.Entity objects"

func (k *kvSecretFieldMapMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kvsecret.Entity)
	if !ok {
		return false, ErrKvSecretFieldMapMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	fields, err := data.GetFields()
	if err != nil {
		return false, fmt.Errorf("failed to get service accounts from role entity: %w", err)
	}

	return reflect.DeepEqual(fields, k.Expected), nil
}

func (k *kvSecretFieldMapMatcher) FailureMessage(actual interface{}) (message string) {
	data, ok := actual.(kvsecret.Entity)
	if ok {
		fields, _ := data.GetFields()
		if fields != nil {
			return format.MessageWithDiff(fmt.Sprintf("%v", fields), "to equal", fmt.Sprintf("%v", k.Expected))
		}
	}

	return format.Message(actual, "to have fields", k.Expected)
}

func (k *kvSecretFieldMapMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have fields", k.Expected)
}
