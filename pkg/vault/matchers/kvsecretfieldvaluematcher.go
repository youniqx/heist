package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
)

type kvSecretFieldValueMatcher struct {
	Field    string
	Expected string
}

const ErrKvSecretFieldValueMatcherTypeMismatch TypeMismatchError = "can only match against kvsecret.Entity objects"

func (k *kvSecretFieldValueMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kvsecret.Entity)
	if !ok {
		return false, ErrKvSecretFieldValueMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	fields, err := data.GetFields()
	if err != nil {
		return false, fmt.Errorf("failed to get service accounts from role entity: %w", err)
	}

	return fields[k.Field] == k.Expected, nil
}

func (k *kvSecretFieldValueMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(kvsecret.Entity); ok {
		if data != nil && !reflect.ValueOf(data).IsNil() {
			fields, _ := data.GetFields()
			if fields != nil {
				return format.MessageWithDiff(fields[k.Field], "to equal", k.Expected)
			}
		}
	}

	return format.Message(actual, fmt.Sprintf("to have field %s with value", k.Field), k.Expected)
}

func (k *kvSecretFieldValueMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to have field %s with value", k.Field), k.Expected)
}
