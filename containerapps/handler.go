package containerapps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"containerapps/manifest"
	"shared/webservers"
	"shared/repository"
	"shared/shell"

	"github.com/gin-gonic/gin"
)

type InstallRequest struct {
	Path             string            `json:"path" binding:"required"`
	Domain           string            `json:"domain" binding:"required"`
	AppName          string            `json:"appName" binding:"required"`
	Vars             map[string]string `json:"vars"`             // User input variables
	DeploymentMethod string            `json:"deploymentMethod"` // git, none, image, compose
	GitRepo          string            `json:"gitRepo"`
	GitBranch        string            `json:"gitBranch"`
	Image            string            `json:"image"`
	ComposeTemplate  string            `json:"compose_template"`
	ContainerPort    int               `json:"container_port"`
}

type CommandRequest struct {
	AppPath string `json:"app_path" binding:"required"`
	Command string `json:"command" binding:"required"`
	Args    string `json:"args"`
}

type AppActionRequest struct {
	AppPath string `json:"app_path" binding:"required"`
}

func getPrimaryServiceName(m *manifest.Manifest) string {
	lines := strings.Split(m.ComposeTemplate, "\n")
	inServices := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "services:" {
			inServices = true
			continue
		}
		if inServices && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "-") {
			return strings.TrimSuffix(trimmed, ":")
		}
	}
	return "app" // Default fallback
}

func buildCustomManifest(req InstallRequest) *manifest.Manifest {
	var compose string
	port := req.ContainerPort
	if port <= 0 {
		port = 80
	}

	if req.ComposeTemplate != "" {
		compose = req.ComposeTemplate
	} else if req.Image != "" {
		compose = fmt.Sprintf(`version: '3.8'
services:
  app:
    image: %s
    restart: always
    ports:
      - "{HOST_PORT}:%d"
    extra_hosts:
      - "host.containers.internal:host-gateway"
`, req.Image, port)
	} else {
		compose = fmt.Sprintf(`version: '3.8'
services:
  app:
    build: .
    restart: always
    ports:
      - "{HOST_PORT}:%d"
    extra_hosts:
      - "host.containers.internal:host-gateway"
`, port)
	}

	return &manifest.Manifest{
		Name:            "custom-container",
		DisplayName:     req.AppName,
		Version:         "1.0",
		ComposeTemplate: compose,
	}
}

// ListRegistryHandler handles listing available container manifests
func ListRegistryHandler(c *gin.Context) {
	index, err := FetchRegistryIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "apps": index})
}

// GetRequirementsHandler details requirements and variables for container installs
func GetRequirementsHandler(c *gin.Context) {
	slug := c.Param("slug")
	var m *manifest.Manifest
	var err error

	if slug == "custom" {
		m = &manifest.Manifest{
			Name:        "custom",
			DisplayName: "Custom Container App",
			Description: "Custom application built via Dockerfile, image, or compose",
			Variables: []manifest.Variable{
				{Key: "CONTAINER_PORT", Prompt: "Container Port", Default: "80", Required: true, Helper: "Port exposed inside the container"},
			},
		}
	} else {
		m, err = FetchManifest(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"app":              m.DisplayName,
		"variables":        m.Variables,
		"webserver":        m.Webserver,
		"podman_installed": IsPodmanInstalled(),
	})
}

