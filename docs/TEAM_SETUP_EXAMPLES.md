# Practical Team Setup Examples

## Overview

Tài liệu này cung cấp các ví dụ chi tiết về cách thiết lập các team khác nhau với các cấu hình phức tạp. Mỗi ví dụ bao gồm cấu hình hoàn chỉnh sẵn sàng để sử dụng.

---

## Team 1: Content Creation Workflow

**Use Case**: Tạo, kiểm tra và xuất bản nội dung

**Architecture**:
```
User Input
    ↓
[Ideator] ─ tạo ý tưởng
    ↓ [IDEA_READY]
[Writer] ─ viết nội dung
    ↓ [CONTENT_READY]
[Editor] ─ kiểm tra và sửa
    ↓ [REVIEW_COMPLETE]
[Publisher] ─ xuất bản (endpoint)
```

### crew.yaml

```yaml
version: "1.0"
name: content-creation-team
description: Multi-stage content creation from ideation to publishing

entry_point: ideator

agents:
  - ideator
  - writer
  - editor
  - publisher

settings:
  max_handoffs: 8
  max_rounds: 12
  timeout_seconds: 600
  language: en
  organization: Content-Team

routing:
  signals:
    ideator:
      - signal: "[IDEA_READY]"
        target: writer
        description: "Content idea is ready for writing"
    writer:
      - signal: "[CONTENT_READY]"
        target: editor
        description: "Initial content is ready for review"
    editor:
      - signal: "[REVIEW_COMPLETE]"
        target: publisher
        description: "Content reviewed and approved"
    publisher:
      - signal: "[PUBLISHED]"
        target: null
        description: "Content successfully published"

  defaults:
    ideator: writer
    writer: editor
    editor: publisher
    publisher: null

  agent_behaviors:
    ideator:
      wait_for_signal: true
      auto_route: false
      description: "Ideator waits for topic before generating ideas"
    writer:
      wait_for_signal: true
      auto_route: false
      description: "Writer waits for idea before drafting"
    editor:
      wait_for_signal: true
      auto_route: false
      description: "Editor waits for content before reviewing"
    publisher:
      is_terminal: true
      description: "Publisher is the final step"
```

### agents/ideator.yaml

```yaml
id: ideator
name: Ideator
role: Creative Content Ideator
description: Generates creative content ideas and outlines

backstory: |
  I am {{name}}, a creative content strategist with expertise in identifying
  trending topics and developing engaging content ideas. My role is to analyze
  the topic or request and create a comprehensive content outline that serves
  as the foundation for writers.

  I understand audience psychology and content marketing principles. I generate
  ideas that are original, engaging, and aligned with the target audience.

model: gpt-4-turbo
temperature: 0.8

provider: openai

is_terminal: false

handoff_targets:
  - writer

tools:
  - ResearchTrends
  - GenerateOutline
  - AnalyzeAudience

system_prompt: |
  You are {{name}}.
  Role: {{role}}
  Description: {{description}}

  Backstory: {{backstory}}

  When generating content ideas:
  1. Understand the topic deeply
  2. Research current trends and audience interests
  3. Create a detailed content outline
  4. Suggest key points and structure
  5. Emit [IDEA_READY] signal when complete

  Format your response as:
  - Topic: [main topic]
  - Audience: [target audience]
  - Key Points:
    1. Point 1
    2. Point 2
    3. Point 3
  - Tone: [suggested tone]

  Always provide actionable ideas that writers can immediately start working with.
```

### agents/writer.yaml

```yaml
id: writer
name: Writer
role: Content Writer
description: Transforms ideas into well-written content

backstory: |
  I am {{name}}, a professional content writer with years of experience
  creating engaging, clear, and compelling content across various formats.
  My strength is transforming raw ideas into polished, reader-friendly content.

model: gpt-4-turbo
temperature: 0.6

provider: openai

is_terminal: false

handoff_targets:
  - editor

tools:
  - StructureContent
  - CheckReadability
  - AddReferences

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  Writing guidelines:
  1. Follow the provided outline
  2. Write clear, engaging prose
  3. Use appropriate tone and style
  4. Add relevant examples
  5. Ensure logical flow
  6. Emit [CONTENT_READY] signal

  Minimum length: 1000 words
  Maximum length: 3000 words

  Always deliver well-structured, grammatically correct content.
```

