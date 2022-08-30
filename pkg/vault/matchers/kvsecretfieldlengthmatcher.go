package matchers

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
)

type kvSecretFieldLengthMatcher struct {
	Field  string
	Length int
}

const ErrKvSecretFieldLengthMatcherTypeMismatch TypeMismatchError = "can only match against kvsecret.Entity objects"

func (k *kvSecretFieldLengthMatcher) Match(actual interface{}) (success bool, err error) {
	data, ok := actual.(kvsecret.Entity)
	if !ok {
		return false, ErrKvSecretFieldLengthMatcherTypeMismatch
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return false, nil
	}

	fields, err := data.GetFields()
	if err != nil {
		return false, fmt.Errorf("failed to get service accounts from role entity: %w", err)
	}

	return len(fields[k.Field]) == k.Length, nil
}

func (k *kvSecretFieldLengthMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(kvsecret.Entity); ok {
		if data != nil && !reflect.ValueOf(data).IsNil() {
			fields, _ := data.GetFields()
			if fields != nil {
				return format.MessageWithDiff(fields[k.Field], "to have length", strconv.Itoa(k.Length))
			}
		}
	}

	return format.Message(actual, fmt.Sprintf("to have field %s with length", k.Field), k.Length)
}

func (k *kvSecretFieldLengthMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to have field %s with length", k.Field), k.Length)
}
