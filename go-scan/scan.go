package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Function Type Constants
const (
	FuncTypeConstructor = "constructor"
	FuncTypeGetter      = "getter"
	FuncTypeSetter      = "setter"
	FuncTypeHandler     = "handler"
	FuncTypeMiddleware  = "middleware"
	FuncTypeHigherOrder = "higher_order"
	FuncTypeValidator   = "validator"
	FuncTypeConverter   = "converter"
	FuncTypeInit        = "init"
	FuncTypeCallback    = "callback"
	FuncTypeTest        = "test"
	FuncTypeHelper      = "helper"
	FuncTypeLifecycle   = "lifecycle"
	FuncTypeStandard    = "standard"
)

// FunctionInfo holds information about a function
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

// FileInfo holds information about a Go file
type FileInfo struct {
	FilePath  string         `json:"file_path"`
	RelPath   string         `json:"relative_path"`
	Package   string         `json:"package"`
	Functions []FunctionInfo `json:"functions"`
	Count     int            `json:"function_count"`
	TestCount int            `json:"test_count"`
}

// PackageStats holds package statistics
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

// ScanReport holds the complete scan report
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
	pathFlag := flag.String("path", ".", "Path to scan")
	outputFlag := flag.String("output", "full_project_report.json", "Output file")
	flag.Parse()

	rootPath, err := filepath.Abs(*pathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	report := scanDirectory(rootPath)

	// Write JSON output
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFlag, data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Scan complete: %d files, %d functions, %d methods, %d tests\n",
		report.TotalFiles, report.TotalFunctions, report.TotalMethods, report.TotalTests)
	fmt.Printf("✓ Report saved: %s\n", *outputFlag)
}

func scanDirectory(rootPath string) *ScanReport {
	report := &ScanReport{
		RootPath: rootPath,
		Files:    []FileInfo{},
		Packages: []PackageStats{},
		Summary:  make(map[string]int),
	}

	packageMap := make(map[string]*PackageStats)

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and common non-source directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)

		fileInfo := parseGoFile(path, relPath)
		if fileInfo != nil {
			report.Files = append(report.Files, *fileInfo)
			report.TotalFiles++

			// Update package stats
			pkgName := fileInfo.Package
			if _, exists := packageMap[pkgName]; !exists {
				packageMap[pkgName] = &PackageStats{Name: pkgName}
			}
			pkg := packageMap[pkgName]
			pkg.Files++

			for _, fn := range fileInfo.Functions {
				report.TotalFunctions++
				pkg.Functions++

				if fn.IsMethod {
					report.TotalMethods++
					pkg.Methods++
				}

				if fn.IsTest {
					report.TotalTests++
					pkg.Tests++
				}

				// Update function type stats
				updateFunctionTypeStats(&report.FunctionTypeStats, fn.FuncType)
			}
		}

		return nil
	})

	// Convert package map to slice
	for _, pkg := range packageMap {
		report.Packages = append(report.Packages, *pkg)
	}

	// Sort packages by function count
	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].Functions > report.Packages[j].Functions
	})

	// Build directory tree
	report.DirectoryTree = buildDirectoryTree(rootPath, report.Files)

	return report
}

func updateFunctionTypeStats(stats *FunctionTypeStats, funcType string) {
	switch funcType {
	case FuncTypeConstructor:
		stats.Constructors++
	case FuncTypeGetter:
		stats.Getters++
	case FuncTypeSetter:
		stats.Setters++
	case FuncTypeHandler:
		stats.Handlers++
	case FuncTypeMiddleware:
		stats.Middlewares++
	case FuncTypeHigherOrder:
		stats.HigherOrder++
	case FuncTypeValidator:
		stats.Validators++
	case FuncTypeConverter:
		stats.Converters++
	case FuncTypeInit:
		stats.Inits++
	case FuncTypeCallback:
		stats.Callbacks++
	case FuncTypeTest:
		stats.Tests++
	case FuncTypeHelper:
		stats.Helpers++
	case FuncTypeLifecycle:
		stats.Lifecycle++
	case FuncTypeStandard:
		stats.Standard++
	}
}

func parseGoFile(filePath, relPath string) *FileInfo {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	fileInfo := &FileInfo{
		FilePath:  filePath,
		RelPath:   relPath,
		Package:   node.Name.Name,
		Functions: []FunctionInfo{},
	}

	// Create comment map
	commentMap := ast.NewCommentMap(fset, node, node.Comments)

	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcInfo := extractFunctionInfo(fn, fset, commentMap)
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
			fileInfo.Count++

			if funcInfo.IsTest {
				fileInfo.TestCount++
			}
		}
	}

	return fileInfo
}

