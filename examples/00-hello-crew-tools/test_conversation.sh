#!/bin/bash

# Test script for hello-crew-tools
# This script runs a test conversation to validate memory tool capability

echo "=========================================="
echo "hello-crew-tools: Test Conversation"
echo "=========================================="
echo ""
echo "This test validates whether LLM tools can help with:"
echo "1. Message counting"
echo "2. Fact extraction"
echo "3. Conversation search"
echo ""
echo "Expected behavior:"
echo "- With tools: Agent calls get_message_count(), returns accurate count"
echo "- Without tools: Agent ignores tools, returns 'I don't know'"
echo ""
echo "Test conversation:"
echo "1. Tôi tên gì? (What's my name?)"
echo "2. Tôi là John Doe (I am John Doe)"
echo "3. Tôi tên gì? (What's my name?)"
echo "4. Tôi đã hỏi mấy câu? (How many questions did I ask?)"
echo ""
echo "=========================================="
echo ""

# Run the application with test input
(
echo "Tôi tên gì?"
sleep 1
echo "Tôi là John Doe"
sleep 1
echo "Tôi tên gì?"
sleep 1
echo "Tôi đã hỏi mấy câu?"
sleep 1
echo "exit"
) | ./hello-crew-tools 2>&1
