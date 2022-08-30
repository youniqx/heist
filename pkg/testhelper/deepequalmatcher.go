package testhelper

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-test/deep"
)

type deepEqualMatcher struct {
	Expected interface{}
}

func (d *deepEqualMatcher) Match(actual interface{}) (success bool, err error) {
	diff := deep.Equal(actual, d.Expected)
	return diff == nil, nil
}

func (d *deepEqualMatcher) FailureMessage(actual interface{}) (message string) {
	diff := deep.Equal(actual, d.Expected)
	return fmt.Sprintf("Found the following differences when comparing %s to %s:\n%s", reflect.TypeOf(actual), reflect.TypeOf(d.Expected), strings.Join(diff, "\n"))
}

func (d *deepEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	diff := deep.Equal(actual, d.Expected)
	return fmt.Sprintf("Did not find differences when comparing %s to %s:\n%s", actual, d.Expected, strings.Join(diff, "\n"))
}
