package languages

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup) {
	langs := r.Group("/languages")
	langs.GET("/", CheckAll)
	// Generic endpoints for languages
	langs.GET("/:lang/versions", ListVersions)
	langs.GET("/:lang/current", GetCurrentVersion)
	langs.POST("/:lang/install", InstallLanguage)
	langs.POST("/:lang/uninstall", UninstallLanguage)
	langs.POST("/:lang/uninstallSystem", UninstallLanguage)
	langs.POST("/:lang/activate", ActivateLanguage)
	langs.POST("/:lang/deactivate", DeactivateLanguage)

	// Keep specific legacy routes that handle extra config (like venv/config arrays)
	langs.GET("/php/config", GetPHPConfig)
	langs.POST("/php/config", UpdatePHPConfig)

	langs.POST("/python/venv", CreatePythonVenv)



	langs.GET("/hardCheck", HardCheckAll)
}
