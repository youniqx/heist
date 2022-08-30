package matchers

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/core"
)

type pathMatcher struct {
	Expected string
}

const ErrPathMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing core.MountPathEntity or core.SecretPathEntity"

func (p *pathMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(core.MountPathEntity); ok {
		if reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		path, err := data.GetMountPath()
		if err != nil {
			return false, fmt.Errorf("failed to get path from mount: %w", err)
		}

		return p.Expected == path, nil
	}

	if data, ok := actual.(core.SecretPathEntity); ok {
		if reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		path, err := data.GetSecretPath()
		if err != nil {
			return false, fmt.Errorf("failed to get path from secret: %w", err)
		}

		return p.Expected == path, nil
	}

	return false, ErrPathMatcherTypeMismatch
}

func (p *pathMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(core.MountPathEntity); ok && !reflect.ValueOf(data).IsNil() {
		path, _ := data.GetMountPath()
		return format.MessageWithDiff(path, "to equal", p.Expected)
	}

	if data, ok := actual.(core.SecretPathEntity); ok && !reflect.ValueOf(data).IsNil() {
		path, _ := data.GetSecretPath()
		return format.MessageWithDiff(path, "to equal", p.Expected)
	}

	return format.Message(actual, "to have name", p.Expected)
}

func (p *pathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have name", p.Expected)
}
