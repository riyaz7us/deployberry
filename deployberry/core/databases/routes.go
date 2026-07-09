package databases

import (
	"deployberry/core/dbops"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	dbs := r.Group("/databases")
	{
		dbs.GET("", CheckAll)
		// Generic database lifecycle endpoints
		dbs.GET("/:db/versions", ListVersions)
		dbs.GET("/:db/version", GetCurrentVersion)
		dbs.POST("/:db/install", InstallDatabase)
		dbs.POST("/:db/uninstall", UninstallDatabase)
		dbs.POST("/:db/uninstallSystem", UninstallDatabase)
		dbs.POST("/:db/activate", ActivateDatabase)
		dbs.POST("/:db/deactivate", DeactivateDatabase)
		dbs.POST("/:db/configure_password", ConfigurePassword)

		// Dynamic Database Engine Operations
		engineGroup := dbs.Group("/:db", dbops.ProviderMiddleware())
		{
			engineGroup.GET("/installed", dbops.CheckInstalled)
			engineGroup.GET("/credentials", dbops.GetCredentials)
			engineGroup.POST("/credentials/update", dbops.UpdateCredentials)
			engineGroup.GET("/databases", dbops.ListDatabases)
			engineGroup.POST("/database/create", dbops.CreateDatabase)
			engineGroup.POST("/database/delete", dbops.DeleteDatabase)

			engineGroup.GET("/users", dbops.ListUsers)
			engineGroup.POST("/user/create", dbops.CreateUser)
			engineGroup.POST("/user/delete", dbops.DeleteUser)
			engineGroup.POST("/privileges/grant", dbops.GrantPrivileges)
			engineGroup.POST("/privileges/revoke", dbops.RevokePrivileges)
			engineGroup.POST("/query", dbops.ExecuteControl)
			engineGroup.POST("/exec", dbops.ExecuteControl)
		}

		dbs.GET("/hardCheck", HardCheckAll)
	}
}
