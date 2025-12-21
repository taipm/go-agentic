package crewai

import (
	"fmt"
	"strings"
	"time"
)

// HTMLReport generates an HTML test report
type HTMLReport struct {
	TestResults []*TestResult
	GeneratedAt time.Time
	TotalTests  int
	PassedTests int
	FailedTests int
}

// NewHTMLReport creates a new HTML report
func NewHTMLReport(results []*TestResult) *HTMLReport {
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	return &HTMLReport{
		TestResults: results,
		GeneratedAt: time.Now(),
		TotalTests:  len(results),
		PassedTests: passed,
		FailedTests: len(results) - passed,
	}
}

// ToHTML generates the complete HTML report
func (r *HTMLReport) ToHTML() string {
	var sb strings.Builder

	sb.WriteString(r.htmlHeader())
	sb.WriteString(r.htmlSummary())
	sb.WriteString(r.htmlTestDetails())
	sb.WriteString(r.htmlFooter())

	return sb.String()
}

func (r *HTMLReport) htmlHeader() string {
	return `<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>go-crewai Test Report</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
            color: #333;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.2);
        }

        .header p {
            font-size: 1.1em;
            opacity: 0.9;
        }

        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            padding: 40px;
            background: #f8f9fa;
            border-bottom: 2px solid #e9ecef;
        }

        .summary-card {
            background: white;
            padding: 20px;
            border-radius: 6px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            text-align: center;
        }

        .summary-card h3 {
            font-size: 0.9em;
            color: #666;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }

        .summary-card .value {
            font-size: 2.5em;
            font-weight: bold;
            color: #333;
        }

        .summary-card.passed .value {
            color: #28a745;
        }

        .summary-card.failed .value {
            color: #dc3545;
        }

        .summary-card.total .value {
            color: #667eea;
        }

        .timestamp {
            padding: 20px 40px;
            background: white;
            border-bottom: 1px solid #e9ecef;
            font-size: 0.9em;
            color: #666;
        }

        .tests-container {
            padding: 40px;
        }

        .test-case {
            margin-bottom: 30px;
            border: 2px solid #e9ecef;
            border-radius: 6px;
            overflow: hidden;
            transition: all 0.3s ease;
        }

        .test-case:hover {
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
            border-color: #667eea;
        }

        .test-header {
            padding: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: #f8f9fa;
            cursor: pointer;
            transition: background 0.2s ease;
        }

        .test-header:hover {
            background: #e9ecef;
        }

        .test-title {
            flex: 1;
        }

        .test-id {
            font-size: 0.85em;
            color: #999;
            margin-right: 15px;
        }

        .test-name {
            font-size: 1.2em;
            font-weight: bold;
            color: #333;
            margin-bottom: 5px;
        }

        .test-desc {
            font-size: 0.9em;
            color: #666;
        }

        .test-status {
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .status-badge {
            padding: 6px 12px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 0.85em;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .status-badge.passed {
            background: #d4edda;
            color: #155724;
        }

        .status-badge.failed {
            background: #f8d7da;
            color: #721c24;
        }

        .duration {
            font-size: 0.85em;
            color: #999;
            margin-left: 10px;
        }

        .test-body {
            padding: 20px;
            display: none;
            background: white;
            border-top: 1px solid #e9ecef;
        }

        .test-case.expanded .test-body {
            display: block;
        }

        .test-section {
            margin-bottom: 20px;
        }

        .test-section:last-child {
            margin-bottom: 0;
        }

        .section-title {
            font-size: 0.95em;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .assertion-list, .error-list, .flow-list {
            list-style: none;
            padding-left: 0;
        }

        .assertion-list li, .error-list li, .flow-list li {
            padding: 8px 12px;
            margin-bottom: 8px;
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            border-radius: 2px;
            font-size: 0.9em;
        }

        .assertion-list li.passed {
            border-left-color: #28a745;
            background: #f0f9f6;
        }

        .assertion-list li.failed {
            border-left-color: #dc3545;
            background: #fdf7f7;
        }

        .error-list li {
            border-left-color: #dc3545;
            background: #fdf7f7;
            color: #721c24;
        }

        .flow-list li {
            display: inline-block;
            background: white;
            border: 1px solid #667eea;
            border-radius: 4px;
            padding: 8px 12px;
            margin-right: 10px;
            margin-bottom: 8px;
            color: #667eea;
            font-weight: 500;
        }

        .flow-list li::after {
            content: ' →';
            margin-left: 10px;
        }

        .flow-list li:last-child::after {
            content: '';
        }

        .response-box {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 12px;
            margin-top: 10px;
            font-family: 'Monaco', 'Courier New', monospace;
            font-size: 0.85em;
            max-height: 200px;
            overflow-y: auto;
            word-break: break-word;
            white-space: pre-wrap;
        }

        .footer {
            background: #f8f9fa;
            padding: 30px 40px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
            border-top: 1px solid #e9ecef;
        }

        .stats {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin-bottom: 15px;
        }

        .stat-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .stat-item .dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }

        .stat-item.passed .dot {
            background: #28a745;
        }

        .stat-item.failed .dot {
            background: #dc3545;
        }

        .stat-item.total .dot {
            background: #667eea;
        }

        @media (max-width: 768px) {
            .header h1 {
                font-size: 1.8em;
            }

            .test-header {
                flex-direction: column;
                align-items: flex-start;
            }

            .test-status {
                margin-top: 10px;
            }

            .summary {
                grid-template-columns: 1fr;
            }
        }

        .toggle-all {
            margin-bottom: 20px;
            padding: 10px 20px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
            transition: background 0.2s ease;
        }

        .toggle-all:hover {
            background: #5568d3;
        }
    </style>
</head>
<body>
    <div class="container">
`
}

