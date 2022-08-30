package core

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"

	"github.com/youniqx/heist/pkg/httpclient"
)

func WithAuthProvider(provider AuthProvider) Option {
	return func(api *api) error {
		api.AuthProvider = provider
		return api.authenticate()
	}
}

func WithCACerts(cas ...string) Option {
	return func(api *api) error {
		vaultURL, err := url.Parse(api.VaultAddress)
		if err != nil {
			return ErrSetupFailed.WithDetails(fmt.Sprintf("Failed to parse vault address: %s", api.VaultAddress)).WithCause(err)
		}

		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		for _, ca := range cas {
			if ok := rootCAs.AppendCertsFromPEM([]byte(ca)); !ok {
				return ErrSetupFailed.WithDetails(fmt.Sprintf("Failed to append CA to root ca pool: %s", ca))
			}
		}

		api.Client = httpclient.NewClientWithHttpClient(vaultURL, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:    rootCAs,
					MinVersion: tls.VersionTLS12,
				},
			},
		})
		return nil
	}
}
