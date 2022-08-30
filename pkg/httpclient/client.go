package httpclient

import (
	"net/http"
	"net/url"

	"github.com/youniqx/heist/pkg/erx"
)

type Client interface {
	Perform(options ...Option) (*Result, error)
}

var (
	// ErrRequestError is returned when a request could not be made
	ErrRequestError = erx.New("HTTP Client", "failed to make request")

	// ErrCodingError is returned when encoding or decoding request or response bodies has failed
	ErrCodingError = erx.New("HTTP Client", "failed to marshal values")
)

type client struct {
	Client  *http.Client
	BaseURL *url.URL
}

type Result struct {
	StatusCode int
}

func NewClient(base *url.URL) Client {
	instance := &client{
		Client:  &http.Client{},
		BaseURL: base,
	}

	return instance
}

func NewClientWithHttpClient(base *url.URL, httpClient *http.Client) Client {
	instance := &client{
		Client:  httpClient,
		BaseURL: base,
	}

	return instance
}
