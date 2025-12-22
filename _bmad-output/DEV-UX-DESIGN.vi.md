# Thiết kế Trải nghiệm Nhà phát triển (DEV UX) cho go-agentic

**Ngày tạo:** 22-12-2025
**Góc nhìn:** Người dùng thư viện (không phải người đóng góp)
**Mục tiêu:** Thiết kế cách các nhà phát triển NÊN trải nghiệm go-agentic

---

## Tóm tắt điều hành

Trạng thái hiện tại: go-agentic **đã sẵn sàng cho sản xuất nhưng cảm giác chưa hoàn thiện** (mức độ trưởng thành 7/10). Chỉ có 1 ví dụ hoạt động; 4 ví dụ khác chưa hoàn thiện hoặc còn thiếu. Người dùng có thể bắt đầu nhanh chóng nhưng sau đó gặp khó khăn khi tùy chỉnh.

**Vấn đề cốt lõi:** Thiết kế thư viện ưu tiên **kiến trúc nội bộ** hơn là **trải nghiệm người dùng**.

**Giải pháp:** Thiết kế DEV UX trước, sau đó điều chỉnh kiến trúc xoay quanh nó.

---

## Phần 1: Người dùng của chúng ta là ai?

### Các chân dung người dùng chính

**Chân dung 1: Nhà phát triển Backend Go** (50% người dùng)
- Thành thạo Go, goroutines, interfaces
- Xây dựng các hệ thống sản xuất
- Cần điều phối đa tác nhân (multi-agent) cho logic nghiệp vụ
- Ràng buộc thời gian: Cao (muốn làm việc hiệu quả trong vài giờ, không phải vài ngày)
- Điểm đau (Pain point): Mã lặp lại (boilerplate), schema rườm rà
- Chỉ số thành công: "Tôi đã sao chép một ví dụ và nó hoạt động ngay lập tức"

**Chân dung 2: Nhà phát triển Go mới làm quen với AI** (30% người dùng)
- Chuyên gia Go, nhưng chưa quen với LLM/agent
- Muốn học các mẫu (pattern) agent
- Xây dựng PoC hoặc dự án nghiên cứu
- Ràng buộc thời gian: Trung bình
- Điểm đau: Các khái niệm không rõ ràng, thiếu các mẫu thiết kế
- Chỉ số thành công: "Tôi hiểu cách các agent điều hướng (route) cho nhau"

**Chân dung 3: Kỹ sư DevOps/Cơ sở hạ tầng** (15% người dùng)
- Quen thuộc với YAML, triển khai, khả năng quan sát (observability)
- Triển khai các hệ thống go-agentic
- Cần cách tiếp cận cấu hình dưới dạng mã (configuration-as-code)
- Ràng buộc thời gian: Trung bình
- Điểm đau: Lỗi xác thực, thiết lập môi trường
- Chỉ số thành công: "Tôi có thể triển khai các hệ thống đa nhóm (multi-crew) từ YAML"

**Chân dung 4: Kỹ sư Dữ liệu/ML trong nhóm Go** (5% người dùng)
- Nền tảng Python, đang học Go
- Xây dựng các agent xử lý dữ liệu
- Cần các ví dụ API rõ ràng
- Ràng buộc thời gian: Thấp (có thể chấp nhận đường cong học tập)
- Điểm đau: Cú pháp Go, rào cản ngôn ngữ
- Chỉ số thành công: "Các mẫu thiết kế đủ rõ ràng để áp dụng vào lĩnh vực của tôi"

---

## Phần 2: Các điểm gây khó khăn hiện tại

### Khó khăn 1: "Chỉ có một ví dụ hoạt động"
```
Người dùng tìm kiếm: "Ví dụ trợ lý nghiên cứu"
Tìm thấy: thư mục examples/research-assistant/
Hy vọng: Mã hoạt động để chạy
Thực tế: main.go trống, cấu hình chưa hoàn thiện
Cảm thấy: Bị bỏ rơi, bối rối về mức độ hoàn thiện của thư viện
```

**Tác động:** Người dùng nghi ngờ sự trưởng thành của thư viện trước khi bắt đầu.

