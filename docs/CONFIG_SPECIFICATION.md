# Go-Agentic Configuration Specification

## Tổng Quan

Tài liệu này mô tả đặc tả kỹ thuật cho các tập tin cấu hình YAML trong hệ thống go-agentic. Hệ thống sử dụng hai loại tập tin cấu hình chính:

1. **crew.yaml** - Định nghĩa đội làm việc (crew) và cách các agent tương tác
2. **agent.yaml** - Định nghĩa chi tiết của từng agent riêng lẻ

---

## 1. Crew Configuration (crew.yaml)

### 1.1 Cấu Trúc Tổng Quát

```yaml
version: "1.0"
name: crew-name
description: Mô tả chi tiết về crew
entry_point: agent-id-bắt-đầu

agents:
  - agent-id-1
  - agent-id-2
  - agent-id-3

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: Tên tổ chức

routing:
  signals: {}
  defaults: {}
  agent_behaviors: {}
```

### 1.2 Các Trường Bắt Buộc

#### **version** (String, Bắt buộc)
- **Định nghĩa**: Phiên bản schema cấu hình
- **Giá trị**: `"1.0"` (hiện tại)
- **Ví dụ**:
  ```yaml
  version: "1.0"
  ```

#### **name** (String, Bắt buộc)
- **Định nghĩa**: Tên định danh của crew, được sử dụng cho logging và tracking
- **Quy tắc**:
  - Chỉ chứa ký tự: `a-z`, `0-9`, hyphen (`-`)
  - Độ dài: 3-50 ký tự
  - Viết thường (lowercase)
- **Ví dụ**:
  ```yaml
  name: hello-crew
  name: it-support-system
  name: data-analysis-team
  ```

#### **description** (String, Bắt buộc)
- **Định nghĩa**: Mô tả chi tiết về chức năng của crew
- **Quy tắc**: Nên mô tả rõ ràng mục đích và phạm vi
- **Ví dụ**:
  ```yaml
  description: A minimal crew with a single Hello agent for learning basics
  description: Multi-agent system for IT troubleshooting and system diagnostics
  ```

#### **entry_point** (String, Bắt buộc)
- **Định nghĩa**: ID của agent được sử dụng làm điểm khởi đầu của crew
- **Quy tắc**:
  - Phải khớp với một trong các agent IDs được liệt kê trong `agents` array
  - Đây là agent sẽ nhận request đầu tiên từ người dùng
- **Ví dụ**:
  ```yaml
  entry_point: hello-agent        # Single-agent crew
  entry_point: orchestrator       # Multi-agent crew
  ```

#### **agents** (Array of Strings, Bắt buộc)
- **Định nghĩa**: Danh sách các agent IDs thuộc crew
- **Quy tắc**:
  - Mỗi item phải là ID của agent được định nghĩa trong các tập tin YAML riêng
  - Thứ tự liệt kê có thể ảnh hưởng đến routing logic
  - Mỗi ID phải là duy nhất trong crew
- **Ví dụ**:
  ```yaml
  agents:
    - hello-agent

  agents:
    - orchestrator
    - clarifier
    - executor
  ```

### 1.3 Các Trường Tùy Chọn

#### **settings** (Object, Tùy chọn)
- **Định nghĩa**: Cấu hình toàn cục cho crew
- **Trường con**:

##### settings.max_handoffs (Integer)
- **Mặc định**: 5
- **Định nghĩa**: Số lần handoff tối đa giữa các agent
- **Ví dụ**: `max_handoffs: 10` cho complex workflows

##### settings.max_rounds (Integer)
- **Mặc định**: 10
- **Định nghĩa**: Số lần tối đa mà mỗi agent có thể xử lý request
- **Ví dụ**: `max_rounds: 20` cho deep analysis

##### settings.timeout_seconds (Integer)
- **Mặc định**: 300 (5 phút)
- **Định nghĩa**: Thời gian timeout tối đa cho toàn bộ execution
- **Đơn vị**: Giây (seconds)
- **Ví dụ**:
  ```yaml
  settings:
    timeout_seconds: 600  # 10 phút
  ```

##### settings.language (String)
- **Mặc định**: `en`
- **Định nghĩa**: Ngôn ngữ chính của crew
- **Giá trị hỗ trợ**: `en`, `vi` (có thể mở rộng)
- **Ví dụ**: `language: vi` cho crew tiếng Việt

