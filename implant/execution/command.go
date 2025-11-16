package execution

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// Command Execution Engine - Cross-platform command runner
//
// BLUE TEAM DETECTION INDICATORS:
// 1. Process: Suspicious parent-child relationships (unsigned binary spawning cmd.exe/powershell)
// 2. EDR: Command-line logging shows unusual arguments
// 3. Sysmon: Event ID 1 (Process Creation) from unexpected parent
// 4. Windows: Event ID 4688 (Process Creation) with suspicious commands
// 5. Behavioral: Process accessing network + spawning shells
//
// MITIGATIONS:
// - Application whitelisting (prevent unsigned binary execution)
// - Command-line logging and analysis
// - EDR behavioral monitoring
// - Sysmon process creation rules
// - PowerShell ScriptBlock logging

// CommandExecutor handles cross-platform command execution
type CommandExecutor struct {
	timeout      time.Duration
	maxOutputLen int
}

// CommandResult contains the output and metadata from command execution
type CommandResult struct {
	Output   string
	Error    string
	ExitCode int
	Duration time.Duration
	Success  bool
}

// NewCommandExecutor creates a new command executor with configuration
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	if timeout == 0 {
		timeout = 30 * time.Second // Default 30s timeout
	}

	return &CommandExecutor{
		timeout:      timeout,
		maxOutputLen: 1024 * 1024, // 1MB max output
	}
}

// Execute runs a command with timeout and output capture
//
// OPSEC Notes:
// - Uses native shell for each platform (cmd.exe on Windows, sh on Linux)
// - Captures STDOUT and STDERR
// - Enforces timeout to prevent hung processes
// - Returns sanitized output (truncated if too large)
func (ce *CommandExecutor) Execute(command string) *CommandResult {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ce.timeout)
	defer cancel()

	// Prepare command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: Use cmd.exe
		// OPSEC: cmd.exe spawning is highly suspicious if parent is unknown binary
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", command)
	} else {
		// Linux/macOS: Use sh
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
	}

	// Setup output capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Hide console window on Windows
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	// Execute command
	err := cmd.Run()

	// Calculate duration
	duration := time.Since(startTime)

	// Build result
	result := &CommandResult{
		Output:   ce.sanitizeOutput(stdout.String()),
		Error:    ce.sanitizeOutput(stderr.String()),
		Duration: duration,
	}

	// Handle execution errors
	if err != nil {
		result.Success = false
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	return result
}

// ExecutePowerShell executes PowerShell commands (Windows only)
//
// BLUE TEAM DETECTION:
// - PowerShell ScriptBlock Logging (Event ID 4104)
// - AMSI (Antimalware Scan Interface) inspection
// - PowerShell Transcription logging
// - Unusual PowerShell parent processes
func (ce *CommandExecutor) ExecutePowerShell(script string) *CommandResult {
	if runtime.GOOS != "windows" {
		return &CommandResult{
			Success:  false,
			Error:    "PowerShell only available on Windows",
			ExitCode: -1,
		}
	}

	// Use PowerShell with execution policy bypass
	// OPSEC WARNING: -ExecutionPolicy Bypass is highly suspicious!
	command := fmt.Sprintf("powershell.exe -ExecutionPolicy Bypass -NoProfile -NonInteractive -Command \"%s\"", script)

	return ce.Execute(command)
}

// ExecuteWithInput executes a command with stdin input
// Useful for interactive commands or piping data
func (ce *CommandExecutor) ExecuteWithInput(command string, input string) *CommandResult {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), ce.timeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
	}

	// Setup stdin
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	err := cmd.Run()
	duration := time.Since(startTime)

	result := &CommandResult{
		Output:   ce.sanitizeOutput(stdout.String()),
		Error:    ce.sanitizeOutput(stderr.String()),
		Duration: duration,
	}

	if err != nil {
		result.Success = false
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	return result
}

