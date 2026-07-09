package installer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"shared/system/manifest"
	"regexp"
	"strconv"
	"strings"
)

// SystemStepResult tracks the outcome and details of executing a manifest action step
type SystemStepResult struct {
	Label  string `json:"label"`
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Runner is responsible for executing the steps defined in a SystemManifest.
type Runner struct {
	PackageManager string // "apt" or "dnf"
}

// NewRunner creates a new installer runner by detecting the OS package manager.
func NewRunner() *Runner {
	pm := "apt"
	if _, err := exec.LookPath("dnf"); err == nil {
		pm = "dnf"
	}
	return &Runner{
		PackageManager: pm,
	}
}

// RunAction executes a specific action block defined in the manifest.
func (r *Runner) RunAction(m *manifest.SystemManifest, actionName string) ([]SystemStepResult, error) {
	var results []SystemStepResult

	// Execute global dependencies first if this is an install action
	if actionName == "install" {
		if len(m.Repositories) > 0 {
			label := fmt.Sprintf("Add repositories: %d repos", len(m.Repositories))
			err := r.addRepositories(m.Repositories)
			res := SystemStepResult{Label: label}
			if err != nil {
				res.Error = err.Error()
				results = append(results, res)
				return results, fmt.Errorf("%s: %w", label, err)
			}
			results = append(results, res)
		}
		if len(m.Packages) > 0 {
			pkgs := r.resolvePackages(m.Packages)
			label := fmt.Sprintf("Install global packages: %s", strings.Join(pkgs, ", "))
			err := r.installPackages(m.Packages)
			res := SystemStepResult{Label: label}
			if err != nil {
				res.Error = err.Error()
				results = append(results, res)
				return results, fmt.Errorf("%s: %w", label, err)
			}
			results = append(results, res)
		}
	}

	var block manifest.ActionBlock
	switch actionName {
	case "install":
		block = m.Install
	case "uninstall":
		block = m.Uninstall
	case "activate":
		block = m.Activate
	case "deactivate":
		block = m.Deactivate
	default:
		// Check custom commands
		if customBlock, exists := m.Commands[actionName]; exists {
			block = customBlock
		} else {
			return results, fmt.Errorf("action '%s' not found in manifest", actionName)
		}
	}

	subResults, err := r.executeActionBlock(m, block, actionName)
	results = append(results, subResults...)
	return results, err
}

func (r *Runner) executeActionBlock(m *manifest.SystemManifest, block manifest.ActionBlock, actionName string) ([]SystemStepResult, error) {
	var results []SystemStepResult

	// 1. Install or remove packages for this action
	if len(block.Packages) > 0 {
		pkgs := r.resolvePackages(block.Packages)
		label := fmt.Sprintf("Install packages: %s", strings.Join(pkgs, ", "))
		if actionName == "uninstall" {
			label = fmt.Sprintf("Remove packages: %s", strings.Join(pkgs, ", "))
		}

		var err error
		if actionName == "uninstall" {
			err = r.removePackages(block.Packages)
		} else {
			err = r.installPackages(block.Packages)
		}

		res := SystemStepResult{Label: label}
		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			return results, fmt.Errorf("%s: %w", label, err)
		}
		results = append(results, res)
	}

	// 2. Execute Steps
	for _, step := range block.Steps {
		label := step.Label
		if label == "" {
			label = fmt.Sprintf("Configure step: %s", step.Type)
		}

		out, err := r.runConfigureStep(m, step)
		res := SystemStepResult{
			Label:  label,
			Output: out,
		}
		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			if !step.IgnoreErrors {
				return results, fmt.Errorf("step '%s' failed: %w", label, err)
			}
		} else {
			results = append(results, res)
		}
	}

	// 3. Manage Services
	if len(block.Services) > 0 {
		for _, svc := range block.Services {
			resolvedName := r.resolveService(svc)
			if resolvedName == "" {
				continue
			}
			label := fmt.Sprintf("Service %s: %s", resolvedName, svc.Action)
			err := r.manageServices([]manifest.Service{svc})
			res := SystemStepResult{Label: label}
			if err != nil {
				res.Error = err.Error()
				results = append(results, res)
				if !svc.IgnoreErrors {
					return results, fmt.Errorf("%s: %w", label, err)
				}
			} else {
				results = append(results, res)
			}
		}
	}

	return results, nil
}

func (r *Runner) runConfigureStep(m *manifest.SystemManifest, step manifest.ConfigureStep) (string, error) {
	switch step.Type {
	case "cmd":
		return r.executeCmdStep(step)
	case "sql_exec":
		return "", r.executeSqlStep(step)
	case "file_replace":
		return "", r.executeFileReplaceStep(step)
	case "file_append":
		return "", r.executeFileAppendStep(step)
	case "kill_port":
		return "", r.executeKillPortStep(step)
	case "run_action":
		if step.Action == "" {
			return "", fmt.Errorf("run_action step requires an 'action' to run")
		}
		subRes, err := r.RunAction(m, step.Action)
		var summaries []string
		for _, sr := range subRes {
			status := "success"
			if sr.Error != "" {
				status = "failed: " + sr.Error
			}
			summaries = append(summaries, fmt.Sprintf("- %s (%s)", sr.Label, status))
		}
		return strings.Join(summaries, "\n"), err
	default:
		return "", fmt.Errorf("unknown step type: %s", step.Type)
	}
}

func (r *Runner) executeCmdStep(step manifest.ConfigureStep) (string, error) {
	if step.Command == "" {
		return "", fmt.Errorf("cmd step requires 'command'")
	}
	cmd := exec.Command(step.Command, step.Args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	if len(step.Env) > 0 {
		cmd.Env = append(cmd.Env, step.Env...)
	}
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("command '%s %s' failed: %v\nStderr: %s", step.Command, strings.Join(step.Args, " "), err, stderr.String())
	}
	return out.String(), nil
}

