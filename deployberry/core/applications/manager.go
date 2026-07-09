package applications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"deployberry/core/applications/appinstaller"
	"containerapps"
	"shared/repository"

	"github.com/gin-gonic/gin"
)

// ApplicationStatus represents the current status of an application
type ApplicationStatus struct {
	Status      string `json:"status"`
	IsRunning   bool   `json:"is_running"`
	IsHealthy   bool   `json:"is_healthy"`
	LastChecked string `json:"last_checked"`
	Uptime      string `json:"uptime"`
	MemoryUsage string `json:"memory_usage,omitempty"`
	CPUUsage    string `json:"cpu_usage,omitempty"`
}

// ApplicationConfig represents configuration for an application
type ApplicationConfig struct {
	Variables map[string]string `json:"variables"`
	Webserver WebserverConfig   `json:"webserver"`
	Process   ProcessConfig     `json:"process"`
	Database  DatabaseConfig    `json:"database"`
}

// WebserverConfig represents webserver configuration
type WebserverConfig struct {
	Type    string                 `json:"type"` // nginx, caddy, apache
	Version string                 `json:"version"`
	Mode    string                 `json:"mode"` // proxy, php-fpm
	Port    int                    `json:"port"`
	Root    string                 `json:"root"`
	SSL     bool                   `json:"ssl"`
	Config  map[string]interface{} `json:"config"`
}

// ProcessConfig represents process manager configuration
type ProcessConfig struct {
	Manager   string `json:"manager"` // pm2, systemd, none
	PID       int    `json:"pid,omitempty"`
	Status    string `json:"status"`
	StartTime string `json:"start_time"`
	Restarts  int    `json:"restarts"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type    string `json:"type"` // mysql, postgres, sqlite
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Name    string `json:"name"`
	User    string `json:"user"`
	Version string `json:"version"`
	SSL     bool   `json:"ssl"`
}

// ListApplications returns all installed applications
func ListApplications(c *gin.Context) {
	var applications []repository.Application
	db := repository.GetDB()

	// Get query parameters
	search := c.Query("search")
	status := c.Query("status")
	provider := c.Query("provider")

	query := db.Model(&repository.Application{})

	// Apply filters
	if search != "" {
		query = query.Where("title LIKE ? OR display_name LIKE ? OR domain LIKE ? OR path LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}

	err := query.Order("created_at DESC").Find(&applications).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	// Map to response objects with runtime from language field
	type AppResponse struct {
		ID           uint      `json:"id"`
		Path         string    `json:"path"`
		Domain       string    `json:"domain"`
		Provider     string    `json:"provider"`
		Title        string    `json:"title"`
		DisplayName  string    `json:"display_name"`
		Version      string    `json:"version"`
		Runtime      string    `json:"runtime"`
		Database     string    `json:"database"`
		DeployMethod string    `json:"deploy_method"`
		Status       string    `json:"status"`
		Variables    string    `json:"variables"`
		Language     string    `json:"language"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	responseData := make([]AppResponse, len(applications))
	for i, app := range applications {
		responseData[i] = AppResponse{
			ID:           app.ID,
			Path:         app.Path,
			Domain:       app.Domain,
			Provider:     app.Provider,
			Title:        app.Title,
			DisplayName:  app.DisplayName,
			Version:      app.Version,
			Runtime:      app.Language,
			Database:     app.Database,
			DeployMethod: app.DeployMethod,
			Status:       app.Status,
			Variables:    app.Variables,
			Language:     app.Language,
			CreatedAt:    app.CreatedAt,
			UpdatedAt:    app.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
		"count":   len(responseData),
	})
}