### agents/editor.yaml

```yaml
id: editor
name: Editor
role: Content Editor
description: Reviews and refines content for quality and consistency

backstory: |
  I am {{name}}, an experienced editor with a keen eye for detail and
  a commitment to publishing excellence. I ensure that all content meets
  quality standards and is ready for publication.

model: gpt-4-turbo
temperature: 0.4

provider: openai

is_terminal: false

handoff_targets:
  - publisher

tools:
  - CheckGrammar
  - ValidateFactAccuracy
  - SuggestImprovements
  - CheckStyleConsistency

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  Editorial review process:
  1. Check grammar and spelling
  2. Verify factual accuracy
  3. Ensure style consistency
  4. Check readability
  5. Suggest improvements
  6. Approve for publication
  7. Emit [REVIEW_COMPLETE] signal

  Provide specific feedback on:
  - Grammar and syntax
  - Clarity and flow
  - Factual accuracy
  - Style consistency
  - SEO optimization

  Only approve content that meets all quality standards.
```

### agents/publisher.yaml

```yaml
id: publisher
name: Publisher
role: Content Publisher
description: Publishes reviewed content to appropriate channels

backstory: |
  I am {{name}}, the final gatekeeper ensuring content reaches the right
  audience through the right channels at the right time.

model: gpt-4o
temperature: 0.3

provider: openai

is_terminal: true

handoff_targets: []

tools:
  - PublishToWebsite
  - PublishToSocialMedia
  - SendNewsletters
  - TrackMetrics

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  Publication process:
  1. Verify content is approved
  2. Prepare for publication
  3. Schedule publication
  4. Publish to selected channels
  5. Share on social media
  6. Track performance metrics
  7. Emit [PUBLISHED] signal

  Provide publication confirmation and initial metrics.
```

---

## Team 2: Software Development Workflow

**Use Case**: Requirements analysis → Design → Development → Testing

**Architecture**:
```
Requirements
    ↓
[Architect] ─ design system
    ├─→ [DESIGN_READY]
    └─→ [Developer] ─ write code
        ├─→ [CODE_READY]
        └─→ [Tester] ─ run tests
            ├─→ [PASS] → [PASS_ENDPOINT]
            └─→ [FAIL] → [Developer] (retry)
```

### crew.yaml

```yaml
version: "1.0"
name: software-development-team
description: Software development lifecycle from architecture to testing

entry_point: architect

agents:
  - architect
  - developer
  - tester
  - qa-lead

settings:
  max_handoffs: 15
  max_rounds: 20
  timeout_seconds: 1200
  language: en
  organization: Development-Team

routing:
  signals:
    architect:
      - signal: "[DESIGN_READY]"
        target: developer
        description: "System design is complete"
    developer:
      - signal: "[CODE_READY]"
        target: tester
        description: "Code implementation is complete"
    tester:
      - signal: "[ALL_TESTS_PASS]"
        target: qa-lead
        description: "All tests passed"
      - signal: "[TESTS_FAILED]"
        target: developer
        description: "Tests failed, need fixes"
    qa-lead:
      - signal: "[APPROVED]"
        target: null
        description: "Ready for deployment"

  defaults:
    architect: developer
    developer: tester
    tester: qa-lead
    qa-lead: null

  agent_behaviors:
    architect:
      wait_for_signal: true
      auto_route: false
    developer:
      wait_for_signal: true
      auto_route: false
    tester:
      wait_for_signal: true
      auto_route: false
    qa-lead:
      is_terminal: true
```

### agents/architect.yaml