##### settings.organization (String)
- **Mặc định**: Không có
- **Định nghĩa**: Tên tổ chức quản lý crew
- **Ví dụ**:
  ```yaml
  settings:
    organization: IT-Support-Team
    organization: Data-Analysis-Department
  ```

#### **routing** (Object, Tùy chọn)
- **Định nghĩa**: Cấu hình định tuyến tín hiệu giữa các agent
- **Trường con**:

##### routing.signals
- **Định nghĩa**: Bản đồ các tín hiệu mà agent có thể phát
- **Cấu trúc**:
  ```yaml
  routing:
    signals:
      agent-id:
        - signal: "[SIGNAL_NAME]"
          target: target-agent-id
          description: Mô tả tín hiệu
  ```
- **Ví dụ**: Xem phần 2.3

##### routing.defaults
- **Định nghĩa**: Routing mặc định khi không có tín hiệu explicit
- **Cấu trúc**:
  ```yaml
  routing:
    defaults:
      agent-id: target-agent-id-hoặc-null
  ```

##### routing.parallel_groups

- **Định nghĩa**: Nhóm agent được kích hoạt đồng thời từ một signal
- **Cấu trúc**:
  ```yaml
  routing:
    parallel_groups:
      group-name:
        agents: [agent1, agent2]       # Danh sách agents chạy song song
        wait_for_all: boolean          # true=chờ tất cả, false=chỉ cần 1
        timeout_seconds: integer       # Timeout cho nhóm
  ```
- **Ví dụ**:
  ```yaml
  routing:
    signals:
      teacher:
        - signal: "[QUESTION]"
          target: parallel_question    # Target là tên parallel group

    parallel_groups:
      parallel_question:
        agents: [student, reporter]
        wait_for_all: false
        timeout_seconds: 30
  ```

##### routing.agent_behaviors
- **Định nghĩa**: Hành động riêng của từng agent
- **Cấu trúc**:
  ```yaml
  routing:
    agent_behaviors:
      agent-id:
        wait_for_signal: boolean
        auto_route: boolean
        is_terminal: boolean
        description: Mô tả hành động
  ```

### 1.4 Ví Dụ Chi Tiết

#### Single-Agent Crew (Hello Crew)

```yaml
version: "1.0"
name: hello-crew
description: A minimal crew with a single Hello agent

entry_point: hello-agent

agents:
  - hello-agent

settings:
  max_handoffs: 1
  max_rounds: 1
  timeout_seconds: 30
  language: en
  organization: Learning-Example
```

#### Multi-Agent Crew (IT Support)

```yaml
version: "1.0"
description: "Go-CrewAI IT Support Crew Configuration"

entry_point: orchestrator

agents:
  - orchestrator
  - clarifier
  - executor

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
  organization: IT-Support-Team

routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
        description: "Route to executor for immediate diagnosis"
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
        description: "Route to clarifier for information gathering"
    clarifier:
      - signal: "[KẾT THÚC]"
        target: executor
        description: "Hand off to executor after gathering info"
    executor:
      - signal: "[COMPLETE]"
        target: null
        description: "Diagnosis complete, return result"

  defaults:
    orchestrator: clarifier
    clarifier: executor
    executor: null

  agent_behaviors:
    orchestrator:
      wait_for_signal: true
      auto_route: false
      description: "Orchestrator always waits for explicit routing signal"
    clarifier:
      wait_for_signal: true
      auto_route: false
      description: "Clarifier waits for signal before handoff"
    executor:
      is_terminal: true
      description: "Executor is terminal, returns immediately"
```

#### Parallel Execution Crew (Quiz Exam)

```yaml
version: "1.0"
name: quiz-exam-crew
description: Multi-agent oral exam with parallel execution

entry_point: teacher

agents:
  - teacher
  - student
  - reporter

settings:
  max_rounds: 30
  max_handoffs: 30

routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question  # Routes to Student + Reporter
      - signal: "[END]"
        target: reporter           # Final report save
    student:
      - signal: "[ANSWER]"
        target: parallel_answer    # Routes to Teacher + Reporter
    reporter:
      - signal: "[OK]"
        target: ""                 # Acknowledge, no routing
      - signal: "[DONE]"
        target: ""                 # Terminate workflow

  parallel_groups:
    parallel_question:
      agents: [student, reporter]
      wait_for_all: false
      timeout_seconds: 30
    parallel_answer:
      agents: [teacher, reporter]
      wait_for_all: false
      timeout_seconds: 30

  agent_behaviors:
    teacher:
      wait_for_signal: false
    student:
      wait_for_signal: false
    reporter:
      wait_for_signal: false
```

