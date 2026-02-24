package client

import (
	"errors"
	"fmt"
	"strings"
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

// Sentinel errors for errors.Is() checking.
var (
	ErrAPI            = errors.New("lagoon api error")
	ErrConnection     = errors.New("lagoon connection error")
	ErrNotFound       = errors.New("lagoon resource not found")
	ErrValidation     = errors.New("lagoon validation error")
	ErrDuplicateEntry = errors.New("lagoon duplicate entry")
)

// IsDuplicateEntry returns true if the error is a Lagoon API error indicating
// the resource already exists. This can be a MySQL duplicate entry constraint
// violation or an application-level "already exists" error.
func IsDuplicateEntry(err error) bool {
	var apiErr *LagoonAPIError
	if errors.As(err, &apiErr) {
		msg := strings.ToLower(apiErr.Message)
		return strings.Contains(msg, "duplicate entry") ||
			strings.Contains(msg, "already exists")
	}
	return false
}

// Is enables errors.Is() support for typed errors.
func (e *LagoonAPIError) Is(target error) bool        { return target == ErrAPI }
func (e *LagoonConnectionError) Is(target error) bool  { return target == ErrConnection }
func (e *LagoonNotFoundError) Is(target error) bool    { return target == ErrNotFound }
func (e *LagoonValidationError) Is(target error) bool  { return target == ErrValidation }
