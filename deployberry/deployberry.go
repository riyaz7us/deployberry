package main

import (
	"log"
	"net/http"
	"os"
	"shared/globals"
	"shared/shell"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var routeRegistries []func(r *gin.Engine, api *gin.RouterGroup)
var authMiddleware gin.HandlerFunc
var initCallback func()

func main() {
	if len(os.Args) > 1 {
		command := os.Args[1]
		if !globals.IsDevelopment() && os.Getuid() != 0 {
			log.Fatal("This program must be run as root")
		}
		shell.HandleCommand(command)
		return
	}

	if err := os.Chdir(globals.BASE_DIR); err != nil {
		log.Fatal("Failed to change to app directory: " + err.Error())
	}
	if err := shell.StartupCheck(); err != nil {
		log.Fatal("Startup Check Failed: " + err.Error())
	}

	if !globals.IsDevelopment() && os.Getuid() != 0 {
		log.Fatal("This program must be run as root")
	}

	// Initialize compiled modules if any
	if initCallback != nil {
		initCallback()
	}

	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup protected API group
	api := r.Group("/api")
	if authMiddleware != nil {
		api.Use(authMiddleware)
	}

	// Register routes from compiled modules
	for _, reg := range routeRegistries {
		reg(r, api)
	}

	// Default fallback handler for standalone mode
	if authMiddleware == nil {
		r.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":7717"
	}
	log.Fatal(r.Run(port))
}
