package tools

import (
	"testing"
)

// TestNotEmpty tests NotEmpty validator
func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantError bool
	}{
		{"non_empty", "hello", false},
		{"empty_string", "", true},
		{"whitespace_only", "   ", true},
		{"not_string", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NotEmpty(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("NotEmpty() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestEmail tests Email validator
func TestEmail(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_email", "user@example.com", false},
		{"valid_email_subdomain", "user@mail.example.com", false},
		{"invalid_no_at", "userexample.com", true},
		{"invalid_no_domain", "user@.com", true},
		{"invalid_format", "user@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Email() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMinLength tests MinLength validator
func TestMinLength(t *testing.T) {
	validator := MinLength(5)

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
	}{
		{"valid_length", "hello", false},
		{"too_short", "hi", true},
		{"exact_length", "12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("MinLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestRange tests Range validator
func TestRange(t *testing.T) {
	validator := Range(1, 100)

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
	}{
		{"in_range", 50, false},
		{"min_boundary", 1, false},
		{"max_boundary", 100, false},
		{"below_range", 0, true},
		{"above_range", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Range() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestOneOf tests OneOf validator
func TestOneOf(t *testing.T) {
	validator := OneOf("red", "green", "blue")

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_red", "red", false},
		{"valid_green", "green", false},
		{"invalid_yellow", "yellow", true},
		{"invalid_empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("OneOf() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestCombine tests combining multiple validators
func TestCombine(t *testing.T) {
	validator := Combine(
		NotEmpty,
		Email,
	)

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_email", "user@example.com", false},
		{"empty_string", "", true},
		{"invalid_format", "notanemail", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Combine() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidationRules tests ValidationRules collection
func TestValidationRules(t *testing.T) {
	rules := NewValidationRules().
		AddRule("email", Email).
		AddRule("minlen", MinLength(3))

	if !rules.HasRule("email") {
		t.Error("Expected rule 'email' to exist")
	}

	if rules.HasRule("nonexistent") {
		t.Error("Expected rule 'nonexistent' to not exist")
	}

	if rules.GetRule("email") == nil {
		t.Error("Expected GetRule to return non-nil")
	}
}
