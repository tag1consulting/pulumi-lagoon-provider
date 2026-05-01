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

func TestIsDuplicateEntry(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "duplicate entry error",
			err:      &LagoonAPIError{Message: "(conn:10, no: 1062, SQLState: 23000) Duplicate entry 'lagoon-prod' for key 'name'"},
			expected: true,
		},
		{
			name:     "other API error",
			err:      &LagoonAPIError{Message: "permission denied"},
			expected: false,
		},
		{
			name:     "wrapped duplicate entry",
			err:      fmt.Errorf("create failed: %w", &LagoonAPIError{Message: "Duplicate entry 'test' for key 'name'"}),
			expected: true,
		},
		{
			name:     "non-API error",
			err:      fmt.Errorf("network timeout"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "project already exists",
			err:      &LagoonAPIError{Message: "Error creating project 'drupal-example'. Project already exists."},
			expected: true,
		},
		{
			name:     "wrapped already exists",
			err:      fmt.Errorf("create failed: %w", &LagoonAPIError{Message: "Resource already exists"}),
			expected: true,
		},
		{
			name: "duplicate in nested Errors slice",
			err: &LagoonAPIError{
				Message: "graphql error",
				Errors:  []GraphQLError{{Message: "Duplicate entry 'foo' for key 'name'"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateEntry(tt.err); got != tt.expected {
				t.Errorf("IsDuplicateEntry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContainsNotFound(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"user not found", true},
		{"User not found", true},
		{"USER NOT FOUND", true},
		{"user does not exist", true},
		{"no user found", true},
		{"No user found with that email", true},
		{"access denied: no user permissions for this operation", false},
		{"permission denied", false},
		{"not found", false},
		{"group not found", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			if got := containsNotFound(tt.msg); got != tt.want {
				t.Errorf("containsNotFound(%q) = %v, want %v", tt.msg, got, tt.want)
			}
		})
	}
}

func TestLagoonAPIError_Is_ErrDuplicateEntry(t *testing.T) {
	// Test errors.Is(apiErr, ErrDuplicateEntry) via the Is() method directly.
	tests := []struct {
		name string
		err  *LagoonAPIError
		want bool
	}{
		{
			name: "top-level duplicate message",
			err:  &LagoonAPIError{Message: "Duplicate entry 'foo' for key 'name'"},
			want: true,
		},
		{
			name: "top-level already exists message",
			err:  &LagoonAPIError{Message: "Resource already exists"},
			want: true,
		},
		{
			name: "nested Errors duplicate message",
			err: &LagoonAPIError{
				Message: "graphql error",
				Errors:  []GraphQLError{{Message: "Duplicate entry 'bar'"}},
			},
			want: true,
		},
		{
			name: "unrelated API error",
			err:  &LagoonAPIError{Message: "some other error"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, ErrDuplicateEntry)
			if got != tt.want {
				t.Errorf("errors.Is(%v, ErrDuplicateEntry) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
