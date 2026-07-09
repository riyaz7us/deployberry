package webservers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"shared/globals"
	"shared/repository"
	"shared/shell"
	"shared/system/installer"
	"shared/system/manifest"

	"github.com/gin-gonic/gin"
)

// WebConfig holds all the necessary parameters for generating
// web server configuration files for a domain.
type WebConfig struct {
	Domain          string `json:"domain"` // e.g., "example.com"
	RootPath        string `json:"root_path"` // The public root directory, e.g., "/var/www/html/example.com"
	PHPVersion      string `json:"php_version"` // PHP version for FPM, e.g., "8.2". If empty, PHP is disabled.
	ReverseProxyURL string `json:"reverse_proxy_url"` // URL for reverse proxy, e.g., "http://localhost:3000". If empty, disabled.
	EnableGzip      bool   `json:"enable_gzip"`   // Flag to enable gzip compression.
	EnableCache     bool   `json:"enable_cache"`  // Flag to enable basic caching.
}

// CreateWebConfigs creates both Nginx and Caddy configuration files for the given configuration.
// It acts as an orchestrator, calling the specific generator for each web server.
// Now saves to local storage first, then deploys to available servers.
func CreateWebConfigs(cfg WebConfig) error {
	// Create nginx config (this saves to local storage and deploys if available)
	if err := CreateNginxConfig(cfg); err != nil {
		return err
	}

	// Create caddy config (this deploys if available, storage already handled by nginx)
	if err := CreateCaddyConfig(cfg); err != nil {
		return err
	}

	return nil
}

// DeleteWebConfigInternal drops all web server configurations natively without gin.Context
func DeleteWebConfigInternal(domain string) {
	// Remove from database
	db := repository.GetDB()
	db.Where("domain = ?", domain).Delete(&repository.WebConfig{})

	// Remove local files
	os.Remove(fmt.Sprintf("%s/nginx_%s", globals.WEBCONFIGS_DIR, domain))
	os.Remove(fmt.Sprintf("%s/caddy_%s", globals.WEBCONFIGS_DIR, domain))

	// Remove deployed Nginx files and reload if available
	os.Remove(fmt.Sprintf("%s/%s", NGINX_CONF_DIR, domain))
	os.Remove(fmt.Sprintf("%s/%s", NGINX_ENABLED_DIR, domain))

	// Remove deployed Caddy files and reload if available
	os.Remove(fmt.Sprintf("%s/%s", CADDY_CONF_DIR, domain))

	// Reload webservers if they were managing this domain (best effort)
	if shell.IsServiceActive("nginx") {
		shell.ExecuteCommand("nginx -s reload")
	}
	if shell.IsServiceActive("caddy") {
		shell.ExecuteCommand("caddy reload --config /etc/caddy/Caddyfile")
	}
}

// SaveWebConfig saves configuration to local storage files and DB
func SaveWebConfig(cfg WebConfig) error {
	// Create webconfigs directory
	os.MkdirAll(globals.WEBCONFIGS_DIR, 0755)

	// Generate both nginx and caddy configs
	nginxConfig, err := GenerateNginxConfig(cfg)
	if err != nil {
		return err
	}

	caddyConfig, err := GenerateCaddyConfig(cfg)
	if err != nil {
		return err
	}

	// Save webconfig to database
	db := repository.GetDB()
	webConfig := repository.WebConfig{
		Domain:          cfg.Domain,
		RootPath:        cfg.RootPath,
		PHPVersion:      cfg.PHPVersion,
		ReverseProxyURL: cfg.ReverseProxyURL,
		EnableGzip:      cfg.EnableGzip,
		EnableCache:     cfg.EnableCache,
	}
	db.Save(&webConfig)

	// Save individual config files
	os.WriteFile(fmt.Sprintf("%s/nginx_%s", globals.WEBCONFIGS_DIR, cfg.Domain), []byte(nginxConfig), 0644)
	os.WriteFile(fmt.Sprintf("%s/caddy_%s", globals.WEBCONFIGS_DIR, cfg.Domain), []byte(caddyConfig), 0644)

	return nil
}

// --- Gin Handlers ---

func InstallWebserver(c *gin.Context, serverName string) {
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", serverName), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "install")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	db := repository.GetDB()
	dbRecord := repository.WebServer{
		Type: serverName,
	}
	_ = db.Where("type = ?", serverName).FirstOrCreate(&dbRecord).Error
	dbRecord.Name = serverName
	dbRecord.Version = "installed"
	dbRecord.Status = "installed"
	dbRecord.Active = true
	dbRecord.Port = 80
	dbRecord.SSLPort = 443
	if serverName == "nginx" {
		dbRecord.ConfigDir = "/etc/nginx"
	} else if serverName == "caddy" {
		dbRecord.ConfigDir = "/etc/caddy"
	}
	db.Save(&dbRecord)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s installed", serverName), "steps": steps})
}

