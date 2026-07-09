package applications

import (
	"github.com/gin-gonic/gin"
)

// RegisterApplicationManagerRoutes registers all application management routes
func RegisterRoutes(r *gin.RouterGroup) {
	apps := r.Group("/applications")
	{
		// Application CRUD
		apps.GET("", ListApplications)
		apps.GET("/:id", GetApplication)
		apps.GET("/:id/status", GetApplicationStatus)
		apps.GET("/:id/config", GetApplicationConfig)
		apps.PUT("/:id/config", UpdateApplication)

		// Application lifecycle - all use generic command execution
		apps.POST("/:id/start", ExecuteApplicationCommand("start"))
		apps.POST("/:id/stop", ExecuteApplicationCommand("stop"))
		apps.POST("/:id/restart", ExecuteApplicationCommand("restart"))
		apps.DELETE("/:id", DeleteApplication)

		// Application operations
		apps.POST("/:id/update", UpdateApplicationManifest)
		apps.POST("/:id/command", RunApplicationCommand)
		apps.GET("/:id/logs", GetApplicationLogs)
		apps.GET("/:id/files", ListEditableFiles)
	}
}
