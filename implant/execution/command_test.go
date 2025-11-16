package execution

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestExecuteCommand tests basic command execution
func TestExecuteCommand(t *testing.T) {
	var cmd string
	var expectedOutput string

	// Platform-specific test commands
	if runtime.GOOS == "windows" {
		cmd = "echo test"
		expectedOutput = "test"
	} else {
		cmd = "echo test"
		expectedOutput = "test"
	}

	output, err := ExecuteCommand(cmd)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Output doesn't contain expected string: got %s, want substring %s",
			output, expectedOutput)
	}
}

// TestExecuteCommandWithTimeout tests command execution with timeout
func TestExecuteCommandWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 1 127.0.0.1"
	} else {
		cmd = "sleep 1"
	}

	output, err := ExecuteCommandWithContext(ctx, cmd)
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if output == "" {
		t.Error("Expected output, got empty string")
	}
}

// TestExecuteCommandTimeout tests timeout handling
func TestExecuteCommandTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 10 127.0.0.1" // Should take ~10 seconds
	} else {
		cmd = "sleep 10" // Should take 10 seconds
	}

	_, err := ExecuteCommandWithContext(ctx, cmd)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// TestExecuteInvalidCommand tests error handling for invalid commands
func TestExecuteInvalidCommand(t *testing.T) {
	cmd := "nonexistentcommand12345"

	_, err := ExecuteCommand(cmd)
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}
}

// TestExecutePowerShell tests PowerShell execution on Windows
func TestExecutePowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell test only runs on Windows")
	}

	cmd := "Get-Date"
	output, err := ExecutePowerShell(cmd)
	if err != nil {
		t.Fatalf("PowerShell execution failed: %v", err)
	}

	if output == "" {
		t.Error("Expected PowerShell output, got empty string")
	}
}

// TestExecuteShell tests shell execution
func TestExecuteShell(t *testing.T) {
	var cmd string
	var expectedSubstring string

	if runtime.GOOS == "windows" {
		cmd = "echo %COMPUTERNAME%"
		expectedSubstring = "" // Variable content, just check non-empty
	} else {
		cmd = "echo $HOME"
		expectedSubstring = "/" // Should contain a path
	}

	output, err := ExecuteShell(cmd)
	if err != nil {
		t.Fatalf("Shell execution failed: %v", err)
	}

	if output == "" {
		t.Error("Expected shell output, got empty string")
	}

	if expectedSubstring != "" && !strings.Contains(output, expectedSubstring) {
		t.Errorf("Output doesn't contain expected substring: got %s, want substring %s",
			output, expectedSubstring)
	}
}

// TestChangeDirectory tests directory change command
func TestChangeDirectory(t *testing.T) {
	var testDir string

	if runtime.GOOS == "windows" {
		testDir = "C:\\"
	} else {
		testDir = "/tmp"
	}

	err := ChangeDirectory(testDir)
	if err != nil {
		t.Fatalf("ChangeDirectory failed: %v", err)
	}

	// Verify we're in the new directory
	var pwdCmd string
	if runtime.GOOS == "windows" {
		pwdCmd = "cd"
	} else {
		pwdCmd = "pwd"
	}

	output, err := ExecuteCommand(pwdCmd)
	if err != nil {
		t.Fatalf("pwd command failed: %v", err)
	}

	if !strings.Contains(strings.ToLower(output), strings.ToLower(testDir)) {
		t.Errorf("Directory change failed: got %s, want %s", output, testDir)
	}
}

// TestListDirectory tests directory listing
func TestListDirectory(t *testing.T) {
	var testDir string

	if runtime.GOOS == "windows" {
		testDir = "C:\\Windows"
	} else {
		testDir = "/etc"
	}

	output, err := ListDirectory(testDir)
	if err != nil {
		t.Fatalf("ListDirectory failed: %v", err)
	}

	if output == "" {
		t.Error("Expected directory listing, got empty string")
	}
}

// TestDownloadFile tests file download functionality
func TestDownloadFile(t *testing.T) {
	t.Skip("Skipping file download test - requires network access")

	url := "https://example.com"
	destPath := "/tmp/test_download.txt"

	err := DownloadFile(url, destPath)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
}

// TestUploadFile tests file upload functionality
func TestUploadFile(t *testing.T) {
	t.Skip("Skipping file upload test - requires server")

	srcPath := "/tmp/test_file.txt"
	url := "https://example.com/upload"

	err := UploadFile(srcPath, url)
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}
}

// TestGetProcessList tests process listing
func TestGetProcessList(t *testing.T) {
	output, err := GetProcessList()
	if err != nil {
		t.Fatalf("GetProcessList failed: %v", err)
	}

	if output == "" {
		t.Error("Expected process list, got empty string")
	}

	// Check for expected process list indicators
	if runtime.GOOS == "windows" {
		if !strings.Contains(output, "PID") && !strings.Contains(output, "Image Name") {
			t.Error("Output doesn't look like Windows process list")
		}
	} else {
		if !strings.Contains(output, "PID") || !strings.Contains(output, "CMD") {
			t.Error("Output doesn't look like Unix process list")
		}
	}
}

// TestKillProcess tests process termination
func TestKillProcess(t *testing.T) {
	t.Skip("Skipping kill process test - requires careful setup")

	// Would need to create a test process first
	pid := 12345
	err := KillProcess(pid)
	if err == nil {
		t.Error("Expected error for non-existent PID")
	}
}

// TestGetSystemInfo tests system information gathering
func TestGetSystemInfo(t *testing.T) {
	info, err := GetSystemInfo()
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	if info.Hostname == "" {
		t.Error("Expected hostname, got empty string")
	}

	if info.Username == "" {
		t.Error("Expected username, got empty string")
	}

	if info.OS == "" {
		t.Error("Expected OS, got empty string")
	}

	if info.Architecture == "" {
		t.Error("Expected architecture, got empty string")
	}
}

// TestCommandValidation tests command validation
func TestCommandValidation(t *testing.T) {
	testCases := []struct {
		name      string
		command   string
		shouldErr bool
	}{
		{"Valid simple command", "echo test", false},
		{"Valid piped command", "echo test | grep test", false},
		{"Empty command", "", true},
		{"Whitespace only", "   ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCommand(tc.command)
			if tc.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// BenchmarkExecuteCommand benchmarks command execution
func BenchmarkExecuteCommand(b *testing.B) {
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo test"
	} else {
		cmd = "echo test"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteCommand(cmd)
	}
}

// BenchmarkGetSystemInfo benchmarks system info gathering
func BenchmarkGetSystemInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSystemInfo()
	}
}

// BenchmarkGetProcessList benchmarks process listing
func BenchmarkGetProcessList(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetProcessList()
	}
}
