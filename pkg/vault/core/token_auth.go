package core

import (
	"os"
)

func WithTokenAuth() Option {
	return WithAuthProvider(&tokenAuth{})
}

type tokenAuth struct{}

func (t *tokenAuth) Authenticate(API) (*AuthResponse, error) {
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return nil, ErrAPIError.WithDetails("required env var VAULT_TOKEN is not set")
	}

	return &AuthResponse{
		Auth: AuthData{
			ClientToken: token,
		},
	}, nil
}

func WithToken(token string) Option {
	return WithAuthProvider(&staticTokenAuth{
		Token: token,
	})
}

type staticTokenAuth struct {
	Token string
}

func (s *staticTokenAuth) Authenticate(API) (*AuthResponse, error) {
	if s.Token == "" {
		return nil, ErrAPIError.WithDetails("vault token is empty")
	}

	return &AuthResponse{
		Auth: AuthData{
			ClientToken: s.Token,
		},
	}, nil
}
