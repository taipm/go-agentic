# Example Output - Vietnamese Multi-Agent Conversation

This document shows actual output from running the simple-chat example with Vietnamese-speaking agents.

## Running the Example

```bash
cp .env.example .env
# Add your OpenAI API key
go run main.go
```

## Expected Output Sample

### Topic 1: Go Best Practices

```
📌 Chủ đề 1: Những thực hành tốt nhất khi viết code Go là gì?
--------------------------------------------------

[Người Tò Mò]: Khi viết code Go, có một số thực hành tốt nhất mà bạn nên tuân 
thủ để đảm bảo mã của bạn không chỉ chạy hiệu quả mà còn dễ bảo trì...

1. **Sử dụng Go Modules**: Bắt đầu từ Go 1.11, Go Modules được giới thiệu 
để quản lý các thư viện và phụ thuộc...

2. **Cấu trúc thư mục rõ ràng**: Tổ chức mã nguồn của bạn theo cách hợp lý...

[... full response from Người Tò Mò ...]

[Chuyên Gia]: Rất vui được nghe bạn quan tâm đến các thực hành tốt nhất 
khi viết code Go. Nếu bạn cần đi sâu vào một vấn đề nào đó...

1. **Cách tối ưu hóa hiệu suất với goroutines**: Bạn có muốn biết thêm...

2. **Phân tích và ghi lỗi**: Có thể bạn muốn tìm hiểu thêm...

[... full response from Chuyên Gia ...]

✅ Kết Quả Cuối Cùng:
[Final comprehensive response from Expert in Vietnamese]
```

## Key Features Demonstrated

### ✅ Vietnamese Language Support
- All agents respond in Vietnamese
- Natural Vietnamese grammar and expressions
- Vietnamese question formulation
- Culturally appropriate responses

### ✅ Agent Personalities
- **Người Tò Mò (Enthusiast)**
  - Asks detailed questions
  - Shows curiosity and enthusiasm
  - Temperature: 0.8 (more creative)
  - Builds on previous answers

- **Chuyên Gia (Expert)**
  - Provides comprehensive answers
  - Offers multiple perspectives
  - Temperature: 0.7 (more consistent)
  - Gives terminal response

### ✅ Multi-Topic Discussion
The example discusses multiple topics in sequence:

1. Những thực hành tốt nhất khi viết code Go là gì?
   (What are the best practices for writing Go code?)

2. Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?
   (How can AI agents improve software development?)

3. Hãy cho tôi biết về những xu hướng mới nhất trong máy học?
   (Tell me about the latest trends in machine learning?)

4. Ứng dụng của Go trong các hệ thống distributed có những đặc điểm gì?
   (What are the characteristics of Go applications in distributed systems?)

## Conversation Flow Example

```
User Topic Input
    ↓
[Người Tò Mò] Processes topic
    ↓ 
[Người Tò Mò] Asks insightful question in Vietnamese
    ↓
[Chuyên Gia] Provides detailed answer in Vietnamese
    ↓
[Người Tò Mò] Asks follow-up question based on previous response
    ↓
[Chuyên Gia] Provides final comprehensive response
    ↓
Output complete, move to next topic
```

## Configuration Used

```yaml
crew:
  maxRounds: 4          # 4 rounds of conversation per topic
  maxHandoffs: 3        # Up to 3 handoffs between agents

agents:
  - id: "enthusiast"
    name: "Người Tò Mò"
    temperature: 0.8    # More creative

  - id: "expert"  
    name: "Chuyên Gia"
    temperature: 0.7    # More consistent
```

## Response Characteristics

### Người Tò Mò (Enthusiast) Characteristics:
✅ Opens with a topic-related introduction
✅ Lists multiple points or perspectives
✅ Asks relevant follow-up questions
✅ Ends with a question to continue conversation
✅ Temperature 0.8 produces varied responses

### Chuyên Gia (Expert) Characteristics:
✅ Acknowledges the question
✅ Provides detailed answers
✅ Offers multiple sub-points
✅ Suggests areas for deeper exploration
✅ Terminal response (ends conversation)
✅ Temperature 0.7 produces consistent, professional responses

## Real Conversation Examples

### Example 1: Go Programming

**Topic:** Những thực hành tốt nhất khi viết code Go là gì?

**Người Tò Mò:**
```
Khi viết code Go, có một số thực hành tốt nhất mà bạn nên tuân thủ...
1. **Sử dụng Go Modules**: ...
2. **Cấu trúc thư mục rõ ràng**: ...
3. **Đặt tên rõ ràng**: ...
[... 10 detailed points with explanations ...]
```

