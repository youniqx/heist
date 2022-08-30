package core

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/httpclient"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var vaultAPILogger = logf.Log.WithName("vault-api")

type api struct {
	Client         httpclient.Client
	Logger         logr.Logger
	AuthProvider   AuthProvider
	AuthLock       sync.Mutex
	AuthValidUntil time.Time
	Token          string
	VaultAddress   string
}

func (a *api) GetVaultAddress(path ...string) string {
	result := a.VaultAddress

	for _, segment := range path {
		segment = strings.TrimPrefix(segment, "/")
		segment = strings.TrimSuffix(segment, "/")
		result += "/"
		result += segment
	}

	return result
}

type Option func(api *api) error

const (
	authValidityFactor  time.Duration = 3
	authValidityDivisor time.Duration = 4
)

func NewCoreAPI(address string, opts ...Option) (API, error) {
	log := vaultAPILogger.WithValues("address", address)

	vaultURL, err := url.Parse(address)
	if err != nil {
		log.Info("failed to parse vault address", "error", err)
		return nil, ErrSetupFailed.WithDetails("failed to parse vault address").WithCause(err)
	}

	instance := &api{
		Client:       httpclient.NewClient(vaultURL),
		Logger:       log,
		VaultAddress: vaultURL.String(),
	}

	for _, opt := range opts {
		if err := opt(instance); err != nil {
			return nil, ErrSetupFailed.WithDetails("unable to configure vault api").WithCause(err)
		}
	}

	return instance, nil
}

func (a *api) authenticate() error {
	a.AuthLock.Lock()
	defer a.AuthLock.Unlock()

	if time.Now().Before(a.AuthValidUntil) {
		return nil
	}

	response, err := a.AuthProvider.Authenticate(a)
	if err != nil {
		return ErrAPIError.WithDetails("failed to authenticate in Vault").WithCause(err)
	}

	authValidity := time.Duration(response.Auth.LeaseDuration) * time.Second
	authValidity *= authValidityFactor
	authValidity /= authValidityDivisor

	a.AuthValidUntil = time.Now().Add(authValidity)
	a.Token = response.Auth.ClientToken

	return nil
}

func (a *api) Log() logr.Logger {
	return a.Logger
}
