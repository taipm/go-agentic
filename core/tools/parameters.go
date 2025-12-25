package tools

import (
	"fmt"
	"strings"
)

// ParameterExtractor provides a fluent interface for extracting and validating
// tool parameters with deferred error checking.
//
// Usage:
//
//	pe := NewParameterExtractor(args).WithTool("MyTool")
//	name := pe.RequireString("name")
//	age := pe.OptionalInt("age", 0)
//	if err := pe.Errors(); err != nil {
//	    return "", err
//	}
//	// Proceed with extracted parameters...
type ParameterExtractor struct {
	args   map[string]interface{}
	errors []error
	tool   string
}

// NewParameterExtractor creates a new parameter extractor for the given arguments.
func NewParameterExtractor(args map[string]interface{}) *ParameterExtractor {
	if args == nil {
		args = make(map[string]interface{})
	}
	return &ParameterExtractor{
		args:   args,
		errors: make([]error, 0),
	}
}

// WithTool sets the tool name for error context.
// This is optional but helps with error messages.
func (pe *ParameterExtractor) WithTool(name string) *ParameterExtractor {
	pe.tool = name
	return pe
}

// RequireString extracts a required string parameter.
// Accumulates error if parameter is missing or cannot be coerced to string.
// Returns empty string on error (error collected for later checking).
func (pe *ParameterExtractor) RequireString(key string) string {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}
	return val
}

// RequireInt extracts a required int parameter.
// Accumulates error if parameter is missing or cannot be coerced to int.
// Returns 0 on error (error collected for later checking).
func (pe *ParameterExtractor) RequireInt(key string) int {
	val, err := MustGetInt(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return 0
	}
	return val
}

// RequireBool extracts a required bool parameter.
// Accumulates error if parameter is missing or cannot be coerced to bool.
// Returns false on error (error collected for later checking).
func (pe *ParameterExtractor) RequireBool(key string) bool {
	val, err := MustGetBool(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return false
	}
	return val
}

// RequireFloat extracts a required float parameter.
// Accumulates error if parameter is missing or cannot be coerced to float.
// Returns 0 on error (error collected for later checking).
func (pe *ParameterExtractor) RequireFloat(key string) float64 {
	val, err := MustGetFloat(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return 0
	}
	return val
}

// OptionalString extracts an optional string parameter with a default.
// Returns default value on missing parameter or coercion error.
// Does NOT accumulate errors (silent failure).
func (pe *ParameterExtractor) OptionalString(key string, def string) string {
	return OptionalGetString(pe.args, key, def)
}

// OptionalInt extracts an optional int parameter with a default.
// Returns default value on missing parameter or coercion error.
// Does NOT accumulate errors (silent failure).
func (pe *ParameterExtractor) OptionalInt(key string, def int) int {
	return OptionalGetInt(pe.args, key, def)
}

// OptionalBool extracts an optional bool parameter with a default.
// Returns default value on missing parameter or coercion error.
// Does NOT accumulate errors (silent failure).
func (pe *ParameterExtractor) OptionalBool(key string, def bool) bool {
	return OptionalGetBool(pe.args, key, def)
}

// OptionalFloat extracts an optional float parameter with a default.
// Returns default value on missing parameter or coercion error.
// Does NOT accumulate errors (silent failure).
func (pe *ParameterExtractor) OptionalFloat(key string, def float64) float64 {
	return OptionalGetFloat(pe.args, key, def)
}

// Errors returns a combined error of all accumulated errors, or nil if no errors.
// Should be called after all parameter extraction to validate.
func (pe *ParameterExtractor) Errors() error {
	if len(pe.errors) == 0 {
		return nil
	}
	if len(pe.errors) == 1 {
		return pe.errors[0]
	}

	// Multiple errors: combine with newlines
	var sb strings.Builder
	for i, err := range pe.errors {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(err.Error())
	}
	return fmt.Errorf("validation errors:\n%s", sb.String())
}

// HasErrors returns true if any errors have been accumulated.
func (pe *ParameterExtractor) HasErrors() bool {
	return len(pe.errors) > 0
}

// contextError adds tool name context to an error message if set.
func (pe *ParameterExtractor) contextError(key string, err error) error {
	if pe.tool != "" {
		return fmt.Errorf("tool %q, parameter %q: %w", pe.tool, key, err)
	}
	return fmt.Errorf("parameter %q: %w", key, err)
}

// ExtractAll validates that all accumulated errors are nil.
// Useful as a shorthand: defer pe.ExtractAll() after building
// (though Errors() method is more explicit).
// Note: This is a read-only helper; use Errors() for actual error checking.
func (pe *ParameterExtractor) AllValid() bool {
	return !pe.HasErrors()
}

