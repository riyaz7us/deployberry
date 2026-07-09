//go:build !containerapps_only

package main

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"deployberry/core/applications"
	"deployberry/core/applications/appinstaller"
	corecron "deployberry/core/cronmanager"
	"deployberry/core/databases"
	"deployberry/core/files"
	"deployberry/core/handlers"
	"deployberry/core/languages"
	"deployberry/core/middleware"
	"deployberry/core/processes"
	"shared/cronmanager"
	"shared/webservers"
	"ui"

	"github.com/gin-gonic/gin"
)

func init() {
	authMiddleware = middleware.AuthMiddleware()
	initCallback = func() {
		if err := cronmanager.LoadCrons(); err != nil {
			log.Printf("Warning: Failed to load cron jobs: %v", err)
		}
	}
	routeRegistries = append(routeRegistries, func(r *gin.Engine, api *gin.RouterGroup) {
		// Public routes
		r.POST("/api/login", handlers.Login)

		// Protected core routes
		api.GET("stats", handlers.CheckResourceUsage)
		api.GET("/logged-in", handlers.LoggedIn)
		corecron.RegisterRoutes(api)
		files.RegisterRoutes(api)
		processes.RegisterRoutes(api)
		applications.RegisterRoutes(api)
		webservers.RegisterRoutes(api)
		languages.RegisterRoutes(api)
		databases.RegisterRoutes(api)
		appinstaller.RegisterRoutes(api)

		// Static UI file server
		subFS, err := fs.Sub(ui.Assets, "dist")
		if err != nil {
			log.Fatal("Failed to load embedded web assets: ", err)
		}
		fileServer := http.FileServer(http.FS(subFS))

		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}
			f, err := subFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
			c.FileFromFS("index.html", http.FS(subFS))
		})
	})
}