func ActionWebserver(c *gin.Context, serverName string, action string) {
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", serverName), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	db := repository.GetDB()
	if action == "uninstall" {
		db.Where("type = ?", serverName).Delete(&repository.WebServer{})
	} else if action == "activate" {
		db.Model(&repository.WebServer{}).Where("type = ?", serverName).Update("active", true)
	} else if action == "deactivate" {
		db.Model(&repository.WebServer{}).Where("type = ?", serverName).Update("active", false)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s %s successful", serverName, action), "steps": steps})
}

func WebConfigList(c *gin.Context) {
	db := repository.GetDB()
	var webConfigs []repository.WebConfig
	err := db.Find(&webConfigs).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var configList []map[string]interface{}
	for _, config := range webConfigs {
		configMap := map[string]interface{}{
			"domain":            config.Domain,
			"root_path":         config.RootPath,
			"php_version":       config.PHPVersion,
			"reverse_proxy_url": config.ReverseProxyURL,
			"enable_gzip":       config.EnableGzip,
			"enable_cache":      config.EnableCache,
			"webserver":         config.WebServer,
			"ssl":               config.SSL,
			"created_at":        config.CreatedAt,
			"updated_at":        config.UpdatedAt,
		}
		configList = append(configList, configMap)
	}

	c.JSON(200, gin.H{"success": true, "configs": configList})
}

func WebConfigDeploy(c *gin.Context) {
	domain := c.Param("domain")
	server := c.DefaultQuery("server", "both")

	db := repository.GetDB()
	var webConfig repository.WebConfig
	err := db.Where("domain = ?", domain).First(&webConfig).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Configuration not found"})
		return
	}

	webCfg := WebConfig{
		Domain:          webConfig.Domain,
		RootPath:        webConfig.RootPath,
		PHPVersion:      webConfig.PHPVersion,
		ReverseProxyURL: webConfig.ReverseProxyURL,
		EnableGzip:      webConfig.EnableGzip,
		EnableCache:     webConfig.EnableCache,
	}

	var errors []string

	if server == "nginx" || server == "both" {
		if IsNginxAvailable() {
			if err := DeployNginxConfig(webCfg); err != nil {
				errors = append(errors, fmt.Sprintf("nginx: %v", err))
			}
		} else {
			errors = append(errors, "nginx: not available")
		}
	}

	if server == "caddy" || server == "both" {
		if IsCaddyAvailable() {
			if err := DeployCaddyConfig(webCfg); err != nil {
				errors = append(errors, fmt.Sprintf("caddy: %v", err))
			}
		} else {
			errors = append(errors, "caddy: not available")
		}
	}

	if len(errors) > 0 {
		c.JSON(500, gin.H{"error": "Deployment failed", "details": strings.Join(errors, ", ")})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Configuration deployed successfully"})
}

func WebConfigRecreate(c *gin.Context) {
	domain := c.Param("domain")
	server := c.DefaultQuery("server", "both")

	db := repository.GetDB()
	var webConfig repository.WebConfig
	err := db.Where("domain = ?", domain).First(&webConfig).Error
	if err != nil {
		c.JSON(404, gin.H{"error": "Configuration not found"})
		return
	}

	webCfg := WebConfig{
		Domain:          webConfig.Domain,
		RootPath:        webConfig.RootPath,
		PHPVersion:      webConfig.PHPVersion,
		ReverseProxyURL: webConfig.ReverseProxyURL,
		EnableGzip:      webConfig.EnableGzip,
		EnableCache:     webConfig.EnableCache,
	}

	var errors []string

	if server == "nginx" || server == "both" {
		os.Remove(fmt.Sprintf("%s/%s", NGINX_CONF_DIR, domain))
		os.Remove(fmt.Sprintf("%s/%s", NGINX_ENABLED_DIR, domain))

		if IsNginxAvailable() {
			if err := DeployNginxConfig(webCfg); err != nil {
				errors = append(errors, fmt.Sprintf("nginx: %v", err))
			}
		} else {
			errors = append(errors, "nginx: not available")
		}
	}

	if server == "caddy" || server == "both" {
		os.Remove(fmt.Sprintf("%s/%s", CADDY_CONF_DIR, domain))

		if IsCaddyAvailable() {
			if err := DeployCaddyConfig(webCfg); err != nil {
				errors = append(errors, fmt.Sprintf("caddy: %v", err))
			}
		} else {
			errors = append(errors, "caddy: not available")
		}
	}

	if len(errors) > 0 {
		c.JSON(500, gin.H{"error": "Recreation failed", "details": strings.Join(errors, ", ")})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Configuration recreated successfully"})
}

func WebConfigDelete(c *gin.Context) {
	domain := c.Param("domain")

	DeleteWebConfigInternal(domain)

	c.JSON(200, gin.H{"success": true, "message": "Configuration deleted"})
}

func WebConfigCreate(c *gin.Context) {
	var cfg WebConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if cfg.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Domain is required"})
		return
	}

	if cfg.RootPath == "" {
		cfg.RootPath = "/var/www/html" // Default fallback if not specified
	}

	if err := CreateWebConfigs(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Configuration created successfully"})
}

