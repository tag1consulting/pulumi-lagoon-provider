package resources

import "testing"

func TestSetOptional(t *testing.T) {
	m := map[string]any{}
	setOptional(m, "key", nil)
	if _, ok := m["key"]; ok {
		t.Error("nil pointer should not set key")
	}

	val := "hello"
	setOptional(m, "key", &val)
	if m["key"] != "hello" {
		t.Errorf("expected 'hello', got %v", m["key"])
	}
}

func TestSetOptionalInt(t *testing.T) {
	m := map[string]any{}
	setOptionalInt(m, "key", nil)
	if _, ok := m["key"]; ok {
		t.Error("nil pointer should not set key")
	}

	val := 42
	setOptionalInt(m, "key", &val)
	if m["key"] != 42 {
		t.Errorf("expected 42, got %v", m["key"])
	}
}

func TestSetOptionalBool(t *testing.T) {
	m := map[string]any{}
	setOptionalBool(m, "key", nil)
	if _, ok := m["key"]; ok {
		t.Error("nil pointer should not set key")
	}

	val := true
	setOptionalBool(m, "key", &val)
	if m["key"] != true {
		t.Errorf("expected true, got %v", m["key"])
	}
}

func TestPtrDiffers(t *testing.T) {
	tests := []struct {
		name string
		a, b *string
		want bool
	}{
		{"both nil", nil, nil, false},
		{"a nil", nil, strPtr("x"), true},
		{"b nil", strPtr("x"), nil, true},
		{"same", strPtr("x"), strPtr("x"), false},
		{"different", strPtr("x"), strPtr("y"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrDiffers(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ptrDiffers(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestPtrIntDiffers(t *testing.T) {
	tests := []struct {
		name string
		a, b *int
		want bool
	}{
		{"both nil", nil, nil, false},
		{"a nil", nil, intPtr(1), true},
		{"b nil", intPtr(1), nil, true},
		{"same", intPtr(1), intPtr(1), false},
		{"different", intPtr(1), intPtr(2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrIntDiffers(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ptrIntDiffers(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestPtrBoolDiffers(t *testing.T) {
	tests := []struct {
		name string
		a, b *bool
		want bool
	}{
		{"both nil", nil, nil, false},
		{"a nil", nil, boolPtr(true), true},
		{"b nil", boolPtr(true), nil, true},
		{"same true", boolPtr(true), boolPtr(true), false},
		{"same false", boolPtr(false), boolPtr(false), false},
		{"different", boolPtr(true), boolPtr(false), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrBoolDiffers(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ptrBoolDiffers(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// helpers for tests
func strPtr(s string) *string  { return &s }
func intPtr(i int) *int        { return &i }
func boolPtr(b bool) *bool     { return &b }
