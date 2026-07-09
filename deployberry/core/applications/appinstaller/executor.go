package appinstaller

import (
	"deployberry/core/applications/appinstaller/manifest"
	"fmt"
	"log"
	"os/exec"
	"shared/shell"
	"strings"
	"time"
)

// StepResult tracks the output and status of each executed command
type StepResult struct {
	Index  int
	Name   string
	Output string
	Error  error
}

// ─────────────────────────────────────────
// MAIN EXECUTION FLOWS
// ─────────────────────────────────────────

// RunInstall executes the full installation lifecycle defined in the manifest.
func RunInstall(m *manifest.Manifest, vars map[string]string) ([]StepResult, error) {
	var results []StepResult
	appName := vars["APP_NAME"]
	log.Printf("[executor] Installing app %s...", appName)
	installStart := time.Now()

	// 1. Install System Packages (Cross-platform)
	pkgStart := time.Now()
	if err := InstallSystemPackages(m.SystemPackages); err != nil {
		log.Printf("[executor] Stage 1/6 (Packages) failed after %v: %v", time.Since(pkgStart), err)
		return nil, fmt.Errorf("system packages failed: %w", err)
	}
	log.Printf("[executor] Stage 1/6 (Packages) completed in %v", time.Since(pkgStart))

	// 2. Acquire Tools
	toolsStart := time.Now()
	if err := AcquireTools(m.Tools); err != nil {
		log.Printf("[executor] Stage 2/6 (Tools) failed after %v: %v", time.Since(toolsStart), err)
		return nil, fmt.Errorf("tool acquisition failed: %w", err)
	}
	log.Printf("[executor] Stage 2/6 (Tools) completed in %v", time.Since(toolsStart))

	// 3. Deployment
	if m.Deploy.Default != "" || m.Deploy.Git != nil || m.Deploy.Upload || m.Deploy.Manual {
		deployStart := time.Now()
		selectedMethod := vars["DEPLOYMENT_METHOD"]
		if err := RunDeployment(m, vars, selectedMethod); err != nil {
			log.Printf("[executor] Stage 3/6 (Deploy) failed after %v: %v", time.Since(deployStart), err)
			return nil, fmt.Errorf("deployment failed: %w", err)
		}
		// Ensure the app directory and all deployed files are owned by the restricted app user
		appPath := vars["APP_PATH"]
		if appPath != "" {
			shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(appPath)))
		}
		log.Printf("[executor] Stage 3/6 (Deploy: %s) completed in %v", selectedMethod, time.Since(deployStart))
	}

	// 4. Pre-Install Hook
	if m.Hooks.PreInstall != "" {
		hookStart := time.Now()
		out, err := ExecWithSubstitute(m.Hooks.PreInstall, vars)
		results = append(results, StepResult{Name: "Pre-install Hook", Output: out, Error: err})
		if err != nil {
			log.Printf("[executor] Stage 4/6 (Pre-Install Hook) failed after %v: %v", time.Since(hookStart), err)
			return results, err
		}
		log.Printf("[executor] Stage 4/6 (Pre-Install Hook) completed in %v", time.Since(hookStart))
	}

	// 5. Install Steps
	if len(m.Install) > 0 {
		stepsStart := time.Now()
		for i, step := range m.Install {
			stepStart := time.Now()
			out, err := ExecWithSubstitute(step.Run, vars)
			results = append(results, StepResult{Index: i, Name: step.Name, Output: out, Error: err})
			if err != nil {
				log.Printf("[executor] Stage 5/6 (Step %d/%d: %s) failed after %v: %v", i+1, len(m.Install), step.Name, time.Since(stepStart), err)
				return results, fmt.Errorf("step '%s' failed: %w", step.Name, err)
			}
			log.Printf("[executor] Stage 5/6 (Step %d/%d: %s) completed in %v", i+1, len(m.Install), step.Name, time.Since(stepStart))
		}
		log.Printf("[executor] Stage 5/6 (All Steps) completed in %v", time.Since(stepsStart))
	}

	// 6. Post-Install Hook
	if m.Hooks.PostInstall != "" {
		hookStart := time.Now()
		out, err := ExecWithSubstitute(m.Hooks.PostInstall, vars)
		results = append(results, StepResult{Name: "Post-install Hook", Output: out, Error: err})
		if err != nil {
			log.Printf("[executor] Stage 6/6 (Post-Install Hook) failed after %v: %v", time.Since(hookStart), err)
			return results, err
		}
		log.Printf("[executor] Stage 6/6 (Post-Install Hook) completed in %v", time.Since(hookStart))
	}

	log.Printf("[executor] App %s installed successfully in %v", appName, time.Since(installStart))
	return results, nil
}