func (r *Runner) executeSqlStep(step manifest.ConfigureStep) error {
	if step.Executable == "" || step.Query == "" {
		return fmt.Errorf("sql_exec step requires 'executable' and 'query'")
	}
	cmd := exec.Command(step.Executable, "-u", "root", "-e", step.Query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sql_exec failed: %v\nStderr: %s", err, stderr.String())
	}
	return nil
}

func (r *Runner) executeFileReplaceStep(step manifest.ConfigureStep) error {
	if step.File == "" || step.Search == "" {
		return fmt.Errorf("file_replace step requires 'file' and 'search'")
	}
	contentBytes, err := os.ReadFile(step.File)
	if err != nil {
		return err
	}
	content := string(contentBytes)
	re, err := regexp.Compile(step.Search)
	if err != nil {
		return fmt.Errorf("invalid regex search '%s': %w", step.Search, err)
	}
	newContent := re.ReplaceAllLiteralString(content, step.Replace)
	return os.WriteFile(step.File, []byte(newContent), 0644)
}

func (r *Runner) executeFileAppendStep(step manifest.ConfigureStep) error {
	if step.File == "" || step.Block == "" {
		return fmt.Errorf("file_append step requires 'file' and 'block'")
	}
	f, err := os.OpenFile(step.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + step.Block + "\n")
	return err
}

func (r *Runner) executeKillPortStep(step manifest.ConfigureStep) error {
	if step.Port == 0 {
		return fmt.Errorf("kill_port step requires 'port'")
	}
	// Try lsof first
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", step.Port))
	out, err := cmd.Output()
	if err == nil && len(bytes.TrimSpace(out)) > 0 {
		pids := strings.Fields(string(out))
		for _, pidStr := range pids {
			pid, _ := strconv.Atoi(pidStr)
			if pid > 0 {
				proc, _ := os.FindProcess(pid)
				_ = proc.Kill()
			}
		}
	}
	return nil
}

func (r *Runner) addRepositories(repos []manifest.Repository) error {
	for _, repo := range repos {
		if repo.OS != r.PackageManager {
			continue // Skip repos not meant for this OS
		}
		if r.PackageManager == "apt" {
			if strings.HasPrefix(repo.URL, "ppa:") {
				if err := r.executeCommand("add-apt-repository", "-y", repo.URL); err != nil {
					return err
				}
			}
			_ = r.executeCommand("apt-get", "update", "-y")
		} else if r.PackageManager == "dnf" {
			if err := r.executeCommand("dnf", "install", "-y", repo.URL); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner) resolvePackages(packages []manifest.SystemPackage) []string {
	var resolved []string
	for _, p := range packages {
		if r.PackageManager == "apt" && p.Apt != "" {
			resolved = append(resolved, p.Apt)
		} else if r.PackageManager == "dnf" && p.Dnf != "" {
			resolved = append(resolved, p.Dnf)
		} else if p.Name != "" {
			resolved = append(resolved, p.Name)
		}
	}
	return resolved
}

func (r *Runner) installPackages(packages []manifest.SystemPackage) error {
	toInstall := r.resolvePackages(packages)
	if len(toInstall) == 0 {
		return nil
	}
	if r.PackageManager == "dnf" {
		args := append([]string{"install", "-y"}, toInstall...)
		return r.executeCommand("dnf", args...)
	}
	// For apt, enforce non-interactive option overrides
	args := append([]string{"-y", "-o", "Dpkg::Options::=--force-confdef", "-o", "Dpkg::Options::=--force-confold", "install"}, toInstall...)
	return r.executeCommand("apt-get", args...)
}

func (r *Runner) removePackages(packages []manifest.SystemPackage) error {
	toRemove := r.resolvePackages(packages)
	if len(toRemove) == 0 {
		return nil
	}
	if r.PackageManager == "apt" {
		args := append([]string{"remove", "-y", "--purge"}, toRemove...)
		return r.executeCommand("apt-get", args...)
	}
	args := append([]string{"remove", "-y"}, toRemove...)
	return r.executeCommand("dnf", args...)
}

func (r *Runner) resolveService(s manifest.Service) string {
	if r.PackageManager == "apt" {
		if s.Apt != "" {
			return s.Apt
		}
		if s.Dnf != "" && s.Name == "" {
			return ""
		}
	} else if r.PackageManager == "dnf" {
		if s.Dnf != "" {
			return s.Dnf
		}
		if s.Apt != "" && s.Name == "" {
			return ""
		}
	}
	return s.Name
}

func (r *Runner) manageServices(services []manifest.Service) error {
	for _, svc := range services {
		name := r.resolveService(svc)
		if name == "" {
			continue
		}
		switch svc.Action {
		case "enable_start":
			if err := r.executeCommand("systemctl", "enable", name); err != nil {
				return err
			}
			if err := r.executeCommand("systemctl", "start", name); err != nil {
				return err
			}
		case "disable_stop":
			_ = r.executeCommand("systemctl", "stop", name)
			_ = r.executeCommand("systemctl", "disable", name)
		case "restart":
			if err := r.executeCommand("systemctl", "restart", name); err != nil {
				return err
			}
		case "reload":
			if err := r.executeCommand("systemctl", "reload", name); err != nil {
				return err
			}
		case "start":
			if err := r.executeCommand("systemctl", "start", name); err != nil {
				return err
			}
		case "stop":
			if err := r.executeCommand("systemctl", "stop", name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Runner) executeCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command '%s %s' failed: %v\nStderr: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return nil
}