**Workflow Diagram:**

```text
┌──────────────┐                      ┌──────────────┐
│   Teacher    │ ────[QUESTION]────▶  │   Student    │
│              │                      │              │
│  Tools:      │                      │  (No tools)  │
│  - GetStatus │                      │              │
│  - Record    │ ◀────[ANSWER]─────   │              │
└──────────────┘                      └──────────────┘
       │                                     │
       │ [QUESTION]                          │ [ANSWER]
       ▼                                     ▼
┌──────────────┐                      ┌──────────────┐
│   Reporter   │ ◀───────────────────│   Reporter   │
│              │                      │   (parallel) │
│  Tools:      │                      │              │
│  - GetStatus │                      │              │
│  - WriteRpt  │                      │              │
└──────────────┘                      └──────────────┘
       │
       │ [END] → Reporter → [DONE] → Terminate
       ▼
```

---

## 2. Agent Configuration (agent.yaml)

### 2.1 Cấu Trúc Tổng Quát

```yaml
# Identifiers
id: unique-agent-id
name: Tên hiển thị
role: Vai trò chính
description: Mô tả chi tiết

# Backstory
backstory: |
  Thông tin bối cảnh multi-line

# Model Configuration
model: model-name
temperature: 0.7
provider: ollama|openai
provider_url: http://localhost:11434

# Agent Behavior
is_terminal: true|false
handoff_targets: []

# Tools & System
tools:
  - ToolName1
  - ToolName2

system_prompt: |
  System prompt template
```

### 2.2 Các Trường Bắt Buộc

#### **id** (String, Bắt buộc)
- **Định nghĩa**: Định danh duy nhất của agent trong crew
- **Quy tắc**:
  - Chỉ chứa ký tự: `a-z`, `0-9`, hyphen, underscore
  - Độ dài: 3-50 ký tự
  - Viết thường (lowercase)
  - Phải khớp với tên tập tin (không có `.yaml`)
- **Ví dụ**:
  ```yaml
  id: hello-agent
  id: orchestrator
  id: data_analyzer
  ```

#### **name** (String, Bắt buộc)
- **Định nghĩa**: Tên hiển thị của agent (có thể dùng tiếng Việt)
- **Quy tắc**: Có thể chứa bất kỳ ký tự nào, bao gồm ký tự Unicode
- **Ví dụ**:
  ```yaml
  name: Hello Agent
  name: My              # Orchestrator trong IT Support
  name: Ngân            # Clarifier trong IT Support
  name: Trang           # Executor trong IT Support
  ```

#### **role** (String, Bắt buộc)
- **Định nghĩa**: Vai trò chính của agent trong team
- **Quy tắc**: Mô tả rõ ràng trách nhiệm
- **Ví dụ**:
  ```yaml
  role: Friendly Assistant
  role: Điều phối viên hệ thống
  role: Người thu thập thông tin
  role: Chuyên gia khắc phục sự cố IT
  ```

#### **description** (String, Bắt buộc)
- **Định nghĩa**: Mô tả chi tiết về agent
- **Quy tắc**: 1-2 câu, rõ ràng và ngắn gọn
- **Ví dụ**:
  ```yaml
  description: A simple and friendly assistant that greets users
  description: Người điều phối hệ thống và điểm vào cho các yêu cầu hỗ trợ IT
  ```

#### **backstory** (String, Bắt buộc)
- **Định nghĩa**: Bối cảnh chi tiết về agent, có thể sử dụng multiline
- **Quy tắc**:
  - Sử dụng syntax `|` hoặc `|-` cho multiline
  - Có thể chứa template variables: `{{name}}`, `{{role}}`, `{{description}}`
  - Nên chi tiết, sống động, và giúp LLM hiểu ngữ cảnh
