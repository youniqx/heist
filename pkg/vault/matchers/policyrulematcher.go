package matchers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-test/deep"
	"github.com/onsi/gomega/format"
	"github.com/youniqx/heist/pkg/vault/policy"
)

type policyRuleMatcher struct {
	Expected []*policy.Rule
}

const ErrPolicyRuleMatcherTypeMismatch TypeMismatchError = "can only match against objects implementing policy.Body"

func (p *policyRuleMatcher) Match(actual interface{}) (success bool, err error) {
	if data, ok := actual.(policy.Body); ok {
		if data == nil || reflect.ValueOf(data).IsNil() {
			return false, nil
		}

		rules, err := data.GetPolicyRules()
		if err != nil {
			return false, fmt.Errorf("failed to get rules from policy: %w", err)
		}

		return deep.Equal(p.Expected, rules) == nil, nil
	}

	return false, ErrPolicyRuleMatcherTypeMismatch
}

func (p *policyRuleMatcher) FailureMessage(actual interface{}) (message string) {
	if data, ok := actual.(policy.Body); ok {
		rules, _ := data.GetPolicyRules()
		if rules != nil {
			var builder strings.Builder

			builder.WriteString("Expected rules of policy to match ")
			builder.WriteString(fmt.Sprintf("%v", p.Expected))
			builder.WriteString(". Found the following differences:")
			builder.WriteString("\n")

			for _, diff := range deep.Equal(p.Expected, rules) {
				builder.WriteString(" - ")
				builder.WriteString(diff)
				builder.WriteString("\n")
			}

			return builder.String()
		}
	}

	return format.Message(actual, "to have policies", p.Expected)
}

func (p *policyRuleMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have policies", p.Expected)
}
