package resources

import (
	"fmt"
	"strings"
)

var validVariableScopes = map[string]bool{
	"build":                         true,
	"runtime":                       true,
	"global":                        true,
	"container_registry":            true,
	"internal_container_registry":   true,
}

var validTaskPermissions = map[string]bool{
	"guest":      true,
	"developer":  true,
	"maintainer": true,
}

var validTaskArgTypes = map[string]bool{
	"string":                            true,
	"environment_source_name":           true,
	"environment_source_name_exclude_self": true,
}

// validateVariableScope returns an error if scope is not one of the accepted values.
func validateVariableScope(scope string) error {
	if !validVariableScopes[strings.ToLower(scope)] {
		return fmt.Errorf("invalid variable scope %q: must be one of build, runtime, global, container_registry, internal_container_registry", scope)
	}
	return nil
}

// validateTaskPermission returns an error if permission is not one of the accepted values.
func validateTaskPermission(permission string) error {
	if !validTaskPermissions[strings.ToLower(permission)] {
		return fmt.Errorf("invalid task permission %q: must be one of guest, developer, maintainer", permission)
	}
	return nil
}

// validateTaskArgType returns an error if the argument type is not one of the accepted values.
func validateTaskArgType(argType string) error {
	if !validTaskArgTypes[strings.ToLower(argType)] {
		return fmt.Errorf("invalid task argument type %q: must be one of STRING, ENVIRONMENT_SOURCE_NAME, ENVIRONMENT_SOURCE_NAME_EXCLUDE_SELF", argType)
	}
	return nil
}

// validatePositiveID returns an error if id is not positive.
func validatePositiveID(field string, id int) error {
	if id <= 0 {
		return fmt.Errorf("%s must be a positive integer, got %d", field, id)
	}
	return nil
}

// validateRouteListLimits checks annotation, alternativeNames, and pathRoute counts.
func validateRouteListLimits(args RouteArgs) error {
	if args.Annotations != nil && len(*args.Annotations) > 10 {
		return fmt.Errorf("annotations exceeds maximum of 10 entries (got %d)", len(*args.Annotations))
	}
	if args.AlternativeNames != nil && len(*args.AlternativeNames) > 25 {
		return fmt.Errorf("alternativeNames exceeds maximum of 25 entries (got %d)", len(*args.AlternativeNames))
	}
	if args.PathRoutes != nil && len(*args.PathRoutes) > 10 {
		return fmt.Errorf("pathRoutes exceeds maximum of 10 entries (got %d)", len(*args.PathRoutes))
	}
	return nil
}
