package shell

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"shared/globals"
	"shared/repository"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// HandleCommand processes CLI commands like "register", "install", "uninstall", etc.
func HandleCommand(command string) {
	switch command {
	case "install":
		runInstall()
	case "uninstall":
		runUninstall()
	case "start", "stop", "restart":
		runServiceCmd(command)
	case "status":
		runStatus()
	case "logs":
		runLogs()
	case "exec":
		runExec()
	case "register":
		runRegister()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: install, uninstall, start, stop, restart, status, logs, exec, register")
	}
}

func runRegister() {
	var username, password string

	// Check if username and password are provided as arguments
	if len(os.Args) >= 4 {
		username = os.Args[2]
		password = os.Args[3]
	} else {
		// Prompt for username
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)

		// Prompt for password
		fmt.Print("Enter password: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}

	// Store username and hashed password in auth bucket
	db := repository.GetDB()
	if err := db.Create(&repository.User{Username: username, Password: string(hashedPassword)}).Error; err != nil {
		fmt.Println("Error registering user:", err)
		return
	}

	fmt.Printf("Registering user: %s\n", username)
	fmt.Println("Registration successful!")

	// Print all users to verify insert
	users := []repository.User{}
	err = db.Find(&users).Error
	if err != nil {
		fmt.Println("Error getting users:", err)
	} else {
		fmt.Println("in database:")
		for _, user := range users {
			fmt.Printf("Following users are registered: %s\n", user.Username)
		}
	}
}

func runInstall() {
	if !globals.IsDevelopment() && os.Getuid() != 0 {
		log.Fatal("Error: This installer must be run as root (sudo)")
	}

	fmt.Println("======================================")
	fmt.Println("DeployBerry Installer")
	fmt.Println("======================================")
	fmt.Println()

	// 1. Setup system user/group
	fmt.Println("Setting up system user and group...")
	if err := exec.Command("getent", "group", "www-data").Run(); err != nil {
		if err := exec.Command("groupadd", "-r", "www-data").Run(); err != nil {
			log.Printf("Warning: Failed to create group www-data: %v", err)
		}
	}

	if err := exec.Command("id", "-u", "panel_apps").Run(); err != nil {
		cmd := exec.Command("useradd", "--system", "--create-home", "--home-dir", "/home/panel_apps", "--shell", "/bin/bash", "-G", "www-data", "panel_apps")
		if err := cmd.Run(); err != nil {
			log.Fatalf("Error: Failed to create system user panel_apps: %v", err)
		}
	} else {
		exec.Command("usermod", "-s", "/bin/bash", "panel_apps").Run()
	}

	// 2. Create directories
	fmt.Println("Creating directories...")
	dirs := []string{
		"/opt/deployberry/bin",
		"/opt/deployberry/scripts",
		"/opt/deployberry/web",
		"/var/lib/deployberry/webconfigs",
		"/var/lib/deployberry/tools",
		"/var/lib/deployberry/data",
		"/var/lib/deployberry/backups",
		"/var/log/deployberry",
		"/etc/deployberry",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			log.Fatalf("Error: Failed to create directory %s: %v", d, err)
		}
	}

	// 3. Copy executing binary to /opt/deployberry/bin/deployberry
	fmt.Println("Installing DeployBerry executable...")
	selfPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error: Failed to get path of current executable: %v", err)
	}

	targetBin := "/opt/deployberry/bin/deployberry"
	if err := copyFile(selfPath, targetBin); err != nil {
		log.Fatalf("Error: Failed to copy executable to %s: %v", targetBin, err)
	}
	if err := os.Chmod(targetBin, 0755); err != nil {
		log.Fatalf("Error: Failed to make executable: %v", err)
	}

	// 4. Create symlink /usr/local/bin/deployberry
	fmt.Println("Creating command line symlink...")
	symlinkPath := "/usr/local/bin/deployberry"
	os.Remove(symlinkPath) // remove if already exists
	if err := os.Symlink(targetBin, symlinkPath); err != nil {
		log.Printf("Warning: Failed to create symlink: %v. You can link manually with: ln -s %s %s", err, targetBin, symlinkPath)
	}

	// Also symlink as 'panel17' for backward compatibility
	symlinkPanel17Path := "/usr/local/bin/panel17"
	os.Remove(symlinkPanel17Path)
	_ = os.Symlink(targetBin, symlinkPanel17Path)

	// 5. Create env file
	fmt.Println("Configuring environment...")
	envFile := "/etc/deployberry/env.sh"
	envContent := `#!/bin/bash
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
export PATH="/root/.config/composer/vendor/bin:/root/.local/bin:$PATH"
export UV_NO_BUILD_ISOLATION=1
`
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		log.Printf("Warning: Failed to write environment file %s: %v", envFile, err)
	}

	// Wire into panel_apps .bashrc
	wireBashrc("/home/panel_apps/.bashrc", "source /etc/deployberry/env.sh 2>/dev/null", "panel_apps", "panel_apps")
	// Wire into root's .bashrc
	wireBashrc("/root/.bashrc", "source /etc/deployberry/env.sh 2>/dev/null", "root", "root")

	// 6. Create systemd service
	fmt.Println("Creating systemd service...")
	serviceFile := "/etc/systemd/system/deployberry.service"
	serviceContent := `[Unit]
Description=DeployBerry Control Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/deployberry
ExecStart=/opt/deployberry/bin/deployberry
Restart=on-failure
NoNewPrivileges=false
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		log.Fatalf("Error: Failed to write systemd service file: %v", err)
	}

	// 7. Reload systemd daemon & enable service
	fmt.Println("Activating DeployBerry service...")
	exec.Command("systemctl", "daemon-reload").Run()
	exec.Command("systemctl", "enable", "deployberry").Run()

	isActive := exec.Command("systemctl", "is-active", "--quiet", "deployberry").Run() == nil
	if isActive {
		fmt.Println("Restarting active DeployBerry service...")
		exec.Command("systemctl", "restart", "deployberry").Run()
	} else {
		fmt.Println("Starting DeployBerry service...")
		exec.Command("systemctl", "start", "deployberry").Run()
	}

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("Installation Complete!")
	fmt.Println("======================================")
	fmt.Println()
	fmt.Println("DeployBerry has been installed successfully.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Register an admin user: deployberry register <username> <password>")
	fmt.Println("  2. Access panel: http://your-server-ip:7717")
	fmt.Println()
}

func runUninstall() {
	if !globals.IsDevelopment() && os.Getuid() != 0 {
		log.Fatal("Error: This uninstaller must be run as root (sudo)")
	}

	fmt.Println("======================================")
	fmt.Println("DeployBerry Uninstaller")
	fmt.Println("======================================")
	fmt.Println()
	fmt.Println("This will completely remove DeployBerry from your system.")
	fmt.Println("The following will be removed:")
	fmt.Println("  - /opt/deployberry/ (application files)")
	fmt.Println("  - /etc/systemd/system/deployberry.service (systemd service)")
	fmt.Println("  - /usr/local/bin/deployberry (CLI symlink)")
	fmt.Println("  - /etc/deployberry/ (configuration)")
	fmt.Println("  - /var/log/deployberry/ (logs)")
	fmt.Println()

	fmt.Print("Are you sure you want to continue? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if !strings.EqualFold(text, "y") && !strings.EqualFold(text, "yes") {
		fmt.Println("Uninstallation cancelled.")
		return
	}

	// 1. Stop and disable service
	fmt.Println("Stopping DeployBerry service...")
	exec.Command("systemctl", "stop", "deployberry").Run()
	exec.Command("systemctl", "disable", "deployberry").Run()

	// 2. Remove systemd service
	fmt.Println("Removing systemd service...")
	os.Remove("/etc/systemd/system/deployberry.service")
	exec.Command("systemctl", "daemon-reload").Run()

	// 3. Remove symlinks
	fmt.Println("Removing CLI symlinks...")
	os.Remove("/usr/local/bin/deployberry")
	os.Remove("/usr/local/bin/panel17")

	// 4. Remove application folder
	fmt.Println("Removing application files...")
	os.RemoveAll("/opt/deployberry")

	// 5. Ask for data removal
	fmt.Print("Remove data directory /var/lib/deployberry? [y/N]: ")
	textData, _ := reader.ReadString('\n')
	textData = strings.TrimSpace(textData)
	if strings.EqualFold(textData, "y") || strings.EqualFold(textData, "yes") {
		fmt.Println("Removing data directory...")
		os.RemoveAll("/var/lib/deployberry")
	} else {
		fmt.Println("Data directory preserved.")
	}

	// 6. Remove log files
	fmt.Println("Removing log files...")
	os.RemoveAll("/var/log/deployberry")

	// 7. Remove environment configuration
	fmt.Println("Removing environment files...")
	os.RemoveAll("/etc/deployberry")

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("Uninstallation Complete!")
	fmt.Println("======================================")
	fmt.Println()
}

func runServiceCmd(action string) {
	cmd := exec.Command("sudo", "systemctl", action, "deployberry")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running systemctl %s deployberry: %v", action, err)
	}
}

func runStatus() {
	cmd := exec.Command("systemctl", "status", "deployberry")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runLogs() {
	cmd := exec.Command("journalctl", "-u", "deployberry", "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runExec() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: deployberry exec <command> [args...]")
	}
	args := os.Args[2:]
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "/"
	}
	
	var cmd *exec.Cmd
	escapedPwd := EscapeShellArg(pwd)
	cmdString := fmt.Sprintf("source %s 2>/dev/null && { cd %s 2>/dev/null || { echo \"Warning: Directory %s is not accessible. Falling back to home directory.\" >&2; cd ~panel_apps 2>/dev/null || cd /; }; } && exec %s", 
		globals.ENV_PATH, escapedPwd, escapedPwd, strings.Join(EscapeShellArgs(args), " "))

	cmd = exec.Command("sudo", "-u", "panel_apps", "-H", "bash", "-c", cmdString)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func wireBashrc(filePath, sourceCmd, owner, group string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.WriteFile(filePath, []byte(""), 0644)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	if !strings.Contains(string(content), sourceCmd) {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString("\n" + sourceCmd + "\n")
			f.Close()
		}
	}

	if owner != "root" {
		exec.Command("chown", owner+":"+group, filePath).Run()
	}
}

func StartupCheck() error {
	if err := repository.InitDatabase(); err != nil {
		panic("Failed to initialize database: " + err.Error())
	}
	if err := InitializeDeployBerryEnv(); err != nil {
		fmt.Println("⚠️ Warning: Failed to initialize master DeployBerry environment:", err)
	}
	jwtSecret := GetJWTSecret()
	os.Setenv("JWT_SECRET", jwtSecret)
	return nil
}

func GetJWTSecret() string {
	db := repository.GetDB()
	var jwtSecret repository.Config
	err := db.Where("key = ?", "jwt_secret").First(&jwtSecret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			secret := make([]byte, 32)
			if _, err := rand.Read(secret); err != nil {
				panic("Failed to generate JWT secret: " + err.Error())
			}
			secretBase64 := base64.StdEncoding.EncodeToString(secret)
			SetJWTSecret(secretBase64)
			return secretBase64
		}
		panic("Failed to get JWT secret: " + err.Error())
	}
	return jwtSecret.Value
}

func SetJWTSecret(secret string) {
	db := repository.GetDB()
	if err := db.Create(&repository.Config{Key: "jwt_secret", Value: secret}).Error; err != nil {
		panic("Failed to set JWT secret: " + err.Error())
	}
}

func InitializeDeployBerryEnv() error {
	envPath := globals.ENV_PATH

	if globals.IsDevelopment() {
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(envPath), 0755)
			baseEnv := "#!/bin/bash\nexport PATH=\"$PATH\"\n"
			if err := os.WriteFile(envPath, []byte(baseEnv), 0644); err != nil {
				return err
			}
		}
		return nil
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("global environment file %s is missing. Please run the installer first", envPath)
	}

	if err := exec.Command("getent", "group", "www-data").Run(); err != nil {
		return fmt.Errorf("required system group 'www-data' is missing. Please run the installer first")
	}

	if err := exec.Command("id", "-u", "panel_apps").Run(); err != nil {
		return fmt.Errorf("required system user 'panel_apps' is missing. Please run the installer first")
	}

	return nil
}
