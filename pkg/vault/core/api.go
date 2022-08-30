package core

import (
	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/erx"
	"github.com/youniqx/heist/pkg/httpclient"
)

// ErrDoesNotExist is returned by all apis when a vault object doesn't exist.
var ErrDoesNotExist = erx.New("Vault API", "vault object does not exist")

// ErrSetupFailed is returned when the Vault API instanced failed to initialize.
var ErrSetupFailed = erx.New("Vault API", "failed to initialize Vault API")

// ErrAPIError is returned for any error encountered while interacting with the Vault API.
var ErrAPIError = erx.New("Vault API", "failed to interact with Vault API")

// ErrHTTPError is returned for any error encountered while sending or receiving HTTP requests.
var ErrHTTPError = erx.New("Vault API", "failed to send or receive HTTP request")

type erxError = erx.Error

type VaultHTTPError struct {
	erxError

	StatusCode  int
	VaultErrors []string
}

type MountPathEntity interface {
	GetMountPath() (string, error)
}

type SecretPathEntity interface {
	GetSecretPath() (string, error)
}

type RequestType string

const (
	// MethodGet is a http GET request.
	MethodGet RequestType = "GET"
	// MethodDelete is a http DELETE request.
	MethodDelete RequestType = "DELETE"
	// MethodPost is a http POST request.
	MethodPost RequestType = "POST"
	// MethodPut is a http PUT request.
	MethodPut RequestType = "PUT"
	// MethodList is a http LIST request.
	MethodList RequestType = "LIST"
)

type RequestAPI interface {
	MakeRequest(method RequestType, path string, request httpclient.Encodable, response httpclient.Decodeable) error
	GetVaultAddress(path ...string) string
}

type LoggingAPI interface {
	Log() logr.Logger
}

type API interface {
	RequestAPI
	LoggingAPI
}

type MountPath string

func (k MountPath) GetMountPath() (string, error) {
	return string(k), nil
}

type SecretPath string

func (k SecretPath) GetSecretPath() (string, error) {
	return string(k), nil
}

type PolicyNameEntity interface {
	GetPolicyName() (string, error)
}

type PolicyName string

func (p PolicyName) GetPolicyName() (string, error) {
	return string(p), nil
}

type RoleNameEntity interface {
	GetRoleName() (string, error)
}

type RoleName string

func (r RoleName) GetRoleName() (string, error) {
	return string(r), nil
}

type RolePoliciesEntity interface {
	GetRolePolicies() ([]PolicyName, error)
}
