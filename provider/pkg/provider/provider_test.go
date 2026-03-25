package provider

import (
	"testing"
)

func TestNewProvider_BuildsWithoutError(t *testing.T) {
	p, err := NewProvider("0.0.0-test")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	_ = p // non-nil interface value; successful build is sufficient
}
