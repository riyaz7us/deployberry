package utils

import (
	"fmt"
	"os"
	"os/exec"
	"shared/repository"
	"time"
)

// GetMySQLRootCredentials returns the root password and port for MySQL
// This function provides a centralized way to get MySQL root credentials
func GetMySQLRootCredentials() (string, int, error) {
	db := repository.GetDB()
	var dbServer repository.DatabaseServer
	err := db.Where("type IN ?", []string{"mysql", "mariadb"}).First(&dbServer).Error
	if err == nil {
		return dbServer.RootPassword, dbServer.Port, nil
	}

	// Fallback to environment variable
	rootPassword := getEnvOrDefault("MYSQL_ROOT_PASSWORD", "")
	if rootPassword == "" {
		return "", 0, fmt.Errorf("MySQL root password not found in database or environment")
	}

	return rootPassword, 3306, nil
}

// TestMySQLConnection tests if MySQL is accessible with the current credentials
func TestMySQLConnection() error {
	rootPassword, port, err := GetMySQLRootCredentials()
	if err != nil {
		return err
	}

	// Test connection using mysql command line, passing password in env variable to hide it from process list
	cmd := exec.Command("mysql", "-u", "root", "-h", "localhost", fmt.Sprintf("-P%d", port), "-e", "SELECT 1")
	cmd.Env = append(os.Environ(), fmt.Sprintf("MYSQL_PWD=%s", rootPassword))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("MySQL connection test failed: %v, output: %s", err, string(output))
	}

	return nil
}

// ResetMySQLRootPassword resets the MySQL root password if needed
func ResetMySQLRootPassword(newPassword string) error {
	// Stop MySQL service
	stopCmd := exec.Command("systemctl", "stop", "mysql")
	if err := stopCmd.Run(); err != nil {
		// Try mysqld for RHEL/CentOS
		stopCmd = exec.Command("systemctl", "stop", "mysqld")
		_ = stopCmd.Run()
	}

	// Defer restarting the normal MySQL service
	defer func() {
		// Start MySQL service normally
		startCmd := exec.Command("systemctl", "start", "mysql")
		if err := startCmd.Run(); err != nil {
			// Try mysqld for RHEL/CentOS
			startCmd = exec.Command("systemctl", "start", "mysqld")
			_ = startCmd.Run()
		}
	}()

	// Start MySQL in safe mode
	safeCmd := exec.Command("mysqld_safe", "--skip-grant-tables", "--skip-networking")
	if err := safeCmd.Start(); err != nil {
		return fmt.Errorf("failed to start MySQL in safe mode: %v", err)
	}

	// Wait a bit for MySQL to start
	time.Sleep(5 * time.Second)

	// Reset password
	resetCmd := exec.Command("mysql", "-u", "root", "-e",
		fmt.Sprintf("FLUSH PRIVILEGES; ALTER USER 'root'@'localhost' IDENTIFIED BY '%s'; FLUSH PRIVILEGES;", newPassword))
	output, err := resetCmd.CombinedOutput()

	// Kill safe mode process
	_ = safeCmd.Process.Kill()
	_ = safeCmd.Wait()
	// Clean up any remaining safe-mode mysqld processes
	_ = exec.Command("pkill", "-f", "mysqld").Run()

	if err != nil {
		return fmt.Errorf("failed to reset MySQL root password: %v, output: %s", err, string(output))
	}

	return nil
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
