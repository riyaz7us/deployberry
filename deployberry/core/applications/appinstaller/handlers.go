package appinstaller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"deployberry/core/applications/appinstaller/manifest"
	"deployberry/core/databases"
	"deployberry/core/dbops"
	"deployberry/core/languages"
	"shared/webservers"
	"shared/repository"
	"shared/shell"

	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────
// VERSION HELPER FUNCTIONS
// ─────────────────────────────────────────

// parseVersionConstraint parses a version constraint and returns required and recommended versions
func parseVersionConstraint(constraint string) (required, recommended string) {
	clean := strings.TrimSpace(constraint)
	if clean == "" {
		return "", ""
	}

	// Handle >=
	if strings.Contains(clean, ">=") {
		parts := strings.Fields(clean)
		for _, part := range parts {
			if strings.HasPrefix(part, ">=") {
				ver := strings.TrimPrefix(part, ">=")
				if ver == "" {
					for idx, p := range parts {
						if p == ">=" && idx+1 < len(parts) {
							return parts[idx+1], "latest"
						}
					}
				} else {
					return ver, "latest"
				}
			}
		}
		return clean, "latest"
	}

	// Handle >
	if strings.Contains(clean, ">") {
		parts := strings.Fields(clean)
		for _, part := range parts {
			if strings.HasPrefix(part, ">") {
				ver := strings.TrimPrefix(part, ">")
				if ver == "" {
					for idx, p := range parts {
						if p == ">" && idx+1 < len(parts) {
							return parts[idx+1], "latest"
						}
					}
				} else {
					return ver, "latest"
				}
			}
		}
		return clean, "latest"
	}

	// Handle caret ^ or tilde ~
	if strings.HasPrefix(clean, "^") {
		required = strings.TrimSpace(strings.TrimPrefix(clean, "^"))
		return required, required
	}
	if strings.HasPrefix(clean, "~") {
		required = strings.TrimSpace(strings.TrimPrefix(clean, "~"))
		return required, required
	}

	// For ranges like "8.1 - 8.4"
	if strings.Contains(clean, "-") {
		parts := strings.Split(clean, "-")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	return clean, clean
}

// ─────────────────────────────────────────
// REQUIREMENTS STRUCTURES
// ─────────────────────────────────────────

type RuntimeRequirement struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	Required    string `json:"required"`    // Required version from manifest
	Recommended string `json:"recommended"` // Recommended version (latest or highest)
}

type DatabaseRequirement struct {
	Selected string      `json:"selected"`
	Options  []DBOptions `json:"options"`
}

type DBOptions struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Available   bool   `json:"available"`
	Required    string `json:"required"`    // Required version from manifest
	Recommended string `json:"recommended"` // Recommended version (latest or highest)
}

// ─────────────────────────────────────────
// REQUEST STRUCTURES
// ─────────────────────────────────────────

type InstallRequest struct {
	Path             string            `json:"path" binding:"required"`
	Domain           string            `json:"domain" binding:"required"`
	AppName          string            `json:"appName" binding:"required"`
	Vars             map[string]string `json:"vars"` // User-provided vars (e.g., ADMIN_EMAIL)
	DeploymentMethod string            `json:"deploymentMethod"`
	DatabaseEngine   string            `json:"databaseEngine"`
	GitRepo          string            `json:"gitRepo"`
	GitBranch        string            `json:"gitBranch"`
	ManualPath       string            `json:"manualPath"`
}

type CommandRequest struct {
	AppPath string `json:"app_path" binding:"required"`
	Command string `json:"command" binding:"required"` // The key from manifest.commands
	Args    string `json:"args"`                       // Optional args if custom command args are defined
}

type AppActionRequest struct {
	AppPath string `json:"app_path" binding:"required"`
}

// ─────────────────────────────────────────
// VERSION CHECKING
// ─────────────────────────────────────────

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	for len(parts) < 3 {
		parts = append(parts, "0")
	}
	return strings.Join(parts[:3], ".")
}

