package crewai

const exampleHTMLClient = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>go-crewai SSE Streaming Client</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 1000px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin: 0 0 10px 0;
        }
        .subtitle {
            color: #666;
            margin-bottom: 20px;
        }
        .input-group {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        input {
            flex: 1;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        button {
            padding: 10px 20px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
        }
        button:hover {
            background: #0056b3;
        }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .stream-output {
            background: #f9f9f9;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 15px;
            height: 400px;
            overflow-y: auto;
            font-family: "Courier New", monospace;
            font-size: 12px;
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        .event {
            margin-bottom: 10px;
            padding: 8px;
            border-left: 3px solid #ddd;
            padding-left: 10px;
        }
        .event.agent_start {
            border-left-color: #007bff;
            color: #007bff;
        }
        .event.agent_response {
            border-left-color: #28a745;
            color: #28a745;
        }
        .event.tool_start {
            border-left-color: #ffc107;
            color: #856404;
        }
        .event.tool_result {
            border-left-color: #17a2b8;
            color: #0c5460;
        }
        .event.pause {
            border-left-color: #ff6b6b;
            color: #721c24;
            background: #f8d7da;
        }
        .event.error {
            border-left-color: #dc3545;
            color: #721c24;
            background: #f8d7da;
        }
        .status {
            margin-top: 10px;
            padding: 10px;
            border-radius: 4px;
            font-size: 12px;
        }
        .status.running {
            background: #d1ecf1;
            color: #0c5460;
        }
        .status.done {
            background: #d4edda;
            color: #155724;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 go-crewai SSE Streaming Client</h1>
        <p class="subtitle">Real-time crew execution with Server-Sent Events</p>

        <div class="input-group">
            <input type="text" id="queryInput" placeholder="Enter your IT support request (e.g., 'Tôi không vào được Internet')" />
            <button id="sendBtn" onclick="sendRequest()">Send</button>
        </div>

        <div id="streamOutput" class="stream-output"></div>
        <div id="status" class="status" style="display:none;"></div>
    </div>

    <script>
        let history = [];
        let eventSource = null;

        function sendRequest() {
            const query = document.getElementById('queryInput').value.trim();
            if (!query) {
                alert('Please enter a query');
                return;
            }

            const btn = document.getElementById('sendBtn');
            btn.disabled = true;

            const status = document.getElementById('status');
            status.className = 'status running';
            status.textContent = '⏳ Executing...';
            status.style.display = 'block';

            const output = document.getElementById('streamOutput');
            output.innerHTML += '<div class="event">You: ' + query + '</div>\n';

            // Add to history
            history.push({ role: 'user', content: query });

            const payload = {
                query: query,
                history: history
            };

            // Close existing connection
            if (eventSource) {
                eventSource.close();
            }

            // Open SSE connection
            eventSource = new EventSource('/api/crew/stream?q=' + encodeURIComponent(JSON.stringify(payload)));

            eventSource.onmessage = function(event) {
                const data = event.data;
                if (!data) return;

                try {
                    const streamEvent = JSON.parse(data);
                    handleStreamEvent(streamEvent);
                } catch (e) {
                    console.error('Failed to parse event:', e);
                }
            };

            eventSource.onerror = function(error) {
                console.error('SSE Error:', error);
                eventSource.close();
                btn.disabled = false;
                status.className = 'status';
                status.textContent = '❌ Stream closed';
                status.style.display = 'block';
            };
        }

        function handleStreamEvent(event) {
            const output = document.getElementById('streamOutput');
            const status = document.getElementById('status');

            let displayText = '';
            switch (event.type) {
                case 'agent_start':
                    displayText = '🔄 ' + event.content + ' [' + event.agent + ']';
                    break;
                case 'agent_response':
                    displayText = '💬 Agent (' + event.agent + '): ' + event.content;
                    history.push({ role: 'assistant', content: event.content });
                    break;
                case 'tool_start':
                    displayText = '🔧 ' + event.content;
                    break;
                case 'tool_result':
                    displayText = '✅ ' + event.content;
                    break;
                case 'pause':
                    displayText = '⏸️  WAITING FOR INPUT';
                    document.getElementById('sendBtn').disabled = false;
                    status.className = 'status done';
                    status.textContent = '✅ Ready for next input';
                    status.style.display = 'block';
                    eventSource.close();
                    return;
                case 'done':
                    displayText = '✅ ' + event.content;
                    document.getElementById('sendBtn').disabled = false;
                    status.className = 'status done';
                    status.textContent = '✅ Completed';
                    status.style.display = 'block';
                    eventSource.close();
                    return;
                case 'error':
                    displayText = '❌ Error: ' + event.content;
                    document.getElementById('sendBtn').disabled = false;
                    break;
                default:
                    displayText = '[' + event.type + '] ' + event.content;
            }

            const eventDiv = document.createElement('div');
            eventDiv.className = 'event ' + event.type;
            eventDiv.textContent = displayText;
            output.appendChild(eventDiv);
            output.scrollTop = output.scrollHeight;
        }

        // Allow Enter key to send
        document.getElementById('queryInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !document.getElementById('sendBtn').disabled) {
                sendRequest();
            }
        });
    </script>
</body>
</html>`
