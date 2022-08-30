package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/mount"
	"github.com/youniqx/heist/pkg/vault/transit"
)

type typeMatcher struct {
	Expected string
}

const ErrTypeMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing auth.MethodEntity, mount.Entity or transit.KeyEntity"

func (p *typeMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(auth.MethodEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		method, err := data.GetMethod()
		if err != nil {
			return false, fmt.Errorf("failed to get type from auth method: %w", err)
		}

		return p.Expected == string(method), nil
	}

	if data, ok := actual.(mount.Entity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		method, err := data.GetMountType()
		if err != nil {
			return false, fmt.Errorf("failed to get type from mount: %w", err)
		}

		return p.Expected == string(method), nil
	}

	if data, ok := actual.(transit.KeyEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		keyType, err := data.GetTransitKeyType()
		if err != nil {
			return false, fmt.Errorf("failed to get type from transit key: %w", err)
		}

		return p.Expected == string(keyType), nil
	}

	return false, ErrTypeMatcherTypeMismatch
}

func (p *typeMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(auth.MethodEntity); ok {
		method, _ := data.GetMethod()
		if string(method) != "" {
			return format.MessageWithDiff(string(method), "to equal", fmt.Sprintf("%v", p.Expected))
		}
	}

	if data, ok := actual.(mount.Entity); ok {
		mountType, _ := data.GetMountType()
		if string(mountType) != "" {
			return format.MessageWithDiff(string(mountType), "to equal", fmt.Sprintf("%v", p.Expected))
		}
	}

	if data, ok := actual.(transit.KeyEntity); ok {
		keyType, _ := data.GetTransitKeyType()
		if string(keyType) != "" {
			return format.MessageWithDiff(string(keyType), "to equal", fmt.Sprintf("%v", p.Expected))
		}
	}

	return format.Message(actual, "to have type", p.Expected)
}

func (p *typeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have type", p.Expected)
}
