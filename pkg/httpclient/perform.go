package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultMethod  = http.MethodGet
	defaultTimeout = 60 * time.Second
)

//nolint:cyclop
func (c *client) Perform(options ...Option) (*Result, error) {
	config := &requestConfig{
		Path:     "",
		Method:   defaultMethod,
		Request:  nil,
		Response: nil,
		Headers:  nil,
		Query:    nil,
		Timeout:  defaultTimeout,
	}

	for _, option := range options {
		if err := option(config); err != nil {
			return nil, ErrRequestError.WithDetails("failed to configure request").WithCause(err)
		}
	}

	requestURL, err := url.Parse(c.BaseURL.String())
	if err != nil {
		return nil, ErrRequestError.WithDetails("failed to parse base url").WithCause(err)
	}

	requestURL.Path = config.Path
	requestURL.RawQuery = strings.Join(config.Query, "&")

	var requestBody io.Reader

	if config.Request != nil {
		buffer := &bytes.Buffer{}
		if err := config.Request.Encode(buffer); err != nil {
			return nil, ErrRequestError.WithDetails("failed to encode request body").WithCause(err)
		}

		requestBody = buffer
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, config.Method, requestURL.String(), requestBody)
	if err != nil {
		return nil, ErrRequestError.WithDetails("failed to create request").WithCause(err)
	}

	for _, header := range config.Headers {
		req.Header.Add(header.Key, header.Value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, ErrRequestError.WithDetails("failed to send request").WithCause(err)
	}

	defer resp.Body.Close()

	for _, bodyDecoder := range config.Response {
		if bodyDecoder.ShouldDecode(resp) {
			if err := bodyDecoder.Decode(resp.Body); err != nil {
				return nil, ErrRequestError.WithDetails("failed to decode response body").WithCause(err)
			}

			break
		}
	}

	return &Result{
		StatusCode: resp.StatusCode,
	}, nil
}