// InstallHandler handles deploying the containerized app
func InstallHandler(c *gin.Context) {
	slug := c.Param("slug")

	domainPattern := `(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`
	appNamePattern := `^[a-zA-Z0-9_-]+( [a-zA-Z0-9_-]+)*$`

	domainRegex := regexp.MustCompile(domainPattern)
	appNameRegex := regexp.MustCompile(appNamePattern)

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if !domainRegex.MatchString(req.Domain) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid domain name!"})
		return
	}

	if !appNameRegex.MatchString(req.AppName) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid application name!"})
		return
	}

	// 1. Check and install Podman if missing
	if err := InstallPodman(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to setup Podman dependencies: %v", err)})
		return
	}

	// 2. Fetch Manifest
	var m *manifest.Manifest
	var err error
	if slug == "custom" {
		m = buildCustomManifest(req)
	} else {
		m, err = FetchManifest(slug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to load manifest: %v", err)})
			return
		}
	}

	appPath := filepath.Join(req.Path, req.Domain)
	if err := os.MkdirAll(appPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to create app directory: %v", err)})
		return
	}

	// Ensure panel_apps:www-data owns it initially
	shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(appPath)))

	// 3. Ports Management & Auto-Allocation
	var hostPort int
	if portVal, ok := req.Vars["APP_PORT"]; ok && portVal != "" {
		if p, err := strconv.Atoi(portVal); err == nil && p > 0 {
			hostPort = p
		}
	}
	if hostPort <= 0 {
		hostPort, err = GetFreePort()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	// 4. Build Variables
	vars := map[string]string{
		"APP_PATH":          appPath,
		"APP_NAME":          req.AppName,
		"DOMAIN":            req.Domain,
		"HOST_PORT":         strconv.Itoa(hostPort),
		"APP_PORT":          strconv.Itoa(hostPort),
		"DEPLOYMENT_METHOD": req.DeploymentMethod,
		"GIT_REPO":          req.GitRepo,
		"GIT_BRANCH":        req.GitBranch,
	}

	// Load manifest variables default values
	for _, v := range m.Variables {
		if v.Default != "" {
			vars[v.Key] = v.Default
		}
	}
	// Override with user variables
	for k, v := range req.Vars {
		vars[k] = v
	}

	primaryService := getPrimaryServiceName(m)
	varsJSON, _ := json.Marshal(req.Vars)

	// 5. Register application in database
	db := repository.GetDB()
	app := &repository.Application{
		Path:         appPath,
		Domain:       req.Domain,
		Provider:     slug,
		Title:        req.AppName,
		DisplayName:  m.DisplayName,
		Version:      m.Version,
		Database:     "", // no host database links
		DeployMethod: req.DeploymentMethod,
		Status:       "installing",
		Language:     "podman",
		Variables:    string(varsJSON),
	}

	if err := db.Create(app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to register app record: %v", err)})
		return
	}

	containerApp := &repository.ContainerApp{
		ApplicationID: app.ID,
		ComposeFile:   shell.SubstituteVars(m.ComposeTemplate, vars),
		HostPort:      hostPort,
		ContainerName: primaryService,
	}
	if err := db.Create(containerApp).Error; err != nil {
		db.Model(app).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to register container app metadata: %v", err)})
		return
	}

	// 6. Run Executor install pipeline
	steps, err := RunInstall(m, vars, primaryService)
	if err != nil {
		db.Model(app).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	// 7. Configure Nginx/Caddy proxy
	webCfg := webservers.WebConfig{
		Domain:          req.Domain,
		RootPath:        appPath,
		PHPVersion:      "",
		EnableGzip:      true,
		EnableCache:     true,
		ReverseProxyURL: fmt.Sprintf("http://127.0.0.1:%d", hostPort),
	}

	if err := webservers.CreateWebConfigs(webCfg); err != nil {
		db.Model(app).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to deploy webserver configs: %v", err), "steps": steps})
		return
	}

	// Force ownership check inside host folder
	shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(appPath)))

	db.Model(app).Update("status", "running")

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Installed successfully", "steps": steps})
}

// DeleteContainerAppInternal cleans up a container application
func DeleteContainerAppInternal(app *repository.Application) error {
	db := repository.GetDB()
	var containerApp repository.ContainerApp
	if err := db.Where("application_id = ?", app.ID).First(&containerApp).Error; err != nil {
		return fmt.Errorf("container metadata not found: %w", err)
	}

	var m *manifest.Manifest
	var err error
	if app.Provider == "custom" {
		m = &manifest.Manifest{
			ComposeTemplate: containerApp.ComposeFile,
		}
	} else {
		m, err = FetchManifest(app.Provider)
		if err != nil {
			return fmt.Errorf("failed to fetch manifest: %w", err)
		}
	}

	var appVars map[string]string
	json.Unmarshal([]byte(app.Variables), &appVars)
	vars := map[string]string{
		"APP_PATH":  app.Path,
		"APP_NAME":  app.Title,
		"DOMAIN":    app.Domain,
		"HOST_PORT": strconv.Itoa(containerApp.HostPort),
	}
	for k, v := range appVars {
		vars[k] = v
	}

	// 1. Run compose down and host files removal
	if err := RunDelete(m, vars, containerApp.ContainerName); err != nil {
		return fmt.Errorf("failed to run delete: %w", err)
	}

	// 2. Drop Webserver configs
	webservers.DeleteWebConfigInternal(app.Domain)

	// 3. Remove GORM records
	db.Delete(&containerApp)
	db.Delete(&app)

	return nil
}

// UpdateContainerAppInternal executes update steps for container applications
func UpdateContainerAppInternal(app *repository.Application) ([]StepResult, error) {
	db := repository.GetDB()
	var containerApp repository.ContainerApp
	if err := db.Where("application_id = ?", app.ID).First(&containerApp).Error; err != nil {
		return nil, fmt.Errorf("container metadata not found: %w", err)
	}

	var m *manifest.Manifest
	var err error
	if app.Provider == "custom" {
		m = &manifest.Manifest{
			ComposeTemplate: containerApp.ComposeFile,
		}
	} else {
		m, err = FetchManifest(app.Provider)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch manifest: %w", err)
		}
	}

	var appVars map[string]string
	json.Unmarshal([]byte(app.Variables), &appVars)
	vars := map[string]string{
		"APP_PATH":  app.Path,
		"APP_NAME":  app.Title,
		"DOMAIN":    app.Domain,
		"HOST_PORT": strconv.Itoa(containerApp.HostPort),
	}
	for k, v := range appVars {
		vars[k] = v
	}

	return RunUpdate(m, vars, containerApp.ContainerName)
}

