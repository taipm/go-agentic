#!/bin/bash

# Go Code Scanner Script
# Usage: ./scan.sh [--path <dir>] [--output <file>] [--open]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCAN_PATH="${SCRIPT_DIR}/.."
OUTPUT_FILE="full_project_report.json"
OPEN_REPORT=false

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}          ${BLUE}🔍 Go Code Scanner${NC}                            ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_usage() {
    echo -e "${YELLOW}Usage:${NC} $0 [options]"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo "  --path <dir>      Directory to scan (default: parent of script dir)"
    echo "  --output <file>   Output JSON file (default: full_project_report.json)"
    echo "  --open            Open HTML report in browser after generation"
    echo "  --help            Show this help message"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 --path ./myproject"
    echo "  $0 --open"
    echo "  $0 --path ./src --output report.json --open"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --path)
            SCAN_PATH="$2"
            shift 2
            ;;
        --output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        --open)
            OPEN_REPORT=true
            shift
            ;;
        --help|-h)
            print_header
            print_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown option $1${NC}"
            print_usage
            exit 1
            ;;
    esac
done

# Main execution
print_header

echo -e "${YELLOW}📂 Scan Path:${NC} $SCAN_PATH"
echo -e "${YELLOW}📄 Output:${NC} $OUTPUT_FILE"
echo ""

# Check if scan.go exists
if [[ ! -f "${SCRIPT_DIR}/scan.go" ]]; then
    echo -e "${RED}Error: scan.go not found in ${SCRIPT_DIR}${NC}"
    exit 1
fi

# Run scanner
echo -e "${BLUE}⏳ Running scanner...${NC}"
cd "$SCRIPT_DIR"
go run scan.go -path "$SCAN_PATH" -output "$OUTPUT_FILE"

if [[ $? -ne 0 ]]; then
    echo -e "${RED}❌ Scan failed!${NC}"
    exit 1
fi

echo ""

# Show statistics using jq if available
if command -v jq &> /dev/null; then
    echo -e "${GREEN}📊 Scan Statistics:${NC}"
    echo "────────────────────────────────────────"
    
    TOTAL_FILES=$(jq '.total_files' "$OUTPUT_FILE")
    TOTAL_FUNCS=$(jq '.total_functions' "$OUTPUT_FILE")
    TOTAL_METHODS=$(jq '.total_methods' "$OUTPUT_FILE")
    TOTAL_TESTS=$(jq '.total_tests' "$OUTPUT_FILE")
    
    echo -e "  ${CYAN}Files:${NC}     $TOTAL_FILES"
    echo -e "  ${CYAN}Functions:${NC} $TOTAL_FUNCS"
    echo -e "  ${CYAN}Methods:${NC}   $TOTAL_METHODS"
    echo -e "  ${CYAN}Tests:${NC}     $TOTAL_TESTS"
    
    echo ""
    echo -e "${GREEN}🏷️ Function Types:${NC}"
    echo "────────────────────────────────────────"
    
    CONSTRUCTORS=$(jq '.function_type_stats.constructors' "$OUTPUT_FILE")
    GETTERS=$(jq '.function_type_stats.getters' "$OUTPUT_FILE")
    SETTERS=$(jq '.function_type_stats.setters' "$OUTPUT_FILE")
    HANDLERS=$(jq '.function_type_stats.handlers' "$OUTPUT_FILE")
    MIDDLEWARES=$(jq '.function_type_stats.middlewares' "$OUTPUT_FILE")
    HIGHER_ORDER=$(jq '.function_type_stats.higher_order' "$OUTPUT_FILE")
    VALIDATORS=$(jq '.function_type_stats.validators' "$OUTPUT_FILE")
    CONVERTERS=$(jq '.function_type_stats.converters' "$OUTPUT_FILE")
    INITS=$(jq '.function_type_stats.inits' "$OUTPUT_FILE")
    CALLBACKS=$(jq '.function_type_stats.callbacks' "$OUTPUT_FILE")
    TESTS=$(jq '.function_type_stats.tests' "$OUTPUT_FILE")
    HELPERS=$(jq '.function_type_stats.helpers' "$OUTPUT_FILE")
    LIFECYCLE=$(jq '.function_type_stats.lifecycle' "$OUTPUT_FILE")
    STANDARD=$(jq '.function_type_stats.standard' "$OUTPUT_FILE")
    
    echo -e "  📦 Constructors: $CONSTRUCTORS"
    echo -e "  🔧 Getters:      $GETTERS"
    echo -e "  ✏️  Setters:      $SETTERS"
    echo -e "  🎯 Handlers:     $HANDLERS"
    echo -e "  🔗 Middlewares:  $MIDDLEWARES"
    echo -e "  ⚡ Higher-Order: $HIGHER_ORDER"
    echo -e "  ✅ Validators:   $VALIDATORS"
    echo -e "  🔄 Converters:   $CONVERTERS"
    echo -e "  🚀 Inits:        $INITS"
    echo -e "  📞 Callbacks:    $CALLBACKS"
    echo -e "  🧪 Tests:        $TESTS"
    echo -e "  🔨 Helpers:      $HELPERS"
    echo -e "  ♻️  Lifecycle:    $LIFECYCLE"
    echo -e "  📝 Standard:     $STANDARD"
    echo ""
fi

# Generate HTML report
echo -e "${BLUE}⏳ Generating HTML report...${NC}"
HTML_OUTPUT="${OUTPUT_FILE%.json}.html"

if [[ ! -f "${SCRIPT_DIR}/generate-html-report.go" ]]; then
    echo -e "${YELLOW}⚠️ generate-html-report.go not found, skipping HTML generation${NC}"
else
    go run generate-html-report.go -input "$OUTPUT_FILE" -output "$HTML_OUTPUT"
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✅ HTML report generated: ${HTML_OUTPUT}${NC}"
        
        # Open in browser if requested
        if [[ "$OPEN_REPORT" == true ]]; then
            echo -e "${BLUE}🌐 Opening report in browser...${NC}"
            if [[ "$OSTYPE" == "darwin"* ]]; then
                open "$HTML_OUTPUT"
            elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
                xdg-open "$HTML_OUTPUT" 2>/dev/null || sensible-browser "$HTML_OUTPUT"
            elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
                start "$HTML_OUTPUT"
            fi
        fi
    fi
fi

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Scan completed successfully!${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Output files:${NC}"
echo "  📄 JSON: $OUTPUT_FILE"
echo "  🌐 HTML: $HTML_OUTPUT"
