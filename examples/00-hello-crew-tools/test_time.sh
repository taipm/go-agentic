#!/bin/bash

echo "Testing get_current_time() tool"
echo "================================"
echo ""
echo "Current system time:"
date "+Time: %H:%M:%S"
echo "Date: $(date '+%Y-%m-%d')"
echo ""
echo "Starting agent test..."
echo ""

(
echo "Mấy giờ rồi?"
sleep 5
echo "exit"
) | ./hello-crew-tools 2>&1 | grep -E "TOOL EXECUTION|TOOL RESULT|Response:|Conversation state" | head -20