// RunContainerAppCommandInternal runs a command inside a container app
func RunContainerAppCommandInternal(app *repository.Application, commandName string, commandArgs map[string]string) (string, error) {
	db := repository.GetDB()
	var containerApp repository.ContainerApp
	if err := db.Where("application_id = ?", app.ID).First(&containerApp).Error; err != nil {
		return "", fmt.Errorf("container metadata not found: %w", err)
	}

	var m *manifest.Manifest
	var err error
	if app.Provider == "custom" {
		m = &manifest.Manifest{
			ComposeTemplate: containerApp.ComposeFile,
		}
	} else {
		m, err = FetchManifest(app.Provider)
		if err != nil {
			return "", fmt.Errorf("failed to fetch manifest: %w", err)
		}
	}

	cmdString := ""
	targetService := containerApp.ContainerName

	if m.Commands != nil {
		if cmd, exists := m.Commands[commandName]; exists {
			cmdString = cmd.Run
			if cmd.Service != "" {
				targetService = cmd.Service
			}
		}
	}

	if cmdString == "" {
		cmdString = commandName
	}

	// Construct variables map
	vars := map[string]string{
		"APP_PATH":  app.Path,
		"APP_NAME":  app.Title,
		"DOMAIN":    app.Domain,
		"HOST_PORT": strconv.Itoa(containerApp.HostPort),
	}

	// Load stored variables
	var appVars map[string]string
	json.Unmarshal([]byte(app.Variables), &appVars)
	for k, v := range appVars {
		vars[k] = v
	}

	// Add invocation arguments
	for k, v := range commandArgs {
		vars[k] = v
	}
	if argsStr, ok := commandArgs["ARGS"]; ok {
		vars["ARGS"] = argsStr
	}

	cmdString = shell.SubstituteVars(cmdString, vars)
	res := ExecuteInContainer(app.Path, targetService, cmdString)
	if !res.Success {
		return res.Output, fmt.Errorf("command execution failed: %v", res.Error)
	}

	return res.Output, nil
}

// GetContainerAppLogsInternal retrieves the logs of a container application
func GetContainerAppLogsInternal(app *repository.Application, tailLines string) (string, error) {
	cmdStr := fmt.Sprintf("cd %s && podman-compose logs", shell.EscapeShellArg(app.Path))
	if tailLines != "" {
		linesNum, err := strconv.Atoi(tailLines)
		if err == nil && linesNum > 0 {
			cmdStr = fmt.Sprintf("cd %s && podman-compose logs --tail %d", shell.EscapeShellArg(app.Path), linesNum)
		}
	}
	res := shell.ExecuteCommand(cmdStr)
	if !res.Success {
		return res.Output, fmt.Errorf("failed to fetch container logs: %v", res.Error)
	}
	return res.Output, nil
}

// RunContainerLifecycleCommand runs start/stop/restart for a container application
func RunContainerLifecycleCommand(app *repository.Application, commandName string) (string, error) {
	var cmdStr string
	switch commandName {
	case "start":
		cmdStr = fmt.Sprintf("cd %s && podman-compose start", shell.EscapeShellArg(app.Path))
		res := shell.ExecuteCommand(cmdStr)
		if !res.Success {
			cmdStr = fmt.Sprintf("cd %s && podman-compose up -d", shell.EscapeShellArg(app.Path))
			res = shell.ExecuteCommand(cmdStr)
		}
		if !res.Success {
			return res.Output, fmt.Errorf("failed to start containers: %v", res.Error)
		}
		return res.Output, nil
	case "stop":
		cmdStr = fmt.Sprintf("cd %s && podman-compose stop", shell.EscapeShellArg(app.Path))
	case "restart":
		cmdStr = fmt.Sprintf("cd %s && podman-compose restart", shell.EscapeShellArg(app.Path))
	default:
		return "", fmt.Errorf("unsupported lifecycle command: %s", commandName)
	}

	res := shell.ExecuteCommand(cmdStr)
	if !res.Success {
		return res.Output, fmt.Errorf("lifecycle command '%s' failed: %v", commandName, res.Error)
	}
	return res.Output, nil
}

// UpdateHandler triggers update steps
func UpdateHandler(c *gin.Context) {
	var req AppActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("path = ?", req.AppPath).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "App not found"})
		return
	}

	steps, err := UpdateContainerAppInternal(&app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Updated successfully", "steps": steps})
}

// DeleteHandler stops containers and purges records
func DeleteHandler(c *gin.Context) {
	var req AppActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("path = ?", req.AppPath).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "App not found"})
		return
	}

	if err := DeleteContainerAppInternal(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Application deleted successfully"})
}

// CommandHandler runs commands inside the container
func CommandHandler(c *gin.Context) {
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("path = ?", req.AppPath).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "App not found"})
		return
	}

	argsMap := map[string]string{
		"ARGS": req.Args,
	}

	output, err := RunContainerAppCommandInternal(&app, req.Command, argsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "output": output})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "output": output})
}
