# Configuration Documentation Index

Bộ tài liệu cấu hình hoàn chỉnh cho go-agentic framework, bao gồm các đặc tả chi tiết, ví dụ thực tế, và hướng dẫn setup.

## 📚 Tài Liệu Chính

### 1. [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md)
**Tài liệu đặc tả kỹ thuật chi tiết (Comprehensive Technical Specification)**

Nội dung:
- Cấu trúc tổng quát của crew.yaml và agent.yaml
- Giải thích chi tiết từng trường bắt buộc
- Mô tả các trường tùy chọn
- JSON Schema validation rules
- Ví dụ đầy đủ cho mỗi loại configuration

**Thích hợp cho**: Những ai cần hiểu rõ tất cả các chi tiết kỹ thuật

**Kích thước**: ~150 KB, chi tiết nhất

---

### 2. [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md)
**Hướng dẫn tham khảo nhanh (Quick Reference Guide)**

Nội dung:
- Minimal templates cho crew.yaml và agent.yaml
- Decision trees cho việc chọn provider, temperature, v.v.
- Common patterns (single-agent, linear, branching)
- Validation checklist
- Bảng tham khảo trường (Field Reference Table)
- Quick start examples
- Tips & tricks

**Thích hợp cho**: Những ai cần setup nhanh hoặc tra cứu nhanh

**Kích thước**: ~50 KB, súc tích và dễ dàng

---

### 3. [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md)
**Tham khảo JSON Schema (Schema Reference)**

Nội dung:
- Complete JSON Schema cho crew.yaml
- Complete JSON Schema cho agent.yaml
- Type definitions với examples
- Enumerations (provider, language, version)
- Validation rules chi tiết
- Complete valid examples
- Common mistakes & fixes
- YAML syntax quick tips

**Thích hợp cho**: Developers, tool builders, validation

**Kích thước**: ~80 KB, kỹ thuật

---

### 4. [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md)
**Các ví dụ setup team thực tế (Practical Team Examples)**

Nội dung:
- Team 1: Content Creation Workflow (4 agents)
  - Ideator → Writer → Editor → Publisher

- Team 2: Software Development Workflow (4 agents)
  - Architect → Developer → Tester → QA-Lead

- Team 3: Customer Support Tiếng Việt (4 agents)
  - Triage → FAQ → Support → Escalation

- Team 4: Business Analytics (4 agents)
  - Data-Engineer → Analyst → Insight-Specialist → Reporter

**Thích hợp cho**: Những ai muốn xây dựng team mới có cấu trúc tương tự

**Kích thước**: ~100 KB, đầy đủ code

---

## 🎯 Bộ Tài Liệu Hiện Có

Trong docs/ directory:

```
docs/
├── CONFIG_DOCUMENTATION_INDEX.md         ← Bạn đang đọc
├── CONFIG_SPECIFICATION.md               ← Chi tiết toàn bộ
├── CONFIG_QUICK_REFERENCE.md             ← Tham khảo nhanh
├── CONFIG_SCHEMA_REFERENCE.md            ← Schema & validation
├── TEAM_SETUP_EXAMPLES.md                ← Ví dụ thực tế
├── CORE_LIBRARY_UPDATES.md               ← NEW: Core features
├── AGENT_MODEL_CONFIGURATION.md          ← NEW: Model setup
├── LIBRARY_USAGE.md                      ← Core library guide
├── ARCHITECTURE.md                       ← System architecture
└── ...
```

## 🚀 Hướng Dẫn Sử Dụng

### Cho Người Bắt Đầu (New User)

1. **Start**: Đọc [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md)
   - 5 phút hiểu được minimal template
   - Chọn provider (Ollama vs OpenAI)
   - Setup environment

2. **Follow**: Một trong các [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md)
   - Chọn ví dụ gần với use case của bạn
   - Copy template
   - Customize

3. **Reference**: Dùng [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md)
   - Nếu cần giải thích field cụ thể
   - Check syntax rules
   - Validate configuration

### Cho Người Tìm Hiểu Chi Tiết (Deep Dive)

1. **Read**: [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md)
   - Hiểu mỗi field là gì
   - Quy tắc validation
   - Best practices

2. **Study**: [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md)
   - JSON Schema format
   - Type definitions
   - Common mistakes

3. **Practice**: [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md)
   - Xem các team thực tế
   - Hiểu signal routing
   - Thiết kế team của riêng bạn

### Cho Người Troubleshooting (Problem Solving)

1. **Error?** → [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md) Validation Checklist
2. **Syntax?** → [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md) Common Mistakes
3. **Design?** → [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md) Similar Pattern
4. **Details?** → [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md) Full Explanation

---

## 🎓 Learning Path

