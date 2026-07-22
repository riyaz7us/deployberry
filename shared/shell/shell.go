package shell

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"shared/globals"
	"strings"
	"sync"
)

// BashResult represents the result of a bash command execution
type BashResult struct {
	Output   string    `json:"output"`
	ExitCode int       `json:"exit_code"`
	Success  bool      `json:"success"`
	Error    error     `json:"-"`
}

// ExecuteBashOptions contains options for command execution
type ExecuteBashOptions struct {
	StreamOutput bool // If true, streams output to stdout/stderr in real-time
}

type safeWriter struct {
	mu     sync.Mutex
	buf    bytes.Buffer
	stream bool
	isErr  bool
}

func (w *safeWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.stream {
		if w.isErr {
			_, _ = os.Stderr.Write(p)
		} else {
			_, _ = os.Stdout.Write(p)
		}
	}

	return w.buf.Write(p)
}

func (w *safeWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// ExecuteBash executes a bash command or script and returns the result
// If isScript is true, the command will be treated as a script path with arguments
// If isScript is false, the command will be executed directly as a bash command
func ExecuteBash(command string, isScript bool, options ExecuteBashOptions, args ...string) BashResult {
	var cmd *exec.Cmd

	// Ensures all version managers (NVM, Go, Pyenv) are loaded natively
	envLoader := "source " + globals.ENV_PATH + " 2>/dev/null; "

	if isScript {
		// For scripts, load env then run script
		fullCmd := envLoader + command + " " + strings.Join(args, " ")
		cmd = exec.Command("bash", "-c", fullCmd)
	} else {
		// For commands, load env then execute directly
		fullCmd := envLoader + command
		cmd = exec.Command("bash", "-c", fullCmd)
	}

	stdoutWriter := &safeWriter{stream: options.StreamOutput, isErr: false}
	stderrWriter := &safeWriter{stream: options.StreamOutput, isErr: true}
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err := cmd.Run()

	output := strings.TrimSpace(stdoutWriter.String())
	errOutput := strings.TrimSpace(stderrWriter.String())

	if err != nil && errOutput != "" {
		if output != "" {
			output = fmt.Sprintf("%s\nError: %s", output, errOutput)
		} else {
			output = fmt.Sprintf("Error: %s", errOutput)
		}
	}

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return BashResult{
		Output:   output,
		ExitCode: exitCode,
		Success:  err == nil,
		Error:    err,
	}
}

// ExecuteScript is a convenience wrapper for executing shell scripts
func ExecuteScript(scriptPath string, args ...string) BashResult {
	return ExecuteBash(scriptPath, true, ExecuteBashOptions{}, args...)
}

// ExecuteScriptWithOutput is a convenience wrapper for executing shell scripts with output streaming
func ExecuteScriptWithOutput(scriptPath string, args ...string) BashResult {
	return ExecuteBash(scriptPath, true, ExecuteBashOptions{StreamOutput: true}, args...)
}

// GetCommandVersion checks the version of a command/tool by executing it.
func GetCommandVersion(cmd string) (string, bool) {
	res := ExecuteCommand(cmd)
	if !res.Success {
		return "", false
	}
	v := strings.TrimSpace(res.Output)
	if v == "" || strings.Contains(v, "not found") || strings.Contains(v, "is not recognized") {
		return "", false
	}
	return v, true
}

// ExecuteCommand is a convenience wrapper for executing direct bash commands
func ExecuteCommand(command string) BashResult {
	return ExecuteBash(command, false, ExecuteBashOptions{})
}

// ExecuteCommandWithOutput is a convenience wrapper for executing direct bash commands with output streaming
func ExecuteCommandWithOutput(command string) BashResult {
	return ExecuteBash(command, false, ExecuteBashOptions{StreamOutput: true})
}

// EscapeShellArg escapes a string to be safely used as a command line argument
func EscapeShellArg(arg string) string {
	if arg == "" {
		return "''"
	}
	// Replace single quotes with '\'' and wrap in single quotes
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}

// ExecuteAsAppUser runs a command as the panel_apps user
func ExecuteAsAppUser(command string, appPath string) BashResult {
	envLoader := "source " + globals.ENV_PATH + " 2>/dev/null"

	var fullCmd string
	if appPath != "" && appPath != "." {
		fullCmd = fmt.Sprintf("%s && cd %s && %s", envLoader, EscapeShellArg(appPath), command)
	} else {
		fullCmd = fmt.Sprintf("%s && %s", envLoader, command)
	}

	sudoCmd := fmt.Sprintf("sudo -u panel_apps -H bash -c %s", EscapeShellArg(fullCmd))
	return ExecuteBash(sudoCmd, false, ExecuteBashOptions{})
}

// IsServiceActive checks if a systemd service is currently active.
func IsServiceActive(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)
	err := cmd.Run()
	return err == nil
}

// EscapeShellArgs escapes a list of strings to be safely used as command line arguments
func EscapeShellArgs(args []string) []string {
	escaped := make([]string, len(args))
	for i, arg := range args {
		escaped[i] = EscapeShellArg(arg)
	}
	return escaped
}

// SubstituteVars replaces placeholder variables like {APP_PATH} or {HOST_PORT} with their values
func SubstituteVars(content string, vars map[string]string) string {
	for k, v := range vars {
		placeholder := fmt.Sprintf("{%s}", k)
		content = strings.ReplaceAll(content, placeholder, v)
	}
	return content
}