// RunCommand executes a specific command from the manifest's commands block.
func RunCommand(m *manifest.Manifest, cmdKey string, vars map[string]string) (string, error) {
	cmd, exists := m.Commands[cmdKey]
	if !exists {
		return "", fmt.Errorf("command '%s' not found in manifest", cmdKey)
	}
	return ExecWithSubstitute(cmd.Run, vars)
}

// ─────────────────────────────────────────
// UPDATE & DELETE FLOWS
// ─────────────────────────────────────────

// RunUpdate executes the update lifecycle defined in the manifest.
func RunUpdate(m *manifest.Manifest, vars map[string]string) ([]StepResult, error) {
	var results []StepResult

	if m.Hooks.PreUpdate != "" {
		out, err := ExecWithSubstitute(m.Hooks.PreUpdate, vars)
		results = append(results, StepResult{Name: "Pre-update Hook", Output: out, Error: err})
		if err != nil {
			return results, err
		}
	}

	for i, step := range m.Update {
		out, err := ExecWithSubstitute(step.Run, vars)
		results = append(results, StepResult{Index: i, Name: step.Name, Output: out, Error: err})
		if err != nil {
			return results, fmt.Errorf("update step '%s' failed: %w", step.Name, err)
		}
	}

	if m.Hooks.PostUpdate != "" {
		out, err := ExecWithSubstitute(m.Hooks.PostUpdate, vars)
		results = append(results, StepResult{Name: "Post-update Hook", Output: out, Error: err})
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// RunDelete executes the deletion hooks and removes files if specified.
func RunDelete(m *manifest.Manifest, vars map[string]string) error {
	// 1. Run Pre-Delete Hook (e.g., stopping a custom daemon)
	if m.Delete.PreDelete != "" {
		_, err := ExecWithSubstitute(m.Delete.PreDelete, vars)
		if err != nil {
			return fmt.Errorf("pre_delete hook failed: %w", err)
		}
	}

	// 2. Remove files via shell if requested
	if m.Delete.RemoveFiles {
		appPath := vars["APP_PATH"]
		if appPath != "" && appPath != "/" { // Safety check
			res := shell.ExecuteCommand(fmt.Sprintf("rm -rf %s", shell.EscapeShellArg(appPath)))
			if res.Error != nil {
				return fmt.Errorf("failed to remove files: %w", res.Error)
			}
		}
	}

	// Note: DropDatabase and RemoveWebserverConfig are handled by the Go handler,
	// because DeployBerry tracks these in its own SQLite DB and configs.
	return nil
}

// ─────────────────────────────────────────
// SYSTEM PACKAGE MANAGER (Cross-Platform)
// ─────────────────────────────────────────

// InstallSystemPackages detects the OS package manager and installs required dependencies.
func InstallSystemPackages(packages []manifest.SystemPackage) error {
	if len(packages) == 0 {
		return nil
	}

	pkgManager, installCmd := detectPackageManager()
	if pkgManager == "" {
		return fmt.Errorf("no supported package manager found (apt/dnf)")
	}

	var toInstall []string
	for _, pkg := range packages {
		// Resolve the correct package name for the detected OS
		name := pkg.ResolveForOS(pkgManager)
		if name != "" {
			toInstall = append(toInstall, name)
		}
	}

	if len(toInstall) == 0 {
		return nil
	}

	startTime := time.Now()

	// Example: apt-get install -y libvips-dev
	fullCmd := fmt.Sprintf("%s %s", installCmd, strings.Join(toInstall, " "))
	res := shell.ExecuteCommand(fullCmd)
	if res.Error != nil {
		log.Printf("[executor] Installing packages [%s] failed after %v: %v", strings.Join(toInstall, ", "), time.Since(startTime), res.Error)
		return fmt.Errorf("failed to install packages [%s]: %v\nOutput: %s", strings.Join(toInstall, ", "), res.Error, res.Output)
	}

	log.Printf("[executor] Installed packages [%s] in %v", strings.Join(toInstall, ", "), time.Since(startTime))

	// If redis-server or redis was in the installed package list, try to enable and start it
	for _, pkgName := range toInstall {
		if pkgName == "redis-server" || pkgName == "redis" {
			shell.ExecuteCommand("systemctl enable redis-server && systemctl start redis-server || service redis-server start || systemctl enable redis && systemctl start redis || service redis start")
		}
	}

	return nil
}

func detectPackageManager() (string, string) {
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "apt", "DEBIAN_FRONTEND=noninteractive apt-get -y -o Dpkg::Options::=\"--force-confdef\" -o Dpkg::Options::=\"--force-confold\" install"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf", "dnf install -y"
	}
	return "", ""
}

// ─────────────────────────────────────────
// TOOL ACQUISITION
// ─────────────────────────────────────────

// AcquireTools checks for required binaries and fetches them if missing.
func AcquireTools(tools []manifest.Tool) error {
	for _, tool := range tools {
		// 1. Check if it's already installed via the Verify command
		if tool.Verify != "" {
			if res := shell.ExecuteCommand(tool.Verify); res.Error == nil {
				continue // Tool is already present and working
			}
		} else {
			// Fallback: just check if the binary exists in PATH
			if _, err := exec.LookPath(tool.Name); err == nil {
				continue
			}
		}

		// 2. Not found, we need to install it based on the method provided
		var installCmd string
		switch {
		case tool.Url != "":
			dest := tool.Dest
			if dest == "" {
				dest = "/usr/local/bin/" + tool.Name
			}
			installCmd = fmt.Sprintf("curl -fsSL %s -o %s && chmod +x %s", tool.Url, dest, dest)
		case tool.Npm != "":
			installCmd = fmt.Sprintf("npm install -g %s", tool.Npm)
		case tool.Pip != "":
			installCmd = fmt.Sprintf("pip3 install %s || pip3 install --break-system-packages %s", tool.Pip, tool.Pip)
		case tool.Composer != "":
			installCmd = fmt.Sprintf("composer global require %s", tool.Composer)
		case tool.Run != "":
			installCmd = tool.Run
		default:
			// If tool is missing and no acquisition method was provided
			return fmt.Errorf("tool '%s' is missing and no acquisition method was provided in the manifest", tool.Name)
		}

		log.Printf("[executor] Installing tool '%s'...", tool.Name)
		startTime := time.Now()
		res := shell.ExecuteCommand(installCmd)
		if res.Error != nil {
			log.Printf("[executor] Tool '%s' installation failed after %v: %v", tool.Name, time.Since(startTime), res.Error)
			return fmt.Errorf("failed to acquire tool '%s': %v\nOutput: %s", tool.Name, res.Error, res.Output)
		}
		log.Printf("[executor] Tool '%s' installed in %v", tool.Name, time.Since(startTime))
	}
	return nil
}

// ─────────────────────────────────────────
// UTILITIES
// ─────────────────────────────────────────


// ExecWithSubstitute handles variable substitution and runs the shell command.
func ExecWithSubstitute(cmdStr string, vars map[string]string) (string, error) {
	cmdStr = shell.SubstituteVars(cmdStr, vars)
	res := shell.ExecuteAsAppUser(cmdStr, vars["APP_PATH"])
	if res.Error != nil {
		return res.Output, res.Error
	}
	return res.Output, nil
}