func (r *HTMLReport) htmlSummary() string {
	passRate := 0
	if r.TotalTests > 0 {
		passRate = (r.PassedTests * 100) / r.TotalTests
	}

	return fmt.Sprintf(`
        <div class="header">
            <h1>🧪 go-crewai Test Report</h1>
            <p>Intelligent Agent Routing System Test Suite</p>
        </div>

        <div class="summary">
            <div class="summary-card total">
                <h3>Total Tests</h3>
                <div class="value">%d</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="value">%d</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="value">%d</div>
            </div>
            <div class="summary-card total">
                <h3>Pass Rate</h3>
                <div class="value">%d%%</div>
            </div>
        </div>

        <div class="timestamp">
            <strong>Generated:</strong> %s<br>
            <strong>Test Duration:</strong> %s
        </div>
`,
		r.TotalTests,
		r.PassedTests,
		r.FailedTests,
		passRate,
		r.GeneratedAt.Format("2006-01-02 15:04:05"),
		r.getTotalDuration(),
	)
}

func (r *HTMLReport) htmlTestDetails() string {
	var sb strings.Builder

	sb.WriteString(`
        <div class="tests-container">
            <button class="toggle-all" onclick="toggleAllTests()">📂 Expand/Collapse All Tests</button>

            <div id="tests">
`)

	for _, result := range r.TestResults {
		sb.WriteString(r.htmlTestCase(result))
	}

	sb.WriteString(`
            </div>

            <script>
                function toggleAllTests() {
                    const tests = document.querySelectorAll('.test-case');
                    const allExpanded = Array.from(tests).every(t => t.classList.contains('expanded'));
                    tests.forEach(test => {
                        if (allExpanded) {
                            test.classList.remove('expanded');
                        } else {
                            test.classList.add('expanded');
                        }
                    });
                }

                // Make test headers clickable
                document.querySelectorAll('.test-header').forEach(header => {
                    header.addEventListener('click', function() {
                        this.closest('.test-case').classList.toggle('expanded');
                    });
                });
            </script>
        </div>
`)

	return sb.String()
}

