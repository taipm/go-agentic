package tools

import (
	"testing"
	"time"
)

// TestUUID tests UUID validation and conversion
func TestUUID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid_format", "not-a-uuid", true},
		{"invalid_version", "550e8400-e29b-31d4-a716-446655440000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CoerceToUUID(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToUUID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestEmail tests email validation and conversion
func TestCoerceEmailType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_email", "user@example.com", false},
		{"invalid_format", "not_an_email", true},
		{"empty_string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CoerceToEmail(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToEmail() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestURL tests URL validation and conversion
func TestCoerceURLType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_http", "http://example.com", false},
		{"valid_https", "https://example.com/path", false},
		{"no_protocol", "example.com", true},
		{"ftp_protocol", "ftp://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CoerceToURL(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToURL() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestPhone tests phone validation and conversion
func TestCoercePhoneType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_phone", "1234567890", false},
		{"valid_with_plus", "+1234567890", false},
		{"too_short", "123", true},
		{"invalid_chars", "123abc7890", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CoerceToPhone(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToPhone() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestDate tests date parsing and conversion
func TestCoerceDateType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		expectedY int
		expectedM int
		expectedD int
	}{
		{"valid_date", "2024-12-25", false, 2024, 12, 25},
		{"invalid_format", "2024/12/25", true, 0, 0, 0},
		{"invalid_month", "2024-13-01", true, 0, 0, 0},
		{"invalid_day", "2024-12-32", true, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := CoerceToDate(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToDate() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError {
				if date.Year != tt.expectedY || date.Month != tt.expectedM || date.Day != tt.expectedD {
					t.Errorf("CoerceToDate() = %v, want %d-%d-%d", date, tt.expectedY, tt.expectedM, tt.expectedD)
				}
			}
		})
	}
}

// TestDateTime tests datetime parsing and conversion
func TestCoerceDateTimeType(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid_datetime", "2024-12-25T10:30:00Z", false},
		{"with_offset", "2024-12-25T10:30:00+07:00", false},
		{"invalid_format", "2024-12-25 10:30:00", true},
		{"empty_string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CoerceToDateTime(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("CoerceToDateTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestSlug tests slug conversion
func TestCoerceSlugType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple_text", "Hello World", "hello-world"},
		{"with_special", "Hello @ World!", "hello-world"},
		{"multiple_spaces", "Hello    World", "hello-world"},
		{"already_slug", "hello-world", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := CoerceToSlug(tt.input)
			if string(result) != tt.expected {
				t.Errorf("CoerceToSlug() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestEnum tests enum validation
func TestEnum(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		allowed   []string
		wantError bool
	}{
		{"valid_enum", "red", []string{"red", "green", "blue"}, false},
		{"invalid_enum", "yellow", []string{"red", "green", "blue"}, true},
		{"empty_allowed", "", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEnum(tt.value, tt.allowed)
			if (err != nil) != tt.wantError {
				t.Errorf("NewEnum() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestDateToTime tests Date to time.Time conversion
func TestDateToTime(t *testing.T) {
	date := Date{Year: 2024, Month: 12, Day: 25}
	tm := date.ToTime()

	if tm.Year() != 2024 || tm.Month() != time.December || tm.Day() != 25 {
		t.Errorf("Date.ToTime() = %v, want 2024-12-25", tm)
	}
}

// TestDateString tests Date string representation
func TestDateString(t *testing.T) {
	date := Date{Year: 2024, Month: 3, Day: 5}
	expected := "2024-03-05"
	if date.String() != expected {
		t.Errorf("Date.String() = %q, want %q", date.String(), expected)
	}
}
