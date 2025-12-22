# 🚨 Hardcoded Values Audit - Tiêu chí Thư viện Lõi (Core Library Standard)

**Ngày:** 2025-12-22
**Phạm vi:** `core/` directory - Thư viện lõi go-agentic
**Tiêu chí:** Thư viện lõi **KHÔNG được phép HARDCODE** - Phải có validation và báo lỗi

---

## 📋 Tổng Kết Điều Chỉnh

**Kết quả trước:** 8 hardcoded "đúng", 5 hardcoded "tùy chọn"
**Kết quả sau:** ❌ **KHÔNG CHẤP NHẬN** - Thư viện lõi phải **validate tham số** chứ không được hardcode

**Phân loại lại:**
- 🔴 **PHẢI KHẮC PHỤC (Critical):** 7 giá trị
- 🟡 **CÓ THỂ CẢNH BÁO (Warning):** 4 giá trị
- 🟢 **CHẤP NHẬN ĐƯỢC (Internal):** 2 giá trị

---

## 🔴 PHẢI KHẮC PHỤC - Hardcoded Khi Nên Có Tham Số

### ❌ 1. Default Provider Selection

**Vị trí:** `core/agent.go:23, 67` + `core/providers/provider.go:86`

**Hiện tại:**
```go
providerType := agent.Provider
if providerType == "" {
    providerType = "ollama" // ❌ HARDCODE!
}
```

**Vấn đề:**
- ❌ Thư viện lõi đặt mặc định `"ollama"` cho tất cả người dùng
- ❌ Không thể đổi thành `"openai"` nếu Agent không chỉ định
- ❌ Nên báo lỗi nếu thiếu tham số, không phải mặc định

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Validate, không hardcode
if agent.Provider == "" {
    return nil, fmt.Errorf("Agent.Provider không được để trống - phải chỉ định 'openai' hoặc 'ollama'")
}

providerType := agent.Provider
provider, err := providerFactory.GetProvider(providerType, agent.ProviderURL, apiKey)
if err != nil {
    return nil, fmt.Errorf("provider '%s' không hợp lệ: %w", providerType, err)
}
```

**Khuyến nghị:**
- 🟢 Cho phép default trong **ứng dụng (app)**, nhưng
- 🔴 **Thư viện lõi phải báo lỗi** nếu không được cung cấp

---

### ❌ 2. Default Ollama URL

**Vị trí:** `core/providers/ollama/provider.go:57` + `core/providers/provider.go:120`

**Hiện tại:**
```go
if baseURL == "" {
    baseURL = "http://localhost:11434" // ❌ HARDCODE!
}
```

**Vấn đề:**
- ❌ Nếu Ollama chạy trên máy chủ khác, nó sẽ lỗi im lặng
- ❌ Không biết tại sao kết nối thất bại (nó đã tự đặt URL rồi)
- ❌ Người dùng không biết URL nào đang được sử dụng

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Require hoặc lấy từ environment
if baseURL == "" {
    baseURL = os.Getenv("OLLAMA_URL")
}
if baseURL == "" {
    return nil, fmt.Errorf("Ollama URL không được cấp - thiết lập provider_url trong YAML hoặc biến môi trường OLLAMA_URL")
}

// Validate URL format
if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
    return nil, fmt.Errorf("Ollama URL phải bắt đầu bằng http:// hoặc https://: %s", baseURL)
}
```

**Khuyến nghị:**
- 🟢 Hỗ trợ biến môi trường `OLLAMA_URL` làm fallback
- 🔴 Nếu không tìm thấy, **báo lỗi rõ ràng** chứ không tự đặt default

---

### ❌ 3. OpenAI Client TTL Cache

**Vị trí:** `core/providers/openai/provider.go:27`

**Hiện tại:**
```go
const clientTTL = 1 * time.Hour // ❌ HARDCODE!
```

**Vấn đề:**
- ❌ Không thể điều chỉnh TTL cho các use case khác nhau
- ❌ Không có cách nào để cấu hình nếu 1 giờ không phù hợp
- ❌ Thư viện lõi quyết định cho ứng dụng

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Làm configurable
type OpenAIProvider struct {
    apiKey        string
    client        openai.Client
    clientTTL     time.Duration  // ✅ Configurable
}

// Cho phép cấu hình hoặc sử dụng default
if clientTTL == 0 {
    clientTTL = 1 * time.Hour  // Default, không phải hardcode
}
```

**Khuyến nghị:**
- 🔴 Thêm `ClientTTL` vào OpenAI provider config
- 🟢 Cung cấp default hợp lý nếu không được chỉ định

---

### ❌ 4. Parallel Agent Timeout

**Vị trí:** `core/crew.go:1183`

**Hiện tại:**
```go
const ParallelAgentTimeout = 60 * time.Second // ❌ HARDCODE!
```

**Vấn đề:**
- ❌ Không thể thay đổi timeout cho từng crew
- ❌ Một số task cần >60s, số khác cần <60s
- ❌ Nên là tham số của Crew, không phải constant

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Làm field của Crew
type Crew struct {
    Agents                []Agent
    MaxRounds             int
    ParallelAgentTimeout  time.Duration  // ✅ Configurable
    // ...
}

// Validate và cung cấp default
if c.ParallelAgentTimeout == 0 {
    c.ParallelAgentTimeout = 60 * time.Second  // Default
}
if c.ParallelAgentTimeout < 5*time.Second {
    return fmt.Errorf("ParallelAgentTimeout phải ≥ 5 giây, nhận: %v", c.ParallelAgentTimeout)
}
```

