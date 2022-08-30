package testhelper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/onsi/gomega/format"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type conditionMatcher struct {
	Name    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}

//nolint:cyclop
func parseConditionStatusFromObject(actual interface{}, name string) (string, string, string, error) {
	data, ok := actual.(*unstructured.Unstructured)
	if !ok {
		return "", "", "", fmt.Errorf("can only match against *unstructured.Unstructured objects")
	}

	if data == nil || reflect.ValueOf(data).IsNil() {
		return "", "", "", fmt.Errorf("object is nil")
	}

	status := data.Object["status"]
	if status == nil || reflect.ValueOf(status).IsNil() {
		return "", "", "", fmt.Errorf("object %s/%s does not have a status object", data.GetNamespace(), data.GetName())
	}

	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("object %s/%s does not have a valid status object", data.GetNamespace(), data.GetName())
	}
	if statusMap == nil || reflect.ValueOf(statusMap).IsNil() {
		return "", "", "", fmt.Errorf("object %s/%s does not have a valid status object", data.GetNamespace(), data.GetName())
	}

	conditions := statusMap["conditions"]
	if conditions == nil || reflect.ValueOf(conditions).IsNil() {
		return "", "", "", fmt.Errorf("object %s/%s does not have a conditions field", data.GetNamespace(), data.GetName())
	}

	conditionsArray, ok := conditions.([]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
	}

	for _, item := range conditionsArray {
		condition, ok := item.(map[string]interface{})
		if !ok {
			return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
		}

		conditionType, ok := condition["type"].(string)
		if !ok {
			return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
		}

		if conditionType != name {
			continue
		}

		conditionStatus, ok := condition["status"].(string)
		if !ok {
			return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
		}

		conditionReason, ok := condition["reason"].(string)
		if !ok {
			return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
		}

		conditionMessage, ok := condition["message"].(string)
		if !ok {
			return "", "", "", fmt.Errorf("object %s/%s does not have a valid conditions slice", data.GetNamespace(), data.GetName())
		}

		return conditionStatus, conditionReason, conditionMessage, nil
	}

	return "", "", "", fmt.Errorf("object %s/%s does not have the specified condition", data.GetNamespace(), data.GetName())
}

func (p *conditionMatcher) Match(actual interface{}) (success bool, err error) {
	status, reason, message, err := parseConditionStatusFromObject(actual, p.Name)
	if err != nil {
		return false, err
	}

	return status == string(p.Status) && reason == p.Reason && strings.HasPrefix(message, p.Message), nil
}

func (p *conditionMatcher) FailureMessage(actual interface{}) (message string) {
	status, reason, msg, err := parseConditionStatusFromObject(actual, p.Name)
	if err == nil {
		return format.MessageWithDiff(status+", "+reason+", "+msg, "to equal", string(p.Status)+", "+p.Reason+", "+p.Message)
	}

	return format.Message(actual, "to have condition", string(p.Status)+", "+p.Reason+", "+p.Message)
}

func (p *conditionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have condition", string(p.Status)+", "+p.Reason+", "+p.Message)
}