### Khó khăn 2: "Tôi thêm công cụ của mình ở đâu?"
```
Người dùng muốn: Công cụ tùy chỉnh cho logic nghiệp vụ của tôi
Đường dẫn: Đọc mã nguồn tools.go → hiểu mẫu thiết kế → sao chép cấu trúc
Thực tế: Mẫu thiết kế rõ ràng, nhưng yêu cầu đọc mã nguồn
Tốt hơn: Bản mẫu (template) trong tài liệu hoặc trình tạo mã
```

**Tác động:** Mất 15-20 phút để hiểu so với 2 phút nếu có tài liệu tốt.

### Khó khăn 3: "Làm thế nào để thay đổi điều hướng?"
```
Người dùng muốn: Agent khác nhau dựa trên loại khách hàng
Đường dẫn: Chỉnh sửa crew.yaml → nhưng điều hướng sử dụng hệ thống tín hiệu phức tạp
Thực tế: Phải đọc orchestrator.yaml (hơn 160 dòng) để hiểu các tín hiệu
Tốt hơn: Tài liệu rõ ràng về các mẫu điều hướng dựa trên tín hiệu
```

**Tác động:** Người dùng không sửa đổi gì cả, chỉ sử dụng nguyên trạng.

### Khó khăn 4: "Schema tham số quá rườm rà"
```
Người dùng định nghĩa tham số công cụ:
  "properties": map[string]interface{}{
      "path": map[string]interface{}{
          "type": "string",
          "description": "...",
      },
  }
Thực tế: Nhiều mã lặp lại, dễ sai sót
Tốt hơn: Hàm hỗ trợ như: tool.StringParam("path", "description")
```

**Tác động:** Hơn 30 dòng mã cho một công cụ đơn giản có 2 tham số.

### Khó khăn 5: "Có chuyện gì đã xảy ra?"
```
Công cụ thất bại, người dùng thấy: "lỗi thực thi công cụ"
Câu hỏi: Công cụ nào? Tại sao? Làm sao để sửa?
Thực tế: Phải đọc nhật ký (logs), hiểu phân loại lỗi
Tốt hơn: Thông báo lỗi rõ ràng kèm theo các bước khắc phục
```

**Tác động:** Thời gian gỡ lỗi tăng gấp 5 lần.

### Khó khăn 6: "Vị trí tệp cấu hình"
```
Người dùng sửa mã, truyền configDir: "config"
Thực tế: Framework mong đợi config/crew.yaml và config/agents/*.yaml
Câu hỏi: Làm sao tôi biết được? Không có trong tài liệu.
Tốt hơn: Thông báo lỗi hướng dẫn cấu trúc tệp
```

**Tác động:** Bối rối trong lần thiết lập đầu tiên.

---

## Phần 3: Hành trình người dùng mong muốn

### Mục tiêu: Từ con số không đến khi hoạt động trong 5 phút

```
Phút 0: Người dùng mở README.md
  ↓ Thấy: "Chạy ví dụ Hỗ trợ IT trong 2 phút"
  ↓ Hướng dẫn rõ ràng để thiết lập môi trường

Phút 1: git clone + cd examples/it-support

Phút 2: cp .env.example .env + thiết lập API key

Phút 3: go run ./cmd/main.go

Phút 4: Hệ thống yêu cầu nhập liệu, đưa ra phản hồi

Phút 5: Người dùng nghĩ: "Tuyệt! Bây giờ hãy để tôi sửa đổi nó..."
```

**Trạng thái hiện tại:** ✅ Điều này đã hoạt động cho ví dụ Hỗ trợ IT.

### Mục tiêu: Từ "Hello World" đến "Phiên bản sửa đổi" trong 15 phút