- **Ví dụ**:
  ```yaml
  backstory: |
    You are {{name}}, a warm and welcoming assistant.
    Your role is to greet users and understand their needs.
  ```

#### **model** (String, Bắt buộc)
- **Định nghĩa**: Tên model LLM sẽ sử dụng
- **Giá trị cho Ollama**: `gemma3:1b`, `deepseek-r1:1.5b`, `mistral:latest`
- **Giá trị cho OpenAI**: `gpt-4o`, `gpt-4-turbo`, `gpt-4o-mini`
- **Ví dụ**:
  ```yaml
  model: gemma3:1b                # Ollama
  model: deepseek-r1:1.5b         # Ollama
  model: gpt-4-turbo              # OpenAI
  ```

#### **temperature** (Number, Bắt buộc)
- **Định nghĩa**: Độ sáng tạo của model (0.0 - 1.0)
- **Giá trị**:
  - `0.0` - Kết quả xác định, không sáng tạo
  - `0.5-0.7` - Cân bằng giữa sáng tạo và xác định (khuyến nghị)
  - `1.0` - Rất sáng tạo, không thường xuyên
- **Ví dụ**:
  ```yaml
  temperature: 0.7        # Cân bằng
  temperature: 0.3        # Chính xác (cho tasks logic)
  temperature: 0.9        # Sáng tạo (cho brainstorming)
  ```

#### **provider** (String, Bắt buộc)
- **Định nghĩa**: Nhà cung cấp LLM
- **Giá trị hỗ trợ**: `ollama`, `openai`
- **Ví dụ**:
  ```yaml
  provider: ollama
  provider: openai
  ```

#### **provider_url** (String, Bắt buộc khi provider=ollama)
- **Định nghĩa**: URL của Ollama server hoặc OpenAI endpoint
- **Giá trị mặc định cho Ollama**: `http://localhost:11434`
- **Giá trị cho OpenAI**: Thường không cần (sử dụng API key thay vì URL)
- **Ví dụ**:
  ```yaml
  provider: ollama
  provider_url: http://localhost:11434

  provider: ollama
  provider_url: http://ollama.example.com:11434
  ```

### 2.3 Các Trường Tùy Chọn

#### **is_terminal** (Boolean, Tùy chọn)
- **Mặc định**: `false`
- **Định nghĩa**: Agent này có phải là endpoint (điểm cuối) không?
  - `true`: Execution dừng sau agent này, không route sang agent khác
  - `false`: Agent có thể emit signal để route sang agent khác
- **Quy tắc**:
  - Single-agent crews: Nên là `true`
  - Multi-agent crews: Thường `false` cho intermediate agents, `true` cho terminal agent
- **Ví dụ**:
  ```yaml
  is_terminal: true       # Executor hoặc single-agent crews
  is_terminal: false      # Orchestrator, Clarifier
  ```

#### **handoff_targets** (Array of Strings, Tùy chọn)
- **Mặc định**: `[]` (trống)
- **Định nghĩa**: Danh sách các agent mà agent này có thể route sang
- **Quy tắc**:
  - Mỗi item phải là ID của agent khác trong crew
  - Nên liệt kê chi tiết các target khả dĩ
  - Agent terminal nên để trống
- **Ví dụ**:
  ```yaml
  handoff_targets:
    - clarifier
    - executor

  handoff_targets: []     # Terminal agent
  ```

#### **tools** (Array of Strings, Tùy chọn)
- **Mặc định**: `[]` (không có tool)
- **Định nghĩa**: Danh sách các tool mà agent có thể gọi
- **Quy tắc**:
  - Tên tool phải khớp với các tool được đăng ký trong system
  - Các tool phải có hàm tương ứng với signature chính xác
  - Tên tool viết CamelCase
- **Ví dụ**:
  ```yaml
  tools: []              # Không tool

  tools:
    - GetCPUUsage
    - GetMemoryUsage
    - PingHost
  ```

#### **system_prompt** (String, Tùy chọn)
- **Mặc định**: Auto-generated từ fields khác
- **Định nghĩa**: Custom system prompt cho agent
- **Quy tắc**:
  - Sử dụng `|` hoặc `|-` cho multiline
  - Có thể sử dụng template variables: `{{name}}`, `{{role}}`, `{{description}}`, `{{backstory}}`
  - Có độ ưu tiên cao hơn prompt auto-generated
