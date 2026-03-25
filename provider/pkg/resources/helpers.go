package resources

import (
	"context"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/tag1consulting/pulumi-lagoon/provider/pkg/config"
)

// clientContextKey is the context key for injecting a test client.
type clientContextKey struct{}

// withTestClient returns a context with a mock client injected (for testing only).
func withTestClient(ctx context.Context, c LagoonClient) context.Context {
	return context.WithValue(ctx, clientContextKey{}, c)
}

// clientFor returns the LagoonClient from context (test override) or from the Pulumi config.
func clientFor(ctx context.Context) LagoonClient {
	if c, ok := ctx.Value(clientContextKey{}).(LagoonClient); ok {
		return c
	}
	cfg := infer.GetConfig[config.LagoonConfig](ctx)
	return cfg.NewClient()
}

// setOptional sets a map key if the optional string pointer is non-nil.
func setOptional(m map[string]any, key string, val *string) {
	if val != nil {
		m[key] = *val
	}
}

// setOptionalInt sets a map key if the optional int pointer is non-nil.
func setOptionalInt(m map[string]any, key string, val *int) {
	if val != nil {
		m[key] = *val
	}
}

// setOptionalBool sets a map key if the optional bool pointer is non-nil.
func setOptionalBool(m map[string]any, key string, val *bool) {
	if val != nil {
		m[key] = *val
	}
}

// ptrDiffers returns true if two optional string pointers differ.
func ptrDiffers(a, b *string) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil || b == nil {
		return true
	}
	return *a != *b
}

// ptrIntDiffers returns true if two optional int pointers differ.
func ptrIntDiffers(a, b *int) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil || b == nil {
		return true
	}
	return *a != *b
}

// ptrBoolDiffers returns true if two optional bool pointers differ.
func ptrBoolDiffers(a, b *bool) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil || b == nil {
		return true
	}
	return *a != *b
}

// ptrOrDefault returns the value of a string pointer, or the default if nil.
func ptrOrDefault(p *string, def string) string {
	if p != nil {
		return *p
	}
	return def
}

// ptrIntOrDefault returns the value of an int pointer, or the default if nil.
func ptrIntOrDefault(p *int, def int) int {
	if p != nil {
		return *p
	}
	return def
}