```
Phút 0: Người dùng chạy xong ví dụ Hỗ trợ IT
  ↓ Đọc: "Các bước tiếp theo - Tùy chỉnh ví dụ này"
  ↓ Thấy: 3 sửa đổi phổ biến với hướng dẫn rõ ràng

Phút 5: Người dùng sửa đổi backstory của agent (chỉ chỉnh sửa YAML)
  ↓ Khởi động lại ứng dụng, thấy hành vi khác đi
  ↓ Nghĩ: "Tuyệt vời, cấu hình YAML thật hữu ích"

Phút 10: Người dùng thêm công cụ mới
  ↓ Làm theo: Hướng dẫn "Thêm công cụ" (không phải lục lọi mã nguồn)
  ↓ Sử dụng: Bản mẫu mã với các phần rõ ràng
  ↓ Đăng ký: Công cụ trong agents/executor.yaml
  ↓ Khởi động lại, kiểm tra công cụ

Phút 15: Người dùng thêm agent tùy chỉnh
  ↓ Sao chép: agents/executor.yaml sang agents/specialist.yaml
  ↓ Sửa đổi: id, name, role, backstory
  ↓ Cập nhật: điều hướng trong crew.yaml
  ↓ Kiểm tra: Agent mới xuất hiện
```

**Trạng thái hiện tại:** ❌ Người dùng phải đọc mã nguồn để biết về công cụ và điều hướng.

### Mục tiêu: Từ "Phiên bản sửa đổi" đến "Hệ thống của riêng tôi" trong 2 giờ

```
Phút 30: Người dùng muốn: "Tôi cần điều hướng tùy chỉnh cho lĩnh vực của mình"
  ↓ Đọc: Tài liệu "Điều hướng dựa trên tín hiệu"
  ↓ Hiểu: Mẫu thiết kế [ROUTE_SPECIALIST]
  ↓ Định nghĩa: custom_orchestrator.yaml với vai trò rõ ràng
  ↓ Kiểm tra: Điều hướng hoạt động

Phút 60: Người dùng muốn: "Triển khai đa nhóm"
  ↓ Đọc: Hướng dẫn "Nhiều nhóm (Multiple Crews)"
  ↓ Tạo: crew-team-a.yaml, crew-team-b.yaml
  ↓ Học: Cách phối hợp giữa các nhóm
  ↓ Triển khai: Hệ thống nhận biết nhóm

Phút 120: Người dùng có: Hệ thống đang hoạt động, đã tùy chỉnh và triển khai
```

**Trạng thái hiện tại:** ❌ Chỉ có mẫu Hỗ trợ IT được ghi chép; các mẫu khác còn thiếu.

---

## Phần 4: Các nguyên tắc thiết kế DEV UX

### Nguyên tắc 1: **Tài liệu > Mã nguồn**

Người dùng KHÔNG BAO GIỜ cần phải đọc mã nguồn để hiểu:
- Cách thêm một công cụ
- Cách sửa đổi điều hướng
- Cách thêm một agent
- Cách xử lý lỗi
- Cách triển khai

**Thực hiện:**
- Viết docs/GUIDE_ADDING_TOOLS.md TRƯỚC KHI tạo PR cho tools.go
- Viết docs/GUIDE_SIGNAL_ROUTING.md trước khi đóng các vấn đề xác thực tín hiệu
- Viết docs/GUIDE_DEPLOYMENT.md trước khi đánh dấu là "sẵn sàng cho sản xuất"

### Nguyên tắc 2: **Các mẫu Sao chép-Dán (Copy-Paste Patterns)**

Mọi thao tác phổ biến nên có:
1. Bản mẫu rõ ràng trong docs/
2. Ví dụ trong examples/
3. Chú thích mã nội dòng (inline) hiển thị mẫu thiết kế

**Ví dụ:**
```
# Thêm một công cụ
docs/GUIDE_ADDING_TOOLS.md
  → Bản mẫu mã (sẵn sàng để sao chép-dán)
  → Các phần được giải thích
  → Các lỗi thường gặp

examples/it-support/internal/tools.go
  → Mẫu GetCPUUsage (đơn giản nhất)
  → Mẫu GetMemoryUsage (có tham số)
  → Mẫu GetDiskSpace (có xác thực)
```

### Nguyên tắc 3: **Học tập theo lớp (Layered Learning)**

