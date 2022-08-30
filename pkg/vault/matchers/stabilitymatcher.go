package matchers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	"github.com/youniqx/heist/pkg/vault/testenv"
)

type stabilityMatcher struct {
	Duration time.Duration
}

type TypeMismatchError string

func (t TypeMismatchError) Error() string {
	return string(t)
}

const ErrStabilityMatcherTypeMismatch TypeMismatchError = "can only match against objects impleneting the testenv.TestSecret interface"

func (p *stabilityMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(testenv.TestSecret); ok {
		if reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		return p.testStability(data)
	}

	return false, ErrStabilityMatcherTypeMismatch
}

const stepFactor = 5

func (p *stabilityMatcher) testStability(data testenv.TestSecret) (bool, error) {
	interval := p.Duration / stepFactor
	timeout := time.Now().Add(p.Duration * stepFactor)

	var (
		stableSecret *kvsecret.KvSecret
		stableCount  int
	)

	for time.Now().Before(timeout) {
		time.Sleep(interval)

		newSecret, err := data.API().ReadKvSecret(data.Engine(), data)
		if err != nil {
			return false, fmt.Errorf("failed to read kv secret: %w", err)
		}

		if reflect.DeepEqual(newSecret, stableSecret) {
			stableCount++
		} else {
			stableSecret = newSecret
			stableCount = 0
		}

		if stableCount >= stepFactor {
			return true, nil
		}
	}

	return false, nil
}

func (p *stabilityMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be stable for", p.Duration)
}

func (p *stabilityMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be stable for", p.Duration)
}