// checks if available version matches the constraint using semver library
// Supports: ">=8.1", "<=8.4", ">8.1", "<8.4", "8.1-8.4", "8.2", etc.
func isVersionCompatible(availableVersion, constraint string) bool {
	if constraint == "" || constraint == "latest" {
		return true
	}
	fmt.Println("Checking version compatibility:", availableVersion, constraint)

	// Parse available version
	// "8.2" is automatically interpreted as "8.2.0"
	normalized := normalizeVersion(availableVersion)
	v, err := semver.NewVersion(normalized)
	if err != nil {
		fmt.Println("Error parsing available version:", err)
		return false
	}

	// Parse constraint
	// Natively supports ">=8.0", "8.1-8.4", etc. without regex!
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		fmt.Println("Error parsing constraint:", err)
		return false
	}

	// Check if the version satisfies the constraint
	return c.Check(v)
}

// ─────────────────────────────────────────
// HANDLERS
// ─────────────────────────────────────────

func ListRegistryHandler(c *gin.Context) {
	index, err := FetchRegistryIndex()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "apps": index})
}

func GetRequirementsHandler(c *gin.Context) {
	slug := c.Param("slug")
	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Split requirements into runtime and database

	var runtimeReq *RuntimeRequirement
	var databaseReq *DatabaseRequirement

	var runtimeReqs []RuntimeRequirement
	for _, rt := range m.Runtime {
		if rt.Language != "" {
			required, recommended := parseVersionConstraint(rt.Version)
			req := RuntimeRequirement{
				Name:        rt.Language,
				Available:   false,
				Version:     "",
				Required:    required,
				Recommended: recommended,
			}

			if rt.Language == "static" {
				req.Available = true
				req.Version = "N/A"
			} else {
				// Check if language is available and matches version constraint
				langData, _ := languages.CheckOneService(rt.Language)
				if langData.Version != "" {
					req.Available = isVersionCompatible(langData.Version, rt.Version)
					req.Version = langData.Version
				}
			}
			runtimeReqs = append(runtimeReqs, req)
		}
	}

	if len(runtimeReqs) > 0 {
		runtimeReq = &runtimeReqs[0]
	}

	// Check database requirement if required
	if m.Database.Required {
		databaseReq = &DatabaseRequirement{
			Selected: "",
			Options:  []DBOptions{},
		}

		// Show all database options (installed and not installed)
		// m.Database.Engines is []DatabaseEngine with name and version
		for _, engine := range m.Database.Engines {
			dbData, _ := databases.CheckOneService(engine.Name)

			// Parse version constraint for this database
			required, recommended := parseVersionConstraint(engine.Version)

			// Default values for not installed databases
			version := ""
			available := false

			// Check if database is installed
			if dbData.Version != "" {
				version = dbData.Version
				available = isVersionCompatible(dbData.Version, engine.Version) && dbData.Active
			}

			// Always add the option so users can see what's required
			databaseReq.Options = append(databaseReq.Options, DBOptions{
				Name:        engine.Name,
				Version:     version,
				Available:   available,
				Required:    required,
				Recommended: recommended,
			})

			// Auto-select first available and compatible database
			if available && databaseReq.Selected == "" {
				databaseReq.Selected = engine.Name
			}
		}

		// Fallback: if no compatible database is selected, select the first option
		if databaseReq.Selected == "" && len(databaseReq.Options) > 0 {
			databaseReq.Selected = databaseReq.Options[0].Name
		}
	}

	var inMemoryReqs []DBOptions
	for _, db := range m.InMemory {
		dbData, _ := databases.CheckOneService(db.Name)
		required, recommended := parseVersionConstraint(db.Version)
		version := ""
		available := false
		if dbData.Version != "" {
			version = dbData.Version
			available = isVersionCompatible(dbData.Version, db.Version) && dbData.Active
		}
		inMemoryReqs = append(inMemoryReqs, DBOptions{
			Name:        db.Name,
			Version:     version,
			Available:   available,
			Required:    required,
			Recommended: recommended,
		})
	}

	// Get deployment methods if deployment is configured
	deployment := GetDeploymentMethods(m)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"app":        m.DisplayName,
		"runtime":    runtimeReq,
		"runtimes":   runtimeReqs,
		"database":   databaseReq,
		"in_memory":  inMemoryReqs,
		"variables":  m.Variables,
		"deployment": deployment,
	})
}

