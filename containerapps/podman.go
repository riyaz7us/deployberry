package containerapps

import (
	"fmt"
	"os/exec"
	"shared/shell"
)

// IsPodmanInstalled checks if both podman and podman-compose binaries are available in PATH
func IsPodmanInstalled() bool {
	if _, err := exec.LookPath("podman"); err != nil {
		return false
	}
	if _, err := exec.LookPath("podman-compose"); err != nil {
		return false
	}
	return true
}

// InstallPodman installs podman and podman-compose using the detected system package manager
func InstallPodman() error {
	if IsPodmanInstalled() {
		return nil
	}

	var installCmd string
	if _, err := exec.LookPath("apt-get"); err == nil {
		installCmd = "DEBIAN_FRONTEND=noninteractive apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y podman podman-compose"
	} else if _, err := exec.LookPath("dnf"); err == nil {
		installCmd = "dnf install -y podman podman-compose"
	} else {
		return fmt.Errorf("no supported package manager found (apt-get/dnf) to install podman")
	}

	res := shell.ExecuteCommand(installCmd)
	if res.Error != nil {
		return fmt.Errorf("failed to install podman/podman-compose: %w. Output: %s", res.Error, res.Output)
	}

	return nil
}
