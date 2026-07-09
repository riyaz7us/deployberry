package main

import (
	"containerapps"

	"github.com/gin-gonic/gin"
)

func init() {
	routeRegistries = append(routeRegistries, func(r *gin.Engine, api *gin.RouterGroup) {
		containerapps.RegisterRoutes(api)
	})
}