```
Level 1: Beginner (30 minutes)
├─ CONFIG_QUICK_REFERENCE.md (Minimal Template)
├─ Copy single-agent example
└─ Run it

Level 2: Intermediate (2 hours)
├─ CONFIG_SPECIFICATION.md (Sections 1-2)
├─ AGENT_MODEL_CONFIGURATION.md (Understand models)
├─ TEAM_SETUP_EXAMPLES.md (Pick one team)
├─ Customize the team
└─ Test it

Level 3: Advanced (4 hours)
├─ CONFIG_SPECIFICATION.md (Full)
├─ CONFIG_SCHEMA_REFERENCE.md
├─ CORE_LIBRARY_UPDATES.md (New features)
├─ Design custom team from scratch
├─ Implement signal routing
└─ Test edge cases

Level 4: Expert (Ongoing)
├─ All documents
├─ CORE_LIBRARY_UPDATES.md (Advanced features)
├─ LIBRARY_USAGE.md (Tool integration)
├─ ARCHITECTURE.md
└─ Build production systems
```

---

## 📋 Tìm Nhanh (Quick Lookup)

### Tôi muốn...

**...thiết lập nhanh một crew**
→ [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md) - Minimal Template section

**...hiểu field cụ thể nào đó**
→ [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md) - Search for field name

**...check YAML syntax đúng hay sai**
→ [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md) - Type Definitions section

**...copy một team có cấu trúc tương tự**
→ [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md) - Pick matching pattern

**...validate configuration của tôi**
→ [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md) - Validation Checklist

**...fix lỗi trong config**
→ [CONFIG_SCHEMA_REFERENCE.md](CONFIG_SCHEMA_REFERENCE.md) - Common Mistakes & Fixes

**...thiết kế multi-agent workflow**
→ [TEAM_SETUP_EXAMPLES.md](TEAM_SETUP_EXAMPLES.md) - Study Team 1 & 2

**...hiểu signal routing như thế nào**
→ [CONFIG_SPECIFICATION.md](CONFIG_SPECIFICATION.md) - Section 1.3 & 1.4

**...lựa chọn provider nào (Ollama vs OpenAI)**
→ [AGENT_MODEL_CONFIGURATION.md](AGENT_MODEL_CONFIGURATION.md) - Provider Setup

**...biết temperature nên set bao nhiêu**
→ [AGENT_MODEL_CONFIGURATION.md](AGENT_MODEL_CONFIGURATION.md) - Temperature Configuration

**...tìm hiểu tính năng mới của core library**
→ [CORE_LIBRARY_UPDATES.md](CORE_LIBRARY_UPDATES.md) - All new features

**...cấu hình model fallback (primary & backup)**
→ [AGENT_MODEL_CONFIGURATION.md](AGENT_MODEL_CONFIGURATION.md) - Backup Model section

**...thiết lập metrics và monitoring**
→ [CORE_LIBRARY_UPDATES.md](CORE_LIBRARY_UPDATES.md) - Section 5 Metrics

---

## 📊 Tài Liệu So Sánh

| Yêu cầu | Quick Reference | Specification | Schema | Examples |
|---------|-----------------|----------------|--------|----------|
| Setup nhanh | ⭐⭐⭐ | ⭐ | ⭐ | ⭐⭐ |
| Chi tiết kỹ thuật | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Ví dụ code | ⭐⭐ | ⭐⭐ | ⭐ | ⭐⭐⭐ |
| Troubleshooting | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Validation | ⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐ |
| Learning | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |

---

## 🔗 Related Documentation

**Existing Docs**:
- [LIBRARY_USAGE.md](LIBRARY_USAGE.md) - Core library API reference
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture overview
- [examples/00-hello-crew/README.md](../examples/00-hello-crew/README.md) - Hello Crew example
- [examples/it-support/README.md](../examples/it-support/README.md) - IT Support example

**Examples in Repository**:
- `examples/00-hello-crew/config/` - Single-agent example
- `examples/it-support/config/` - Multi-agent example (Tiếng Việt)

---

## 📝 Ghi Chú

### Tài Liệu Này Bao Gồm:

✅ crew.yaml specification chi tiết
✅ agent.yaml specification chi tiết
✅ JSON Schema validation rules
✅ Minimal templates sẵn dùng
✅ 4 complete team examples
✅ Best practices & guidelines
✅ Troubleshooting common issues
✅ Quick reference tables
✅ Learning paths for all levels

### Không Bao Gồm:

❌ API reference (xem LIBRARY_USAGE.md)
❌ System architecture deep dive (xem ARCHITECTURE.md)
❌ Tool implementation guide
❌ Deployment guide
❌ Performance tuning guide

---

## 💡 Tips

1. **Bookmark**: Bookmark [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md) cho tra cứu nhanh
2. **Template**: Sao lưu minimal template từ Quick Reference
3. **Validate**: Luôn chạy validation checklist trước khi test
4. **Learn**: Bắt đầu từ Quick Reference, sau đó đi vào chi tiết
5. **Practice**: Copy ví dụ trước, sau đó customize

---

## Version Info

- **Documentation Version**: 1.0
- **Schema Version**: 1.0
- **Last Updated**: 2025-12-22
- **Compatible with**: go-agentic core library v1.0+

---

## Feedback & Contributions

Để cải thiện tài liệu:
- Report issues: GitHub Issues
- Suggest improvements: GitHub Discussions
- Contribute: Pull Requests

---

**Happy Configuration! 🚀**

Để bắt đầu ngay, đi đến [CONFIG_QUICK_REFERENCE.md](CONFIG_QUICK_REFERENCE.md)
