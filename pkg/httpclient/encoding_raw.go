package httpclient

import (
	"bytes"
	"io"
	"net/http"
)

type rawEncoder struct {
	Buffer      *bytes.Buffer
	Constraints []Constraint
}

func (r *rawEncoder) ShouldDecode(response *http.Response) bool {
	for _, constraint := range r.Constraints {
		if !constraint(response) {
			return false
		}
	}

	return true
}

func (r *rawEncoder) Encode(writer io.Writer) error {
	if _, err := r.Buffer.WriteTo(writer); err != nil {
		return ErrCodingError.WithDetails("failed to write data to buffer").WithCause(err)
	}

	return nil
}

func (r *rawEncoder) Decode(reader io.Reader) error {
	if _, err := r.Buffer.ReadFrom(reader); err != nil {
		return ErrCodingError.WithDetails("failed to read data from buffer").WithCause(err)
	}

	return nil
}
