package processes

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup) {
	pm2 := r.Group("/pm2")
	{
		// PM2 routes
		pm2.GET("/installed", Pm2Installed)
		pm2.GET("/list", Pm2List)
		pm2.POST("/start", Pm2StartAll)
		pm2.POST("/stop", Pm2StopAll)
		pm2.POST("/restart", Pm2RestartAll)
		pm2.POST("/process/start", Pm2ProcessStart)
		pm2.POST("/process/stop", Pm2ProcessStop)
		pm2.POST("/process/restart", Pm2ProcessRestart)
		pm2.POST("/process/delete", Pm2ProcessDelete)
		pm2.POST("/process/create", Pm2ProcessStart)
		pm2.POST("/save", Pm2Save)
		pm2.POST("/resurrect", Pm2Resurrect)
	}
}
