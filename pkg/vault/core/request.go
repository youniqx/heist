package core

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/youniqx/heist/pkg/httpclient"
)

type request struct {
	Method       string
	Path         string
	RequestBody  httpclient.Encodable
	ResponseBody httpclient.Decodeable
	Attempts     int
	MaxAttempts  int
	Parameters   map[string]string
}

func (r *request) incrementAttempts() {
	r.Attempts++
}

func (r *request) getDelay() time.Duration {
	if r.Attempts == 0 {
		return 0
	}

	delay := minimumRetryDelay

	for i := 1; i < r.Attempts; i++ {
		delay *= 2
	}

	if delay > maximumRetryDelay {
		delay = maximumRetryDelay
	}

	return delay
}

const (
	minimumRetryDelay = 1 * time.Second
	maximumRetryDelay = 16 * time.Second
	maxAttempts       = 2
)

func (a *api) MakeRequest(method RequestType, path string, requestBody httpclient.Encodable, responseBody httpclient.Decodeable) error {
	log := a.Logger.WithName("request").
		WithValues("method", method).
		WithValues("path", path)

	info := &request{
		Method:       string(method),
		Path:         path,
		RequestBody:  requestBody,
		ResponseBody: responseBody,
		Parameters:   make(map[string]string),
		Attempts:     0,
		MaxAttempts:  maxAttempts,
	}

	if info.Method == string(MethodList) {
		info.Method = http.MethodGet
		info.Parameters["list"] = "true"
	}

	if err := a.tryMakingRequest(info); err != nil {
		if isAuthError(err) {
			log.Info("Client is not authenticated, attempting to authenticate again", "error", err)
			return a.authenticateAndRetryRequest(info)
		}

		if isBadRequestError(err) {
			log.Info("Client got a bad request error, retrying...", "error", err)
			return a.retryBadRequest(info)
		}

		log.Info("failed to make api request", "error", err)

		return err
	}

	return nil
}

func (a *api) tryMakingRequest(requestInfo *request) error {
	defer requestInfo.incrementAttempts()

	errorResponse := &ErrorResponse{}

	options := []httpclient.Option{
		httpclient.Path(requestInfo.Path),
		httpclient.Method(requestInfo.Method),
		httpclient.Header("X-Vault-Token", a.Token),
		httpclient.Request(requestInfo.RequestBody),
		httpclient.Response(requestInfo.ResponseBody),
		httpclient.Response(httpclient.JSON(errorResponse, httpclient.ConstraintFailed)),
	}

	for k, v := range requestInfo.Parameters {
		options = append(options, httpclient.Parameter(k, v))
	}

	result, err := a.Client.Perform(options...)
	if err != nil {
		return ErrHTTPError.WithDetails("failed to make api request to Vault").WithCause(err)
	}

	if result.StatusCode >= http.StatusOK && result.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	return &VaultHTTPError{
		erxError:    ErrHTTPError.WithDetails(fmt.Sprintf("received error status code: %d\n\t%s", result.StatusCode, strings.Join(errorResponse.Errors, "\n\t"))),
		StatusCode:  result.StatusCode,
		VaultErrors: errorResponse.Errors,
	}
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}
