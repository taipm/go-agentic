package tools

import (
	"strings"
)

// Middleware định nghĩa một function xử lý tham số trước khi trích xuất
type Middleware func(interface{}) interface{}

// MiddlewareChain quản lý pipeline middleware
type MiddlewareChain struct {
	middlewares []Middleware
}

// NewMiddlewareChain tạo một middleware chain rỗng
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Add thêm một middleware vào chain
func (mc *MiddlewareChain) Add(m Middleware) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, m)
	return mc
}

// Execute chạy toàn bộ middleware chain
func (mc *MiddlewareChain) Execute(value interface{}) interface{} {
	result := value
	for _, m := range mc.middlewares {
		result = m(result)
	}
	return result
}

// ============================================================================
// BUILT-IN MIDDLEWARES
// ============================================================================

// TrimWhitespace xoá whitespace ở đầu và cuối string
func TrimWhitespace(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return v
}

// ToLower convert string sang chữ thường
func ToLower(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		return strings.ToLower(s)
	}
	return v
}

// ToUpper convert string sang chữ hoa
func ToUpper(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		return strings.ToUpper(s)
	}
	return v
}

// RemoveSpecialChars xoá các ký tự đặc biệt, chỉ giữ alphanumeric
func RemoveSpecialChars(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		var result strings.Builder
		for _, r := range s {
			if (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '_' {
				result.WriteRune(r)
			}
		}
		return result.String()
	}
	return v
}

// Sanitize xoá leading/trailing whitespace + lowercase
func Sanitize(v interface{}) interface{} {
	v = TrimWhitespace(v)
	return ToLower(v)
}

// RemoveQuotes xoá quote characters từ string
func RemoveQuotes(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		s = strings.TrimPrefix(s, `"`)
		s = strings.TrimSuffix(s, `"`)
		s = strings.TrimPrefix(s, "'")
		s = strings.TrimSuffix(s, "'")
		return s
	}
	return v
}

// ReplaceSpaces thay thế spaces bằng underscores
func ReplaceSpaces(replacement string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			return strings.ReplaceAll(s, " ", replacement)
		}
		return v
	}
}

// Prefix thêm prefix vào string nếu chưa có
func Prefix(prefix string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			if !strings.HasPrefix(s, prefix) {
				return prefix + s
			}
			return s
		}
		return v
	}
}

// Suffix thêm suffix vào string nếu chưa có
func Suffix(suffix string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			if !strings.HasSuffix(s, suffix) {
				return s + suffix
			}
			return s
		}
		return v
	}
}

// Replace thay thế text trong string
func Replace(old, new string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			return strings.ReplaceAll(s, old, new)
		}
		return v
	}
}

// TrimPrefix xoá prefix từ string
func TrimPrefix(prefix string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			return strings.TrimPrefix(s, prefix)
		}
		return v
	}
}

// TrimSuffix xoá suffix từ string
func TrimSuffix(suffix string) Middleware {
	return func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			return strings.TrimSuffix(s, suffix)
		}
		return v
	}
}

// Custom tạo middleware custom
func CustomMiddleware(fn Middleware) Middleware {
	return fn
}
