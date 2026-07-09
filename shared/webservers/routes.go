package webservers

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the webserver endpoints to the protected API router group
func RegisterRoutes(r *gin.RouterGroup) {
	nginx := r.Group("/nginx")
	caddy := r.Group("/caddy")
	webconfigs := r.Group("/webconfigs")
	{
		// Nginx routes
		nginx.GET("/installed", NginxInstalled)
		nginx.POST("/install", func(c *gin.Context) { InstallWebserver(c, "nginx") })
		nginx.POST("/start", func(c *gin.Context) { ActionWebserver(c, "nginx", "activate") })
		nginx.POST("/stop", func(c *gin.Context) { ActionWebserver(c, "nginx", "deactivate") })
		nginx.POST("/restart", func(c *gin.Context) { ActionWebserver(c, "nginx", "restart") })
		nginx.POST("/uninstall", func(c *gin.Context) { ActionWebserver(c, "nginx", "uninstall") })
		nginx.GET("/status", NginxStatus)
		nginx.GET("/configs", NginxConfigs)
		nginx.POST("/config/add", NginxConfigAdd)
		nginx.POST("/config/delete", NginxConfigDelete)
		nginx.POST("/config/enable", NginxConfigEnable)
		nginx.POST("/config/disable", NginxConfigDisable)

		// Caddy routes
		caddy.POST("/install", func(c *gin.Context) { InstallWebserver(c, "caddy") })
		caddy.POST("/start", func(c *gin.Context) { ActionWebserver(c, "caddy", "activate") })
		caddy.POST("/stop", func(c *gin.Context) { ActionWebserver(c, "caddy", "deactivate") })
		caddy.POST("/restart", func(c *gin.Context) { ActionWebserver(c, "caddy", "restart") })
		caddy.POST("/uninstall", func(c *gin.Context) { ActionWebserver(c, "caddy", "uninstall") })
		caddy.GET("/status", GetCaddyStatus)

		// Shared web configs management routes
		webconfigs.GET("", WebConfigList)
		webconfigs.POST("", WebConfigCreate)
		webconfigs.POST("/:domain/deploy", WebConfigDeploy)
		webconfigs.POST("/:domain/recreate", WebConfigRecreate)
		webconfigs.DELETE("/:domain", WebConfigDelete)
	}
}
