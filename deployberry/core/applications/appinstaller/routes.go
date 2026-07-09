package appinstaller

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes attaches the appinstaller endpoints to your Gin router
func RegisterRoutes(r *gin.RouterGroup) {
	apps := r.Group("/registry")
	{
		apps.GET("/", ListRegistryHandler)
		apps.GET("/:slug/requirements", GetRequirementsHandler)
		apps.GET("/:slug/commands", ListCommandsHandler)

		apps.POST("/:slug/install", InstallHandler)
		apps.POST("/:slug/update", UpdateHandler)
		apps.POST("/:slug/command", CommandHandler)
		apps.DELETE("/:slug/delete", DeleteHandler)
	}
}
