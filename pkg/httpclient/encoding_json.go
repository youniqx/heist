package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
)

type jsonEncoder struct {
	Target      interface{}
	Constraints []Constraint
}

func (j *jsonEncoder) ShouldDecode(response *http.Response) bool {
	for _, constraint := range j.Constraints {
		if !constraint(response) {
			return false
		}
	}

	return true
}

func (j *jsonEncoder) Encode(writer io.Writer) error {
	encoder := json.NewEncoder(writer)

	if err := encoder.Encode(j.Target); err != nil {
		return ErrCodingError.WithDetails("failed to encode value in json").WithCause(err)
	}

	return nil
}

func (j *jsonEncoder) Decode(reader io.Reader) error {
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(j.Target); err != nil {
		return ErrCodingError.WithDetails("failed to decode value from json").WithCause(err)
	}

	return nil
}