- **Ví dụ**:
  ```yaml
  system_prompt: |
    You are {{name}}.
    Role: {{role}}
    Description: {{description}}

    Backstory: {{backstory}}

    Important rules:
    - Always respond in Vietnamese
    - Be helpful and concise
  ```

### 2.4 Ví Dụ Chi Tiết

#### Simple Agent (Hello Crew)

```yaml
id: hello-agent
name: Hello Agent
role: Friendly Assistant
description: A simple and friendly assistant that greets users and provides helpful responses

backstory: |
  You are a warm and welcoming assistant. Your role is to greet users, understand their needs,
  and provide helpful, friendly responses. You keep your answers concise and friendly.

model: gemma3:1b
temperature: 0.7
is_terminal: true

provider: ollama
provider_url: http://localhost:11434

tools: []

system_prompt: |
  You are {{name}}.
  Role: {{role}}
  Description: {{description}}

  Backstory: {{backstory}}

  Be friendly, helpful, and concise in your responses.
```

#### Orchestrator Agent (IT Support)

```yaml
id: orchestrator
name: My
role: Điều phối viên hệ thống
description: Điều phối hệ thống và điểm vào cho các yêu cầu hỗ trợ IT

backstory: |
  Tôi là Điều Phối Viên - điểm vào cho mọi yêu cầu hỗ trợ IT của bạn.
  Vai trò của tôi là phân tích mô tả vấn đề và quyết định liệu cần thêm thông tin
  hay có thể tiến hành chẩn đoán ngay lập tức.

model: deepseek-r1:1.5b
temperature: 0.7
is_terminal: false

provider: ollama
provider_url: http://localhost:11434

tools: []

handoff_targets:
  - clarifier
  - executor

system_prompt: |
  Bạn là {{name}}.
  Vai trò: {{role}}
  Mô tả: {{description}}

  [Custom instructions chi tiết...]
```

#### Executor Agent (IT Support)

```yaml
id: executor
name: Trang
role: Chuyên gia khắc phục sự cố IT
description: Chuyên gia khắc phục sự cố IT và chẩn đoán hệ thống

backstory: |
  Tôi là {{name}} - chuyên gia khắc phục sự cố IT có kiến thức sâu về chẩn đoán hệ thống.
  Vai trò của tôi là sử dụng các công cụ khả dụng để chẩn đoán vấn đề và cung cấp giải pháp.

model: deepseek-r1:1.5b
temperature: 0.7
is_terminal: true

provider: ollama
provider_url: http://localhost:11434

tools:
  - GetCPUUsage
  - GetMemoryUsage
  - GetDiskSpace
  - GetSystemInfo
  - GetRunningProcesses
  - PingHost
  - CheckServiceStatus
  - ResolveDNS

handoff_targets: []

system_prompt: |
  Bạn là {{name}}.
  Vai trò: {{role}}
  Mô tả: {{description}}

  [Detailed instructions cho diagnostics...]
```

---

## 3. Một Số Team Ví Dụ Chi Tiết

### 3.1 Team Phân Tích Dữ Liệu (Data Analysis Team)

#### crew.yaml
```yaml
version: "1.0"
name: data-analysis-team
description: Multi-agent team for data loading, analysis, and visualization

entry_point: coordinator

agents:
  - coordinator
  - data-loader
  - analyzer
  - visualizer

settings:
  max_handoffs: 10
  max_rounds: 15
  timeout_seconds: 600
  language: en
  organization: Analytics-Department

routing:
  signals:
    coordinator:
      - signal: "[ANALYZE_DATA]"
        target: analyzer
        description: "Start data analysis"
      - signal: "[LOAD_DATA]"
        target: data-loader
        description: "Load data first"
    data-loader:
      - signal: "[DATA_READY]"
        target: analyzer
        description: "Data loaded, proceed to analysis"
    analyzer:
      - signal: "[VISUALIZATION_NEEDED]"
        target: visualizer
        description: "Create visualizations"
      - signal: "[ANALYSIS_COMPLETE]"
        target: coordinator
        description: "Analysis done"
    visualizer:
      - signal: "[VISUALIZATION_COMPLETE]"
        target: null
        description: "Done"

  defaults:
    coordinator: data-loader
    data-loader: analyzer
    analyzer: visualizer
    visualizer: null
```

