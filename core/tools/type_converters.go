package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// TYPE-SAFE CONVERTERS
// ============================================================================

// UUID type alias for UUID string validation
type UUID string

// IsValidUUID kiểm tra UUID format v4
func IsValidUUID(v string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(v))
}

// CoerceToUUID convert interface{} to UUID string
func CoerceToUUID(v interface{}) (UUID, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return "", err
	}
	if !IsValidUUID(s) {
		return "", fmt.Errorf("invalid UUID format: %s", s)
	}
	return UUID(s), nil
}

// ============================================================================
// EMAIL TYPE
// ============================================================================

// EmailAddress type alias for validated email
type EmailAddress string

// IsValidEmail kiểm tra email format
func IsValidEmail(v string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(v)
}

// CoerceToEmail convert interface{} to EmailAddress
func CoerceToEmail(v interface{}) (EmailAddress, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return "", err
	}
	if !IsValidEmail(s) {
		return "", fmt.Errorf("invalid email format: %s", s)
	}
	return EmailAddress(s), nil
}

// ============================================================================
// URL TYPE
// ============================================================================

// URL type alias for validated URL
type URL string

// IsValidURL kiểm tra URL format
func IsValidURL(v string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s]+`)
	return urlRegex.MatchString(v)
}

// CoerceToURL convert interface{} to URL
func CoerceToURL(v interface{}) (URL, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return "", err
	}
	if !IsValidURL(s) {
		return "", fmt.Errorf("invalid URL format: %s (must start with http:// or https://)", s)
	}
	return URL(s), nil
}

// ============================================================================
// PHONE TYPE
// ============================================================================

// PhoneNumber type alias for validated phone number
type PhoneNumber string

// IsValidPhone kiểm tra phone format
func IsValidPhone(v string) bool {
	phoneRegex := regexp.MustCompile(`^[+]?[0-9]{10,15}$`)
	return phoneRegex.MatchString(v)
}

// CoerceToPhone convert interface{} to PhoneNumber
func CoerceToPhone(v interface{}) (PhoneNumber, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return "", err
	}
	s = strings.TrimSpace(s)
	if !IsValidPhone(s) {
		return "", fmt.Errorf("invalid phone format: %s (10-15 digits, optional +)", s)
	}
	return PhoneNumber(s), nil
}

// ============================================================================
// DATE TYPE
// ============================================================================

// Date type alias for validated date
type Date struct {
	Year  int
	Month int
	Day   int
}

// ParseDate parses date string (YYYY-MM-DD format)
func ParseDate(s string) (Date, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return Date{}, fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD)", s)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return Date{}, fmt.Errorf("invalid year: %s", parts[0])
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return Date{}, fmt.Errorf("invalid month: %s", parts[1])
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return Date{}, fmt.Errorf("invalid day: %s", parts[2])
	}

	if month < 1 || month > 12 {
		return Date{}, fmt.Errorf("invalid month: %d (1-12)", month)
	}

	if day < 1 || day > 31 {
		return Date{}, fmt.Errorf("invalid day: %d (1-31)", day)
	}

	return Date{Year: year, Month: month, Day: day}, nil
}

// CoerceToDate convert interface{} to Date
func CoerceToDate(v interface{}) (Date, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return Date{}, err
	}
	return ParseDate(strings.TrimSpace(s))
}

// String returns date as YYYY-MM-DD
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// ToTime converts Date to time.Time (midnight UTC)
func (d Date) ToTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

// ============================================================================
// DATETIME TYPE
// ============================================================================

// DateTime type alias for validated datetime (RFC3339)
type DateTime struct {
	time.Time
}

// ParseDateTime parses datetime string (RFC3339 format)
func ParseDateTime(s string) (DateTime, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return DateTime{}, fmt.Errorf("invalid datetime format: %s (expected RFC3339: 2006-01-02T15:04:05Z07:00)", s)
	}
	return DateTime{t}, nil
}

// CoerceToDateTime convert interface{} to DateTime
func CoerceToDateTime(v interface{}) (DateTime, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return DateTime{}, err
	}
	return ParseDateTime(strings.TrimSpace(s))
}

// ============================================================================
// SLUG TYPE (URL-friendly)
// ============================================================================

// Slug type alias for URL-friendly slug
type Slug string

// IsValidSlug kiểm tra slug format (lowercase, alphanumeric, hyphens)
func IsValidSlug(v string) bool {
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return slugRegex.MatchString(v)
}

// ToSlug converts string to slug format
func ToSlug(s string) Slug {
	// Lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")
	// Remove special characters
	s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "")
	// Remove multiple consecutive hyphens
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	// Trim hyphens
	s = strings.Trim(s, "-")
	return Slug(s)
}

// CoerceToSlug convert interface{} to Slug
func CoerceToSlug(v interface{}) (Slug, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return "", err
	}
	return ToSlug(s), nil
}

// ============================================================================
// JSON TYPE (reusable JSON object)
// ============================================================================

// JSONObject represents a validated JSON object
type JSONObject map[string]interface{}

// CoerceToJSONObject tries to parse string as JSON object
func CoerceToJSONObject(v interface{}) (JSONObject, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return nil, err
	}

	var obj JSONObject
	err = CoerceFromJSON(s, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON object: %w", err)
	}
	return obj, nil
}

// ============================================================================
// ENUM TYPE (constrained values)
// ============================================================================

// Enum validates that value is in allowed set
type Enum struct {
	Value   string
	Allowed []string
}

// NewEnum creates a new enum with validation
func NewEnum(value string, allowed []string) (*Enum, error) {
	for _, a := range allowed {
		if value == a {
			return &Enum{Value: value, Allowed: allowed}, nil
		}
	}
	return nil, fmt.Errorf("invalid enum value: %s (allowed: %v)", value, allowed)
}

// CoerceToEnum converts and validates to enum
func CoerceToEnum(v interface{}, allowed []string) (*Enum, error) {
	s, err := CoerceToString(v)
	if err != nil {
		return nil, err
	}
	return NewEnum(s, allowed)
}

// String returns the enum value
func (e *Enum) String() string {
	return e.Value
}
