#!/bin/bash

# 🎬 Interactive SSE Streaming Demo Script
# Usage: chmod +x demo.sh && ./demo.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8081"
HEALTH_URL="$SERVER_URL/health"
STREAM_URL="$SERVER_URL/api/crew/stream"

# Header
clear
echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                                                            ║"
echo "║     🎬 SSE Streaming Demo - Interactive Test Suite        ║"
echo "║                                                            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Function: Check server health
check_health() {
    echo -e "\n${BLUE}[CHECK] Verifying server health...${NC}"

    if curl -s "$HEALTH_URL" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is healthy and ready!${NC}"
        return 0
    else
        echo -e "${RED}❌ Server is not responding on $SERVER_URL${NC}"
        echo -e "${YELLOW}Start the server first:${NC}"
        echo "  cd go-crewai"
        echo "  go run ./cmd/main.go --server --port 8081"
        return 1
    fi
}

# Function: Demo header
demo_header() {
    echo -e "\n${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}Demo: $1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
}

# Function: Pretty print streaming events
print_stream() {
    local query=$1
    local history=${2:-""}

    echo -e "\n${YELLOW}📤 Request:${NC}"
    if [ -z "$history" ]; then
        echo "  Query: \"$query\""
    else
        echo "  Query: \"$query\""
        echo "  History: (conversation context included)"
    fi

    echo -e "\n${YELLOW}📡 Streaming response:${NC}"
    echo ""

    # Prepare payload
    if [ -z "$history" ]; then
        PAYLOAD="{\"query\":\"$query\",\"history\":[]}"
    else
        PAYLOAD=$history
    fi

    # Stream events
    curl -s -X POST "$STREAM_URL" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" | while IFS= read -r line; do

        if [[ $line == data:* ]]; then
            # Extract JSON from SSE format
            json_data="${line#data: }"

            # Parse and format event
            if command -v jq &> /dev/null; then
                event_type=$(echo "$json_data" | jq -r '.type' 2>/dev/null)
                content=$(echo "$json_data" | jq -r '.content' 2>/dev/null)
                agent=$(echo "$json_data" | jq -r '.agent' 2>/dev/null)

                case "$event_type" in
                    start)
                        echo -e "${GREEN}🚀 $content${NC}"
                        ;;
                    agent_start)
                        echo -e "${YELLOW}🔄 [$agent] $content${NC}"
                        ;;
                    agent_response)
                        echo -e "${CYAN}💬 [$agent] $content${NC}"
                        ;;
                    tool_start)
                        echo -e "${BLUE}🔧 $content${NC}"
                        ;;
                    tool_result)
                        echo -e "${GREEN}✅ $content${NC}"
                        ;;
                    pause)
                        echo -e "${YELLOW}⏸️  [PAUSE] $content${NC}"
                        ;;
                    done)
                        echo -e "${GREEN}✅ $content${NC}"
                        ;;
                    error)
                        echo -e "${RED}❌ $content${NC}"
                        ;;
                    ping)
                        echo -e "${CYAN}(keep-alive ping)${NC}"
                        ;;
                    *)
                        echo -e "${NC}[$event_type] $content"
                        ;;
                esac
            else
                # Fallback if jq not available
                echo "$json_data"
            fi
        fi
    done
}

# Function: Demo 1 - Simple query
demo_1() {
    demo_header "1️⃣  Simple Query - Machine Slow"

    echo -e "${YELLOW}Scenario:${NC} User reports machine is slow."
    echo -e "${YELLOW}Expected:${NC} Orchestrator → Executor (direct routing)"
    echo ""
    echo -e "${YELLOW}⏳ Processing...${NC}"
    sleep 1

    print_stream "Máy chậm lắm"

    echo ""
    echo -e "${GREEN}✅ Demo 1 completed!${NC}"
}

# Function: Demo 2 - Network issue
demo_2() {
    demo_header "2️⃣  Network Connectivity - Direct Problem"

    echo -e "${YELLOW}Scenario:${NC} User can't access Internet - specific problem."
    echo -e "${YELLOW}Expected:${NC} Orchestrator → Executor with network tools"
    echo ""
    echo -e "${YELLOW}⏳ Processing...${NC}"
    sleep 1

    print_stream "Server 192.168.1.50 không ping được, check cho tôi"

    echo ""
    echo -e "${GREEN}✅ Demo 2 completed!${NC}"
}

# Function: Demo 3 - Vague question (pause/resume)
demo_3() {
    demo_header "3️⃣  Vague Question - Pause/Resume Flow (Part 1)"

    echo -e "${YELLOW}Scenario:${NC} User asks vague question."
    echo -e "${YELLOW}Expected:${NC} Orchestrator → Clarifier → PAUSE (waiting for input)"
    echo ""
    echo -e "${YELLOW}⏳ Processing...${NC}"
    sleep 1

    print_stream "Tôi không vào được Internet"

    echo ""
    echo -e "${YELLOW}ℹ️  Notice:${NC} Stream paused, waiting for user to answer Clarifier's question"
    echo ""
    echo -e "${GREEN}✅ Demo 3 Part 1 completed!${NC}"
    echo -e "${YELLOW}🔄 Would you like to continue with Part 2 (Resume with clarification)?${NC}"
}