Người dùng có thể hiểu thư viện ở 3 cấp độ:
1. **Cơ bản:** Sử dụng ví dụ Hỗ trợ IT nguyên trạng
2. **Trung cấp:** Sửa đổi prompt, thêm công cụ
3. **Nâng cao:** Kiến trúc tùy chỉnh, hệ thống đa nhóm

Mỗi lớp nên rõ ràng, có tài liệu và ví dụ đi kèm.

### Nguyên tắc 4: **Tiết lộ dần dần (Progressive Disclosure)**

Đừng hiển thị mọi thứ cùng một lúc. Hướng dẫn người dùng:
1. Đầu tiên: "Chạy ví dụ này" (thông tin tối thiểu)
2. Sau đó: "Tùy chỉnh ví dụ này" (sửa đổi có hướng dẫn)
3. Cuối cùng: "Xây dựng hệ thống của riêng bạn" (toàn bộ sức mạnh)

### Nguyên tắc 5: **Thông báo lỗi là tài liệu**

Khi có lỗi xảy ra, thông báo lỗi nên:
1. Nêu rõ vấn đề
2. Hiển thị những gì được mong đợi
3. Gợi ý cách khắc phục
4. Liên kết đến tài liệu

**Tệ:** `"lỗi cấu hình"`
**Tốt:** 
```
Lỗi xác thực Schema tín hiệu:
  Agent 'orchestrator' sử dụng tín hiệu [ROUTE_EXECUTOR] trong system_prompt
  nhưng [ROUTE_EXECUTOR] KHÔNG nằm trong allowed_signals
  
  Khắc phục: Thêm '[ROUTE_EXECUTOR]' vào orchestrator.yaml:allowed_signals
  
  Xem: docs/YAML-SIGNALS.md
```

### Nguyên tắc 6: **Tối thiểu hóa cấu hình**

Các giá trị mặc định nên hoạt động cho 90% trường hợp. Người dùng chỉ cấu hình khi cần thiết.

**Tệ:**
```yaml
settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: IT-Support-Team
```

**Tốt:**
```yaml
settings:
  max_handoffs: 5          # Mặc định là 5, thay đổi nếu cần
  # Các mặc định khác: max_rounds=10, timeout=300s, language=en
```

---

## Phần 5: Cấu trúc thư viện được thiết kế lại cho người dùng

### Cấu trúc hiện tại (Tập trung vào nội bộ)
```
examples/
├── it-support/          (⭐⭐⭐⭐ hoàn thiện)
├── research-assistant/  (❌ chưa hoàn thiện)
├── vector-search/       (⭐⭐ một phần)
└── ... thêm 2 khung sườn

docs/
├── (tối thiểu - chủ yếu nằm trong README.md)
```

### Cấu trúc đề xuất (Tập trung vào người dùng)

