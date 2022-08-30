package matchers

import (
	"time"

	"github.com/onsi/gomega/types"
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
	"github.com/youniqx/heist/pkg/vault/policy"
	"github.com/youniqx/heist/pkg/vault/transit"
)

func HavePolicies(policies ...core.PolicyNameEntity) types.GomegaMatcher {
	return &rolePolicyMatcher{Expected: policies}
}

func BeBoundToServiceAccounts(serviceAccounts ...string) types.GomegaMatcher {
	return &boundServiceAccountMatcher{Expected: serviceAccounts}
}

func BeBoundToNamespaces(namespaces ...string) types.GomegaMatcher {
	return &boundServiceAccountNamespaceMatcher{Expected: namespaces}
}

func HaveKvSecretFieldWithValue(field string, value string) types.GomegaMatcher {
	return &kvSecretFieldValueMatcher{
		Field:    field,
		Expected: value,
	}
}

func HaveKvSecretField(field string) types.GomegaMatcher {
	return &kvSecretFieldPresenceMatcher{Field: field}
}

func HaveKvSecretFieldFieldWithLength(field string, length int) types.GomegaMatcher {
	return &kvSecretFieldLengthMatcher{Field: field, Length: length}
}

func HaveKvSecretFields(fields map[string]string) types.GomegaMatcher {
	return &kvSecretFieldMapMatcher{Expected: fields}
}

func HaveAuthType(value auth.Type) types.GomegaMatcher {
	return &typeMatcher{Expected: string(value)}
}

func HaveMountType(value mount.Type) types.GomegaMatcher {
	return &typeMatcher{Expected: string(value)}
}

func HaveKeyType(value transit.KeyType) types.GomegaMatcher {
	return &typeMatcher{Expected: string(value)}
}

func HaveConfig(config interface{}) types.GomegaMatcher {
	return &configMatcher{Expected: config}
}

func HaveName(name string) types.GomegaMatcher {
	return &nameMatcher{Expected: name}
}

func HavePath(path string) types.GomegaMatcher {
	return &pathMatcher{Expected: path}
}

func HaveRules(rules ...*policy.Rule) types.GomegaMatcher {
	return &policyRuleMatcher{Expected: rules}
}

func BeStableFor(duration time.Duration) types.GomegaMatcher {
	return &stabilityMatcher{
		Duration: duration,
	}
}
