package core

import (
	"errors"
	"net/http"
	"time"
)

func isAuthError(err error) bool {
	var responseError *VaultHTTPError
	if errors.As(err, &responseError) {
		if responseError.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return false
}

func (a *api) authenticateAndRetryRequest(req *request) error {
	log := a.Logger.WithName("authenticateAndRetryRequest").
		WithValues("method", req.Method).
		WithValues("path", req.Path)

	if a.AuthValidUntil.After(time.Now()) {
		if err := a.tryMakingRequest(req); err != nil {
			if !isAuthError(err) {
				log.Info("failed to make api request to Vault", "error", err)
				return err
			}
		} else {
			return nil
		}
	}

	delay := req.getDelay()
	attemptLogger := log.WithValues("attempt", req.Attempts, "delay", delay)

	if delay > 0 {
		attemptLogger.Info("Waiting between authorization attempts")
		time.Sleep(delay)
	}

	if err := a.authenticate(); err != nil {
		attemptLogger.Info("failed to authenticate in Vault")
		return a.authenticateAndRetryRequest(req)
	}

	if err := a.tryMakingRequest(req); err != nil {
		if req.Attempts >= req.MaxAttempts {
			attemptLogger.Info("Too many retries, giving up...", "error", err)
			return err
		}

		if isAuthError(err) {
			attemptLogger.Info("Authentication in Vault failed", "error", err)
			return a.authenticateAndRetryRequest(req)
		}

		if isBadRequestError(err) {
			attemptLogger.Info("Client received a Bad Request Error, retrying...", "error", err)
			return a.retryBadRequest(req)
		}

		attemptLogger.Info("failed to make api request to Vault", "error", err)

		return err
	}

	return nil
}
