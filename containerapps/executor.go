package containerapps

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"containerapps/manifest"
	"shared/files"
	"shared/repository"
	"shared/shell"
)

// StepResult tracks the output and status of each executed command
type StepResult struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// GetFreePort finds an available TCP port on the host in the range 9000-10000
func GetFreePort() (int, error) {
	db := repository.GetDB()
	for port := 9000; port <= 10000; port++ {
		// 1. Check GORM Database first to prevent conflicts with other configured sites
		var count int64
		db.Model(&repository.ContainerApp{}).Where("host_port = ?", port).Count(&count)
		if count > 0 {
			continue
		}

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue // Port is in use
		}
		l.Close()
		return port, nil
	}
	return 0, fmt.Errorf("no free ports available in range 9000-10000")
}


// RunComposeUp writes the compose file and starts the services via podman-compose
func RunComposeUp(appPath string, composeContent string) error {
	composePath := filepath.Join(appPath, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	cmd := fmt.Sprintf("cd %s && podman-compose up -d", shell.EscapeShellArg(appPath))
	res := shell.ExecuteCommand(cmd)
	if !res.Success {
		return fmt.Errorf("podman-compose up failed: %v\nOutput: %s", res.Error, res.Output)
	}

	return nil
}

// ExecuteInContainer executes a command inside the primary container of the compose stack
func ExecuteInContainer(appPath string, serviceName string, command string) shell.BashResult {
	cmdStr := fmt.Sprintf("cd %s && podman-compose exec -T %s sh -c %s",
		shell.EscapeShellArg(appPath),
		shell.EscapeShellArg(serviceName),
		shell.EscapeShellArg(command),
	)
	return shell.ExecuteCommand(cmdStr)
}

// ExecHostWithSubstitute runs a host command after replacing variables
func ExecHostWithSubstitute(cmdStr string, vars map[string]string) (string, error) {
	cmdStr = shell.SubstituteVars(cmdStr, vars)
	res := shell.ExecuteCommand(cmdStr)
	if !res.Success {
		return res.Output, res.Error
	}
	return res.Output, nil
}

// RunSourceDeployment fetches the application source code into the APP_PATH before building the container
func RunSourceDeployment(method string, vars map[string]string) error {
	appPath := vars["APP_PATH"]
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	if method != "git" {
		return fmt.Errorf("unsupported deployment method: %s", method)
	}

	source := vars["GIT_REPO"]
	if source == "" {
		return fmt.Errorf("git repository URL not configured")
	}
	branch := vars["GIT_BRANCH"]

	// If it's already a cloned git repository
	if _, err := os.Stat(filepath.Join(appPath, ".git")); err == nil {
		result := files.GitPull(files.GitOperation{Path: appPath})
		if !result.Success {
			return fmt.Errorf("git pull failed: %s", result.Error)
		}
		return nil
	}

	args := []string{"-c", "safe.directory=*", "clone", "--single-branch"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, source, appPath)

	cmd := exec.Command("git", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	return nil
}

// RunInstall executes the compose installation and steps
func RunInstall(m *manifest.Manifest, vars map[string]string, containerName string) ([]StepResult, error) {
	var results []StepResult
	appPath := vars["APP_PATH"]

	// 1. Fetch source repository if configured
	deployMethod := vars["DEPLOYMENT_METHOD"]
	if deployMethod == "git" {
		if err := RunSourceDeployment("git", vars); err != nil {
			return results, fmt.Errorf("source deployment failed: %w", err)
		}
	}

	// 2. Write docker-compose.yml and start container stack
	composeContent := shell.SubstituteVars(m.ComposeTemplate, vars)
	if err := RunComposeUp(appPath, composeContent); err != nil {
		return results, fmt.Errorf("failed to start compose stack: %w", err)
	}

	// Wait 5 seconds for containers to initialize and start their internal services
	time.Sleep(5 * time.Second)

	// 3. Install Steps
	for i, step := range m.Install {
		stepRunSubbed := shell.SubstituteVars(step.Run, vars)
		var output string
		var err error

		if step.RunOnHost {
			output, err = ExecHostWithSubstitute(stepRunSubbed, vars)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			results = append(results, StepResult{
				Index:  i,
				Name:   step.Name,
				Output: output,
				Error:  errMsg,
			})
			if err != nil {
				return results, fmt.Errorf("step '%s' failed on host: %w", step.Name, err)
			}
		} else {
			targetService := containerName
			if step.Service != "" {
				targetService = step.Service
			}
			res := ExecuteInContainer(appPath, targetService, stepRunSubbed)
			var errMsg string
			if res.Error != nil {
				errMsg = res.Error.Error()
			}
			results = append(results, StepResult{
				Index:  i,
				Name:   step.Name,
				Output: res.Output,
				Error:  errMsg,
			})
			if !res.Success {
				return results, fmt.Errorf("step '%s' failed inside container service '%s': %v", step.Name, targetService, res.Error)
			}
		}
	}

	return results, nil
}

// RunUpdate executes update lifecycle inside compose stack
func RunUpdate(m *manifest.Manifest, vars map[string]string, containerName string) ([]StepResult, error) {
	var results []StepResult
	appPath := vars["APP_PATH"]

	// Pull new images and restart container stack
	cmd := fmt.Sprintf("cd %s && podman-compose down && podman-compose pull && podman-compose up -d --build", shell.EscapeShellArg(appPath))
	res := shell.ExecuteCommand(cmd)
	if !res.Success {
		return results, fmt.Errorf("failed to pull/restart containers: %v\nOutput: %s", res.Error, res.Output)
	}

	time.Sleep(5 * time.Second)

	for i, step := range m.Update {
		stepRunSubbed := shell.SubstituteVars(step.Run, vars)
		var output string
		var err error

		if step.RunOnHost {
			output, err = ExecHostWithSubstitute(stepRunSubbed, vars)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			results = append(results, StepResult{
				Index:  i,
				Name:   step.Name,
				Output: output,
				Error:  errMsg,
			})
			if err != nil {
				return results, fmt.Errorf("update step '%s' failed on host: %w", step.Name, err)
			}
		} else {
			targetService := containerName
			if step.Service != "" {
				targetService = step.Service
			}
			res := ExecuteInContainer(appPath, targetService, stepRunSubbed)
			var errMsg string
			if res.Error != nil {
				errMsg = res.Error.Error()
			}
			results = append(results, StepResult{
				Index:  i,
				Name:   step.Name,
				Output: res.Output,
				Error:  errMsg,
			})
			if !res.Success {
				return results, fmt.Errorf("update step '%s' failed inside container service '%s': %v", step.Name, targetService, res.Error)
			}
		}
	}

	return results, nil
}

// RunDelete shuts down podman-compose services and cleans up local directories
func RunDelete(m *manifest.Manifest, vars map[string]string, containerName string) error {
	appPath := vars["APP_PATH"]

	// 1. Shut down compose stack and drop volumes
	cmd := fmt.Sprintf("cd %s && podman-compose down -v", shell.EscapeShellArg(appPath))
	shell.ExecuteCommand(cmd)

	// 2. Remove files
	if appPath != "" && appPath != "/" {
		shell.ExecuteCommand(fmt.Sprintf("rm -rf %s", shell.EscapeShellArg(appPath)))
	}

	return nil
}
