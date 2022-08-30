package erx

import (
	"errors"
	"fmt"
)

type Error interface {
	error
	GetContext() string
	GetReason() string
	GetDetails() string
	GetCause() error
	WithReason(reason string) Error
	WithCause(cause error) Error
	WithDetails(details string) Error
	Copy() Error
}

type instance struct {
	Context string
	Reason  string
	Details string
	Cause   error
}

func (e *instance) GetContext() string {
	return e.Context
}

func (e *instance) GetReason() string {
	return e.Reason
}

func (e *instance) GetDetails() string {
	return e.Details
}

func (e *instance) GetCause() error {
	return e.Cause
}

func New(context string, reason string) Error {
	return &instance{
		Context: context,
		Reason:  reason,
		Details: "",
		Cause:   nil,
	}
}

func (e *instance) WithReason(reason string) Error {
	return &instance{
		Context: e.Context,
		Reason:  reason,
		Details: e.Details,
		Cause:   e.Cause,
	}
}

func (e *instance) WithCause(cause error) Error {
	return &instance{
		Context: e.Context,
		Reason:  e.Reason,
		Details: e.Details,
		Cause:   cause,
	}
}

func (e *instance) WithDetails(details string) Error {
	return &instance{
		Context: e.Context,
		Reason:  e.Reason,
		Details: details,
		Cause:   e.Cause,
	}
}

func (e *instance) Copy() Error {
	return &instance{
		Context: e.Context,
		Reason:  e.Reason,
		Details: e.Details,
		Cause:   e.Cause,
	}
}

func FormatError(err Error) string {
	switch {
	case err.GetDetails() != "" && err.GetCause() != nil:
		return fmt.Sprintf("\n[%s] %s: %s -> %v", err.GetContext(), err.GetReason(), err.GetDetails(), err.GetCause())
	case err.GetDetails() != "":
		return fmt.Sprintf("\n[%s] %s: %s", err.GetContext(), err.GetReason(), err.GetDetails())
	case err.GetCause() != nil:
		return fmt.Sprintf("\n[%s] %s -> %v", err.GetContext(), err.GetReason(), err.GetCause())
	default:
		return fmt.Sprintf("\n[%s] %s", err.GetContext(), err.GetReason())
	}
}

func Is(err Error, other error) bool {
	var otherErx Error
	if errors.As(other, &otherErx) {
		return err.GetContext() == otherErx.GetContext() && err.GetReason() == otherErx.GetReason()
	}

	return false
}

func Unwrap(err Error) error {
	return err.GetCause()
}

func (e *instance) Error() string {
	return FormatError(e)
}

func (e *instance) Is(err error) bool {
	return Is(e, err)
}

func (e *instance) Unwrap() error {
	return Unwrap(e)
}
