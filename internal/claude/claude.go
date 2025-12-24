package claude

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// State represents the state of a Claude process
type State string

const (
	StateRunning State = "running"
	StateWaiting State = "waiting"
	StateIdle    State = "idle"
	StateUnknown State = "unknown"
)

// Info contains information about a Claude process
type Info struct {
	State      State
	PID        int
	LastActive string // e.g., "12m ago"
	TokensIn   int
	TokensOut  int
}

// GetState returns the Claude state for a session path
func GetState(sessionPath string) State {
	pid, err := GetProcessPID(sessionPath)
	if err != nil || pid == 0 {
		return StateIdle
	}

	// Check if process is waiting for input (simplified check)
	// A more accurate check would inspect the process's file descriptors
	if isWaitingForInput(pid) {
		return StateWaiting
	}

	return StateRunning
}

// GetInfo returns full Claude info for a session path
func GetInfo(sessionPath string) *Info {
	info := &Info{
		State: GetState(sessionPath),
	}

	pid, err := GetProcessPID(sessionPath)
	if err == nil {
		info.PID = pid
	}

	// TODO: Parse .claude directory for token usage and last prompt
	// This requires understanding Claude's session storage format

	return info
}

// GetProcessPID finds the Claude process running in the given session path
func GetProcessPID(sessionPath string) (int, error) {
	// Use ps to find claude processes
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return 0, err
	}

	absPath, _ := filepath.Abs(sessionPath)

	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, "claude") {
			continue
		}
		// Skip grep processes
		if strings.Contains(line, "grep") || strings.Contains(line, "ps aux") {
			continue
		}

		// Parse the PID (second field)
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		// Check if this process's cwd matches our session path
		cwd := getProcessCwd(pid)
		if cwd == absPath {
			return pid, nil
		}
	}

	return 0, nil
}

// getProcessCwd gets the current working directory of a process
func getProcessCwd(pid int) string {
	// On macOS, use lsof
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "n") {
			return line[1:]
		}
	}
	return ""
}

// isWaitingForInput checks if a process is waiting for input
func isWaitingForInput(pid int) bool {
	// Check process state - if it's in interruptible sleep (S), likely waiting
	// This is a simplified heuristic
	out, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return false
	}
	state := strings.TrimSpace(string(out))
	return strings.HasPrefix(state, "S")
}

// StopProcess stops a Claude process
func StopProcess(sessionPath string) error {
	pid, err := GetProcessPID(sessionPath)
	if err != nil {
		return err
	}
	if pid == 0 {
		return nil // No process to stop
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(syscall.SIGTERM)
}

// StartProcess starts Claude in the given session path with optional args
func StartProcess(sessionPath string, args []string) error {
	cmd := exec.Command("claude", args...)
	cmd.Dir = sessionPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}
