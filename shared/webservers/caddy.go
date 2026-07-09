package webservers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"shared/shell"
	"strings"

	"github.com/gin-gonic/gin"
)

const CADDY_CONF_DIR = "/etc/caddy/sites"

const caddyTemplate = `
{{.Domain}} {
    root * {{.RootPath}}
    
    {{if .PHPVersion}}
    php_fastcgi unix//run/php/php{{.PHPVersion}}-fpm.sock
    {{end}}

    {{if .ReverseProxyURL}}
    reverse_proxy {{.ReverseProxyURL}}
    {{end}}

    {{if .EnableGzip}}
    encode gzip
    {{end}}

    file_server
}
`

func CreateCaddyConfig(cfg WebConfig) error {
	// Try to deploy to caddy if available
	if IsCaddyAvailable() {
		return DeployCaddyConfig(cfg)
	}

	// If caddy not available, configs are already saved by nginx handler
	return nil
}

func GenerateCaddyConfig(cfg WebConfig) (string, error) {
	tmpl, err := template.New("caddy").Parse(caddyTemplate)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, cfg); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func DeployCaddyConfig(cfg WebConfig) error {
	caddyConfig, err := GenerateCaddyConfig(cfg)
	if err != nil {
		return err
	}

	// Create the sites-available directory if it doesn't exist
	if err := os.MkdirAll(CADDY_CONF_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create caddy config directory: %w", err)
	}

	// Ensure placeholder file exists to prevent empty glob reload failures in Caddy 2
	placeholderPath := fmt.Sprintf("%s/placeholder.conf", CADDY_CONF_DIR)
	if _, err := os.Stat(placeholderPath); os.IsNotExist(err) {
		_ = os.WriteFile(placeholderPath, []byte("# Placeholder config to prevent Caddy glob errors\n"), 0644)
	}

	configPath := fmt.Sprintf("%s/%s", CADDY_CONF_DIR, cfg.Domain)
	if err := os.WriteFile(configPath, []byte(caddyConfig), 0644); err != nil {
		return fmt.Errorf("failed to write caddy config file: %w", err)
	}

	// Ensure the main Caddyfile includes the sites-available directory
	if err := ensureCaddyfileImports(); err != nil {
		return fmt.Errorf("failed to update caddyfile imports: %w", err)
	}

	// Reload Caddy to apply changes, but only if it's active
	if shell.IsServiceActive("caddy") {
		if err := exec.Command("caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run(); err != nil {
			fmt.Printf("warning: failed to reload caddy: %v\n", err)
		}
	}

	return nil
}

// ensureCaddyfileImports ensures the main Caddyfile includes imports from sites-available
func ensureCaddyfileImports() error {
	caddyfilePath := "/etc/caddy/Caddyfile"

	// Read the current Caddyfile
	content, err := os.ReadFile(caddyfilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read caddyfile: %w", err)
	}

	caddyfileContent := string(content)

	// Check if import statement already exists
	importStatement := fmt.Sprintf("import %s/*", CADDY_CONF_DIR)
	if strings.Contains(caddyfileContent, importStatement) {
		return nil // Already has the import
	}

	// Add the import statement at the end
	var newContent string
	if len(caddyfileContent) > 0 && !strings.HasSuffix(caddyfileContent, "\n") {
		newContent = caddyfileContent + "\n\n" + importStatement + "\n"
	} else {
		newContent = caddyfileContent + "\n" + importStatement + "\n"
	}

	// Ensure target directory exists
	if err := os.MkdirAll("/etc/caddy", 0755); err != nil {
		return fmt.Errorf("failed to create caddy directory: %w", err)
	}

	// Write back the updated Caddyfile
	if err := os.WriteFile(caddyfilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to update caddyfile: %w", err)
	}

	return nil
}

func IsCaddyAvailable() bool {
	_, err := exec.LookPath("caddy")
	return err == nil
}

// --- Gin Handlers ---

// GetCaddyStatus gets the status of caddy service
func GetCaddyStatus(c *gin.Context) {
	// Check if caddy is installed
	if _, err := exec.LookPath("caddy"); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"installed": false,
			"active":    false,
			"message":   "Caddy is not installed",
		})
		return
	}

	// Get caddy version
	versionCmd := exec.Command("caddy", "version")
	versionOutput, err := versionCmd.Output()
	var versionInfo string
	if err == nil {
		versionInfo = strings.TrimSpace(string(versionOutput))
	}
	// Extract clean version number
	re := regexp.MustCompile(`v?([0-9]+\.[0-9]+\.[0-9]+)`)
	versionMatch := re.FindStringSubmatch(versionInfo)
	cleanVersion := ""
	if len(versionMatch) >= 2 {
		cleanVersion = versionMatch[1]
	} else {
		cleanVersion = versionInfo
	}

	// Use shared IsServiceActive check
	active := shell.IsServiceActive("caddy")

	// Execute native systemctl status command for detailed output
	out, err := exec.Command("systemctl", "status", "caddy").CombinedOutput()
	outputStr := string(out)

	response := gin.H{
		"installed": true,
		"active":    active,
		"version":   cleanVersion,
		"success":   err == nil || active,
		"output":    outputStr,
	}

	c.JSON(http.StatusOK, response)
}