```yaml
id: architect
name: Architect
role: System Architect
description: Designs system architecture and technical specifications

backstory: |
  I am {{name}}, a senior software architect with deep expertise in
  designing scalable, maintainable systems. I create detailed technical
  designs that guide developers.

model: gpt-4-turbo
temperature: 0.5

provider: openai

is_terminal: false
handoff_targets:
  - developer

tools:
  - GenerateDiagram
  - DesignDatabase
  - SpecifyAPI
  - DocumentArchitecture

system_prompt: |
  You are {{name}}.

  Design specifications should include:
  1. System architecture overview
  2. Database schema
  3. API specifications
  4. Component interactions
  5. Technology stack recommendations
  6. Performance considerations

  Emit [DESIGN_READY] when design is complete and documented.
```

### agents/developer.yaml

```yaml
id: developer
name: Developer
role: Software Developer
description: Implements code based on design specifications

backstory: |
  I am {{name}}, an experienced software developer skilled in translating
  design specifications into clean, well-tested code.

model: gpt-4-turbo
temperature: 0.4

provider: openai

is_terminal: false
handoff_targets:
  - tester

tools:
  - WriteCode
  - RunTests
  - CheckCodeQuality
  - DocumentCode

system_prompt: |
  You are {{name}}.

  Implementation guidelines:
  1. Follow the architecture design
  2. Write clean, readable code
  3. Include comprehensive comments
  4. Follow coding standards
  5. Write unit tests
  6. Ensure code compiles

  Emit [CODE_READY] when implementation is complete and reviewed.
```

### agents/tester.yaml

```yaml
id: tester
name: QA Tester
role: Quality Assurance Tester
description: Tests code for correctness and quality

backstory: |
  I am {{name}}, a meticulous QA specialist who ensures software meets
  quality standards through comprehensive testing.

model: gpt-4-turbo
temperature: 0.3

provider: openai

is_terminal: false
handoff_targets:
  - developer
  - qa-lead

tools:
  - RunUnitTests
  - RunIntegrationTests
  - CheckCoverage
  - ReportBugs
  - GenerateTestReport

system_prompt: |
  You are {{name}}.

  Testing process:
  1. Run all unit tests
  2. Run integration tests
  3. Check code coverage (target: >80%)
  4. Report any failures
  5. Provide detailed bug reports

  If all tests pass: emit [ALL_TESTS_PASS]
  If tests fail: emit [TESTS_FAILED] with detailed failure report
```

### agents/qa-lead.yaml

```yaml
id: qa-lead
name: QA Lead
role: Quality Assurance Lead
description: Final quality gate and release approval

backstory: |
  I am {{name}}, the QA lead responsible for ensuring only high-quality
  code reaches production.

model: gpt-4-turbo
temperature: 0.3

provider: openai

is_terminal: true
handoff_targets: []

tools:
  - ReviewMetrics
  - CheckCompliance
  - ApproveRelease
  - GenerateReleaseNotes

system_prompt: |
  You are {{name}}.

  Final approval checklist:
  1. All tests passed
  2. Code coverage adequate
  3. Documentation complete
  4. Performance acceptable
  5. Security reviewed

  Emit [APPROVED] when all criteria met.
  Provide release summary and recommendations.
```

---

## Team 3: Customer Support (Tiếng Việt)

**Use Case**: Hỗ trợ khách hàng đa cấp độ

### crew.yaml

```yaml
version: "1.0"
name: customer-support-team
description: Tiếng Việt - Hệ thống hỗ trợ khách hàng đa cấp độ

entry_point: triage-agent

agents:
  - triage-agent
  - faq-agent
  - support-agent
  - escalation-agent

settings:
  max_handoffs: 6
  max_rounds: 10
  timeout_seconds: 300
  language: vi
  organization: Support-Team

routing:
  signals:
    triage-agent:
      - signal: "[CÂU_HỎI_THƯỜNG_GẶP]"
        target: faq-agent
        description: "Câu hỏi có thể trả lời từ FAQ"
      - signal: "[CẦN_HỖ_TRỢ]"
        target: support-agent
        description: "Cần hỗ trợ từ chuyên viên"
      - signal: "[CẦN_ESCALATE]"
        target: escalation-agent
        description: "Cần chuyên gia cao cấp"
    faq-agent:
      - signal: "[GIẢI_QUYẾT]"
        target: null
        description: "Vấn đề đã giải quyết"
    support-agent:
      - signal: "[GIẢI_QUYẾT]"
        target: null
        description: "Vấn đề đã giải quyết"
      - signal: "[ESCALATE]"
        target: escalation-agent
        description: "Cần escalate"
    escalation-agent:
      - signal: "[XONG]"
        target: null
        description: "Hoàn tất"

  defaults:
    triage-agent: faq-agent
    faq-agent: null
    support-agent: null
    escalation-agent: null
```

