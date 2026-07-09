package files

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// GitOperation represents a git operation request
type GitOperation struct {
	Path    string `json:"path" binding:"required"`
	Action  string `json:"action" binding:"required"` // pull, force_pull, stash, apply_stash, reset_hard, status, log
	Branch  string `json:"branch,omitempty"`          // optional branch for pull operations
	Remote  string `json:"remote,omitempty"`          // optional remote name (default: origin)
	Message string `json:"message,omitempty"`         // optional commit message for stash
}

// GitStatus represents the status of a git repository
type GitStatus struct {
	Path          string   `json:"path"`
	IsGitRepo     bool     `json:"is_git_repo"`
	CurrentBranch string   `json:"current_branch,omitempty"`
	RemoteURL     string   `json:"remote_url,omitempty"`
	Status        string   `json:"status,omitempty"`
	LastCommit    string   `json:"last_commit,omitempty"`
	Changes       []string `json:"changes,omitempty"`
	Ahead         int      `json:"ahead,omitempty"`
	Behind        int      `json:"behind,omitempty"`
}

// GitResult represents the result of a git operation
type GitResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SSHKeyResult represents the result of SSH key operations
type SSHKeyResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	PublicKey  string `json:"public_key,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	KeyExists  bool   `json:"key_exists"`
	KeyPath    string `json:"key_path,omitempty"`
	Error      string `json:"error,omitempty"`
}

func ExecGit(dir string, args ...string) (string, error) {
	gitArgs := append([]string{"-c", "safe.directory=*"}, args...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput != "" {
			return strings.TrimSpace(stdout.String()), fmt.Errorf("%s", errOutput)
		}
		return strings.TrimSpace(stdout.String()), err
	}
	return strings.TrimSpace(stdout.String()), nil
}

// GetGitRepoStatus returns comprehensive git status for a path
func GetGitRepoStatus(path string) (GitStatus, error) {
	status := GitStatus{
		Path:      path,
		IsGitRepo: false,
	}

	if !PathExists(path) {
		return status, fmt.Errorf("path does not exist")
	}

	// Check if it's a git repository
	_, err := ExecGit(path, "rev-parse", "--git-dir")
	if err != nil {
		return status, nil
	}

	status.IsGitRepo = true

	// Get current branch
	if branch, err := ExecGit(path, "branch", "--show-current"); err == nil {
		status.CurrentBranch = branch
	}

	// Get remote URL
	if remoteURL, err := ExecGit(path, "remote", "get-url", "origin"); err == nil {
		status.RemoteURL = remoteURL
	}

	// Get status
	if changes, err := ExecGit(path, "status", "--porcelain"); err == nil {
		if changes != "" {
			status.Changes = strings.Split(changes, "\n")
		}
	}

	// Get ahead/behind info
	if abInfo, err := ExecGit(path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD"); err == nil {
		parts := strings.Fields(abInfo)
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%d", &status.Behind)
			fmt.Sscanf(parts[1], "%d", &status.Ahead)
		}
	}

	// Get last commit info
	if lastCommit, err := ExecGit(path, "log", "--oneline", "-1"); err == nil {
		status.LastCommit = lastCommit
	}

	return status, nil
}

// GitPull performs a git pull operation
func GitPull(op GitOperation) GitResult {
	remote := "origin"
	if op.Remote != "" {
		remote = op.Remote
	}

	branch := "main"
	if op.Branch != "" {
		branch = op.Branch
	} else {
		if b, err := ExecGit(op.Path, "branch", "--show-current"); err == nil && b != "" {
			branch = b
		}
	}

	output, err := ExecGit(op.Path, "pull", remote, branch)
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git pull failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git pull completed successfully",
		Output:  output,
	}
}

// GitForcePull performs a force git pull operation
func GitForcePull(op GitOperation) GitResult {
	remote := "origin"
	if op.Remote != "" {
		remote = op.Remote
	}

	branch := "main"
	if op.Branch != "" {
		branch = op.Branch
	} else {
		if b, err := ExecGit(op.Path, "branch", "--show-current"); err == nil && b != "" {
			branch = b
		}
	}

	output, err := ExecGit(op.Path, "fetch", remote)
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git force pull failed during fetch",
			Error:   err.Error(),
			Output:  output,
		}
	}

	resetOutput, err := ExecGit(op.Path, "reset", "--hard", fmt.Sprintf("%s/%s", remote, branch))
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git force pull failed during reset",
			Error:   err.Error(),
			Output:  resetOutput,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git force pull completed successfully",
		Output:  resetOutput,
	}
}

// GitStash stashes current changes
func GitStash(op GitOperation) GitResult {
	args := []string{"stash", "push"}
	if op.Message != "" {
		args = append(args, "-m", fmt.Sprintf("Auto-stash %s", op.Message))
	} else {
		args = append(args, "-m", "Auto-stash")
	}

	output, err := ExecGit(op.Path, args...)
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git stash failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Changes stashed successfully",
		Output:  output,
	}
}

// GitApplyStash applies the latest stash
func GitApplyStash(op GitOperation) GitResult {
	output, err := ExecGit(op.Path, "stash", "apply")
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git apply stash failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Stash applied successfully",
		Output:  output,
	}
}

// GitResetHard performs a hard reset
func GitResetHard(op GitOperation) GitResult {
	output, err := ExecGit(op.Path, "reset", "--hard", "HEAD")
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git reset hard failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git reset hard completed successfully",
		Output:  output,
	}
}

// GitStatusCmd gets the current git status
func GitStatusCmd(op GitOperation) GitResult {
	output, err := ExecGit(op.Path, "status")
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git status failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git status retrieved successfully",
		Output:  output,
	}
}

// GitLog gets the git log
func GitLog(op GitOperation) GitResult {
	output, err := ExecGit(op.Path, "log", "--oneline", "-10")
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git log failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git log retrieved successfully",
		Output:  output,
	}
}

// GitFetch performs a git fetch operation
func GitFetch(op GitOperation) GitResult {
	remote := "origin"
	if op.Remote != "" {
		remote = op.Remote
	}

	output, err := ExecGit(op.Path, "fetch", remote)
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git fetch failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git fetch completed successfully",
		Output:  output,
	}
}

// GitBranch lists available branches
func GitBranch(op GitOperation) GitResult {
	output, err := ExecGit(op.Path, "branch", "-a")
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git branch failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: "Git branches retrieved successfully",
		Output:  output,
	}
}

// GitCheckout switches to a different branch
func GitCheckout(op GitOperation) GitResult {
	if op.Branch == "" {
		return GitResult{
			Success: false,
			Message: "Branch name is required for checkout",
			Error:   "Branch parameter is missing",
		}
	}

	output, err := ExecGit(op.Path, "checkout", op.Branch)
	if err != nil {
		return GitResult{
			Success: false,
			Message: "Git checkout failed",
			Error:   err.Error(),
			Output:  output,
		}
	}

	return GitResult{
		Success: true,
		Message: fmt.Sprintf("Checked out to branch '%s' successfully", op.Branch),
		Output:  output,
	}
}

// GetOrGenerateSSHKey retrieves existing SSH key or generates a new one
func GetOrGenerateSSHKey() SSHKeyResult {
	usr, err := user.Current()
	if err != nil {
		return SSHKeyResult{
			Success:   false,
			Message:   "Failed to get user home directory",
			Error:     err.Error(),
			KeyExists: false,
		}
	}

	sshDir := filepath.Join(usr.HomeDir, ".ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa")
	publicKeyPath := filepath.Join(sshDir, "id_rsa.pub")

	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		err := os.MkdirAll(sshDir, 0700)
		if err != nil {
			return SSHKeyResult{
				Success:   false,
				Message:   "Failed to create SSH directory",
				Error:     err.Error(),
				KeyExists: false,
			}
		}
	}

	if _, err := os.Stat(privateKeyPath); err == nil {
		_, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return SSHKeyResult{
				Success:   false,
				Message:   "Failed to read private key",
				Error:     err.Error(),
				KeyExists: true,
			}
		}

		publicKey, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return SSHKeyResult{
				Success:   false,
				Message:   "Failed to read public key",
				Error:     err.Error(),
				KeyExists: true,
			}
		}

		return SSHKeyResult{
			Success:   true,
			Message:   "SSH key retrieved successfully",
			PublicKey: string(publicKey),
			KeyExists: true,
			KeyPath:   privateKeyPath,
		}
	}

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", privateKeyPath, "-N", "", "-C", "generated-by-deployberry")
	var cmdStderr bytes.Buffer
	cmd.Stderr = &cmdStderr
	err = cmd.Run()

	if err != nil {
		return SSHKeyResult{
			Success:   false,
			Message:   "Failed to generate SSH key",
			Error:     fmt.Sprintf("%v: %s", err, cmdStderr.String()),
			KeyExists: false,
		}
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return SSHKeyResult{
			Success:   false,
			Message:   "SSH key generated but failed to read public key",
			Error:     err.Error(),
			KeyExists: true,
		}
	}

	return SSHKeyResult{
		Success:   true,
		Message:   "SSH key generated successfully",
		PublicKey: string(publicKey),
		KeyExists: true,
		KeyPath:   privateKeyPath,
	}
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
