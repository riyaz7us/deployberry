package cronmanager

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup) {
	sqlcron := r.Group("/sql")
	{
		sqlcron.GET("/backup/crons", GetBackupCrons)
		sqlcron.POST("/backup/cron", CreateBackupCron)
		sqlcron.DELETE("/backup/cron", DeleteBackupCron)

		sqlcron.GET("/backups", GetBackupsList)
		sqlcron.POST("/backup/create", TriggerBackup)
		sqlcron.POST("/backup/restore", RestoreBackupHandler)
		sqlcron.POST("/backup/delete", DeleteBackupFile)
	}
}
