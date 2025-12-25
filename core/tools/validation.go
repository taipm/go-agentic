package tools

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationRule định nghĩa một rule validation có thể tái sử dụng
type ValidationRule struct {
	Name        string
	Description string
	Validate    func(interface{}) error
}

// ValidationRules quản lý tập hợp các validation rules
type ValidationRules struct {
	rules map[string]*ValidationRule
}

// NewValidationRules tạo một tập hợp validation rules rỗng
func NewValidationRules() *ValidationRules {
	return &ValidationRules{
		rules: make(map[string]*ValidationRule),
	}
}

// AddRule thêm một validation rule
func (vr *ValidationRules) AddRule(name string, validator func(interface{}) error) *ValidationRules {
	vr.rules[name] = &ValidationRule{
		Name:     name,
		Validate: validator,
	}
	return vr
}

// AddRuleWithDescription thêm rule với description
func (vr *ValidationRules) AddRuleWithDescription(name, description string, validator func(interface{}) error) *ValidationRules {
	vr.rules[name] = &ValidationRule{
		Name:        name,
		Description: description,
		Validate:    validator,
	}
	return vr
}

// GetRule lấy một validation rule
func (vr *ValidationRules) GetRule(name string) *ValidationRule {
	return vr.rules[name]
}

// HasRule kiểm tra rule có tồn tại không
func (vr *ValidationRules) HasRule(name string) bool {
	_, exists := vr.rules[name]
	return exists
}

// ============================================================================
// BUILT-IN VALIDATORS
// ============================================================================

// NotEmpty kiểm tra string không rỗng
func NotEmpty(v interface{}) error {
	if s, ok := v.(string); ok {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("cannot be empty")
		}
		return nil
	}
	return fmt.Errorf("must be a string")
}

// Email kiểm tra format email
func Email(v interface{}) error {
	if s, ok := v.(string); ok {
		// Simple email regex
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(s) {
			return fmt.Errorf("invalid email format")
		}
		return nil
	}
	return fmt.Errorf("must be a string")
}

// URLValidator kiểm tra format URL (validator function)
// Lưu ý: URL type định nghĩa trong type_converters.go, đây là validator
func URLValidator(v interface{}) error {
	if s, ok := v.(string); ok {
		urlRegex := regexp.MustCompile(`^https?://`)
		if !urlRegex.MatchString(s) {
			return fmt.Errorf("must be a valid URL (http:// or https://)")
		}
		return nil
	}
	return fmt.Errorf("must be a string")
}

// Phone kiểm tra format phone (10-15 digits hoặc +)
func Phone(v interface{}) error {
	if s, ok := v.(string); ok {
		phoneRegex := regexp.MustCompile(`^[+]?[0-9]{10,15}$`)
		if !phoneRegex.MatchString(s) {
			return fmt.Errorf("invalid phone format (10-15 digits)")
		}
		return nil
	}
	return fmt.Errorf("must be a string")
}

// MinLength kiểm tra độ dài tối thiểu
func MinLength(min int) func(interface{}) error {
	return func(v interface{}) error {
		if s, ok := v.(string); ok {
			if len(s) < min {
				return fmt.Errorf("must be at least %d characters", min)
			}
			return nil
		}
		return fmt.Errorf("must be a string")
	}
}

// MaxLength kiểm tra độ dài tối đa
func MaxLength(max int) func(interface{}) error {
	return func(v interface{}) error {
		if s, ok := v.(string); ok {
			if len(s) > max {
				return fmt.Errorf("must be at most %d characters", max)
			}
			return nil
		}
		return fmt.Errorf("must be a string")
	}
}

// Range kiểm tra số nằm trong khoảng
func Range(min, max int) func(interface{}) error {
	return func(v interface{}) error {
		i, err := CoerceToInt(v)
		if err != nil {
			return fmt.Errorf("must be an integer")
		}
		if i < min || i > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}
}

// MinValue kiểm tra giá trị tối thiểu
func MinValue(min int) func(interface{}) error {
	return func(v interface{}) error {
		i, err := CoerceToInt(v)
		if err != nil {
			return fmt.Errorf("must be an integer")
		}
		if i < min {
			return fmt.Errorf("must be at least %d", min)
		}
		return nil
	}
}

// MaxValue kiểm tra giá trị tối đa
func MaxValue(max int) func(interface{}) error {
	return func(v interface{}) error {
		i, err := CoerceToInt(v)
		if err != nil {
			return fmt.Errorf("must be an integer")
		}
		if i > max {
			return fmt.Errorf("must be at most %d", max)
		}
		return nil
	}
}

// OneOf kiểm tra giá trị nằm trong danh sách cho phép
func OneOf(allowed ...string) func(interface{}) error {
	return func(v interface{}) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		for _, a := range allowed {
			if s == a {
				return nil
			}
		}
		return fmt.Errorf("must be one of: %v", allowed)
	}
}

// Regex kiểm tra format theo regex pattern
func Regex(pattern, description string) func(interface{}) error {
	return func(v interface{}) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("must be a string")
		}
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
		if !regex.MatchString(s) {
			return fmt.Errorf("invalid format: %s", description)
		}
		return nil
	}
}

// Custom tạo validator custom từ function
func Custom(validator func(interface{}) error) func(interface{}) error {
	return validator
}

// Combine gộp nhiều validators (tất cả phải pass)
func Combine(validators ...func(interface{}) error) func(interface{}) error {
	return func(v interface{}) error {
		for _, validate := range validators {
			if err := validate(v); err != nil {
				return err
			}
		}
		return nil
	}
}
