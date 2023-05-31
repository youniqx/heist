package core

import (
	"fmt"
	"os"
)

type StringSource interface {
	FetchStringValue() (string, error)
}

type Value string

func (v Value) FetchStringValue() (string, error) {
	return string(v), nil
}

type EnvVar string

func (e EnvVar) FetchStringValue() (string, error) {
	value := os.Getenv(string(e))
	if value == "" {
		return "", ErrSetupFailed.WithDetails(fmt.Sprintf("required env var %s is not set", e))
	}

	return value, nil
}

type File string

func (f File) FetchStringValue() (string, error) {
	data, err := os.ReadFile(string(f))
	if err != nil {
		return "", ErrSetupFailed.WithDetails(fmt.Sprintf("failed to read file at path %s", f)).WithCause(err)
	}

	return string(data), nil
}