func ListCommandsHandler(c *gin.Context) {
	slug := c.Param("slug")
	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Manifest not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "commands": m.Commands})
}

func InstallHandler(c *gin.Context) {
	slug := c.Param("slug") //application name

	pattern := `(?i)^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`
	appNamePattern := `^[a-zA-Z0-9_-]+( [a-zA-Z0-9_-]+)*$`

	domainRegex := regexp.MustCompile(pattern)
	appNameRegex := regexp.MustCompile(appNamePattern)

	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	valid := domainRegex.MatchString(req.Domain)

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid domain! (Use only letters, numbers, and dots, in the form of domain.tld)"})
		return
	}

	validAppName := appNameRegex.MatchString(req.AppName)
	if !validAppName {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid App Name! (Use only letters, numbers, spaces, dashes, and underscores, without leading or trailing spaces)"})
		return
	}

	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	log.Printf("[handler] ======= Incoming Install Request for %s (%s) =======", slug, req.Domain)

	appPath := filepath.Join(req.Path, req.Domain)
	if err := os.MkdirAll(appPath, 0755); err != nil {
		log.Printf("[handler] Failed to create directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to create directory: %v", err)})
		return
	}
	shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(appPath)))

	dbName, dbUser, dbPass := "", "", ""
	dbPort := ""
	dbEngine := ""

	// 1. Setup Database (if required)
	if m.Database.Required {
		dbEngine = req.DatabaseEngine
		if dbEngine == "" {
			if len(m.Database.Engines) > 0 {
				dbEngine = m.Database.Engines[0].Name
			}
		}

		if dbEngine == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Database engine is required but none are specified in the manifest"})
			return
		}

		engineValid := slices.ContainsFunc(m.Database.Engines, func(e manifest.DatabaseEngine) bool {
			return e.Name == dbEngine
		})
		if !engineValid {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("Unsupported database engine: %s", dbEngine)})
			return
		}

		log.Printf("[handler] Setting up database using engine: %s...", dbEngine)
		dbStart := time.Now()
		provider, err := dbops.GetProvider(dbEngine)
		if err != nil {
			log.Printf("[handler] Failed to get database provider: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to get database provider: %v", err)})
			return
		}

		result, err := provider.CreateDatabaseInternal(req.AppName)
		if err != nil {
			log.Printf("[handler] Failed to create database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to create database: %v", err)})
			return
		}
		dbName, dbUser, dbPass = result.Database, result.Username, result.Password
		if result.Port != 0 {
			dbPort = strconv.Itoa(result.Port)
		}
		log.Printf("[handler] DB created in %v", time.Since(dbStart))
	}

	// 2. Build base Variables map
	baseVars := map[string]string{
		"APP_PATH": appPath,
		"APP_NAME": req.AppName,
		"DOMAIN":   req.Domain,
		"DB_HOST":  "localhost",
		"DB_NAME":  dbName,
		"DB_USER":  dbUser,
		"DB_PASS":  dbPass,
	}

	if dbPort != "" {
		baseVars["DB_PORT"] = dbPort
	}

	// Pre-populate variables from manifest with their default values first
	for _, v := range m.Variables {
		if v.Default != "" {
			baseVars[v.Key] = v.Default
		}
	}

	// Merge user-provided variables (overriding defaults)
	for k, v := range req.Vars {
		baseVars[k] = v
	}

	// Add deployment configurations
	if req.DeploymentMethod != "" && req.DeploymentMethod != "none" {
		baseVars["DEPLOYMENT_METHOD"] = req.DeploymentMethod
	}
	if req.GitRepo != "" {
		baseVars["GIT_REPO"] = req.GitRepo
	}
	if req.GitBranch != "" {
		baseVars["GIT_BRANCH"] = req.GitBranch
	}
	if req.ManualPath != "" {
		baseVars["MANUAL_PATH"] = req.ManualPath
	}

	// 3. Merge with system variables
	vars := baseVars

	primaryLang := ""
	if len(m.Runtime) > 0 {
		primaryLang = m.Runtime[0].Language
	}
	varsJSON, _ := json.Marshal(req.Vars)

	db := repository.GetDB()
	app := &repository.Application{
		Path:         appPath,
		Domain:       req.Domain,
		Provider:     slug,
		Title:        req.AppName,
		DisplayName:  m.DisplayName,
		Version:      m.Version,
		Database:     dbEngine, // Save the actual engine type
		DeployMethod: req.DeploymentMethod,
		Status:       "installing",
		Language:     primaryLang,
		Variables:    string(varsJSON),
	}
	if err := db.Create(app).Error; err != nil {
		log.Printf("[handler] Failed to create application record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to register application: %v", err)})
		return
	}

	// Link Database Credentials to this Application if a DB was created
	if m.Database.Required && dbName != "" {
		db.Model(&repository.DatabaseCredential{}).Where("database = ?", dbName).Update("app_id", app.ID)
	}

	// 4. Install package manager on-demand if needed
	if vars["PACKAGE_MANAGER"] != "" && vars["PACKAGE_MANAGER"] != "npm" {
		log.Printf("[handler] Installing custom package manager: %s...", vars["PACKAGE_MANAGER"])
		installPackageManager(vars["PACKAGE_MANAGER"])
	}

	// 5. Run Executor
	execStart := time.Now()
	steps, err := RunInstall(m, vars)
	if err != nil {
		log.Printf("[handler] App %s installation failed in %v: %v", req.AppName, time.Since(execStart), err)
		db.Model(app).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}
	log.Printf("[handler] App %s installed in %v", req.AppName, time.Since(execStart))

	// 5. Configure Webserver
	phpVersion := ""
	hasPHP := false
	for _, rt := range m.Runtime {
		if rt.Language == "php" {
			hasPHP = true
			break
		}
	}
	if hasPHP {
		if phpData, err := languages.CheckOneService("php"); err == nil {
			phpVersion = phpData.Version
		}
	}

	// Web root configuration
	webRoot := appPath
	if vars["WEB_ROOT"] != "" && vars["WEB_ROOT"] != "." {
		webRoot = filepath.Join(appPath, vars["WEB_ROOT"])
	} else if m.Webserver.Root != "" && m.Webserver.Root != "." {
		webRoot = filepath.Join(appPath, m.Webserver.Root)
	}

	webCfg := webservers.WebConfig{
		Domain:      req.Domain,
		RootPath:    webRoot,
		PHPVersion:  phpVersion,
		EnableGzip:  true,
		EnableCache: true,
	}

	// Port configuration for proxy mode
	if m.Webserver.Mode == "proxy" {
		port := m.Webserver.Port
		if vars["APP_PORT"] != "" {
			if parsedPort, err := strconv.Atoi(vars["APP_PORT"]); err == nil && parsedPort > 0 {
				port = parsedPort
			}
		}
		if port != 0 {
			webCfg.ReverseProxyURL = fmt.Sprintf("http://127.0.0.1:%d", port)
		}
	}

	if err := webservers.CreateWebConfigs(webCfg); err != nil {
		db.Model(app).Update("status", "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": fmt.Sprintf("Failed to configure webserver: %v", err), "steps": steps})
		return
	}

	// 6. Setup Process Manager
	if m.Process.Manager == "pm2" {
		entryPoint := m.Process.Start
		if entryPoint == "" {
			entryPoint = vars["ENTRY_POINT"]
		}
		if entryPoint == "" {
			entryPoint = "server.js"
		}

		entryPoint = shell.SubstituteVars(entryPoint, vars)

		pm2Cmd := fmt.Sprintf("pm2 start %s --name %s && pm2 save", entryPoint, shell.EscapeShellArg(req.AppName))
		shell.ExecuteAsAppUser(pm2Cmd, appPath)
	}

	// Ensure all deployed/created/compiled files are owned by the restricted app user and group
	shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(appPath)))

	// Update status to installed
	db.Model(app).Update("status", "installed")

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Installed successfully", "steps": steps})
}

