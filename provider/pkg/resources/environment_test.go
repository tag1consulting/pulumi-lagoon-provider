package resources

import (
	"strings"
	"testing"
)

func TestValidateEnvironmentEnums(t *testing.T) {
	tests := []struct {
		name       string
		deployType string
		envType    string
		wantErr    bool
		errContain string
	}{
		// Valid deploy types (case variations) with valid env type
		{"branch lowercase", "branch", "production", false, ""},
		{"BRANCH uppercase", "BRANCH", "production", false, ""},
		{"Branch mixed", "Branch", "production", false, ""},
		{"pullrequest lowercase", "pullrequest", "production", false, ""},
		{"PULLREQUEST uppercase", "PULLREQUEST", "production", false, ""},
		{"promote lowercase", "promote", "production", false, ""},
		{"PROMOTE uppercase", "PROMOTE", "production", false, ""},

		// Valid env types (case variations) with valid deploy type
		{"production lowercase", "branch", "production", false, ""},
		{"PRODUCTION uppercase", "branch", "PRODUCTION", false, ""},
		{"Production mixed", "branch", "Production", false, ""},
		{"development lowercase", "branch", "development", false, ""},
		{"DEVELOPMENT uppercase", "branch", "DEVELOPMENT", false, ""},
		{"Development mixed", "branch", "Development", false, ""},

		// Invalid deploy type
		{"invalid deployType", "invalid", "production", true, "invalid deployType"},
		{"empty deployType", "", "production", true, "invalid deployType"},
		{"typo deployType", "branc", "production", true, "invalid deployType"},

		// Invalid env type (valid deploy type)
		{"invalid envType", "branch", "staging", true, "invalid environmentType"},
		{"empty envType", "branch", "", true, "invalid environmentType"},
		{"typo envType", "branch", "prod", true, "invalid environmentType"},

		// Both empty
		{"both empty", "", "", true, "invalid deployType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnvironmentEnums(tt.deployType, tt.envType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContain)
				} else if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}
