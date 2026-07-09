package containerapps

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the containerapps endpoints to the protected API router group
func RegisterRoutes(r *gin.RouterGroup) {
	apps := r.Group("/containerapps")
	{
		apps.GET("/registry", ListRegistryHandler)
		apps.GET("/registry/:slug/requirements", GetRequirementsHandler)
		apps.POST("/registry/:slug/install", InstallHandler)
		apps.POST("/registry/:slug/update", UpdateHandler)
		apps.POST("/registry/:slug/command", CommandHandler)
		apps.DELETE("/registry/:slug/delete", DeleteHandler)
	}
}