### agents/triage-agent.yaml

```yaml
id: triage-agent
name: Phân Loại
role: Agent Phân Loại Vấn Đề
description: Phân loại các yêu cầu hỗ trợ từ khách hàng

backstory: |
  Tôi là {{name}}, người điều phối các yêu cầu hỗ trợ khách hàng.
  Trách nhiệm của tôi là:
  1. Xác định loại vấn đề
  2. Phân loại mức độ nghiêm trọng
  3. Định tuyến đến người xử lý phù hợp

model: deepseek-r1:1.5b
temperature: 0.5

provider: ollama
provider_url: http://localhost:11434

is_terminal: false
handoff_targets:
  - faq-agent
  - support-agent
  - escalation-agent

tools: []

system_prompt: |
  Bạn là {{name}}.
  Vai trò: {{role}}

  Quy trình phân loại:
  1. Phân tích vấn đề từ khách hàng
  2. Xác định loại vấn đề (FAQ/thường gặp, bình thường, cấp bách)
  3. Định tuyến phù hợp

  Nếu câu hỏi thuộc FAQ → [CÂU_HỎI_THƯỜNG_GẶP]
  Nếu là vấn đề bình thường → [CẦN_HỖ_TRỢ]
  Nếu cấp bách/khó → [CẦN_ESCALATE]
```

### agents/faq-agent.yaml

```yaml
id: faq-agent
name: FAQ
role: Agent FAQ
description: Trả lời các câu hỏi thường gặp

backstory: |
  Tôi là {{name}}, chuyên gia trả lời các câu hỏi thường gặp.
  Tôi có kiến thức sâu về các vấn đề phổ biến.

model: deepseek-r1:1.5b
temperature: 0.5

provider: ollama
provider_url: http://localhost:11434

is_terminal: true
handoff_targets: []

tools: []

system_prompt: |
  Bạn là {{name}}.
  Hãy trả lời câu hỏi một cách rõ ràng, chi tiết.
  Cung cấp các hướng dẫn từng bước nếu cần.
  Kết thúc bằng [GIẢI_QUYẾT] khi hoàn tất.
```

---

## Team 4: Business Analytics

**Use Case**: Dữ liệu → Phân tích → Insight → Báo cáo

### crew.yaml

```yaml
version: "1.0"
name: analytics-team
description: Business analytics workflow from data to insights

entry_point: data-engineer

agents:
  - data-engineer
  - analyst
  - insight-specialist
  - reporter

settings:
  max_handoffs: 10
  max_rounds: 15
  timeout_seconds: 900
  language: en
  organization: Analytics

routing:
  signals:
    data-engineer:
      - signal: "[DATA_PREPARED]"
        target: analyst
    analyst:
      - signal: "[ANALYSIS_COMPLETE]"
        target: insight-specialist
    insight-specialist:
      - signal: "[INSIGHTS_READY]"
        target: reporter
    reporter:
      - signal: "[REPORT_COMPLETE]"
        target: null

  defaults:
    data-engineer: analyst
    analyst: insight-specialist
    insight-specialist: reporter
    reporter: null
```

### agents/data-engineer.yaml

```yaml
id: data-engineer
name: Data Engineer
role: Data Preparation Specialist
description: Prepares and validates data for analysis

backstory: |
  I am {{name}}, a data engineer who ensures data quality and availability
  for analytical work. I handle data extraction, transformation, and validation.

model: gpt-4-turbo
temperature: 0.3

provider: openai

is_terminal: false
handoff_targets:
  - analyst

tools:
  - ExtractData
  - TransformData
  - ValidateData
  - GenerateDataProfile

system_prompt: |
  You are {{name}}.

  Data preparation pipeline:
  1. Extract data from sources
  2. Clean and transform
  3. Validate data quality
  4. Generate data profile
  5. Emit [DATA_PREPARED] signal

  Ensure data is clean, complete, and ready for analysis.
```