```
examples/
├── 00-hello-world/               ← MỚI: Ví dụ tối thiểu 3 dòng
│   ├── cmd/main.go              # Chỉ gọi Execute()
│   ├── config/
│   │   ├── crew.yaml            # 1 agent đơn giản
│   │   └── agents/simple.yaml
│   └── README.md                # "Hướng dẫn 5 phút"

├── 01-it-support/               ← HIỆN TẠI: Điều hướng đa agent
│   ├── cmd/main.go
│   ├── internal/crew.go
│   ├── internal/tools.go
│   ├── config/crew.yaml
│   ├── config/agents/
│   └── README.md

├── 02-research-assistant/       ← MỚI: Tìm kiếm tài liệu + tổng hợp
│   ├── cmd/main.go
│   ├── internal/tools.go         # Tìm kiếm web, phân tích tài liệu
│   ├── config/crew.yaml
│   ├── config/agents/
│   │   ├── researcher.yaml
│   │   ├── synthesizer.yaml
│   │   └── reviewer.yaml
│   └── README.md

├── 03-customer-service/         ← MỚI: Hỗ trợ đa ngôn ngữ
│   ├── cmd/main.go
│   ├── internal/tools.go
│   ├── config/crew.yaml
│   ├── config/agents/
│   └── README.md

├── 04-data-extraction/          ← MỚI: Xử lý PDF/Tài liệu
│   ├── cmd/main.go
│   ├── internal/tools.go
│   ├── config/crew.yaml
│   └── README.md

├── 05-vector-search/            ← HIỆN TẠI: Tích hợp Qdrant
│   ├── cmd/main.go
│   ├── internal/
│   ├── config/
│   └── README.md

└── templates/                    ← MỚI: Các bản mẫu sao chép-dán
    ├── crew.yaml.template
    ├── agent.yaml.template
    ├── tool.go.template
    └── main.go.template

docs/
├── GUIDE_GETTING_STARTED.md             ← Bắt đầu trong 5 phút
├── GUIDE_HELLO_WORLD.md                 ← Ví dụ tối thiểu
├── GUIDE_MODIFYING_EXAMPLES.md          ← Các tinh chỉnh phổ biến
│
├── GUIDE_ADDING_TOOLS.md                ← Bản mẫu sao chép-dán
├── GUIDE_ADDING_AGENTS.md               ← Từng bước một
├── GUIDE_SIGNAL_ROUTING.md              ← Cách điều hướng hoạt động
├── GUIDE_ERROR_HANDLING.md              ← Những gì có thể sai
│
├── GUIDE_PARAMETER_SCHEMAS.md           ← JSON schema dễ dàng hơn
├── GUIDE_SYSTEM_PROMPTS.md              ← Kỹ thuật viết prompt
├── GUIDE_STREAMING.md                   ← Các sự kiện thời gian thực
│
├── GUIDE_DEPLOYMENT.md                  ← Thiết lập sản xuất
├── GUIDE_MULTI_CREW.md                  ← Nhiều nhóm
├── GUIDE_PERFORMANCE_TUNING.md          ← Tối ưu hóa
│
├── API_REFERENCE.md                     ← Chữ ký hàm
├── ERROR_CODES.md                       ← Ý nghĩa các lỗi
├── FAQ.md                               ← Các câu hỏi thường gặp
│
└── ARCHITECTURE.md                      ← Dành cho người đóng góp

tools/
├── init-crew/                   ← MỚI: CLI để tạo khung sườn crew
├── validate-config/             ← MỚI: Kiểm tra lỗi crew.yaml
├── migrate-v1-to-v2/            ← MỚI: Di chuyển cấu hình
└── ...
```

---

## Phần 6: Các tài liệu chính cần thiết

### Phải có (Đang chặn người dùng)

| Tài liệu | Mục đích | Ví dụ |
|-----|---------|---------|
| **GUIDE_GETTING_STARTED.md** | 5 phút đầu tiên | Clone → Chạy → Sửa đổi |
| **GUIDE_ADDING_TOOLS.md** | Bản mẫu công cụ sao chép-dán | 5 ví dụ công cụ (từ đơn giản đến phức tạp) |
| **GUIDE_SIGNAL_ROUTING.md** | Cách các agent điều hướng cho nhau | 3 mẫu điều hướng |
| **GUIDE_ADDING_AGENTS.md** | Tạo agent từng bước | Sao chép yaml → Sửa đổi → Kiểm tra |
| **ERROR_CODES.md** | Chuyện gì đã sai + cách sửa | 20 lỗi phổ biến |

### Nên có (Gỡ chặn cho người dùng nâng cao)

| Tài liệu | Mục đích | Ví dụ |
|-----|---------|---------|
| **GUIDE_SYSTEM_PROMPTS.md** | Prompt hiệu quả | Các ví dụ tốt/xấu |
| **GUIDE_PARAMETER_SCHEMAS.md** | Hỗ trợ JSON schema | Từ đơn giản đến phức tạp |
| **GUIDE_DEPLOYMENT.md** | Thiết lập sản xuất | Docker, biến môi trường, giám sát |
| **GUIDE_MULTI_CREW.md** | Nhiều nhóm | Điều hướng nhóm, trạng thái chia sẻ |
| **API_REFERENCE.md** | Tất cả các hàm | Có thể tìm kiếm, nhóm theo tác vụ |

### Tốt nếu có (Hoàn thiện)

