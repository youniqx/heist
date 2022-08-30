package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-test/deep"
	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	"github.com/youniqx/heist/pkg/vault/transit"
)

type configMatcher struct {
	Expected interface{}
}

const ErrConfigMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing kvengine.Entity, transit.EngineEntity or transit.KeyEntity"

func (p *configMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(kvengine.Entity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		config, err := data.GetKvEngineConfig()
		if err != nil {
			return false, fmt.Errorf("failed to get config from kv engine: %w", err)
		}

		return deep.Equal(p.Expected, config) == nil, nil
	}

	if data, ok := actual.(transit.EngineEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		config, err := data.GetTransitEngineConfig()
		if err != nil {
			return false, fmt.Errorf("failed to get config from transit engine: %w", err)
		}

		return deep.Equal(p.Expected, config) == nil, nil
	}

	if data, ok := actual.(transit.KeyEntity); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		config, err := data.GetTransitKeyConfig()
		if err != nil {
			return false, fmt.Errorf("failed to get config from transit key: %w", err)
		}

		return deep.Equal(p.Expected, config) == nil, nil
	}

	return false, ErrConfigMatcherTypeMismatch
}

func (p *configMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(kvengine.Entity); ok {
		config, _ := data.GetKvEngineConfig()
		if config != nil {
			return p.buildDiff(config)
		}
	}

	if data, ok := actual.(transit.EngineEntity); ok {
		config, _ := data.GetTransitEngineConfig()
		if config != nil {
			return p.buildDiff(config)
		}
	}

	if data, ok := actual.(transit.KeyEntity); ok {
		config, _ := data.GetTransitKeyConfig()
		if config != nil {
			return p.buildDiff(config)
		}
	}

	return format.Message(actual, "to have config", p.Expected)
}

func (p *configMatcher) buildDiff(config interface{}) string {
	var builder strings.Builder

	builder.WriteString("Expected config of to match ")
	builder.WriteString(fmt.Sprintf("%v", p.Expected))
	builder.WriteString(". Found the following differences:")
	builder.WriteString("\n")

	for _, diff := range deep.Equal(p.Expected, config) {
		builder.WriteString(" - ")
		builder.WriteString(diff)
		builder.WriteString("\n")
	}

	return builder.String()
}

func (p *configMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have config", p.Expected)
}