func extractFunctionInfo(fn *ast.FuncDecl, fset *token.FileSet, commentMap ast.CommentMap) FunctionInfo {
	info := FunctionInfo{
		Name:      fn.Name.Name,
		StartLine: fset.Position(fn.Pos()).Line,
		EndLine:   fset.Position(fn.End()).Line,
	}

	info.TotalLines = info.EndLine - info.StartLine + 1

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		info.IsMethod = true
		info.Receiver = extractReceiverType(fn.Recv.List[0].Type)
	}

	// Check if it's a test
	info.IsTest = strings.HasPrefix(fn.Name.Name, "Test") ||
		strings.HasPrefix(fn.Name.Name, "Benchmark") ||
		strings.HasPrefix(fn.Name.Name, "Example")

	// Count parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if len(field.Names) == 0 {
				info.Parameters++
			} else {
				info.Parameters += len(field.Names)
			}
		}
	}

	// Count returns
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				info.Returns++
			} else {
				info.Returns += len(field.Names)
			}
		}
	}

	// Check if function returns a function
	info.ReturnsFunc = returnsFunctionType(fn.Type.Results)

	// Check if function accepts a function
	info.AcceptsFunc = acceptsFunctionType(fn.Type.Params)

	// Extract comments
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			text := strings.TrimPrefix(comment.Text, "//")
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)
			if text != "" {
				info.Comments = append(info.Comments, text)
			}
		}
	}
	info.CommentCount = len(info.Comments)

	// Classify function type
	info.FuncType, info.FuncTypeReason = classifyFunctionType(fn, info)

	return info
}

func extractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + extractReceiverType(t.X)
	case *ast.IndexExpr:
		return extractReceiverType(t.X)
	case *ast.IndexListExpr:
		return extractReceiverType(t.X)
	}
	return ""
}

// returnsFunctionType checks if function returns a function type
func returnsFunctionType(results *ast.FieldList) bool {
	if results == nil {
		return false
	}
	for _, field := range results.List {
		if isFuncType(field.Type) {
			return true
		}
	}
	return false
}

// acceptsFunctionType checks if function accepts a function as parameter
func acceptsFunctionType(params *ast.FieldList) bool {
	if params == nil {
		return false
	}
	for _, field := range params.List {
		if isFuncType(field.Type) {
			return true
		}
	}
	return false
}

// isFuncType checks if an expression is a function type
func isFuncType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.FuncType:
		return true
	case *ast.Ident:
		name := strings.ToLower(t.Name)
		return strings.Contains(name, "func") ||
			strings.Contains(name, "handler") ||
			strings.Contains(name, "callback")
	case *ast.SelectorExpr:
		return isFuncType(t.Sel)
	}
	return false
}