// ExecuteBackground runs a command in background without waiting
// OPSEC: Useful for long-running tasks, but orphaned processes are suspicious
func (ce *CommandExecutor) ExecuteBackground(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", command)
	} else {
		cmd = exec.Command("/bin/sh", "-c", command)
	}

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
	}

	return cmd.Start()
}

// sanitizeOutput truncates output if it exceeds max length
// Prevents memory issues with large command outputs
func (ce *CommandExecutor) sanitizeOutput(output string) string {
	if len(output) > ce.maxOutputLen {
		return output[:ce.maxOutputLen] + "\n[... output truncated ...]"
	}
	return output
}

// GetSystemInfo gathers basic system information
// Returns OS, architecture, hostname, username, etc.
func (ce *CommandExecutor) GetSystemInfo() map[string]string {
	info := make(map[string]string)

	// OS and Architecture
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH

	// Hostname
	if runtime.GOOS == "windows" {
		result := ce.Execute("hostname")
		if result.Success {
			info["hostname"] = strings.TrimSpace(result.Output)
		}
	} else {
		result := ce.Execute("hostname")
		if result.Success {
			info["hostname"] = strings.TrimSpace(result.Output)
		}
	}

	// Username
	if runtime.GOOS == "windows" {
		result := ce.Execute("whoami")
		if result.Success {
			info["username"] = strings.TrimSpace(result.Output)
		}
	} else {
		result := ce.Execute("whoami")
		if result.Success {
			info["username"] = strings.TrimSpace(result.Output)
		}
	}

	// IP Address
	if runtime.GOOS == "windows" {
		result := ce.Execute("ipconfig")
		info["network"] = ce.extractIPFromIPConfig(result.Output)
	} else {
		result := ce.Execute("ip addr show || ifconfig")
		info["network"] = ce.extractIPFromIfconfig(result.Output)
	}

	// Privilege level (Windows)
	if runtime.GOOS == "windows" {
		result := ce.Execute("net session 2>&1")
		if result.ExitCode == 0 {
			info["privilege"] = "admin"
		} else {
			info["privilege"] = "user"
		}
	} else {
		result := ce.Execute("id -u")
		if result.Success && strings.TrimSpace(result.Output) == "0" {
			info["privilege"] = "root"
		} else {
			info["privilege"] = "user"
		}
	}

	return info
}

// extractIPFromIPConfig parses ipconfig output for IPv4 address
func (ce *CommandExecutor) extractIPFromIPConfig(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "IPv4 Address") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

// extractIPFromIfconfig parses ifconfig/ip output for IPv4 address
func (ce *CommandExecutor) extractIPFromIfconfig(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet ") && !strings.Contains(line, "127.0.0.1") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "inet" && i+1 < len(parts) {
					addr := parts[i+1]
					// Remove CIDR notation if present
					addr = strings.Split(addr, "/")[0]
					return addr
				}
			}
		}
	}
	return "unknown"
}

// IsElevated checks if the current process has elevated privileges
func (ce *CommandExecutor) IsElevated() bool {
	if runtime.GOOS == "windows" {
		result := ce.Execute("net session 2>&1")
		return result.ExitCode == 0
	} else {
		result := ce.Execute("id -u")
		return result.Success && strings.TrimSpace(result.Output) == "0"
	}
}

// ListProcesses returns list of running processes
// OPSEC: Process enumeration is common EDR detection indicator
func (ce *CommandExecutor) ListProcesses() string {
	if runtime.GOOS == "windows" {
		result := ce.Execute("tasklist /V")
		return result.Output
	} else {
		result := ce.Execute("ps aux")
		return result.Output
	}
}

// KillProcess terminates a process by name or PID
func (ce *CommandExecutor) KillProcess(target string) *CommandResult {
	if runtime.GOOS == "windows" {
		return ce.Execute(fmt.Sprintf("taskkill /F /IM %s 2>&1 || taskkill /F /PID %s", target, target))
	} else {
		return ce.Execute(fmt.Sprintf("kill -9 %s", target))
	}
}