func CommandHandler(c *gin.Context) {
	slug := c.Param("slug")
	var req CommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	vars := map[string]string{
		"APP_PATH": req.AppPath,
		"ARGS":     req.Args,
	}

	output, err := RunCommand(m, req.Command, vars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "output": output})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "output": output})
}

func UpdateHandler(c *gin.Context) {
	slug := c.Param("slug")
	var req AppActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	vars := map[string]string{"APP_PATH": req.AppPath}
	steps, err := RunUpdate(m, vars)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Updated successfully", "steps": steps})
}

func DeleteHandler(c *gin.Context) {
	slug := c.Param("slug")
	var req AppActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	m, err := FetchManifest(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 1. Get app details from DB to know what domain/DB to drop
	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("path = ?", req.AppPath).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "App not found in database"})
		return
	}

	// 2. Run shell deletion hooks (pre_delete + remove files)
	vars := map[string]string{"APP_PATH": req.AppPath, "APP_NAME": app.Title}
	if err := RunDelete(m, vars); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 3. Process Manager cleanup (PM2) - securely escaped
	if m.Process.Manager == "pm2" {
		shell.ExecuteAsAppUser(fmt.Sprintf("pm2 delete %s && pm2 save", shell.EscapeShellArg(app.Title)), "")
	}

	// 4. Drop Webserver configs natively
	if m.Delete.RemoveWebserverConfig {
		webservers.DeleteWebConfigInternal(app.Domain)
		shell.ExecuteCommand("systemctl reload nginx || true")
		shell.ExecuteCommand("systemctl reload caddy || true")
	}

	// 5. Drop Database natively via the correct DB provider using credentials
	if m.Delete.DropDatabase {
		var dbCred repository.DatabaseCredential
		if err := db.Where("app_id = ?", app.ID).First(&dbCred).Error; err == nil {
			provider, err := dbops.GetProvider(dbCred.Type)
			if err == nil {
				if err := provider.DeleteDatabaseInternal(dbCred.Database); err != nil {
					log.Printf("Error deleting database %s: %v", dbCred.Database, err)
				}
				if err := provider.DeleteUserInternal(dbCred.Username); err != nil {
					log.Printf("Error deleting database user %s: %v", dbCred.Username, err)
				}
			} else {
				log.Printf("Error getting database provider for drop: %v", err)
			}
		}
	}

	// 6. Remove from DB
	db.Where("path = ?", req.AppPath).Delete(&repository.Application{})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Application deleted successfully"})
}

// installPackageManager installs the specified package manager on-demand
func installPackageManager(pm string) {
	checkCmd := fmt.Sprintf("command -v %s", pm)

	switch pm {
	case "yarn":
		if res := shell.ExecuteCommand(checkCmd); !res.Success {
			shell.ExecuteCommand("npm install -g yarn")
		}
	case "pnpm":
		if res := shell.ExecuteCommand(checkCmd); !res.Success {
			shell.ExecuteCommand("npm install -g pnpm")
		}
	case "bun":
		if res := shell.ExecuteCommand(checkCmd); !res.Success {
			shell.ExecuteCommand("curl -fsSL https://bun.sh/install.sh | bash")
		}
	}
	// npm is assumed to be available with Node.js installation
}