// ============================================================================
// QW#4: ADVANCED FEATURES
// ============================================================================

// WithValidation adds a validation rule to a parameter (chaining helper)
// Usage: pe.RequireString("name").WithValidation(email)
// Note: Due to Go's type system, this returns string. Validation is added internally.
// Actual validation is: store validator, apply when param extracted
// For now, we recommend using validators directly via ValidationRules

// RequireStringWithValidation extracts required string with validation
func (pe *ParameterExtractor) RequireStringWithValidation(key string, validator func(interface{}) error) string {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	// Apply validation
	if validator != nil {
		if err := validator(val); err != nil {
			pe.errors = append(pe.errors, pe.contextError(key, err))
			return ""
		}
	}

	return val
}

// RequireIntWithValidation extracts required int with validation
func (pe *ParameterExtractor) RequireIntWithValidation(key string, validator func(interface{}) error) int {
	val, err := MustGetInt(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return 0
	}

	// Apply validation
	if validator != nil {
		if err := validator(val); err != nil {
			pe.errors = append(pe.errors, pe.contextError(key, err))
			return 0
		}
	}

	return val
}

// RequireWithMiddleware extracts string with middleware pipeline
func (pe *ParameterExtractor) RequireWithMiddleware(key string, middlewares ...Middleware) string {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	// Apply middleware chain
	chain := NewMiddlewareChain()
	for _, m := range middlewares {
		chain.Add(m)
	}
	result := chain.Execute(val)

	// Convert back to string
	if s, ok := result.(string); ok {
		return s
	}
	return ""
}

// OptionalWithMiddleware extracts optional string with middleware pipeline
func (pe *ParameterExtractor) OptionalWithMiddleware(key string, defaultVal string, middlewares ...Middleware) string {
	val := OptionalGetString(pe.args, key, defaultVal)

	// Apply middleware chain
	if val != "" && val != defaultVal {
		chain := NewMiddlewareChain()
		for _, m := range middlewares {
			chain.Add(m)
		}
		result := chain.Execute(val)
		if s, ok := result.(string); ok {
			return s
		}
	}
	return val
}

// RequireUUID extracts required UUID parameter
func (pe *ParameterExtractor) RequireUUID(key string) UUID {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	uuid, err := CoerceToUUID(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	return uuid
}

// RequireEmail extracts required email parameter
func (pe *ParameterExtractor) RequireEmail(key string) EmailAddress {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	email, err := CoerceToEmail(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	return email
}

// RequireURL extracts required URL parameter
func (pe *ParameterExtractor) RequireURL(key string) URL {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	url, err := CoerceToURL(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	return url
}

// RequirePhone extracts required phone parameter
func (pe *ParameterExtractor) RequirePhone(key string) PhoneNumber {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	phone, err := CoerceToPhone(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	return phone
}

// RequireDate extracts required date parameter
func (pe *ParameterExtractor) RequireDate(key string) Date {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return Date{}
	}

	date, err := CoerceToDate(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return Date{}
	}

	return date
}

// RequireDateTime extracts required datetime parameter
func (pe *ParameterExtractor) RequireDateTime(key string) DateTime {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return DateTime{}
	}

	dt, err := CoerceToDateTime(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return DateTime{}
	}

	return dt
}

// RequireSlug extracts required slug parameter
func (pe *ParameterExtractor) RequireSlug(key string) Slug {
	val, err := MustGetString(pe.args, key)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	slug, err := CoerceToSlug(val)
	if err != nil {
		pe.errors = append(pe.errors, pe.contextError(key, err))
		return ""
	}

	return slug
}

// OptionalUUID extracts optional UUID parameter
func (pe *ParameterExtractor) OptionalUUID(key string, defaultVal UUID) UUID {
	val, exists := pe.args[key]
	if !exists {
		return defaultVal
	}

	uuid, err := CoerceToUUID(val)
	if err != nil {
		return defaultVal
	}

	return uuid
}

// OptionalEmail extracts optional email parameter
func (pe *ParameterExtractor) OptionalEmail(key string, defaultVal EmailAddress) EmailAddress {
	val, exists := pe.args[key]
	if !exists {
		return defaultVal
	}

	email, err := CoerceToEmail(val)
	if err != nil {
		return defaultVal
	}

	return email
}

// OptionalDate extracts optional date parameter
func (pe *ParameterExtractor) OptionalDate(key string, defaultVal Date) Date {
	val, exists := pe.args[key]
	if !exists {
		return defaultVal
	}

	date, err := CoerceToDate(val)
	if err != nil {
		return defaultVal
	}

	return date
}
