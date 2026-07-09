package common

import (
	"fmt"
	"net/http"
	"os"
	"shared/shell"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// InstallMethod represents the installation method data
type InstallMethod struct {
	Type         string `form:"install_type" binding:"required"` // "git", "zip", or "existing", or new "Registry"
	GitRepo      string `form:"git_repo,omitempty"`
	GitBranch    string `form:"git_branch,omitempty"`
	ZipFile      string `form:"zip_file,omitempty"`
	ExistingPath string `form:"existing_path,omitempty"`
	RegistrySlug string `form:"registry_slug"`
}

// InstallResult represents the result of an installation method
type InstallResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
	Output  string `json:"output,omitempty"`
}

// HandleGitInstallation handles Git repository installation
func HandleGitInstallation(installPath, gitRepo, gitBranch string) InstallResult {
	if gitRepo == "" {
		return InstallResult{
			Success: false,
			Error:   "Git repository URL is required",
		}
	}

	// Set default branch if not provided
	if gitBranch == "" {
		gitBranch = "main"
	}

	// Create installation directory
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return InstallResult{
			Success: false,
			Error:   "Failed to create installation directory",
			Details: err.Error(),
		}
	}

	// Clone repository
	repoPath := filepath.Join(installPath, "repo")
	result := shell.ExecuteCommand(fmt.Sprintf(
		"git clone -b %s %s %s",
		shell.EscapeShellArg(gitBranch),
		shell.EscapeShellArg(gitRepo),
		shell.EscapeShellArg(repoPath),
	))

	if result.Error != nil {
		return InstallResult{
			Success: false,
			Error:   "Failed to clone repository",
			Details: result.Error.Error(),
			Output:  result.Output,
		}
	}

	return InstallResult{
		Success: true,
		Path:    repoPath,
	}
}

// HandleZipInstallation handles zip file installation
func HandleZipInstallation(installPath, tempZipPath string) InstallResult {
	// Create installation directory
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return InstallResult{
			Success: false,
			Error:   "Failed to create installation directory",
			Details: err.Error(),
		}
	}

	// Extract zip file
	result := shell.ExecuteCommand(fmt.Sprintf(
		"cd %s && unzip -o %s && rm %s",
		shell.EscapeShellArg(installPath),
		shell.EscapeShellArg(tempZipPath),
		shell.EscapeShellArg(tempZipPath),
	))

	if result.Error != nil {
		return InstallResult{
			Success: false,
			Error:   "Failed to extract zip file",
			Details: result.Error.Error(),
			Output:  result.Output,
		}
	}

	return InstallResult{
		Success: true,
		Path:    installPath,
	}
}

// HandleExistingProjectInstallation handles existing project installation
func HandleExistingProjectInstallation(existingPath string) InstallResult {
	if existingPath == "" {
		return InstallResult{
			Success: false,
			Error:   "Existing project path is required",
		}
	}

	// Check if the directory exists
	if _, err := os.Stat(existingPath); os.IsNotExist(err) {
		return InstallResult{
			Success: false,
			Error:   "Project directory does not exist",
			Details: err.Error(),
		}
	}

	return InstallResult{
		Success: true,
		Path:    existingPath,
	}
}

// ProcessInstallationMethod processes the installation method and returns the result
func ProcessInstallationMethod(method InstallMethod, installPath, tempZipPath string) InstallResult {
	switch method.Type {
	case "git":
		return HandleGitInstallation(installPath, method.GitRepo, method.GitBranch)
	case "zip":
		return HandleZipInstallation(installPath, tempZipPath)
	case "existing":
		return HandleExistingProjectInstallation(method.ExistingPath)
	default:
		return InstallResult{
			Success: false,
			Error:   "Invalid installation method",
		}
	}
}

// SendInstallResult sends the installation result as JSON response
func SendInstallResult(c *gin.Context, result InstallResult) {
	if result.Success {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"path":    result.Path,
			"message": "Installation completed successfully",
		})
	} else {
		statusCode := http.StatusInternalServerError
		if result.Error == "Git repository URL is required" ||
			result.Error == "Zip file is required for zip installation" ||
			result.Error == "Existing project path is required" ||
			result.Error == "Invalid installation method" {
			statusCode = http.StatusBadRequest
		}

		response := gin.H{
			"success": false,
			"error":   result.Error,
		}

		if result.Details != "" {
			response["details"] = result.Details
		}
		if result.Output != "" {
			response["output"] = result.Output
		}

		c.JSON(statusCode, response)
	}
}
