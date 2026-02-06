package resources

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
