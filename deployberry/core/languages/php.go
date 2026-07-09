package languages

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// Helper function to check if PHP is installed
func check_php_installed() bool {
	_, err := exec.LookPath("php")
	return err == nil
}

// PHP Configuration Management
type PHPConfig struct {
	MemoryLimit       string            `json:"memory_limit"`
	MaxExecutionTime  string            `json:"max_execution_time"`
	MaxInputTime      string            `json:"max_input_time"`
	PostMaxSize       string            `json:"post_max_size"`
	UploadMaxFilesize string            `json:"upload_max_filesize"`
	MaxFileUploads    string            `json:"max_file_uploads"`
	DisplayErrors     string            `json:"display_errors"`
	ErrorReporting    string            `json:"error_reporting"`
	LogErrors         string            `json:"log_errors"`
	Extensions        map[string]bool   `json:"extensions"`
	CustomSettings    map[string]string `json:"custom_settings"`
}

// GetPHPConfig retrieves current PHP configuration
func GetPHPConfig(c *gin.Context) {
	// Check if PHP is installed
	if !check_php_installed() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "PHP is not installed",
		})
		return
	}

	// Get php.ini path
	phpIniPath, err := getPHPIniPath()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to locate php.ini file",
			"error":   err.Error(),
		})
		return
	}

	// Parse current configuration
	config, err := parsePHPIni(phpIniPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to parse php.ini file",
			"error":   err.Error(),
		})
		return
	}

	// Get available extensions
	availableExtensions := getAvailableExtensions()

	// Get currently loaded extensions
	loadedExtensions := getLoadedExtensions()

	// Merge extension status
	extensions := make(map[string]bool)
	for _, ext := range availableExtensions {
		extensions[ext] = slices.Contains(loadedExtensions, ext)
	}
	config.Extensions = extensions

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     config,
		"ini_path": phpIniPath,
	})
}

// UpdatePHPConfig updates PHP configuration
func UpdatePHPConfig(c *gin.Context) {
	var config PHPConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid configuration data",
			"error":   err.Error(),
		})
		return
	}

	// Check if PHP is installed
	if !check_php_installed() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "PHP is not installed",
		})
		return
	}

	// Get php.ini path
	phpIniPath, err := getPHPIniPath()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to locate php.ini file",
			"error":   err.Error(),
		})
		return
	}

	// Create backup of current php.ini
	backupPath := phpIniPath + ".backup." + fmt.Sprintf("%d", os.Getpid())
	if err := copyFile(phpIniPath, backupPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create backup of php.ini",
			"error":   err.Error(),
		})
		return
	}

	// Update php.ini
	if err := updatePHPIni(phpIniPath, config); err != nil {
		// Restore backup on error
		copyFile(backupPath, phpIniPath)
		os.Remove(backupPath)

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update php.ini",
			"error":   err.Error(),
		})
		return
	}

	// Clean up backup
	os.Remove(backupPath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "PHP configuration updated successfully. You may need to restart your web server for changes to take effect.",
	})
}

// Helper function to get php.ini path
func getPHPIniPath() (string, error) {
	// Try to get php.ini path from PHP
	cmd := exec.Command("php", "--ini")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Loaded Configuration File:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				if path != "(none)" && path != "" {
					return path, nil
				}
			}
		}
	}

	// Fallback: common php.ini locations
	commonPaths := []string{
		"/etc/php.ini",
		"/usr/local/etc/php.ini",
		"/etc/php/8.2/apache2/php.ini",
		"/etc/php/8.2/fpm/php.ini",
		"/etc/php/8.1/apache2/php.ini",
		"/etc/php/8.1/fpm/php.ini",
		"/etc/php/8.0/apache2/php.ini",
		"/etc/php/8.0/fpm/php.ini",
		"/etc/php/7.4/apache2/php.ini",
		"/etc/php/7.4/fpm/php.ini",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("php.ini file not found")
}

// Helper function to parse php.ini
func parsePHPIni(iniPath string) (*PHPConfig, error) {
	content, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	config := &PHPConfig{
		CustomSettings: make(map[string]string),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		switch key {
		case "memory_limit":
			config.MemoryLimit = value
		case "max_execution_time":
			config.MaxExecutionTime = value
		case "max_input_time":
			config.MaxInputTime = value
		case "post_max_size":
			config.PostMaxSize = value
		case "upload_max_filesize":
			config.UploadMaxFilesize = value
		case "max_file_uploads":
			config.MaxFileUploads = value
		case "display_errors":
			config.DisplayErrors = value
		case "error_reporting":
			config.ErrorReporting = value
		case "log_errors":
			config.LogErrors = value
		default:
			config.CustomSettings[key] = value
		}
	}

	return config, nil
}

// Helper function to update php.ini
func updatePHPIni(iniPath string, config PHPConfig) error {
	content, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	// Settings to update
	settings := map[string]string{
		"memory_limit":        config.MemoryLimit,
		"max_execution_time":  config.MaxExecutionTime,
		"max_input_time":      config.MaxInputTime,
		"post_max_size":       config.PostMaxSize,
		"upload_max_filesize": config.UploadMaxFilesize,
		"max_file_uploads":    config.MaxFileUploads,
		"display_errors":      config.DisplayErrors,
		"error_reporting":     config.ErrorReporting,
		"log_errors":          config.LogErrors,
	}

	// Add custom settings
	for key, value := range config.CustomSettings {
		settings[key] = value
	}

	updatedKeys := make(map[string]bool)

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			newLines = append(newLines, originalLine)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, originalLine)
			continue
		}

		key := strings.TrimSpace(parts[0])
		if newValue, exists := settings[key]; exists && newValue != "" {
			newLines = append(newLines, fmt.Sprintf("%s = %s", key, newValue))
			updatedKeys[key] = true
		} else {
			newLines = append(newLines, originalLine)
		}
	}

	// Add any new settings that weren't found in the file
	for key, value := range settings {
		if !updatedKeys[key] && value != "" {
			newLines = append(newLines, fmt.Sprintf("%s = %s", key, value))
		}
	}

	// Write updated content
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(iniPath, []byte(newContent), 0644)
}

// Helper function to get available extensions
func getAvailableExtensions() []string {
	// Common PHP extensions
	return []string{
		"bcmath", "bz2", "calendar", "ctype", "curl", "dom", "exif", "fileinfo",
		"filter", "ftp", "gd", "gettext", "hash", "iconv", "intl", "json",
		"libxml", "mbstring", "mysqli", "mysqlnd", "openssl", "pcre", "pdo",
		"pdo_mysql", "pdo_pgsql", "pdo_sqlite", "phar", "posix", "readline",
		"reflection", "session", "simplexml", "soap", "sockets", "sodium",
		"spl", "sqlite3", "standard", "tokenizer", "xml", "xmlreader",
		"xmlwriter", "xsl", "zip", "zlib", "redis", "memcached", "imagick",
	}
}

// Helper function to get loaded extensions
func getLoadedExtensions() []string {
	cmd := exec.Command("php", "-m")
	output, err := cmd.Output()
	if err != nil {
		return []string{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var extensions []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "[") {
			extensions = append(extensions, line)
		}
	}

	return extensions
}

// Helper function to copy file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