#### coordinator.yaml
```yaml
id: coordinator
name: Data Coordinator
role: Workflow Coordinator
description: Coordinates data analysis workflow and manages handoffs

backstory: |
  I am the Data Coordinator. My responsibility is to:
  1. Understand the data analysis task
  2. Route to appropriate agents
  3. Collect results from all agents
  4. Present comprehensive findings

model: gpt-4-turbo
temperature: 0.5
is_terminal: false
provider: openai

tools: []

handoff_targets:
  - data-loader
  - analyzer
  - visualizer

system_prompt: |
  You are {{name}}.
  Role: {{role}}
  Description: {{description}}

  Responsibilities:
  - Parse user data analysis request
  - Route to data-loader if data needs loading
  - Route to analyzer for analysis
  - Route to visualizer for charts
  - Coordinate handoffs
  - Present final summary

  Always maintain professional communication and ensure all steps are completed.
```

#### data-loader.yaml
```yaml
id: data-loader
name: Data Loader
role: Data Source Manager
description: Loads and validates data from various sources

backstory: |
  I am the Data Loader. I can:
  1. Load data from CSV, JSON, Excel files
  2. Validate data quality
  3. Handle missing values
  4. Prepare data for analysis

model: gpt-4-turbo
temperature: 0.3
is_terminal: false
provider: openai

tools:
  - LoadCSV
  - LoadJSON
  - ValidateData
  - HandleMissingValues
  - GetDataInfo

handoff_targets:
  - analyzer

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  When loading data:
  1. Identify data source and format
  2. Load using appropriate tool
  3. Validate data quality
  4. Report loading status
  5. Emit [DATA_READY] signal

  Always ensure data integrity before passing to analyzer.
```

#### analyzer.yaml
```yaml
id: analyzer
name: Data Analyst
role: Statistical Analyst
description: Performs statistical analysis and pattern recognition

backstory: |
  I am the Data Analyst. I specialize in:
  1. Statistical analysis
  2. Pattern detection
  3. Trend identification
  4. Anomaly detection
  5. Correlation analysis

model: gpt-4-turbo
temperature: 0.5
is_terminal: false
provider: openai

tools:
  - ComputeStatistics
  - DetectPatterns
  - IdentifyTrends
  - AnalyzeClusters
  - ComputeCorrelation

handoff_targets:
  - visualizer

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  Analysis procedure:
  1. Examine loaded data
  2. Compute descriptive statistics
  3. Detect patterns and anomalies
  4. Identify correlations
  5. Emit [VISUALIZATION_NEEDED] signal

  Provide detailed analytical insights.
```

#### visualizer.yaml
```yaml
id: visualizer
name: Data Visualizer
role: Visualization Specialist
description: Creates charts and visualizations from analysis results

backstory: |
  I am the Data Visualizer. I create:
  1. Statistical charts
  2. Trend visualizations
  3. Distribution plots
  4. Correlation heatmaps
  5. Interactive dashboards

model: gpt-4-turbo
temperature: 0.6
is_terminal: true
provider: openai

tools:
  - CreateChart
  - CreateHistogram
  - CreateScatterPlot
  - CreateHeatmap
  - CreateDashboard

handoff_targets: []

system_prompt: |
  You are {{name}}.
  Role: {{role}}

  Visualization procedure:
  1. Receive analysis results
  2. Select appropriate chart types
  3. Create professional visualizations
  4. Add titles, legends, annotations
  5. Return visualization artifacts
  6. Emit [VISUALIZATION_COMPLETE] signal

  Ensure visualizations are clear and interpretable.
```

### 3.2 Team Hỗ Trợ Khách Hàng (Customer Support Team)