func (r *HTMLReport) htmlTestCase(result *TestResult) string {
	statusClass := "passed"
	statusText := "✓ PASSED"
	if !result.Passed {
		statusClass = "failed"
		statusText = "✗ FAILED"
	}

	durationMs := result.Duration.Milliseconds()

	// Build agent response section
	agentResponseHTML := ""
	if result.Response != nil {
		agentResponseHTML = fmt.Sprintf(`
                    <!-- Agent Response -->
                    <div class="test-section">
                        <div class="section-title">🤖 Agent Response (%s - %s)</div>
                        <div class="response-box">%s</div>
                    </div>
`, result.Response.AgentName, result.Response.AgentID, escapeHTML(result.Response.Content))
	}

	html := fmt.Sprintf(`
            <div class="test-case">
                <div class="test-header">
                    <div class="test-title">
                        <div class="test-id">Scenario %s</div>
                        <div class="test-name">%s</div>
                        <div class="test-desc">%s</div>
                    </div>
                    <div class="test-status">
                        <span class="status-badge %s">%s</span>
                        <span class="duration">%dms</span>
                    </div>
                </div>

                <div class="test-body">
                    <!-- User Input -->
                    <div class="test-section">
                        <div class="section-title">📝 User Input</div>
                        <div class="response-box">%s</div>
                    </div>

                    <!-- Expected vs Actual Flow -->
                    <div class="test-section">
                        <div class="section-title">🔄 Agent Flow</div>
                        <div style="margin-bottom: 10px;">
                            <strong style="color: #667eea;">Expected:</strong>
                            <ul class="flow-list">
                                %s
                            </ul>
                        </div>
                        <div>
                            <strong style="color: #667eea;">Actual:</strong>
                            <ul class="flow-list">
                                %s
                            </ul>
                        </div>
                    </div>

                    %s
`,
		result.Scenario.ID,
		result.Scenario.Name,
		escapeHTML(result.Scenario.Description),
		statusClass,
		statusText,
		durationMs,
		escapeHTML(result.Scenario.UserInput),
		r.htmlFlowList(result.Scenario.ExpectedFlow),
		r.htmlFlowList(result.ActualFlow),
		agentResponseHTML,
	)
	html += `</div>`

	// Assertions
	if len(result.Scenario.Assertions) > 0 {
		html += `
                    <!-- Assertions -->
                    <div class="test-section">
                        <div class="section-title">✓ Assertions</div>
                        <ul class="assertion-list">
`
		for _, assertion := range result.Scenario.Assertions {
			html += fmt.Sprintf(`                            <li class="passed">✓ %s</li>`, escapeHTML(assertion))
		}
		html += `
                        </ul>
                    </div>
`
	}

	// Errors
	if len(result.Errors) > 0 {
		html += `
                    <!-- Errors -->
                    <div class="test-section">
                        <div class="section-title">❌ Errors</div>
                        <ul class="error-list">
`
		for _, errMsg := range result.Errors {
			html += fmt.Sprintf(`                            <li>%s</li>`, escapeHTML(errMsg))
		}
		html += `
                        </ul>
                    </div>
`
	}

	// Warnings
	if len(result.Warnings) > 0 {
		html += `
                    <!-- Warnings -->
                    <div class="test-section">
                        <div class="section-title">⚠️ Warnings</div>
                        <ul class="error-list">
`
		for _, warn := range result.Warnings {
			html += fmt.Sprintf(`                            <li style="color: #856404; border-left-color: #ffc107;">%s</li>`, escapeHTML(warn))
		}
		html += `
                        </ul>
                    </div>
`
	}

	html += `
                </div>
            </div>
`

	return html
}

func (r *HTMLReport) htmlFlowList(flow []string) string {
	if len(flow) == 0 {
		return "<li>None</li>"
	}

	var sb strings.Builder
	for _, agent := range flow {
		sb.WriteString(fmt.Sprintf("<li>%s</li>", escapeHTML(agent)))
	}
	return sb.String()
}

func (r *HTMLReport) htmlFooter() string {
	stats := fmt.Sprintf(`
            <div class="stats">
                <div class="stat-item passed">
                    <div class="dot"></div>
                    <span>%d Passed Tests</span>
                </div>
                <div class="stat-item failed">
                    <div class="dot"></div>
                    <span>%d Failed Tests</span>
                </div>
                <div class="stat-item total">
                    <div class="dot"></div>
                    <span>%d Total Tests</span>
                </div>
            </div>
`,
		r.PassedTests,
		r.FailedTests,
		r.TotalTests,
	)

	return fmt.Sprintf(`
        <div class="footer">
            %s
            <p style="margin-top: 15px;">
                <strong>go-crewai Intelligent Agent Routing System</strong><br>
                Tested on: %s<br>
                <small>© 2025 Go CrewAI - Smart Multi-Agent IT Support System</small>
            </p>
        </div>
    </div>
</body>
</html>
`,
		stats,
		time.Now().Format("2006-01-02 15:04:05"),
	)
}

func (r *HTMLReport) getTotalDuration() string {
	var total time.Duration
	for _, result := range r.TestResults {
		total += result.Duration
	}
	return fmt.Sprintf("%dms", total.Milliseconds())
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