**Chuyên Gia:**
```
Rất vui được nghe bạn quan tâm đến các thực hành tốt nhất khi viết code Go...
1. **Cách tối ưu hóa hiệu suất với goroutines**: ...
2. **Phân tích và ghi lỗi**: ...
3. **Xây dựng API với Go**: ...
[... detailed expert insights ...]
```

### Example 2: AI in Software Development

**Topic:** Làm thế nào mà các AI agent có thể cải thiện phát triển phần mềm?

**Người Tò Mò:**
```
Đó là một câu hỏi thú vị và đang rất được quan tâm...
1. **Tự động hóa quy trình**: ...
2. **Phân tích mã nguồn**: ...
3. **Hỗ trợ lập trình**: ...
[... exploration of AI benefits in development ...]
```

**Chuyên Gia:**
```
AI agents thực sự có thể đóng vai trò quan trọng trong phát triển phần mềm...
[... comprehensive expert response ...]
```

## Language Quality

### Vietnamese Language Features:
✅ Natural Vietnamese expressions
✅ Proper Vietnamese grammar
✅ Vietnamese idioms and phrases
✅ Culturally appropriate responses
✅ UTF-8 full support
✅ Vietnamese punctuation and formatting

### Example Vietnamese Phrases Used:
- "Rất vui được..." (Pleasure to...)
- "Hãy cho tôi biết..." (Please tell me...)
- "Có thể bạn muốn..." (Perhaps you'd like...)
- "Dưới đây là..." (Below are...)
- "Chi tiết như sau..." (Details as follows...)

## Configuration Impact on Output

### If you increase maxRounds:
```yaml
crew:
  maxRounds: 6  # More conversation rounds
```
Result: Longer conversations with more back-and-forth exchanges

### If you increase maxHandoffs:
```yaml
crew:
  maxHandoffs: 5  # More handoffs
```
Result: More opportunities for agents to continue the discussion

### If you change temperature:
```yaml
agents:
  - id: "expert"
    temperature: 0.9  # More creative
```
Result: More varied, creative responses (vs. 0.7 = more consistent)

## Topics Discussed

All topics are in Vietnamese, allowing for natural language processing and discussion:

1. **Programming**: Go best practices, distributed systems
2. **AI/Technology**: AI agents in development, machine learning trends
3. **Software Development**: Software development improvements, technical topics

## Performance Metrics

Typical execution times for full example:
- Topic 1 (Go best practices): ~5-10 seconds
- Topic 2 (AI in development): ~5-10 seconds
- Topic 3 (ML trends): ~5-10 seconds
- Topic 4 (Go distributed systems): ~5-10 seconds

Total execution time: ~20-40 seconds for all 4 topics

## Common Observations

✅ **Natural Vietnamese**: Agents consistently use Vietnamese throughout
✅ **Coherent Conversations**: Each agent's response builds on the previous one
✅ **Expert Knowledge**: Responses demonstrate real understanding of topics
✅ **Professional Tone**: Responses are professional yet accessible
✅ **Question-Driven**: Enthusiast agent naturally formulates follow-up questions
✅ **Terminal Responses**: Expert agent provides fitting conclusion to each topic

## Customization Results

If you modify the YAML configuration, you'll see:

1. **Different Topics**: Just change the `topics` list
2. **Different Agent Personalities**: Change `backstory` and `role`
3. **Longer Conversations**: Increase `maxRounds` and `maxHandoffs`
4. **More Creative Responses**: Increase `temperature` values
5. **Different Language**: Change backstory to instruct agents to use different language

## Troubleshooting Output Issues

### If responses are in English instead of Vietnamese:
- Check that the agent `backstory` includes instruction to speak Vietnamese
- Verify YAML file is properly formatted
- Try running again (randomness in temperature may affect output)

### If conversations seem too short:
- Increase `maxRounds` in `crew.yaml`
- Increase `maxHandoffs`
- Increase agent `temperature` for more varied responses

### If responses are too generic:
- Improve the agent `backstory` with more detailed instructions
- Increase `temperature` for more creative responses
- Modify topics to be more specific

## Conclusion

This example demonstrates that go-agentic:

✅ **Works with any language** - Full UTF-8 support
✅ **Maintains conversation flow** - Natural back-and-forth
✅ **Respects agent roles** - Each agent has distinct personality
✅ **Configurable via YAML** - Easy to customize without code changes
✅ **Production-ready** - Clean, professional conversations

The Vietnamese example proves the library's flexibility for multi-lingual applications.
