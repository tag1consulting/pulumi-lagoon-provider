package client

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// LagoonAPIError represents an error returned by the Lagoon GraphQL API.
type LagoonAPIError struct {
	Message string
	Errors  []GraphQLError
}

// GraphQLError represents a single GraphQL error from the API response.
type GraphQLError struct {
	Message string `json:"message"`
}

func (e *LagoonAPIError) Error() string {
	return fmt.Sprintf("Lagoon API error: %s", e.Message)
}

// LagoonConnectionError represents a network/transport error connecting to the Lagoon API.
type LagoonConnectionError struct {
	Message string
	Cause   error
}

func (e *LagoonConnectionError) Error() string {
	return fmt.Sprintf("Lagoon connection error: %s", e.Message)
}

func (e *LagoonConnectionError) Unwrap() error {
	return e.Cause
}

// LagoonNotFoundError represents a resource that was not found.
type LagoonNotFoundError struct {
	ResourceType string
	Identifier   string
}

func (e *LagoonNotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.Identifier)
}

// LagoonValidationError represents an input validation error.
type LagoonValidationError struct {
	Field      string
	Value      any
	Message    string
	Suggestion string
}

func (e *LagoonValidationError) Error() string {
	msg := fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	if e.Suggestion != "" {
		msg += fmt.Sprintf(" (suggestion: %s)", e.Suggestion)
	}
	return msg
}

// LagoonRateLimitError represents an HTTP 429 Too Many Requests response.
// It is retryable; RetryAfter indicates the minimum wait before the next attempt.
type LagoonRateLimitError struct {
	RetryAfter time.Duration
}

func (e *LagoonRateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("Lagoon rate limit exceeded; retry after %s", e.RetryAfter)
	}
	return "Lagoon rate limit exceeded"
}

// Sentinel errors for errors.Is() checking.
var (
	ErrAPI            = errors.New("lagoon api error")
	ErrConnection     = errors.New("lagoon connection error")
	ErrNotFound       = errors.New("lagoon resource not found")
	ErrValidation     = errors.New("lagoon validation error")
	ErrDuplicateEntry = errors.New("lagoon duplicate entry")
	ErrRateLimit      = errors.New("lagoon rate limit")
)

// isDuplicateMessage returns true if a message indicates a duplicate entry.
func isDuplicateMessage(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "duplicate entry") ||
		strings.Contains(lower, "already exists")
}

// IsDuplicateEntry returns true if the error is a Lagoon API error indicating
// the resource already exists. This can be a MySQL duplicate entry constraint
// violation or an application-level "already exists" error.
func IsDuplicateEntry(err error) bool {
	var apiErr *LagoonAPIError
	if errors.As(err, &apiErr) {
		if isDuplicateMessage(apiErr.Message) {
			return true
		}
		for _, e := range apiErr.Errors {
			if isDuplicateMessage(e.Message) {
				return true
			}
		}
	}
	return false
}

// Is enables errors.Is() support for typed errors.
func (e *LagoonAPIError) Is(target error) bool {
	if target == ErrAPI {
		return true
	}
	if target == ErrDuplicateEntry {
		if isDuplicateMessage(e.Message) {
			return true
		}
		for _, ge := range e.Errors {
			if isDuplicateMessage(ge.Message) {
				return true
			}
		}
	}
	return false
}
func (e *LagoonConnectionError) Is(target error) bool  { return target == ErrConnection }
func (e *LagoonNotFoundError) Is(target error) bool    { return target == ErrNotFound }
func (e *LagoonValidationError) Is(target error) bool  { return target == ErrValidation }
func (e *LagoonRateLimitError) Is(target error) bool   { return target == ErrRateLimit }