| Tài liệu | Mục đích |
|-----|---------|
| **FAQ.md** | Các câu hỏi thường gặp |
| **TROUBLESHOOTING.md** | Chiến lược gỡ lỗi |
| **PERFORMANCE_TUNING.md** | Mẹo tối ưu hóa |
| **SECURITY.md** | Xử lý prompt an toàn, quản lý bí mật |
| **CONTRIBUTING.md** | Cách thêm ví dụ/tài liệu |

---

## Phần 7: Lộ trình các ví dụ hướng tới người dùng

### Trạng thái hiện tại
- ✅ Hỗ trợ IT (hoàn thiện)
- ❌ Trợ lý nghiên cứu (chưa hoàn thiện)
- ⚠️ Tìm kiếm Vector (một phần)
- ❌ Dịch vụ khách hàng (chưa bắt đầu)
- ❌ Phân tích dữ liệu (chưa bắt đầu)

### Các ví dụ hoàn thiện đề xuất

**Nhóm 1: Phải hoàn thành (Quý 1 2025)**

1. **hello-world/** (⭐ Đơn giản nhất)
   - Một agent, không công cụ
   - Hiển thị mẫu Execute() cơ bản
   - Tổng cộng 50 dòng mã
   - Thời gian hoạt động: 2 phút

2. **it-support/** (⭐⭐⭐ Hiện tại)
   - Điều hướng đa agent
   - 13 công cụ
   - Hiển thị mẫu tín hiệu
   - Thời gian hoạt động: 3 phút
   - **Trạng thái:** ✅ Hoàn thiện, đang được bảo trì tích cực

3. **research-assistant/** (⭐⭐⭐⭐ Phức tạp)
   - Xử lý đa giai đoạn
   - Phân tích tài liệu
   - Tìm kiếm web
   - Tổng hợp kết quả
   - Thời gian hoạt động: 5 phút
   - **Ưu tiên:** CAO - Trường hợp sử dụng rất phổ biến

**Nhóm 2: Nên có (Quý 2 2025)**

4. **customer-service/** (⭐⭐⭐ Trung bình)
   - Hỗ trợ đa ngôn ngữ
   - Phân tích cảm xúc
   - Điều hướng ý định
   - Hiển thị bản địa hóa

5. **data-extraction/** (⭐⭐⭐⭐ Phức tạp)
   - Xử lý PDF
   - Trích xuất biểu mẫu
   - Xác thực dữ liệu
   - Hiển thị xử lý tài liệu

**Nhóm 3: Tốt nếu có (Quý 3 2025)**

6. **vector-search/** (⭐⭐⭐⭐ Phức tạp)
   - Tích hợp Qdrant
   - Tìm kiếm ngữ nghĩa
   - Tạo văn bản tăng cường truy xuất (RAG)

---

## Phần 8: Chỉ số thành công của người dùng

Làm thế nào chúng ta biết DEV UX đang hoạt động tốt?

### Chỉ số 1: Thời gian đến thành công đầu tiên

**Mục tiêu:** Người dùng có thể chạy ví dụ trong ≤ 3 phút

**Đo lường:**
- Clone repo → go run → ứng dụng hoạt động
- Theo dõi: Thời gian từ khi clone đến khi có kết quả thành công đầu tiên
- Mục tiêu: < 3 phút cho hello-world, < 5 phút cho Hỗ trợ IT

**Hiện tại:** ✅ 2-3 phút (đạt mục tiêu)

### Chỉ số 2: Tỷ lệ thành công khi Sao chép-Dán

**Mục tiêu:** 90% các đoạn mã mẫu trong tài liệu hoạt động mà không cần sửa đổi

**Đo lường:**
- Kiểm tra định kỳ các đoạn mã trong tài liệu
- Phản hồi từ người dùng về các ví dụ bị hỏng

### Chỉ số 3: Giảm thiểu việc đọc mã nguồn

**Mục tiêu:** Người dùng có thể xây dựng hệ thống hoàn chỉnh mà không cần mở thư mục `core/`

**Đo lường:**
- Khảo sát người dùng
- Phân tích các câu hỏi trên GitHub Issues (liệu chúng có liên quan đến các khái niệm không được ghi chép?)
