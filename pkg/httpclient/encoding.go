package httpclient

import (
	"bytes"
	"io"
	"net/http"
)

type Encodable interface {
	Encode(writer io.Writer) error
}

type Decodeable interface {
	ShouldDecode(response *http.Response) bool
	Decode(reader io.Reader) error
}

type Codeable interface {
	Encodable
	Decodeable
}

type Constraint func(response *http.Response) bool

// ConstraintSuccess is a helper function to filter for successful http responses.
var ConstraintSuccess Constraint = func(response *http.Response) bool {
	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices
}

// ConstraintFailed is a constraint to filter for failed http responses.
var ConstraintFailed Constraint = func(response *http.Response) bool {
	return response.StatusCode >= http.StatusBadRequest
}

// ConstraintNone is a helper function that filters for nothing and allows everything.
var ConstraintNone Constraint = func(response *http.Response) bool {
	return true
}

func Raw(buffer *bytes.Buffer, constraints ...Constraint) Codeable {
	if len(constraints) == 0 {
		constraints = []Constraint{ConstraintSuccess}
	}

	return &rawEncoder{
		Buffer:      buffer,
		Constraints: constraints,
	}
}

func JSON(v interface{}, constraints ...Constraint) Codeable {
	if len(constraints) == 0 {
		constraints = []Constraint{ConstraintSuccess}
	}

	return &jsonEncoder{
		Target:      v,
		Constraints: constraints,
	}
}
