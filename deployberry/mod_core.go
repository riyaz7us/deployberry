//go:build !containerapps_only

package main

import (
	"io/fs"
	"log"
	"net/http"
	"path"
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
			reqPath := c.Request.URL.Path
			if strings.HasPrefix(reqPath, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
				return
			}
			
			cleanPath := strings.TrimPrefix(path.Clean(reqPath), "/")
			
			// 1. SSG/Nuxt route match (e.g. /login -> login/index.html)
			// This also perfectly handles the root "/" request (cleanPath="") -> "index.html"
			if data, err := fs.ReadFile(subFS, path.Join(cleanPath, "index.html")); err == nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", data)
				return
			}

			// 2. Exact file match for static assets (e.g. /_nuxt/app.js)
			if f, err := subFS.Open(cleanPath); err == nil && cleanPath != "" {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
			
			// 3. SPA dynamic route fallback (serve root index.html)
			data, _ := fs.ReadFile(subFS, "index.html")
			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	})
}
