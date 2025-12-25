package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

// ScanReport structures match those from scan.go
type FunctionInfo struct {
	Name           string   `json:"name"`
	Receiver       string   `json:"receiver,omitempty"`
	IsMethod       bool     `json:"is_method"`
	IsTest         bool     `json:"is_test"`
	FuncType       string   `json:"func_type"`
	FuncTypeReason string   `json:"func_type_reason,omitempty"`
	ReturnsFunc    bool     `json:"returns_func"`
	AcceptsFunc    bool     `json:"accepts_func"`
	StartLine      int      `json:"start_line"`
	EndLine        int      `json:"end_line"`
	TotalLines     int      `json:"total_lines"`
	Parameters     int      `json:"parameters"`
	Returns        int      `json:"returns"`
	Comments       []string `json:"comments"`
	CommentCount   int      `json:"comment_count"`
}

// FunctionTypeStats contains statistics for function types
type FunctionTypeStats struct {
	Constructors int `json:"constructors"`
	Getters      int `json:"getters"`
	Setters      int `json:"setters"`
	Handlers     int `json:"handlers"`
	Middlewares  int `json:"middlewares"`
	HigherOrder  int `json:"higher_order"`
	Validators   int `json:"validators"`
	Converters   int `json:"converters"`
	Inits        int `json:"inits"`
	Callbacks    int `json:"callbacks"`
	Tests        int `json:"tests"`
	Helpers      int `json:"helpers"`
	Lifecycle    int `json:"lifecycle"`
	Standard     int `json:"standard"`
}

type FileInfo struct {
	FilePath  string         `json:"file_path"`
	RelPath   string         `json:"relative_path"`
	Package   string         `json:"package"`
	Functions []FunctionInfo `json:"functions"`
	Count     int            `json:"function_count"`
	TestCount int            `json:"test_count"`
}

type PackageStats struct {
	Name      string `json:"name"`
	Files     int    `json:"files"`
	Functions int    `json:"functions"`
	Methods   int    `json:"methods"`
	Tests     int    `json:"tests"`
}

// DirectoryNode represents a node in the directory tree
type DirectoryNode struct {
	Name      string           `json:"name"`
	Path      string           `json:"path"`
	IsDir     bool             `json:"is_dir"`
	Children  []*DirectoryNode `json:"children,omitempty"`
	FileCount int              `json:"file_count,omitempty"`
	FuncCount int              `json:"func_count,omitempty"`
	Package   string           `json:"package,omitempty"`
}

type ScanReport struct {
	RootPath          string            `json:"root_path"`
	TotalFiles        int               `json:"total_files"`
	TotalFunctions    int               `json:"total_functions"`
	TotalMethods      int               `json:"total_methods"`
	TotalTests        int               `json:"total_tests"`
	FunctionTypeStats FunctionTypeStats `json:"function_type_stats"`
	DirectoryTree     *DirectoryNode    `json:"directory_tree"`
	Files             []FileInfo        `json:"files"`
	Packages          []PackageStats    `json:"packages"`
	Summary           map[string]int    `json:"summary"`
}

