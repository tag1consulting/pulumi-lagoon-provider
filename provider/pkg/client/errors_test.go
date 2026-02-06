package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestLagoonAPIError(t *testing.T) {
	err := &LagoonAPIError{Message: "something went wrong"}
	if err.Error() != "Lagoon API error: something went wrong" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLagoonAPIError_Is(t *testing.T) {
	err := &LagoonAPIError{Message: "test"}
	if !errors.Is(err, ErrAPI) {
		t.Error("expected LagoonAPIError to match ErrAPI")
	}
	if errors.Is(err, ErrConnection) {
		t.Error("LagoonAPIError should not match ErrConnection")
	}
	if errors.Is(err, ErrNotFound) {
		t.Error("LagoonAPIError should not match ErrNotFound")
	}
}

func TestLagoonConnectionError(t *testing.T) {
	cause := fmt.Errorf("dial timeout")
	err := &LagoonConnectionError{Message: "connection failed", Cause: cause}
	if err.Error() != "Lagoon connection error: connection failed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLagoonConnectionError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("dial timeout")
	err := &LagoonConnectionError{Message: "connection failed", Cause: cause}
	if err.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestLagoonConnectionError_Is(t *testing.T) {
	err := &LagoonConnectionError{Message: "test"}
	if !errors.Is(err, ErrConnection) {
		t.Error("expected LagoonConnectionError to match ErrConnection")
	}
	if errors.Is(err, ErrAPI) {
		t.Error("LagoonConnectionError should not match ErrAPI")
	}
}

func TestLagoonNotFoundError(t *testing.T) {
	err := &LagoonNotFoundError{ResourceType: "Project", Identifier: "myproject"}
	if err.Error() != "Project not found: myproject" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLagoonNotFoundError_Is(t *testing.T) {
	err := &LagoonNotFoundError{ResourceType: "Project", Identifier: "myproject"}
	if !errors.Is(err, ErrNotFound) {
		t.Error("expected LagoonNotFoundError to match ErrNotFound")
	}
	if errors.Is(err, ErrAPI) {
		t.Error("LagoonNotFoundError should not match ErrAPI")
	}
}

func TestLagoonValidationError(t *testing.T) {
	err := &LagoonValidationError{Field: "name", Message: "required"}
	if err.Error() != "validation error on field 'name': required" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLagoonValidationError_WithSuggestion(t *testing.T) {
	err := &LagoonValidationError{
		Field:      "name",
		Message:    "invalid characters",
		Suggestion: "use lowercase alphanumeric",
	}
	expected := "validation error on field 'name': invalid characters (suggestion: use lowercase alphanumeric)"
	if err.Error() != expected {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestLagoonValidationError_Is(t *testing.T) {
	err := &LagoonValidationError{Field: "name", Message: "required"}
	if !errors.Is(err, ErrValidation) {
		t.Error("expected LagoonValidationError to match ErrValidation")
	}
	if errors.Is(err, ErrNotFound) {
		t.Error("LagoonValidationError should not match ErrNotFound")
	}
}

func TestLagoonConnectionError_NilCause(t *testing.T) {
	err := &LagoonConnectionError{Message: "no cause"}
	if err.Unwrap() != nil {
		t.Error("Unwrap should return nil when cause is nil")
	}
}