#### crew.yaml
```yaml
version: "1.0"
name: customer-support-team
description: Automated customer support with ticket routing and resolution

entry_point: ticket-router

agents:
  - ticket-router
  - issue-classifier
  - faq-responder
  - escalation-agent

settings:
  max_handoffs: 8
  max_rounds: 12
  timeout_seconds: 300
  language: en
  organization: Customer-Support

routing:
  signals:
    ticket-router:
      - signal: "[CLASSIFY_ISSUE]"
        target: issue-classifier
        description: "Classify incoming issue"
    issue-classifier:
      - signal: "[SIMPLE_ISSUE]"
        target: faq-responder
        description: "Issue can be handled by FAQ"
      - signal: "[COMPLEX_ISSUE]"
        target: escalation-agent
        description: "Issue needs escalation"
    faq-responder:
      - signal: "[RESOLVED]"
        target: null
        description: "Issue resolved"
    escalation-agent:
      - signal: "[EXPERT_RESOLUTION]"
        target: null
        description: "Expert resolution provided"

  defaults:
    ticket-router: issue-classifier
    issue-classifier: escalation-agent
    faq-responder: null
    escalation-agent: null
```

#### ticket-router.yaml
```yaml
id: ticket-router
name: Support Router
role: Ticket Router
description: Routes customer support tickets to appropriate agents

backstory: |
  I am the Support Router. I:
  1. Receive customer support tickets
  2. Extract key information
  3. Determine initial routing
  4. Ensure tickets don't get lost

model: gpt-4o-mini
temperature: 0.3
is_terminal: false
provider: openai

tools:
  - ExtractTicketInfo
  - ValidateEmail

handoff_targets:
  - issue-classifier

system_prompt: |
  You are {{name}}.
  Extract ticket information and route to classifier.
  Ensure customer details are preserved.
```

#### issue-classifier.yaml
```yaml
id: issue-classifier
name: Issue Classifier
role: Support Classifier
description: Classifies issues into categories

backstory: |
  I am the Issue Classifier. I categorize:
  1. Simple FAQ issues
  2. Complex technical issues
  3. Billing issues
  4. Feature requests

model: gpt-4o
temperature: 0.5
is_terminal: false
provider: openai

tools:
  - ClassifyIssue
  - GetIssueCategory

handoff_targets:
  - faq-responder
  - escalation-agent

system_prompt: |
  Classify the support issue.
  If simple/FAQ → emit [SIMPLE_ISSUE]
  If complex → emit [COMPLEX_ISSUE]
```

---

## 4. Best Practices & Guidelines

### 4.1 Naming Conventions
- **crew name**: lowercase, hyphenated, descriptive (3-50 chars)
- **agent id**: lowercase, hyphenated, unique within crew (3-50 chars)
- **agent name**: readable, can be in any language
- **signal names**: UPPERCASE_WITH_BRACKETS, e.g., `[ROUTE_EXECUTOR]`

### 4.2 Structuring Multi-Agent Teams
1. **Single-Agent Crews**: `is_terminal: true`, no routing
2. **Linear Workflow**: Sequential agent execution with single path
3. **Branching Workflow**: Multiple paths based on conditions (use signals)
4. **Complex Workflow**: Multiple handoffs and feedback loops

### 4.3 Configuration Validation
- All agents in `crew.agents` must have config files
- `entry_point` must exist in agents list
- Signal targets must point to valid agents
- Terminal agents should have empty handoff_targets
- Provider URLs must be accessible

### 4.4 Performance Tuning
- **timeout_seconds**: Increase for complex analysis, decrease for interactive
- **max_handoffs**: Limit circular references, typically 5-10
- **max_rounds**: Enough iterations for thorough analysis but not infinite loops
- **temperature**: Lower for deterministic tasks, higher for creative work

---

## 5. Troubleshooting

### Common Issues

**Problem**: "no entry agent found"
- **Cause**: `entry_point` doesn't match any agent ID in agents list
- **Solution**: Check spelling and ensure agent config file exists

**Problem**: "signal not recognized"
- **Cause**: Signal in response doesn't match routing config
- **Solution**: Verify signal format in `routing.signals` section

**Problem**: "max handoffs exceeded"
- **Cause**: Too many agent-to-agent transfers
- **Solution**: Increase `settings.max_handoffs` or redesign workflow

---

## Kết Luận

Tài liệu này cung cấp đặc tả chi tiết cho cấu hình crew.yaml và agent.yaml. Sử dụng các ví dụ và best practices để xây dựng multi-agent workflows hiệu quả và có thể mở rộng.

Để tìm hiểu thêm, tham khảo:
- `00-hello-crew/` - Single-agent example
- `it-support/` - Multi-agent example với Vietnamese language
- Core library documentation
