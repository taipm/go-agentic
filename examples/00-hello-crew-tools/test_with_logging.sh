#!/bin/bash

# Test script that shows tool logging clearly
# This demonstrates when tools are called during conversation

echo ""
echo "=========================================="
echo "hello-crew-tools: Tool Call Logging Test"
echo "=========================================="
echo ""
echo "This test shows [TOOL CALL] and [TOOL RESULT] logs"
echo "which prove that tools are being invoked by the agent"
echo ""
echo "Test conversation sequence:"
echo "1. Ask name (no tool needed yet)"
echo "2. Tell name (agent learns it)"
echo "3. Ask name again (recall test)"
echo "4. Ask how many questions (TOOL TEST - agent should call tools)"
echo ""
echo "=========================================="
echo ""
echo "Starting test... (watching for [TOOL CALL] logs)"
echo ""
echo "=========================================="
echo ""

# Run the application with test input, filter for relevant lines
(
echo "Tôi tên gì?"
sleep 3
echo "Tôi là John Doe"
sleep 3
echo "Tôi tên gì?"
sleep 3
echo "Tôi đã hỏi mấy câu?"
sleep 3
echo "exit"
) | ./hello-crew-tools 2>&1 | grep -E "^>|Response:|Conversation state:|TOOL CALL|TOOL RESULT|EXTRACTED|Using Ollama|CONFIG SUCCESS"

echo ""
echo "=========================================="
echo "Test Complete!"
echo "=========================================="
echo ""
echo "Look for lines starting with:"
echo "  [TOOL CALL]   - Shows when a tool is invoked"
echo "  [TOOL RESULT] - Shows what the tool returned"
echo ""
echo "If you see these logs, tools are working! ✅"
echo "If you don't see them, the agent isn't calling tools ❌"
echo ""