func main() {
	inputFile := flag.String("input", "full_project_report.json", "Input JSON report file")
	outputFile := flag.String("output", "report.html", "Output HTML file")
	flag.Parse()

	// Read JSON report
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	var report ScanReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate HTML
	html := generateHTML(&report)

	// Write to file
	err = os.WriteFile(*outputFile, []byte(html), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ HTML report generated: %s\n", *outputFile)
}

// getFuncTypeBadge returns HTML badge for function type
func getFuncTypeBadge(funcType string) string {
	badgeClass := "badge-standard"
	icon := "📝"
	label := funcType

	switch funcType {
	case "constructor":
		badgeClass = "badge-constructor"
		icon = "📦"
		label = "Constructor"
	case "getter":
		badgeClass = "badge-getter"
		icon = "🔧"
		label = "Getter"
	case "setter":
		badgeClass = "badge-setter"
		icon = "✏️"
		label = "Setter"
	case "handler":
		badgeClass = "badge-handler"
		icon = "🎯"
		label = "Handler"
	case "middleware":
		badgeClass = "badge-middleware"
		icon = "🔗"
		label = "Middleware"
	case "higher_order":
		badgeClass = "badge-higher-order"
		icon = "⚡"
		label = "Higher-Order"
	case "validator":
		badgeClass = "badge-validator"
		icon = "✅"
		label = "Validator"
	case "converter":
		badgeClass = "badge-converter"
		icon = "🔄"
		label = "Converter"
	case "init":
		badgeClass = "badge-constructor"
		icon = "🚀"
		label = "Init"
	case "callback":
		badgeClass = "badge-callback"
		icon = "📞"
		label = "Callback"
	case "helper":
		badgeClass = "badge-helper"
		icon = "🔨"
		label = "Helper"
	case "lifecycle":
		badgeClass = "badge-lifecycle"
		icon = "♻️"
		label = "Lifecycle"
	case "test":
		badgeClass = "badge-test"
		icon = "🧪"
		label = "Test"
	case "standard":
		badgeClass = "badge-standard"
		icon = "📝"
		label = "Standard"
	}

	return fmt.Sprintf(`<span class="badge %s">%s %s</span>`, badgeClass, icon, label)
}

// generateTreeHTML generates HTML for the directory tree
func generateTreeHTML(node *DirectoryNode, depth int) string {
	if node == nil {
		return ""
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "    "
	}

	html := ""

	if node.IsDir {
		icon := "📁"
		if depth == 0 {
			icon = "🏠"
		}
		stats := ""
		if node.FileCount > 0 {
			stats = fmt.Sprintf(`<span class="tree-stats">(%d files, %d functions)</span>`, node.FileCount, node.FuncCount)
		}
		html += fmt.Sprintf(`%s<div class="tree-node"><span class="tree-icon">%s</span><span class="tree-dir">%s/</span>%s</div>`, indent, icon, node.Name, stats)

		if len(node.Children) > 0 {
			html += fmt.Sprintf(`%s<div class="tree-children">`, indent)
			for _, child := range node.Children {
				html += "\n" + generateTreeHTML(child, depth+1)
			}
			html += fmt.Sprintf(`\n%s</div>`, indent)
		}
	} else {
		icon := "📄"
		fileClass := "tree-file"
		if len(node.Name) > 3 && node.Name[len(node.Name)-3:] == ".go" {
			icon = "🐹"
			fileClass = "tree-file-go"
		}
		stats := ""
		if node.FuncCount > 0 {
			stats = fmt.Sprintf(`<span class="tree-stats">(%d funcs)</span>`, node.FuncCount)
		}
		pkgInfo := ""
		if node.Package != "" {
			pkgInfo = fmt.Sprintf(`<span class="tree-stats">[%s]</span>`, node.Package)
		}
		html += fmt.Sprintf(`%s<div class="tree-node"><span class="tree-icon">%s</span><span class="%s">%s</span>%s%s</div>`, indent, icon, fileClass, node.Name, stats, pkgInfo)
	}

	return html
}

func generateHTML(report *ScanReport) string {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Code Scan Report</title>
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
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
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
        }
        
        .header p {
            font-size: 1.1em;
            opacity: 0.9;
        }
        
        .content {
            padding: 40px;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        
        .stat-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
        }
        
        .stat-card h3 {
            font-size: 0.9em;
            opacity: 0.9;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .stat-card .number {
            font-size: 2.5em;
            font-weight: bold;
        }
        
        .section {
            margin-bottom: 40px;
        }
        
        .section h2 {
            color: #333;
            padding-bottom: 15px;
            border-bottom: 2px solid #667eea;
            margin-bottom: 20px;
            font-size: 1.8em;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        
        th {
            background: #667eea;
            color: white;
            padding: 15px;
            text-align: left;
            font-weight: 600;
        }
        
        td {
            padding: 12px 15px;
            border-bottom: 1px solid #eee;
        }
        
        tr:hover {
            background: #f9f9f9;
        }
        
        .badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.85em;
            font-weight: 600;
        }
        
        .badge-method {
            background: #e3f2fd;
            color: #1976d2;
        }
        
        .badge-test {
            background: #f3e5f5;
            color: #7b1fa2;
        }
        
        .badge-constructor {
            background: #e8f5e9;
            color: #2e7d32;
        }
        
        .badge-getter {
            background: #e3f2fd;
            color: #1565c0;
        }
        
        .badge-setter {
            background: #fff3e0;
            color: #e65100;
        }
        
        .badge-handler {
            background: #fce4ec;
            color: #c2185b;
        }
        
        .badge-middleware {
            background: #ede7f6;
            color: #5e35b1;
        }
        
        .badge-higher-order {
            background: #f3e5f5;
            color: #8e24aa;
        }
        
        .badge-validator {
            background: #e8eaf6;
            color: #3949ab;
        }
        
        .badge-converter {
            background: #e0f7fa;
            color: #00838f;
        }
        
        .badge-callback {
            background: #fff8e1;
            color: #ff8f00;
        }
        
        .badge-helper {
            background: #f5f5f5;
            color: #616161;
        }
        
        .badge-lifecycle {
            background: #efebe9;
            color: #5d4037;
        }
        
        .badge-standard {
            background: #eceff1;
            color: #455a64;
        }
        
        .type-stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 12px;
            margin-bottom: 40px;
        }
        
        .type-stat-card {
            background: #f8f9fa;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            padding: 15px;
            text-align: center;
        }
        
        .type-stat-card .icon {
            font-size: 1.5em;
            margin-bottom: 5px;
        }
        
        .type-stat-card .label {
            font-size: 0.85em;
            color: #666;
            margin-bottom: 5px;
        }
        
        .type-stat-card .count {
            font-size: 1.8em;
            font-weight: bold;
            color: #333;
        }
        
        /* Directory Tree Styles */
        .tree-container {
            background: #1e1e1e;
            border-radius: 8px;
            padding: 20px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 14px;
            overflow-x: auto;
            max-height: 600px;
            overflow-y: auto;
        }
        
        .tree-node {
            color: #d4d4d4;
            line-height: 1.6;
        }
        
        .tree-dir {
            color: #4fc3f7;
            cursor: pointer;
        }
        
        .tree-dir:hover {
            color: #81d4fa;
        }
        
        .tree-file {
            color: #a5d6a7;
        }
        
        .tree-file-go {
            color: #00bcd4;
        }
        
        .tree-stats {
            color: #9e9e9e;
            font-size: 0.85em;
            margin-left: 8px;
        }
        
        .tree-icon {
            margin-right: 6px;
        }
        
        .tree-children {
            margin-left: 20px;
            border-left: 1px dashed #444;
            padding-left: 10px;
        }
        
        .tree-toggle {
            cursor: pointer;
            user-select: none;
            display: inline-block;
            width: 16px;
        }
        
        .file-list {
            display: grid;
            gap: 15px;
        }
        
        .file-item {
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 8px;
            background: #fafafa;
        }
        
        .file-item h4 {
            color: #667eea;
            margin-bottom: 10px;
            font-size: 1.1em;
        }
        
        .file-item-meta {
            display: flex;
            gap: 20px;
            flex-wrap: wrap;
            font-size: 0.9em;
            color: #666;
        }
        
        .footer {
            background: #f5f5f5;
            padding: 20px 40px;
            text-align: center;
            color: #666;
            font-size: 0.9em;
            border-top: 1px solid #ddd;
        }
        
        .functions-list {
            list-style: none;
            padding: 0;
            margin: 10px 0 0 0;
        }
        
        .functions-list li {
            padding: 8px 0;
            padding-left: 20px;
            color: #555;
            font-size: 0.95em;
            border-left: 3px solid #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Go Code Scan Report</h1>
            <p>Comprehensive function and method analysis</p>
        </div>
        
        <div class="content">
            <div class="stats-grid">
                <div class="stat-card">
                    <h3>Total Files</h3>
                    <div class="number">` + fmt.Sprintf("%d", report.TotalFiles) + `</div>
                </div>
                <div class="stat-card">
                    <h3>Total Functions</h3>
                    <div class="number">` + fmt.Sprintf("%d", report.TotalFunctions) + `</div>
                </div>
                <div class="stat-card">
                    <h3>Total Methods</h3>
                    <div class="number">` + fmt.Sprintf("%d", report.TotalMethods) + `</div>
                </div>
                <div class="stat-card">
                    <h3>Total Tests</h3>
                    <div class="number">` + fmt.Sprintf("%d", report.TotalTests) + `</div>
                </div>
            </div>
            
            <div class="section">
                <h2>🏷️ Function Type Classification</h2>
                <div class="type-stats-grid">
                    <div class="type-stat-card">
                        <div class="icon">📦</div>
                        <div class="label">Constructors</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Constructors) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🔧</div>
                        <div class="label">Getters</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Getters) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">✏️</div>
                        <div class="label">Setters</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Setters) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🎯</div>
                        <div class="label">Handlers</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Handlers) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🔗</div>
                        <div class="label">Middlewares</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Middlewares) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">⚡</div>
                        <div class="label">Higher-Order</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.HigherOrder) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">✅</div>
                        <div class="label">Validators</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Validators) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🔄</div>
                        <div class="label">Converters</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Converters) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🚀</div>
                        <div class="label">Inits</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Inits) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">📞</div>
                        <div class="label">Callbacks</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Callbacks) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🧪</div>
                        <div class="label">Tests</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Tests) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">🔨</div>
                        <div class="label">Helpers</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Helpers) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">♻️</div>
                        <div class="label">Lifecycle</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Lifecycle) + `</div>
                    </div>
                    <div class="type-stat-card">
                        <div class="icon">📝</div>
                        <div class="label">Standard</div>
                        <div class="count">` + fmt.Sprintf("%d", report.FunctionTypeStats.Standard) + `</div>
                    </div>
                </div>
            </div>
            
            <div class="section">
                <h2>🌳 Directory Structure</h2>
                <div class="tree-container">
` + generateTreeHTML(report.DirectoryTree, 0) + `
                </div>
            </div>
            
            <div class="section">
                <h2>📦 Packages Overview</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Package</th>
                            <th>Files</th>
                            <th>Functions</th>
                            <th>Methods</th>
                            <th>Tests</th>
                        </tr>
                    </thead>
                    <tbody>
`

	// Sort packages by functions
	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].Functions > report.Packages[j].Functions
	})

	for _, pkg := range report.Packages {
		html += fmt.Sprintf(`                        <tr>
                            <td><strong>%s</strong></td>
                            <td>%d</td>
                            <td>%d</td>
                            <td>%d</td>
                            <td><span class="badge badge-test">%d</span></td>
                        </tr>
`, pkg.Name, pkg.Files, pkg.Functions, pkg.Methods, pkg.Tests)
	}

	html += `                    </tbody>
                </table>
            </div>
            
            <div class="section">
                <h2>📄 Top 20 Files by Function Count</h2>
                <table>
                    <thead>
                        <tr>
                            <th>File</th>
                            <th>Package</th>
                            <th>Functions</th>
                            <th>Tests</th>
                        </tr>
                    </thead>
                    <tbody>
`

	// Sort files by count
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Count > report.Files[j].Count
	})

	count := 20
	if len(report.Files) < 20 {
		count = len(report.Files)
	}

	for i := 0; i < count; i++ {
		file := report.Files[i]
		html += fmt.Sprintf(`                        <tr>
                            <td><strong>%s</strong></td>
                            <td>%s</td>
                            <td>%d</td>
                            <td><span class="badge badge-test">%d</span></td>
                        </tr>
`, file.RelPath, file.Package, file.Count, file.TestCount)
	}

	html += `                    </tbody>
                </table>
            </div>
            
            <div class="section">
                <h2>🔍 Files with Details</h2>
                <div class="file-list">
`

	// Show details for top 10 files
	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Count > report.Files[j].Count
	})

	for i := 0; i < 10 && i < len(report.Files); i++ {
		file := report.Files[i]
		html += `                    <div class="file-item">
                        <h4>` + file.RelPath + `</h4>
                        <div class="file-item-meta">
                            <span>Package: <strong>` + file.Package + `</strong></span>
                            <span>Functions: <strong>` + fmt.Sprintf("%d", file.Count) + `</strong></span>
                            <span>Tests: <strong>` + fmt.Sprintf("%d", file.TestCount) + `</strong></span>
                        </div>
`

		if len(file.Functions) > 0 {
			html += `                        <ul class="functions-list">`
			for _, fn := range file.Functions {
				if !fn.IsTest {
					receiver := ""
					if fn.IsMethod {
						receiver = " (" + fn.Receiver + ")"
					}
					// Function type badge
					funcTypeBadge := getFuncTypeBadge(fn.FuncType)
					commentsHTML := ""
					if fn.CommentCount > 0 {
						commentsHTML = fmt.Sprintf(`<br><em style="color:#999; font-size:0.9em;">📝 %d comments | %d lines</em>`, fn.CommentCount, fn.TotalLines)
						if len(fn.Comments) > 0 && len(fn.Comments) <= 3 {
							commentsHTML += `<br><span style="color:#666; font-size:0.85em; margin-top:4px; display:block;">`
							for idx, comment := range fn.Comments {
								if idx < 2 {
									commentsHTML += fmt.Sprintf(`<em>• %s</em><br>`, comment)
								}
							}
							commentsHTML += `</span>`
						}
					}
					html += fmt.Sprintf(`
                            <li>%s%s %s <span class="badge badge-method">%d params → %d returns | %d lines</span>%s</li>`, fn.Name, receiver, funcTypeBadge, fn.Parameters, fn.Returns, fn.TotalLines, commentsHTML)
				}
			}
			html += `
                        </ul>`
		}

		html += `
                    </div>
`
	}

	html += `                </div>
            </div>
        </div>
        
        <div class="footer">
            <p>Generated by Go Code Scanner | ` + fmt.Sprintf("%d files analyzed | %d functions discovered", report.TotalFiles, report.TotalFunctions) + `</p>
        </div>
    </div>
</body>
</html>`

	return html
}
