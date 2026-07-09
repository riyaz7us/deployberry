package appinstaller

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"deployberry/core/applications/appinstaller/manifest"
	"shared/files"
	"shared/shell"
)

// RunDeployment handles the deployment phase of installation
func RunDeployment(m *manifest.Manifest, vars map[string]string, selectedMethod string) error {
	// Check if deployment is configured
	if m.Deploy.Default == "" && m.Deploy.Git == nil && !m.Deploy.Upload && !m.Deploy.Manual {
		return nil // No deployment specified
	}

	// Determine deployment method
	method := selectedMethod
	if method == "" {
		method = m.Deploy.Default // Use default if no method selected
	}

	appPath := vars["APP_PATH"]

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Determine full deployment path
	deployPath := appPath
	if m.Deploy.Destination != "" {
		deployPath = filepath.Join(appPath, m.Deploy.Destination)
		if err := os.MkdirAll(deployPath, 0755); err != nil {
			return fmt.Errorf("failed to create deployment directory: %w", err)
		}
	}

	log.Printf("[deployment] Starting deployment phase using method: %s, path: %s", method, deployPath)
	startTime := time.Now()

	var err error
	switch method {
	case "git":
		err = deployGitMulti(m.Deploy, deployPath, vars)
	case "upload":
		err = deployUploadMulti(m.Deploy, deployPath, vars)
	case "manual":
		err = deployManualMulti(m.Deploy, deployPath, vars)
	default:
		err = fmt.Errorf("unsupported deployment method: %s", method)
	}

	if err != nil {
		log.Printf("[deployment] Deploy (%s) failed after %v: %v", method, time.Since(startTime), err)
		return err
	}

	log.Printf("[deployment] Deploy (%s) completed in %v", method, time.Since(startTime))
	return nil
}

// deployGitMulti handles git deployment from new multi-method structure
func deployGitMulti(deploy manifest.Deploy, deployPath string, vars map[string]string) error {
	source := vars["GIT_REPO"]
	if source == "" {
		if deploy.Git == nil || deploy.Git.Source == "" {
			return fmt.Errorf("git deployment source not configured")
		}
		source = shell.SubstituteVars(deploy.Git.Source, vars)
	}

	branch := vars["GIT_BRANCH"]
	if branch == "" && deploy.Git != nil {
		branch = shell.SubstituteVars(deploy.Git.Branch, vars)
	}

	// No redundant source declaration
	// Check if directory is already a git repo
	if _, err := os.Stat(filepath.Join(deployPath, ".git")); err == nil {
		// Repository exists, pull latest changes
		result := files.GitPull(files.GitOperation{Path: deployPath})
		if !result.Success {
			return fmt.Errorf("git pull failed: %s", result.Error)
		}
		return nil
	}

	// Fresh clone
	args := []string{"-c", "safe.directory=*", "clone", "--single-branch"}
	if branch != "" {
		args = append(args, "-b", branch)
	}

	if deploy.Git != nil && deploy.Git.Blobless {
		args = append(args, "--filter=blob:none")
	}

	args = append(args, source, deployPath)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}
	return nil
}

// deployUploadMulti handles upload deployment with folder/zip support
func deployUploadMulti(deploy manifest.Deploy, deployPath string, vars map[string]string) error {
	if !deploy.Upload {
		return fmt.Errorf("upload deployment not enabled")
	}

	// Get uploaded file path from variables (set by UI after upload)
	uploadPath := vars["UPLOAD_PATH"]
	if uploadPath == "" {
		return fmt.Errorf("upload path not specified")
	}

	// Check if it's an archive or directory
	info, err := os.Stat(uploadPath)
	if err != nil {
		return fmt.Errorf("cannot access upload path: %w", err)
	}

	if info.IsDir() {
		// Copy directory
		cmd := exec.Command("cp", "-r", uploadPath+"/.", deployPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("copy failed: %w, output: %s", err, string(output))
		}
	} else {
		// Extract archive
		ext := strings.ToLower(filepath.Ext(uploadPath))
		var cmd *exec.Cmd

		switch ext {
		case ".zip":
			cmd = exec.Command("unzip", "-o", uploadPath, "-d", deployPath)
		case ".tar", ".gz", ".tgz":
			cmd = exec.Command("tar", "-xzf", uploadPath, "-C", deployPath)
		default:
			return fmt.Errorf("unsupported archive format: %s", ext)
		}

		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("extraction failed: %w, output: %s", err, string(output))
		}
	}

	return nil
}

// deployManualMulti handles manual path deployment from new structure
func deployManualMulti(deploy manifest.Deploy, deployPath string, vars map[string]string) error {
	if !deploy.Manual {
		return fmt.Errorf("manual deployment not enabled")
	}

	// Manual deployment - use user path directly as app directory
	// No copying needed - the user's provided path IS the app directory
	// DeployBerry will manage the existing code in-place
	return nil
}

// GetDeploymentMethods returns deployment methods info for frontend
func GetDeploymentMethods(m *manifest.Manifest) map[string]interface{} {
	result := make(map[string]interface{})

	// Available methods
	available := make(map[string]bool)
	if m.Deploy.Git != nil {
		available["git"] = true
	}
	if m.Deploy.Upload {
		available["upload"] = true
	}
	if m.Deploy.Manual {
		available["manual"] = true
	}
	result["available"] = available

	// Required flag
	required := true // default to true if methods are present
	if m.Deploy.Required != nil {
		required = *m.Deploy.Required
	}
	result["required"] = required

	// Default method
	if m.Deploy.Default != "" {
		result["default"] = m.Deploy.Default
	}

	// Git config
	if m.Deploy.Git != nil {
		result["git"] = map[string]string{
			"source": m.Deploy.Git.Source,
			"branch": m.Deploy.Git.Branch,
		}
	}

	return result
}
