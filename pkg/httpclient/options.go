package httpclient

import (
	"fmt"
	"net/url"
	"time"
)

type header struct {
	Key   string
	Value string
}

type requestConfig struct {
	Path     string
	Method   string
	Request  Encodable
	Response []Decodeable
	Headers  []*header
	Query    []string
	Timeout  time.Duration
}

type Option func(config *requestConfig) error

func Path(path string) Option {
	return func(config *requestConfig) error {
		config.Path = path
		return nil
	}
}

func Method(method string) Option {
	return func(config *requestConfig) error {
		config.Method = method
		return nil
	}
}

func Request(body Encodable) Option {
	return func(config *requestConfig) error {
		config.Request = body
		return nil
	}
}

func Timeout(timeout time.Duration) Option {
	return func(config *requestConfig) error {
		config.Timeout = timeout
		return nil
	}
}

func Response(body Decodeable) Option {
	return func(config *requestConfig) error {
		if body != nil {
			config.Response = append(config.Response, body)
		}

		return nil
	}
}

func Header(key string, value string) Option {
	return func(config *requestConfig) error {
		config.Headers = append(config.Headers, &header{
			Key:   key,
			Value: value,
		})

		return nil
	}
}

func Parameter(key string, value string) Option {
	return func(config *requestConfig) error {
		config.Query = append(config.Query, fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value)))
		return nil
	}
}
