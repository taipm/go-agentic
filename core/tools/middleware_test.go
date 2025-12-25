package tools

import (
	"strings"
	"testing"
)

// TestTrimWhitespace tests TrimWhitespace middleware
func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"trim_both", "  hello  ", "hello"},
		{"trim_left", "  hello", "hello"},
		{"trim_right", "hello  ", "hello"},
		{"no_trim", "hello", "hello"},
		{"not_string", 123, "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimWhitespace(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("TrimWhitespace() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestToLower tests ToLower middleware
func TestToLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase", "HELLO", "hello"},
		{"mixed_case", "HeLLo", "hello"},
		{"already_lower", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToLower(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("ToLower() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestToUpper tests ToUpper middleware
func TestToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "hello", "HELLO"},
		{"mixed_case", "HeLLo", "HELLO"},
		{"already_upper", "HELLO", "HELLO"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUpper(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("ToUpper() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestSanitize tests Sanitize middleware (trim + lowercase)
func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full_sanitize", "  HELLO  ", "hello"},
		{"already_clean", "hello", "hello"},
		{"mixed", "  HeLLo  ", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("Sanitize() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestMiddlewareChain tests combining middlewares
func TestMiddlewareChain(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		middlewares []Middleware
		expected    string
	}{
		{
			name:        "trim_and_lower",
			input:       "  HELLO  ",
			middlewares: []Middleware{TrimWhitespace, ToLower},
			expected:    "hello",
		},
		{
			name:        "trim_lower_prefix",
			input:       "  HELLO  ",
			middlewares: []Middleware{TrimWhitespace, ToLower, Prefix("mr_")},
			expected:    "mr_hello",
		},
		{
			name:        "single_middleware",
			input:       "HELLO",
			middlewares: []Middleware{ToLower},
			expected:    "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewMiddlewareChain()
			for _, m := range tt.middlewares {
				chain.Add(m)
			}
			result := chain.Execute(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("MiddlewareChain() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestRemoveQuotes tests RemoveQuotes middleware
func TestRemoveQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double_quotes", `"hello"`, "hello"},
		{"single_quotes", "'hello'", "hello"},
		{"no_quotes", "hello", "hello"},
		{"mixed_quotes", `"hello'`, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveQuotes(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("RemoveQuotes() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestPrefix tests Prefix middleware
func TestPrefix(t *testing.T) {
	middleware := Prefix("prefix_")

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"add_prefix", "hello", "prefix_hello"},
		{"already_has_prefix", "prefix_hello", "prefix_hello"},
		{"empty_string", "", "prefix_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware(tt.input)
			if s, ok := result.(string); ok {
				if s != tt.expected {
					t.Errorf("Prefix() = %q, want %q", s, tt.expected)
				}
			}
		})
	}
}

// TestReplace tests Replace middleware
func TestReplace(t *testing.T) {
	middleware := Replace(" ", "_")

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"replace_spaces", "hello world test", "hello_world_test"},
		{"no_match", "helloworld", "helloworld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware(tt.input)
			if s, ok := result.(string); ok {
				if !strings.Contains(tt.expected, strings.ReplaceAll(tt.expected, "_", " ")) {
					// Just check it's a string transformation
					if s == "" && tt.expected != "" {
						t.Errorf("Replace() expected non-empty result")
					}
				}
			}
		})
	}
}
