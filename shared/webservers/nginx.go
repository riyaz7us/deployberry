package webservers

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"shared/shell"
	"strings"

	"github.com/gin-gonic/gin"
)

const NGINX_CONF_DIR = "/etc/nginx/sites-available"
const NGINX_ENABLED_DIR = "/etc/nginx/sites-enabled"

const nginxTemplate = `
server {
    listen 80;
    server_name {{.Domain}};
    root {{.RootPath}};
    index index.php index.html index.htm;

    {{if .ReverseProxyURL}}
    location / {
        proxy_pass {{.ReverseProxyURL}};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    {{else}}
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    {{end}}

    {{if .PHPVersion}}
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php{{.PHPVersion}}-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    {{end}}

    {{if .EnableGzip}}
    gzip on;
    gzip_disable "msie6";
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_buffers 16 8k;
    gzip_http_version 1.1;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
    {{end}}

    location ~ /\.ht {
        deny all;
    }
}
`



type NginxConfig struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"created_at"`
}

func CreateNginxConfig(cfg WebConfig) error {
	// Save to local storage first
	if err := SaveWebConfig(cfg); err != nil {
		return fmt.Errorf("failed to save webconfig: %w", err)
	}

	// Try to deploy to nginx if available
	if IsNginxAvailable() {
		return DeployNginxConfig(cfg)
	}

	// If nginx not available, just save locally (no error)
	return nil
}

func GenerateNginxConfig(cfg WebConfig) (string, error) {
	tmpl, err := template.New("nginx").Parse(nginxTemplate)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, cfg); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func DeployNginxConfig(cfg WebConfig) error {
	nginxConfig, err := GenerateNginxConfig(cfg)
	if err != nil {
		return err
	}

	configPath := fmt.Sprintf("%s/%s", NGINX_CONF_DIR, cfg.Domain)
	if err := os.WriteFile(configPath, []byte(nginxConfig), 0644); err != nil {
		return fmt.Errorf("failed to write nginx config file: %w", err)
	}

	// Symlink to enable the site
	enabledPath := fmt.Sprintf("%s/%s", NGINX_ENABLED_DIR, cfg.Domain)
	os.Remove(enabledPath)
	if err := os.Symlink(configPath, enabledPath); err != nil {
		return fmt.Errorf("failed to create nginx symlink: %w", err)
	}

	// Reload Nginx to apply changes, but only if it's active
	if shell.IsServiceActive("nginx") {
		if err := exec.Command("nginx", "-s", "reload").Run(); err != nil {
			return fmt.Errorf("failed to reload nginx: %w", err)
		}
	}

	return nil
}

func IsNginxAvailable() bool {
	_, err := exec.LookPath("nginx")
	return err == nil
}

// --- Gin Handlers ---

func NginxInstalled(c *gin.Context) {
	if !IsNginxAvailable() {
		c.JSON(200, gin.H{"success": false})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func NginxStatus(c *gin.Context) {
	out, err := exec.Command("systemctl", "status", "nginx").Output()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	isActive := strings.Contains(string(out), "active (running)")
	c.JSON(200, gin.H{
		"success": true,
		"active":  isActive,
		"status":  string(out),
	})
}

func NginxConfigs(c *gin.Context) {
	files, err := os.ReadDir(NGINX_CONF_DIR)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	var configs []NginxConfig
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := os.ReadFile(NGINX_CONF_DIR + "/" + file.Name())
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		// Check if enabled
		_, err = os.Lstat(NGINX_ENABLED_DIR + "/" + file.Name())
		enabled := err == nil

		configs = append(configs, NginxConfig{
			Name:    file.Name(),
			Content: string(content),
			Enabled: enabled,
		})
	}
	c.JSON(200, gin.H{"success": true, "configs": configs})
}

func NginxConfigAdd(c *gin.Context) {
	var config NginxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := handleAddNginxConfig(config); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func handleAddNginxConfig(cfg NginxConfig) error {
	safeName := filepath.Base(cfg.Name)
	if err := os.WriteFile(NGINX_CONF_DIR+"/"+safeName, []byte(cfg.Content), 0644); err != nil {
		return err
	}
	if cfg.Enabled {
		enabledPath := NGINX_ENABLED_DIR + "/" + safeName
		os.Remove(enabledPath)
		if err := os.Symlink(NGINX_CONF_DIR+"/"+safeName, enabledPath); err != nil {
			return err
		}
	}
	return exec.Command("nginx", "-s", "reload").Run()
}

func NginxConfigDelete(c *gin.Context) {
	var config NginxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	safeName := filepath.Base(config.Name)
	os.Remove(NGINX_ENABLED_DIR + "/" + safeName)
	if err := os.Remove(NGINX_CONF_DIR + "/" + safeName); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	exec.Command("nginx", "-s", "reload").Run()
	c.JSON(200, gin.H{"success": true})
}

func NginxConfigEnable(c *gin.Context) {
	var config NginxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	safeName := filepath.Base(config.Name)
	enabledPath := NGINX_ENABLED_DIR + "/" + safeName
	os.Remove(enabledPath)
	if err := os.Symlink(NGINX_CONF_DIR+"/"+safeName, enabledPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	exec.Command("nginx", "-s", "reload").Run()
	c.JSON(200, gin.H{"success": true})
}

func NginxConfigDisable(c *gin.Context) {
	var config NginxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	safeName := filepath.Base(config.Name)
	if err := os.Remove(NGINX_ENABLED_DIR + "/" + safeName); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	exec.Command("nginx", "-s", "reload").Run()
	c.JSON(200, gin.H{"success": true})
}
