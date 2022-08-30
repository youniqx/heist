package core

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

var retryableBadRequests = []string{
	"Upgrading from non-versioned to versioned data",
}

func isBadRequestError(err error) bool {
	var responseError *VaultHTTPError
	if errors.As(err, &responseError) {
		if responseError.StatusCode == http.StatusBadRequest {
			for _, retryableBadRequest := range retryableBadRequests {
				for _, text := range responseError.VaultErrors {
					if strings.Contains(text, retryableBadRequest) {
						return true
					}
				}
			}
		}
	}

	return false
}

func (a *api) retryBadRequest(req *request) error {
	log := a.Logger.WithName("retryRequest").
		WithValues("method", req.Method).
		WithValues("path", req.Path)

	delay := req.getDelay()
	attemptLogger := log.WithValues("attempt", req.Attempts, "delay", delay)

	if delay > 0 {
		attemptLogger.Info("Waiting between retry attempts")
		time.Sleep(delay)
	}

	if err := a.tryMakingRequest(req); err != nil {
		if req.Attempts >= req.MaxAttempts {
			attemptLogger.Info("Too many retries, giving up...", "error", err)
			return err
		}

		if isBadRequestError(err) {
			attemptLogger.Info("Received bad request error, retrying...", "error", err)
			return a.retryBadRequest(req)
		}

		if isAuthError(err) {
			log.Info("Client got a bad request error, retrying...", "error", err)
			return a.authenticateAndRetryRequest(req)
		}

		attemptLogger.Info("failed to make api request to Vault", "error", err)

		return err
	}

	return nil
}