**Khuyến nghị:**
- 🔴 Chuyển từ `const` thành field của `Crew`
- 🟡 Thêm validation min/max

---

### ❌ 5. Max Tool Output Characters

**Vị trí:** `core/crew.go:1425`

**Hiện tại:**
```go
const maxOutputChars = 2000 // ❌ HARDCODE!
```

**Vấn đề:**
- ❌ Một số tools có output rất lớn (>2000 chars)
- ❌ Không thể điều chỉnh theo nhu cầu
- ❌ Nên là tham số của Crew

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Field của Crew với validation
type Crew struct {
    // ...
    MaxToolOutputChars int  // ✅ Configurable
}

// Validate
if c.MaxToolOutputChars == 0 {
    c.MaxToolOutputChars = 2000  // Default
}
if c.MaxToolOutputChars < 100 {
    return fmt.Errorf("MaxToolOutputChars phải ≥ 100, nhận: %d", c.MaxToolOutputChars)
}

// Sử dụng
func (ce *CrewExecutor) formatToolResults(results []ToolResult) string {
    maxChars := ce.crew.MaxToolOutputChars  // ✅ Lấy từ config
    // ...
}
```

**Khuyến nghị:**
- 🔴 Thêm vào `Crew` config
- 🟡 Thêm validation khoảng hợp lệ

---

### ❌ 6. Client Cleanup Interval

**Vị trí:** `core/providers/openai/provider.go:74`

**Hiện tại:**
```go
ticker := time.NewTicker(5 * time.Minute) // ❌ HARDCODE!
```

**Vấn đề:**
- ❌ Không thể điều chỉnh bao thường xuyên cleanup
- ❌ Có thể không phù hợp với các use case khác nhau

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Configurable
type OpenAIProvider struct {
    // ...
    cleanupInterval time.Duration
}

// Init với validation
func NewOpenAIProvider(apiKey string, cleanupInterval time.Duration) (LLMProvider, error) {
    if cleanupInterval == 0 {
        cleanupInterval = 5 * time.Minute  // Default
    }
    if cleanupInterval < 1*time.Minute {
        return nil, fmt.Errorf("cleanup interval phải ≥ 1 phút")
    }
    // ...
}
```

**Khuyến nghị:**
- 🟡 Cho phép cấu hình cleanup interval
- 🟢 Default 5 phút là hợp lý nếu người dùng không chỉ định

---

### ❌ 7. HTTP Client Timeout

**Vị trí:** `core/providers/ollama/provider.go:73-75`

**Hiện tại:**
```go
client: &http.Client{
    Timeout: 0,  // ❌ HARDCODE! (Vô hạn cho streaming)
},
```

**Vấn đề:**
- ❌ Timeout = 0 có thể cho phép request treo vô hạn
- ❌ Nếu mạng chậm, có thể chờ rất lâu mà không timeout
- ❌ Nên có một timeout tối đa, thậm chí cho streaming

**Cách khắc phục:**
```go
// ✅ ĐÚNG: Có timeout tối đa
const maxHTTPTimeout = 30 * time.Minute  // Timeout tối đa

client: &http.Client{
    Timeout: maxHTTPTimeout,  // Không vô hạn
}

// Hoặc tốt hơn, sử dụng context timeout cho từng request
// Mà không có HTTP client timeout
```

**Khuyến nghị:**
- 🔴 Đặt timeout tối đa (~30 phút) thay vì vô hạn
- 🟡 Để context timeout xử lý timeout chi tiết

---

## 🟡 CÓ THỂ CẢNH BÁO - Hardcoded Nhưng Có Thể Chấp Nhận

### ⚠️ 8. System Message Role

**Vị trí:** `core/providers/ollama/provider.go:276`

**Hiện tại:**
```go
Role: "system"  // Hardcoded role name
```

**Đánh giá:**
- 🟢 **Có thể chấp nhận** - Định nghĩa LLM API (system, user, assistant)
- 🟡 Nhưng nên có const hoặc enum, không magic string

**Cách khắc phục (Optional):**
```go
// ✅ TỐT HƠN: Sử dụng const thay vì magic string
const (
    RoleSystem    = "system"
    RoleUser      = "user"
    RoleAssistant = "assistant"
)

// Sử dụng
Role: RoleSystem
```

---

### ⚠️ 9. Tool Name Convention

**Vị trí:** `core/providers/ollama/provider.go:331`

**Hiện tại:**
```go
if toolName[0] >= 'A' && toolName[0] <= 'Z'  // Uppercase check
```

**Đánh giá:**
- 🟢 **Có thể chấp nhận** - Kiểm tra quy ước Go (PascalCase)
- 🟡 Nên tách thành hàm riêng với documentation

