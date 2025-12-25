package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// CoerceToString converts various types to string safely.
func CoerceToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val)), nil
		}
		return fmt.Sprintf("%v", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int32:
		return fmt.Sprintf("%d", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	case nil:
		return "", fmt.Errorf("cannot coerce nil to string")
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// CoerceToInt converts various types to int safely.
func CoerceToInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	case int32:
		return int(val), nil
	case int:
		return val, nil
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as integer: %w", val, err)
		}
		return i, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to int")
	default:
		return 0, fmt.Errorf("cannot coerce %T to int", v)
	}
}

// CoerceToBool converts various types to bool safely.
func CoerceToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "yes", "1", "on", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disabled":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse %q as boolean", val)
		}
	case float64:
		return val != 0, nil
	case int, int32, int64:
		return val != 0, nil
	case nil:
		return false, fmt.Errorf("cannot coerce nil to bool")
	default:
		return false, fmt.Errorf("cannot coerce %T to bool", v)
	}
}

// CoerceToFloat converts various types to float64 safely.
func CoerceToFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as float: %w", val, err)
		}
		return f, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to float")
	default:
		return 0, fmt.Errorf("cannot coerce %T to float", v)
	}
}

// MustGetString extracts and coerces a required string parameter.
func MustGetString(args map[string]interface{}, key string) (string, error) {
	val, exists := args[key]
	if !exists {
		return "", fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToString(val)
	if err != nil {
		return "", fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetInt extracts and coerces a required int parameter.
func MustGetInt(args map[string]interface{}, key string) (int, error) {
	val, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return 0, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetBool extracts and coerces a required bool parameter.
func MustGetBool(args map[string]interface{}, key string) (bool, error) {
	val, exists := args[key]
	if !exists {
		return false, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return false, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetFloat extracts and coerces a required float parameter.
func MustGetFloat(args map[string]interface{}, key string) (float64, error) {
	val, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToFloat(val)
	if err != nil {
		return 0, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// OptionalGetString extracts an optional string parameter with default value.
func OptionalGetString(args map[string]interface{}, key, defaultVal string) string {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToString(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetInt extracts an optional int parameter with default value.
func OptionalGetInt(args map[string]interface{}, key string, defaultVal int) int {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetBool extracts an optional bool parameter with default value.
func OptionalGetBool(args map[string]interface{}, key string, defaultVal bool) bool {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetFloat extracts an optional float parameter with default value.
func OptionalGetFloat(args map[string]interface{}, key string, defaultVal float64) float64 {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToFloat(val)
	if err != nil {
		return defaultVal
	}
	return result
}