// GetApplication returns a specific application by ID or path
func GetApplication(c *gin.Context) {
	id := c.Param("id")
	path := c.Query("path")

	db := repository.GetDB()
	var app repository.Application

	var err error
	if id != "" {
		// Get by ID
		err = db.Where("id = ?", id).First(&app).Error
	} else if path != "" {
		// Get by path
		err = db.Where("path = ?", path).First(&app).Error
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either ID or path parameter is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	var appDescription string
	var appWebserver string
	var appProcessManager string
	var commands = make(map[string]interface{})

	if app.Language == "podman" {
		appDescription = "Containerized application managed by Podman Compose"
		appWebserver = "proxy"
		appProcessManager = "podman-compose"

		if app.Provider != "custom" {
			cappManifest, err := containerapps.FetchManifest(app.Provider)
			if err == nil {
				appDescription = cappManifest.Description
				appWebserver = cappManifest.Webserver.Mode
				if len(cappManifest.Commands) > 0 {
					for cmdKey, cmdDef := range cappManifest.Commands {
						commands[cmdKey] = map[string]interface{}{
							"label": cmdDef.Label,
							"run":   cmdDef.Run,
							"args":  cmdDef.Args,
						}
					}
				}
			}
		}
	} else {
		// Get manifest to fetch available commands
		manifest, err := appinstaller.FetchManifest(app.Provider)
		if err == nil {
			appDescription = manifest.Description
			appWebserver = manifest.Webserver.Mode
			appProcessManager = manifest.Process.Manager
			if len(manifest.Commands) > 0 {
				for cmdKey, cmdDef := range manifest.Commands {
					commands[cmdKey] = map[string]interface{}{
						"label": cmdDef.Label,
						"run":   cmdDef.Run,
						"args":  cmdDef.Args,
					}
				}
			}
		} else {
			appDescription = "Native application (manifest offline or missing)"
			appWebserver = "unknown"
			appProcessManager = "unknown"
		}
	}

	// Build data object with manifest-derived fields
	// These are static in the manifest schema, not stored in DB
	data := gin.H{
		"id":              app.ID,
		"name":            app.Title,
		"title":           app.Title,
		"display_name":    app.DisplayName,
		"description":     appDescription,
		"version":         app.Version,
		"provider":        app.Provider,
		"path":            app.Path,
		"domain":          app.Domain,
		"status":          app.Status,
		"runtime":         app.Language,
		"language":        app.Language,
		"database":        app.Database,
		"webserver":       appWebserver,
		"process_manager": appProcessManager,
		"deploy_method":   app.DeployMethod,
		"variables":       app.Variables,
		"created_at":      app.CreatedAt,
		"updated_at":      app.UpdatedAt,
	}

	// Prepare response with manifest commands
	response := gin.H{
		"success": true,
		"data":    data,
	}

	if len(commands) > 0 {
		response["commands"] = commands
	}

	c.JSON(http.StatusOK, response)
}

// GetApplicationStatus returns the current status of an application
func GetApplicationStatus(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application ID is required"})
		return
	}

	// Get application from database
	db := repository.GetDB()
	var app repository.Application
	err := db.Where("id = ?", appID).First(&app).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Get application status based on provider and configuration
	status := getApplicationStatus(&app)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetApplicationConfig returns the configuration of an application
func GetApplicationConfig(c *gin.Context) {
	appID := c.Param("id")

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Parse variables from JSON
	var variables map[string]string
	if app.Variables != "" {
		if err := json.Unmarshal([]byte(app.Variables), &variables); err != nil {
			variables = make(map[string]string)
		}
	} else {
		variables = make(map[string]string)
	}

	// Build configuration - basic structure, actual config comes from manifest
	config := ApplicationConfig{
		Variables: variables,
		Webserver: WebserverConfig{},
		Process:   ProcessConfig{},
		Database:  DatabaseConfig{},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateApplication updates an application's configuration
func UpdateApplication(c *gin.Context) {
	var req struct {
		Path      string            `json:"path" binding:"required"`
		Variables map[string]string `json:"variables"`
		Webserver WebserverConfig   `json:"webserver"`
		Process   ProcessConfig     `json:"process"`
		Database  DatabaseConfig    `json:"database"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get application from database
	db := repository.GetDB()
	var app repository.Application
	err := db.Where("path = ?", req.Path).First(&app).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Update variables
	if req.Variables != nil {
		variablesJSON, _ := json.Marshal(req.Variables)
		app.Variables = string(variablesJSON)
	}

	// Save changes
	app.UpdatedAt = time.Now()
	err = db.Save(&app).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application updated successfully",
		"data":    app,
	})
}

// ExecuteApplicationCommand is a generic handler for any manifest-defined command
// This replaces custom implementations like start/stop/restart with manifest execution
func ExecuteApplicationCommand(commandName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appID := c.Param("id")

		db := repository.GetDB()
		var app repository.Application
		if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
			return
		}

		if app.Language == "podman" {
			output, err := containerapps.RunContainerLifecycleCommand(&app, commandName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   fmt.Sprintf("Failed to execute lifecycle command '%s'", commandName),
					"details": err.Error(),
					"output":  output,
				})
				return
			}

			// Update status for specific commands
			statusUpdates := map[string]string{
				"start":   "running",
				"stop":    "stopped",
				"restart": "running",
			}
			if newStatus, exists := statusUpdates[commandName]; exists {
				app.Status = newStatus
				app.UpdatedAt = time.Now()
				db.Save(&app)
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": fmt.Sprintf("Command '%s' executed successfully", commandName),
				"output":  output,
			})
			return
		}

		// Load manifest
		manifest, err := appinstaller.FetchManifest(app.Provider)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manifest"})
			return
		}

		// Prepare variables
		var variables map[string]string
		if app.Variables != "" {
			json.Unmarshal([]byte(app.Variables), &variables)
		} else {
			variables = make(map[string]string)
		}
		variables["APP_PATH"] = app.Path
		variables["APP_NAME"] = app.Title
		variables["DOMAIN"] = app.Domain

		// Execute command from manifest
		output, err := appinstaller.RunCommand(manifest, commandName, variables)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   fmt.Sprintf("Failed to execute command '%s'", commandName),
				"details": err.Error(),
			})
			return
		}

		// Update status for specific commands
		statusUpdates := map[string]string{
			"start":   "running",
			"stop":    "stopped",
			"restart": "running",
		}
		if newStatus, exists := statusUpdates[commandName]; exists {
			app.Status = newStatus
			app.UpdatedAt = time.Now()
			db.Save(&app)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Command '%s' executed successfully", commandName),
			"output":  output,
		})
	}
}

// DeleteApplication deletes an application by executing the delete command from manifest
func DeleteApplication(c *gin.Context) {
	appID := c.Param("id")

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if app.Language == "podman" {
		if err := containerapps.DeleteContainerAppInternal(&app); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete container app",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Application deleted successfully",
		})
		return
	}

	// Load manifest
	manifest, err := appinstaller.FetchManifest(app.Provider)
	if err != nil {
		// If we cannot fetch the manifest, we still delete the GORM record as a fallback
		if err := db.Delete(&app).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove application from database"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Application deleted from database (manifest fetch failed, skipped uninstall hooks)",
		})
		return
	}

	// Prepare variables
	var variables map[string]string
	if app.Variables != "" {
		json.Unmarshal([]byte(app.Variables), &variables)
	} else {
		variables = make(map[string]string)
	}
	variables["APP_PATH"] = app.Path
	variables["APP_NAME"] = app.Title
	variables["DOMAIN"] = app.Domain

	// Execute delete command from manifest
	if _, err := appinstaller.RunCommand(manifest, "delete", variables); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute delete command",
			"details": err.Error(),
		})
		return
	}

	// Remove from database
	if err := db.Delete(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove application from database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application deleted successfully",
	})
}

// UpdateApplication updates an application by executing the update command from manifest
func UpdateApplicationManifest(c *gin.Context) {
	appID := c.Param("id")

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if app.Language == "podman" {
		steps, err := containerapps.UpdateContainerAppInternal(&app)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update container app",
				"details": err.Error(),
			})
			return
		}
		app.UpdatedAt = time.Now()
		db.Save(&app)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Application updated successfully",
			"steps":   steps,
		})
		return
	}

	// Load manifest
	manifest, err := appinstaller.FetchManifest(app.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manifest"})
		return
	}

	// Prepare variables
	var variables map[string]string
	if app.Variables != "" {
		json.Unmarshal([]byte(app.Variables), &variables)
	} else {
		variables = make(map[string]string)
	}
	variables["APP_PATH"] = app.Path
	variables["APP_NAME"] = app.Title
	variables["DOMAIN"] = app.Domain

	// Execute update command from manifest
	output, err := appinstaller.RunCommand(manifest, "update", variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute update command",
			"details": err.Error(),
		})
		return
	}

	// Update timestamp
	app.UpdatedAt = time.Now()
	db.Save(&app)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application updated successfully",
		"output":  output,
	})
}

// RunApplicationCommand executes a manifest command for an application
func RunApplicationCommand(c *gin.Context) {
	var id = c.Param("id")
	var req struct {
		Command string            `json:"command" binding:"required"`
		Args    map[string]string `json:"args,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get application from database
	db := repository.GetDB()
	var app repository.Application
	err := db.Where("id = ?", id).First(&app).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if app.Language == "podman" {
		output, err := containerapps.RunContainerAppCommandInternal(&app, req.Command, req.Args)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to execute command",
				"details": err.Error(),
				"output":  output,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Command '%s' executed successfully", req.Command),
			"output":  output,
		})
		return
	}

	// Load manifest
	manifest, err := appinstaller.FetchManifest(app.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manifest"})
		return
	}

	// Prepare variables
	var variables map[string]string
	if app.Variables != "" {
		json.Unmarshal([]byte(app.Variables), &variables)
	} else {
		variables = make(map[string]string)
	}
	variables["APP_PATH"] = app.Path
	variables["APP_NAME"] = app.Title
	variables["DOMAIN"] = app.Domain

	// Add command arguments to variables
	for k, v := range req.Args {
		variables[k] = v
	}

	// Execute command
	output, err := appinstaller.RunCommand(manifest, req.Command, variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute command",
			"details": err.Error(),
			"output":  output,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Command '%s' executed successfully", req.Command),
		"output":  output,
	})
}

// GetApplicationLogs returns application logs from manifest-defined command
func GetApplicationLogs(c *gin.Context) {
	appID := c.Param("id")

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if app.Language == "podman" {
		output, err := containerapps.GetContainerAppLogsInternal(&app, c.Query("lines"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch logs",
				"details": err.Error(),
				"output":  output,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"logs":    output,
		})
		return
	}

	// Load manifest
	manifest, err := appinstaller.FetchManifest(app.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manifest"})
		return
	}

	// Prepare variables
	var variables map[string]string
	if app.Variables != "" {
		json.Unmarshal([]byte(app.Variables), &variables)
	} else {
		variables = make(map[string]string)
	}
	variables["APP_PATH"] = app.Path
	variables["APP_NAME"] = app.Title
	variables["DOMAIN"] = app.Domain

	// Get lines parameter if provided
	if lines := c.Query("lines"); lines != "" {
		variables["LINES"] = lines
	}

	// Execute logs command from manifest
	output, err := appinstaller.RunCommand(manifest, "logs", variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    output,
	})
}

// Helper functions

// getApplicationStatus determines the current status of an application
func getApplicationStatus(app *repository.Application) ApplicationStatus {
	status := ApplicationStatus{
		Status:      app.Status,
		IsRunning:   app.Status == "running",
		IsHealthy:   false,
		LastChecked: time.Now().Format(time.RFC3339),
		Uptime:      "",
	}

	// Check process health based on process manager
	if app.Status == "running" {
		// This would be expanded to check actual process health
		// For now, assume healthy if running
		status.IsHealthy = true
	}

	return status
}

// ListEditableFiles lists all editable files for an application
func ListEditableFiles(c *gin.Context) {
	appID := c.Param("id")

	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application ID is required"})
		return
	}

	db := repository.GetDB()
	var app repository.Application
	if err := db.Where("id = ?", appID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if app.Language == "podman" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"files":   []string{},
		})
		return
	}

	// Get manifest to fetch editable files
	manifest, err := appinstaller.FetchManifest(app.Provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch manifest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"files":   manifest.EditableFiles,
	})
}