**Cách khắc phục (Optional):**
```go
// ✅ TỐT HƠN: Hàm riêng với documentation
// isValidToolName checks if a tool name follows Go naming conventions (PascalCase)
func isValidToolName(name string) bool {
    if len(name) == 0 {
        return false
    }
    return name[0] >= 'A' && name[0] <= 'Z'
}
```

---

### ⚠️ 10. Request ID Context Key

**Vị trí:** `core/request_tracking.go:15`

**Hiện tại:**
```go
const RequestIDKey = "request-id"
```

**Đánh giá:**
- 🟢 **Có thể chấp nhận** - Context key là nội bộ
- 🟡 Nhưng nên cho phép cấu hình nếu cần

**Cách khắc phục (Optional):**
```go
// Cho phép override nếu cần
type ContextConfig struct {
    RequestIDKey string  // Default: "request-id"
}

var contextConfig = &ContextConfig{
    RequestIDKey: "request-id",
}

// Hàm để set custom key
func SetRequestIDKey(key string) {
    if key != "" {
        contextConfig.RequestIDKey = key
    }
}
```

---

## 🟢 CHẤP NHẬN ĐƯỢC - Internal Constants

### ✅ 11. User Role Default

**Vị trí:** `core/providers/ollama/provider.go:286`

**Đánh giá:**
- 🟢 **OK** - Fallback cho unknown roles
- Có thể chấp nhận nhưng nên có const

---

### ✅ 12. Test Data

**Vị trí:** `core/http_test.go:460`

**Đánh giá:**
- 🟢 **OK** - Test fixtures không cần cấu hình

---

## 📋 Bảng Đánh Giá Cải Tạo

| # | Giá trị | Vị trí | Hiện Tại | Đánh Giá | Khắc Phục |
|---|--------|--------|---------|----------|-----------|
| 1 | Default Provider | agent.go:23 | `"ollama"` | 🔴 HARDCODE | Báo lỗi nếu empty |
| 2 | Ollama URL | ollama/provider.go:57 | `"localhost:11434"` | 🔴 HARDCODE | Require hoặc env var |
| 3 | OpenAI TTL | openai/provider.go:27 | `1h` | 🔴 HARDCODE | Field config |
| 4 | Parallel Timeout | crew.go:1183 | `60s` | 🔴 HARDCODE | Field config |
| 5 | Max Output | crew.go:1425 | `2000` | 🔴 HARDCODE | Field config |
| 6 | Cleanup Interval | openai/provider.go:74 | `5m` | 🔴 HARDCODE | Field config |
| 7 | HTTP Timeout | ollama/provider.go:73 | `0` (∞) | 🔴 DANGEROUS | Set max timeout |
| 8 | System Role | ollama/provider.go:276 | `"system"` | 🟡 MAGIC STRING | Use const |
| 9 | Tool Name Check | ollama/provider.go:331 | Uppercase | 🟡 MAGIC RULE | Extract function |
| 10 | Request ID Key | request_tracking.go:15 | `"request-id"` | 🟡 OK | Keep as is |
| 11 | User Role | ollama/provider.go:286 | `"user"` | 🟢 OK | Use const |
| 12 | Test Data | http_test.go:460 | UTF-8 invalid | 🟢 OK | Keep as is |

---

## 🎯 Kế Hoạch Khắc Phục

### Phase 1: Critical Fixes (🔴 - Bắt buộc)
- [ ] Thêm validation cho `agent.Provider` (báo lỗi nếu empty)
- [ ] Thêm validation cho `provider_url` (require hoặc env var)
- [ ] Thêm `ClientTTL` field vào `OpenAIProvider`
- [ ] Thêm `ParallelAgentTimeout` field vào `Crew`
- [ ] Thêm `MaxToolOutputChars` field vào `Crew`
- [ ] Thêm `CleanupInterval` field vào `OpenAIProvider`
- [ ] Đặt max HTTP timeout (không vô hạn)

### Phase 2: Code Quality (🟡 - Nên làm)
- [ ] Dùng const cho Role names
- [ ] Extract `isValidToolName()` hàm
- [ ] Thêm const cho context keys hoặc cho phép override

### Phase 3: Documentation
- [ ] Cập nhật docs: "Thư viện lõi yêu cầu validation"
- [ ] Thêm mô tả lỗi rõ ràng
- [ ] Ví dụ cấu hình cho mỗi field

---

## 🏁 Kết Luận

**Phát hiện:** Thư viện lõi go-agentic hiện đang hardcode 7 giá trị **không nên hardcode**

**Tiêu chí Core Library:**
- ✅ Validation tham số (não báo lỗi)
- ✅ Không mặc định (nên explicit)
- ✅ Cho phép cấu hình (fields hoặc env vars)
- ✅ Rõ ràng error messages

**Khuyến nghị:**
🔴 **Phải sửa 7 giá trị** để làm cho thư viện lõi hoạt động đúng như một thư viện chuyên nghiệp

---

**Ngày cập nhật:** 2025-12-22
**Tiêu chí:** Core Library Standards - Validation, không hardcode
