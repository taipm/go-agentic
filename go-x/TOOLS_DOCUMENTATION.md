# 🛠️ GO-CREWAI System Execution Tools Documentation

## Overview

The Executor agent (Trang) now has comprehensive system execution capabilities to autonomously diagnose and gather information from the local machine without requiring user intervention.

## New Tools Added (Version 2.0)

### 1. **CheckMemoryStatus()** 
- **Description:** Get detailed memory usage information for the local machine
- **Parameters:** None
- **Returns:** Memory statistics (vm_stat on macOS, free -h on Linux)
- **Use Case:** "Kiểm tra dung lượng bộ nhớ", "Máy bộ nhớ có bao nhiêu"

```
Example Output:
Pages free:                    456789
Pages active:                  234567
Pages inactive:                567890
...
```

### 2. **CheckDiskStatus(path)** ⭐ NEW
- **Description:** Get detailed disk space usage for specified path or root directory
- **Parameters:** 
  - `path` (optional): Path to check (default: "/")
- **Returns:** Disk usage with percentages (df -h output)
- **Use Case:** "Check ổ đĩa", "Dung lượng đĩa còn bao nhiêu", "Ổ đĩa server-backup không còn chỗ"

```
Example: CheckDiskStatus(/home)
Returns:
Filesystem      Size  Used Avail Use% Mounted on
/dev/disk1s5   465Gi  300Gi 165Gi  65% /
```

### 3. **CheckNetworkStatus(host, count)** ⭐ NEW
- **Description:** Check network connectivity by pinging a host
- **Parameters:**
  - `host` (required): Hostname or IP to test (e.g., "8.8.8.8", "google.com")
  - `count` (optional): Number of pings (default: 3)
- **Returns:** Ping statistics or failure message
- **Use Case:** "Server 192.168.1.50 không ping được", "Check kết nối internet"

```
Example: CheckNetworkStatus(8.8.8.8)
Returns:
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: icmp_seq=0 ttl=119 time=24.5 ms
...
--- 8.8.8.8 statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
```

### 4. **ExecuteCommand(command)** ⭐ POWERFUL
- **Description:** Execute a shell command and return its output
- **Parameters:**
  - `command` (required): Shell command to execute (e.g., "ps aux", "ls -la", "uname -a")
- **Returns:** Command output or error message
- **Safety:** Built-in dangerous pattern blocking (rm -rf, mkfs, dd if=, fork bombs)
- **Use Case:** "Bạn tự lấy thông tin máy hiện tại", "Check tiến trình nào chạy"

```
Supported Examples:
- ExecuteCommand(ps aux) → List all running processes
- ExecuteCommand(netstat -tulpn) → Check open ports
- ExecuteCommand(df -h) → Disk usage
- ExecuteCommand(top -b -n 1) → System load
- ExecuteCommand(cat /etc/os-release) → OS information
```

**Blocked Commands (Security):**
- rm -rf (destructive)
- mkfs (filesystem destructive)
- dd if= (disk destructive)
- :(){:|:} (fork bomb)

### 5. **GetSystemDiagnostics()** ⭐ COMPREHENSIVE
- **Description:** Get comprehensive system diagnostics report
- **Parameters:** None
- **Returns:** Complete system health report including:
  - System Information (OS, Hostname)
  - CPU Usage
  - Memory Status
  - Disk Status
  - Top 5 Running Processes
- **Use Case:** "Check toàn bộ hệ thống", "Kiểm tra sức khỏe máy"

```
Output Format:
=== SYSTEM DIAGNOSTICS ===

--- System Information ---
OS: darwin
Hostname: Phans-MacBook-Pro-2

--- CPU Usage ---
CPU Usage: 45.2%

--- Memory Status ---
Pages free: 456789
Pages active: 234567
...

--- Disk Status (Root) ---
Filesystem      Size  Used Avail Use% Mounted on
/dev/disk1s5   465Gi  300Gi 165Gi  65% /

--- Top 5 Running Processes ---
USER    PID  %CPU %MEM COMMAND
root    1    0.0  0.1  /sbin/launchd
root    25   0.0  0.2  /usr/libexec/kextd
...
```

## Existing Tools (Still Available)

### Basic Diagnostics
- **GetCPUUsage()** - Current CPU percentage
- **GetMemoryUsage()** - Memory usage
- **GetDiskSpace(path)** - Disk space for path
- **GetSystemInfo()** - OS, hostname info
- **GetRunningProcesses(count)** - Top processes

### Network Tools
- **PingHost(host, count)** - Ping a host
- **ResolveDNS(hostname)** - DNS resolution
- **CheckNetworkStatus(host)** - Network connectivity check

### Service Management
- **CheckServiceStatus(service)** - Service status

## How Executor Uses These Tools

### Workflow Example: "Kiểm tra dung lượng bộ nhớ localhost"

1. **Orchestrator routes to Executor** (due to explicit "localhost" + "check memory" keywords)
2. **Executor receives request:**
   ```
   🔍 Chẩn đoán: Người dùng yêu cầu kiểm tra dung lượng bộ nhớ trên máy localhost
   ```
3. **Executor executes tools:**
   ```
   CheckMemoryStatus()        ← Get detailed memory info
   GetSystemDiagnostics()     ← Get comprehensive report
   ```
4. **Executor analyzes results:**
   ```
   Phân tích: Bộ nhớ có X GB còn trống, Y GB đang sử dụng...
   ```
5. **Executor provides recommendations:**
   ```
   ✅ Khuyến nghị Cuối Cùng:
   - Bộ nhớ hiện tại đang sử dụng X%
   - Còn Y GB bộ nhớ trống
   - Nếu vượt quá Z%, hãy đóng các ứng dụng không cần thiết
   ```

## Routing Logic Update

The Orchestrator now routes to Executor when:
1. ✅ User specifies "localhost" + action (check, test, diagnose)
2. ✅ User provides specific IP/hostname + clear problem
3. ✅ User explicitly requests system information gathering
4. ✅ User says "tự động lấy thông tin" (get info automatically)

Example routing triggers:
- "Kiểm tra dung lượng bộ nhớ localhost" → **EXECUTOR** ✅
- "Máy của tôi chậm lắm" → Clarifier (vague) ❌
- "Server 192.168.1.100 CPU cao, check ngay" → **EXECUTOR** ✅

## Configuration

All tools are configured in:
- [executor.yaml](config/agents/executor.yaml) - Tool list and system prompt
- [example_it_support.go](example_it_support.go) - Tool implementations

## Security Considerations

1. **ExecuteCommand** has built-in safety checks
2. Commands are executed with timeout context
3. Dangerous patterns are blocked automatically
4. Output is captured and returned safely
5. No shell escaping required (uses context-aware execution)

## Test the Tools

### Interactive Mode
```bash
cd go-crewai
./crewai-example

# Try these requests:
You: Kiểm tra dung lượng bộ nhớ localhost
You: CPU trên máy hiện tại bao nhiêu phần trăm
You: Ổ đĩa có còn chỗ không
You: Check toàn bộ hệ thống
```

### Test Mode
```bash
./crewai test
# Runs 10 test scenarios with HTML report
```

## Future Enhancements

Possible additions:
- Remote host execution (SSH support)
- Performance profiling tools
- Log analysis tools
- Service restart capabilities
- Configuration file analysis
