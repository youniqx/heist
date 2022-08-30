package testhelper

import (
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BeInNamespace(namespace string) types.GomegaMatcher {
	return &namespaceMatcher{Expected: namespace}
}

func HaveResourceName(name string) types.GomegaMatcher {
	return &nameMatcher{Expected: name}
}

func HaveCondition(name string, status metav1.ConditionStatus, reason string, message string) types.GomegaMatcher {
	return &conditionMatcher{
		Name:    name,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
}

func DeepEqual(expected interface{}) types.GomegaMatcher {
	return &deepEqualMatcher{Expected: expected}
}
