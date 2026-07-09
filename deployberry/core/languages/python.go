package languages

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"shared/shell"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

// Linux-only utilities for managing Python with pyenv

// pyenvCommandWithEnv runs a pyenv command with environment prepared
func pyenvCommandWithEnv(args ...string) (string, error) {
	cmd := exec.Command("pyenv", args...)

	homeDir, _ := os.UserHomeDir()
	pyenvRoot := filepath.Join(homeDir, ".pyenv")
	bin := filepath.Join(pyenvRoot, "bin")

	env := append(os.Environ(),
		"PYENV_ROOT="+pyenvRoot,
		"PATH="+bin+":"+os.Getenv("PATH"),
	)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ensurePyenvInstalled checks for pyenv and installs it if missing (Linux only)
func ensurePyenvInstalled() error {
	// Check PATH first
	homeDir, _ := os.UserHomeDir()
	pyenvBin := filepath.Join(homeDir, ".pyenv", "bin")
	if _, err := exec.LookPath("pyenv"); err == nil {
		// Always ensure pyenv bin is in PATH for current process
		os.Setenv("PATH", pyenvBin+":"+os.Getenv("PATH"))
		return nil
	}
	// Check default install dir
	if _, statErr := os.Stat(filepath.Join(pyenvBin, "pyenv")); statErr == nil {
		os.Setenv("PATH", pyenvBin+":"+os.Getenv("PATH"))
		return nil
	}
	// Install if missing
	if runtime.GOOS != "linux" {
		return fmt.Errorf("pyenv auto-installation is only supported on Linux servers. Please install pyenv manually on %s", runtime.GOOS)
	}
	installScript := `#!/bin/bash
set -e
curl -sSL https://pyenv.run | bash
`
	cmd := exec.Command("bash", "-c", installScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install pyenv: %v, output: %s", err, string(out))
	}
	// After install, update PATH for current process so pyenv is immediately available
	os.Setenv("PATH", pyenvBin+":"+os.Getenv("PATH"))
	println("pyenv installed", cmd.String())
	return nil
}

// pythonPathForVersion returns the absolute python executable path under pyenv for a given version
func pythonPathForVersion(version string) string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".pyenv", "versions", version, "bin", "python")
}

// Create a virtual environment using a specific Python version
// Query params: version, path (absolute path to venv directory)
func CreatePythonVenv(c *gin.Context) {
	version := c.Query("version")
	venvPath := c.Query("path")
	if version == "" || venvPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both version and path parameters are required"})
		return
	}
	if err := ensurePyenvInstalled(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to install pyenv: " + err.Error()})
		return
	}
	// Ensure version is installed
	if _, err := os.Stat(pythonPathForVersion(version)); os.IsNotExist(err) {
		if _, err := pyenvCommandWithEnv("install", "-s", version); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to install Python " + version + ": " + err.Error()})
			return
		}
	}

	pythonExec := pythonPathForVersion(version)
	if _, err := os.Stat(pythonExec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Python executable not found for version " + version})
		return
	}

	if err := os.MkdirAll(venvPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create venv directory: " + err.Error()})
		return
	}

	// Create venv
	cmd := exec.Command(pythonExec, "-m", "venv", venvPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create venv: " + err.Error(), "output": string(out)})
		return
	}

	// Ensure the virtual environment is owned by panel_apps:www-data
	shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(venvPath)))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Venv created", "path": venvPath, "version": version})
}
