package operator

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	operatorHandleAnnotationKey    = "youniqx.com/operatorHandle"
	operatorHandleAnnotationEnvKey = "OPERATOR_HANDLE_ANNOTATION"
)

type AnnotationFilter interface {
	predicate.Predicate
	Matches(object client.Object) bool
	MatchesString(value string) bool
}

type annotationFilter struct {
	Value string
}

func NewFilter() AnnotationFilter {
	return NewFilterWithValue(os.Getenv(operatorHandleAnnotationEnvKey))
}

func NewFilterWithValue(value string) AnnotationFilter {
	return &annotationFilter{
		Value: value,
	}
}

func (a *annotationFilter) Matches(object client.Object) bool {
	if object.GetAnnotations()[operatorHandleAnnotationKey] != a.Value {
		log.Info("skipping handling of object since handler annotation key doesn't match")
		return false
	}
	return true
}

func (a *annotationFilter) Create(event event.CreateEvent) bool {
	return a.Matches(event.Object)
}

func (a *annotationFilter) Delete(event event.DeleteEvent) bool {
	return a.Matches(event.Object)
}

func (a *annotationFilter) Update(event event.UpdateEvent) bool {
	return a.Matches(event.ObjectNew)
}

func (a *annotationFilter) Generic(event event.GenericEvent) bool {
	return a.Matches(event.Object)
}

func (a *annotationFilter) MatchesString(value string) bool {
	return value == a.Value
}