# Function: Demo 4 - Resume with clarification
demo_4() {
    demo_header "4️⃣  Vague Question - Resume with Clarification (Part 2)"

    echo -e "${YELLOW}Scenario:${NC} User provides clarification from previous question."
    echo -e "${YELLOW}Expected:${NC} Executor with network diagnostics"
    echo ""
    echo -e "${YELLOW}⏳ Processing...${NC}"
    sleep 1

    HISTORY='{
        "query":"Máy 192.168.1.101, Ubuntu Linux, không ping được 8.8.8.8",
        "history":[
            {"role":"user","content":"Tôi không vào được Internet"},
            {"role":"assistant","content":"Tôi sẽ chuyển sang Clarifier để làm rõ vấn đề..."},
            {"role":"assistant","content":"Bạn đang cố kết nối từ máy nào? (Windows/Mac/Linux)"}
        ]
    }'

    print_stream "Máy 192.168.1.101, Ubuntu Linux, không ping được 8.8.8.8" "$HISTORY"

    echo ""
    echo -e "${GREEN}✅ Demo 4 completed!${NC}"
}

# Function: Demo 5 - Load test
demo_5() {
    demo_header "5️⃣  Load Test - Concurrent Requests"

    echo -e "${YELLOW}Scenario:${NC} Send 3 concurrent requests to test server capacity."
    echo -e "${YELLOW}Expected:${NC} All requests handled independently"
    echo ""

    for i in {1..3}; do
        echo -e "${YELLOW}Request $i/3...${NC}" &
        curl -s -X POST "$STREAM_URL" \
            -H "Content-Type: application/json" \
            -d "{\"query\":\"Load test request $i\"}" > /tmp/stream_$i.log &
    done

    wait

    echo -e "${GREEN}✅ All requests completed!${NC}"
    echo -e "${YELLOW}Check logs:${NC}"
    for i in {1..3}; do
        if [ -s /tmp/stream_$i.log ]; then
            echo -e "  ${GREEN}✅ Request $i${NC}: $(wc -l < /tmp/stream_$i.log) events"
        fi
    done
}

# Function: Demo 6 - Health check
demo_6() {
    demo_header "6️⃣  Server Health Check"

    echo -e "${YELLOW}⏳ Checking server...${NC}"

    if curl -s "$HEALTH_URL" | command -v jq &> /dev/null; then
        curl -s "$HEALTH_URL" | jq '.'
    else
        curl -s "$HEALTH_URL"
    fi

    echo ""
    echo -e "${GREEN}✅ Server health check completed!${NC}"
}

# Main menu
show_menu() {
    echo -e "\n${CYAN}📋 Select Demo:${NC}\n"
    echo "  1️⃣  Demo 1: Simple Query (Machine Slow)"
    echo "  2️⃣  Demo 2: Network Issue (Direct Problem)"
    echo "  3️⃣  Demo 3: Vague Question (Pause/Resume Part 1)"
    echo "  4️⃣  Demo 4: Resume with Clarification (Part 2)"
    echo "  5️⃣  Demo 5: Load Test (Concurrent Requests)"
    echo "  6️⃣  Demo 6: Server Health Check"
    echo "  7️⃣  Demo 7: Run All Demos"
    echo "  8️⃣  Demo 8: Open Web Client"
    echo "  9️⃣  Exit"
    echo ""
}

# Run all demos
run_all_demos() {
    echo -e "\n${CYAN}Running all demos...${NC}"
    demo_1
    read -p "Press Enter to continue to Demo 2..."
    demo_2
    read -p "Press Enter to continue to Demo 3..."
    demo_3
    read -p "Press Enter to continue to Demo 4..."
    demo_4
    read -p "Press Enter to continue to Demo 5..."
    demo_5
    read -p "Press Enter to continue to Demo 6..."
    demo_6
    echo -e "\n${GREEN}✅ All demos completed!${NC}"
}

# Main loop
if ! check_health; then
    exit 1
fi

while true; do
    show_menu
    read -p "Enter choice (1-9): " choice

    case $choice in
        1) demo_1 ;;
        2) demo_2 ;;
        3) demo_3 ;;
        4) demo_4 ;;
        5) demo_5 ;;
        6) demo_6 ;;
        7) run_all_demos ;;
        8)
            echo -e "${BLUE}Opening web client: $SERVER_URL${NC}"
            if command -v open &> /dev/null; then
                open "$SERVER_URL"
            elif command -v xdg-open &> /dev/null; then
                xdg-open "$SERVER_URL"
            else
                echo -e "${YELLOW}Please open $SERVER_URL manually${NC}"
            fi
            ;;
        9)
            echo -e "\n${GREEN}Goodbye! 👋${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice. Please try again.${NC}"
            ;;
    esac

    read -p "Press Enter to continue..."
done