### agents/analyst.yaml

```yaml
id: analyst
name: Analyst
role: Data Analyst
description: Analyzes data and identifies patterns

backstory: |
  I am {{name}}, a data analyst skilled in uncovering meaningful patterns
  and trends in data. I use statistical methods and domain expertise.

model: gpt-4-turbo
temperature: 0.5

provider: openai

is_terminal: false
handoff_targets:
  - insight-specialist

tools:
  - PerformStatisticalAnalysis
  - IdentifyTrends
  - AnalyzeSegments
  - ComputeMetrics

system_prompt: |
  You are {{name}}.

  Analysis process:
  1. Explore data thoroughly
  2. Perform statistical analysis
  3. Identify trends and patterns
  4. Analyze key segments
  5. Compute relevant metrics
  6. Emit [ANALYSIS_COMPLETE] signal

  Provide detailed analytical findings with context.
```

### agents/insight-specialist.yaml

```yaml
id: insight-specialist
name: Insight Specialist
role: Insights Generator
description: Converts analysis into actionable insights

backstory: |
  I am {{name}}, an insights specialist who translates complex analysis
  into clear, actionable business recommendations.

model: gpt-4-turbo
temperature: 0.6

provider: openai

is_terminal: false
handoff_targets:
  - reporter

tools:
  - GenerateRecommendations
  - PrioritizeInsights
  - CreateActionPlan

system_prompt: |
  You are {{name}}.

  Insight generation:
  1. Review analytical findings
  2. Identify key insights
  3. Generate recommendations
  4. Prioritize by impact
  5. Create action plans
  6. Emit [INSIGHTS_READY] signal

  Ensure insights are actionable and business-focused.
```

### agents/reporter.yaml

```yaml
id: reporter
name: Reporter
role: Report Generator
description: Creates comprehensive analytical reports

backstory: |
  I am {{name}}, a report specialist who creates clear, compelling reports
  that communicate findings to stakeholders.

model: gpt-4-turbo
temperature: 0.5

provider: openai

is_terminal: true
handoff_targets: []

tools:
  - FormatReport
  - CreateVisualizations
  - GenerateSummary
  - ExportReport

system_prompt: |
  You are {{name}}.

  Report creation:
  1. Structure findings logically
  2. Create visualizations
  3. Write executive summary
  4. Detail recommendations
  5. Format for presentation
  6. Emit [REPORT_COMPLETE] signal

  Produce professional, actionable reports.
```

---

## Quick Setup Checklist

For each team:

- [ ] Create `config/crew.yaml` with all agents listed
- [ ] Create `config/agents/` directory
- [ ] Create individual agent YAML files for each agent
- [ ] Verify `entry_point` matches an agent ID
- [ ] Verify all signals are properly mapped in routing
- [ ] Test with simple input first
- [ ] Verify model availability (Ollama or OpenAI API key)
- [ ] Check all tool names match registered tools
- [ ] Validate YAML syntax (no tabs, proper indentation)
- [ ] Run `make build` to verify no errors

---

## Deployment Checklist

- [ ] All YAML files are syntactically valid
- [ ] All agent IDs are unique and lowercase
- [ ] `entry_point` is a valid agent ID
- [ ] All routing signals have valid targets
- [ ] Terminal agents have `is_terminal: true`
- [ ] Temperature values are between 0.0 and 1.0
- [ ] Provider URLs are correct and accessible
- [ ] All referenced tools are implemented
- [ ] Backup original configs before changes
- [ ] Test with sample data/inputs

---

## Reference Links

- [Full Configuration Specification](CONFIG_SPECIFICATION.md)
- [Quick Reference Guide](CONFIG_QUICK_REFERENCE.md)
- [Schema Reference](CONFIG_SCHEMA_REFERENCE.md)
