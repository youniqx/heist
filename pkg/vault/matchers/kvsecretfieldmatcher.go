package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
)

type kvSecretFieldPresenceMatcher struct {
	Field string
}

const ErrKvSecretFieldPresenceMatcherTypeMismatch TypeMismatchError = "can only match against kvsecret.Entity objects"

func (k *kvSecretFieldPresenceMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kvsecret.Entity)
	if !ok {
		return false, ErrKvSecretFieldPresenceMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	fields, err := data.GetFields()
	if err != nil {
		return false, fmt.Errorf("failed to get service accounts from role entity: %w", err)
	}

	return fields[k.Field] != "", nil
}

func (k *kvSecretFieldPresenceMatcher) FailureMessage(actual interface{}) (message string) {
	data, ok := actual.(kvsecret.Entity)
	if ok {
		fields, _ := data.GetFields()
		return format.MessageWithDiff(fmt.Sprintf("%v", fields), "to have field", k.Field)
	}

	return format.Message(actual, "to have field", k.Field)
}

func (k *kvSecretFieldPresenceMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have field", k.Field)
}
