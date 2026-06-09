package resources

import (
	"strings"
	"testing"
)

func TestValidateVariableScope(t *testing.T) {
	valid := []string{
		"build", "runtime", "global",
		"container_registry", "internal_container_registry",
		"BUILD", "Runtime", "GLOBAL",
	}
	for _, s := range valid {
		if err := validateVariableScope(s); err != nil {
			t.Errorf("scope %q should be valid, got error: %v", s, err)
		}
	}

	invalid := []string{"", "staging", "production", "dev"}
	for _, s := range invalid {
		if err := validateVariableScope(s); err == nil {
			t.Errorf("scope %q should be invalid, got no error", s)
		}
	}
}

func TestValidateTaskPermission(t *testing.T) {
	valid := []string{"guest", "developer", "maintainer", "GUEST", "Developer"}
	for _, p := range valid {
		if err := validateTaskPermission(p); err != nil {
			t.Errorf("permission %q should be valid, got error: %v", p, err)
		}
	}

	invalid := []string{"", "admin", "owner", "viewer"}
	for _, p := range invalid {
		if err := validateTaskPermission(p); err == nil {
			t.Errorf("permission %q should be invalid, got no error", p)
		}
	}
}

func TestValidateTaskArgType(t *testing.T) {
	valid := []string{
		"string", "STRING",
		"environment_source_name", "ENVIRONMENT_SOURCE_NAME",
		"environment_source_name_exclude_self", "ENVIRONMENT_SOURCE_NAME_EXCLUDE_SELF",
	}
	for _, typ := range valid {
		if err := validateTaskArgType(typ); err != nil {
			t.Errorf("arg type %q should be valid, got error: %v", typ, err)
		}
	}

	invalid := []string{"", "int", "bool", "number"}
	for _, typ := range invalid {
		if err := validateTaskArgType(typ); err == nil {
			t.Errorf("arg type %q should be invalid, got no error", typ)
		}
	}
}

func TestValidatePositiveID(t *testing.T) {
	if err := validatePositiveID("projectId", 1); err != nil {
		t.Errorf("ID=1 should be valid, got: %v", err)
	}
	if err := validatePositiveID("projectId", 100); err != nil {
		t.Errorf("ID=100 should be valid, got: %v", err)
	}

	if err := validatePositiveID("projectId", 0); err == nil {
		t.Error("ID=0 should be invalid, got no error")
	}
	if err := validatePositiveID("projectId", -1); err == nil {
		t.Error("ID=-1 should be invalid, got no error")
	}
}

func TestValidateRouteListLimits(t *testing.T) {
	makeStrings := func(n int) []string {
		s := make([]string, n)
		for i := range s {
			s[i] = "item"
		}
		return s
	}
	makeAnnotations := func(n int) []RouteAnnotationInput {
		a := make([]RouteAnnotationInput, n)
		for i := range a {
			a[i] = RouteAnnotationInput{Key: "k", Value: "v"}
		}
		return a
	}
	makePathRoutes := func(n int) []RoutePathRouteInput {
		r := make([]RoutePathRouteInput, n)
		for i := range r {
			r[i] = RoutePathRouteInput{Path: "/", ToService: "svc"}
		}
		return r
	}

	// Valid counts
	ann10 := makeAnnotations(10)
	alt25 := makeStrings(25)
	pr10 := makePathRoutes(10)
	args := RouteArgs{
		ProjectName:      "proj",
		Domain:           "example.com",
		Annotations:      &ann10,
		AlternativeNames: &alt25,
		PathRoutes:       &pr10,
	}
	if err := validateRouteListLimits(args); err != nil {
		t.Errorf("at-limit counts should be valid, got: %v", err)
	}

	// Over-limit annotations
	ann11 := makeAnnotations(11)
	args.Annotations = &ann11
	if err := validateRouteListLimits(args); err == nil {
		t.Error("11 annotations should be invalid, got no error")
	} else if !strings.Contains(err.Error(), "annotations") {
		t.Errorf("error should mention 'annotations', got: %v", err)
	}

	// Over-limit alternativeNames
	args.Annotations = &ann10
	alt26 := makeStrings(26)
	args.AlternativeNames = &alt26
	if err := validateRouteListLimits(args); err == nil {
		t.Error("26 alternativeNames should be invalid, got no error")
	} else if !strings.Contains(err.Error(), "alternativeNames") {
		t.Errorf("error should mention 'alternativeNames', got: %v", err)
	}

	// Over-limit pathRoutes
	args.AlternativeNames = &alt25
	pr11 := makePathRoutes(11)
	args.PathRoutes = &pr11
	if err := validateRouteListLimits(args); err == nil {
		t.Error("11 pathRoutes should be invalid, got no error")
	} else if !strings.Contains(err.Error(), "pathRoutes") {
		t.Errorf("error should mention 'pathRoutes', got: %v", err)
	}
}