// classifyFunctionType determines the function type based on naming and structure
func classifyFunctionType(fn *ast.FuncDecl, info FunctionInfo) (string, string) {
	name := fn.Name.Name
	nameLower := strings.ToLower(name)

	// Test functions
	if info.IsTest {
		return FuncTypeTest, "Function name starts with Test/Benchmark/Example"
	}

	// Init function
	if name == "init" {
		return FuncTypeInit, "Standard Go init function"
	}

	// Higher-order functions (returns or accepts function)
	if info.ReturnsFunc && info.AcceptsFunc {
		return FuncTypeMiddleware, "Accepts and returns function (middleware pattern)"
	}
	if info.ReturnsFunc {
		if strings.HasPrefix(name, "With") ||
			strings.Contains(nameLower, "middleware") ||
			strings.Contains(nameLower, "wrapper") ||
			strings.Contains(nameLower, "decorator") {
			return FuncTypeMiddleware, "Returns function with middleware naming pattern"
		}
		return FuncTypeHigherOrder, "Returns a function"
	}
	if info.AcceptsFunc {
		if strings.Contains(nameLower, "callback") ||
			strings.Contains(nameLower, "notify") ||
			strings.Contains(nameLower, "subscribe") {
			return FuncTypeCallback, "Accepts function callback"
		}
		return FuncTypeHigherOrder, "Accepts a function parameter"
	}

	// Constructor patterns
	if strings.HasPrefix(name, "New") ||
		strings.HasPrefix(name, "Create") ||
		strings.HasPrefix(name, "Make") ||
		strings.HasPrefix(name, "Build") {
		return FuncTypeConstructor, "Constructor naming pattern (New/Create/Make/Build)"
	}

	// Getter patterns
	if strings.HasPrefix(name, "Get") ||
		strings.HasPrefix(name, "Fetch") ||
		strings.HasPrefix(name, "Load") ||
		strings.HasPrefix(name, "Read") ||
		strings.HasPrefix(name, "Find") ||
		strings.HasPrefix(name, "Query") ||
		strings.HasPrefix(name, "List") {
		return FuncTypeGetter, "Getter naming pattern (Get/Fetch/Load/Read/Find)"
	}

	// Setter patterns
	if strings.HasPrefix(name, "Set") ||
		strings.HasPrefix(name, "Update") ||
		strings.HasPrefix(name, "Save") ||
		strings.HasPrefix(name, "Store") ||
		strings.HasPrefix(name, "Write") ||
		strings.HasPrefix(name, "Put") {
		return FuncTypeSetter, "Setter naming pattern (Set/Update/Save/Store/Write)"
	}

	// Handler patterns
	if strings.HasSuffix(name, "Handler") ||
		strings.HasSuffix(name, "Handle") ||
		strings.HasPrefix(name, "Handle") ||
		strings.HasPrefix(name, "On") ||
		strings.HasSuffix(name, "Controller") {
		return FuncTypeHandler, "Handler naming pattern"
	}

	// Validator patterns
	if strings.HasPrefix(name, "Validate") ||
		strings.HasPrefix(name, "Check") ||
		strings.HasPrefix(name, "Verify") ||
		strings.HasPrefix(name, "Is") ||
		strings.HasPrefix(name, "Has") ||
		strings.HasPrefix(name, "Can") ||
		strings.HasPrefix(name, "Should") {
		return FuncTypeValidator, "Validator naming pattern (Validate/Check/Is/Has/Can)"
	}

	// Converter patterns
	if strings.HasPrefix(name, "To") ||
		strings.HasPrefix(name, "From") ||
		strings.HasPrefix(name, "Parse") ||
		strings.HasPrefix(name, "Format") ||
		strings.HasPrefix(name, "Convert") ||
		strings.HasPrefix(name, "Transform") ||
		strings.HasPrefix(name, "Marshal") ||
		strings.HasPrefix(name, "Unmarshal") ||
		strings.HasPrefix(name, "Encode") ||
		strings.HasPrefix(name, "Decode") {
		return FuncTypeConverter, "Converter naming pattern (To/From/Parse/Format/Convert)"
	}

	// Lifecycle patterns
	if strings.HasPrefix(name, "Start") ||
		strings.HasPrefix(name, "Stop") ||
		strings.HasPrefix(name, "Init") ||
		strings.HasPrefix(name, "Close") ||
		strings.HasPrefix(name, "Shutdown") ||
		strings.HasPrefix(name, "Setup") ||
		strings.HasPrefix(name, "Teardown") ||
		strings.HasPrefix(name, "Cleanup") ||
		strings.HasPrefix(name, "Reset") ||
		strings.HasPrefix(name, "Open") {
		return FuncTypeLifecycle, "Lifecycle naming pattern (Start/Stop/Init/Close/Setup)"
	}

	// Helper patterns (lowercase first letter for unexported)
	if len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
		if strings.Contains(nameLower, "helper") ||
			strings.Contains(nameLower, "util") ||
			strings.Contains(nameLower, "internal") {
			return FuncTypeHelper, "Helper function (unexported with helper naming)"
		}
	}

	// Check for do/run/execute/process patterns
	if strings.HasPrefix(name, "Do") ||
		strings.HasPrefix(name, "Run") ||
		strings.HasPrefix(name, "Execute") ||
		strings.HasPrefix(name, "Process") ||
		strings.HasPrefix(name, "Apply") {
		return FuncTypeHandler, "Action handler pattern (Do/Run/Execute/Process)"
	}

	return FuncTypeStandard, "Standard function"
}

// buildDirectoryTree creates a tree structure from scanned files
func buildDirectoryTree(rootPath string, files []FileInfo) *DirectoryNode {
	root := &DirectoryNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    true,
		Children: []*DirectoryNode{},
	}

	dirMap := make(map[string]*DirectoryNode)
	dirMap["."] = root

	for _, file := range files {
		parts := strings.Split(file.RelPath, string(os.PathSeparator))

		currentPath := "."
		parentNode := root

		for i, part := range parts {
			if i == len(parts)-1 {
				fileNode := &DirectoryNode{
					Name:      part,
					Path:      file.RelPath,
					IsDir:     false,
					FuncCount: file.Count,
					Package:   file.Package,
				}
				parentNode.Children = append(parentNode.Children, fileNode)
			} else {
				if currentPath == "." {
					currentPath = part
				} else {
					currentPath = currentPath + string(os.PathSeparator) + part
				}

				if _, exists := dirMap[currentPath]; !exists {
					dirNode := &DirectoryNode{
						Name:     part,
						Path:     currentPath,
						IsDir:    true,
						Children: []*DirectoryNode{},
					}
					dirMap[currentPath] = dirNode
					parentNode.Children = append(parentNode.Children, dirNode)
				}
				parentNode = dirMap[currentPath]
			}
		}
	}

	calculateDirStats(root)
	sortDirectoryChildren(root)

	return root
}

func calculateDirStats(node *DirectoryNode) (int, int) {
	if !node.IsDir {
		return 1, node.FuncCount
	}

	totalFiles := 0
	totalFuncs := 0

	for _, child := range node.Children {
		files, funcs := calculateDirStats(child)
		totalFiles += files
		totalFuncs += funcs
	}

	node.FileCount = totalFiles
	node.FuncCount = totalFuncs

	return totalFiles, totalFuncs
}

func sortDirectoryChildren(node *DirectoryNode) {
	if !node.IsDir {
		return
	}

	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortDirectoryChildren(child)
	}
}
