package applications

import (
	"net/http"
	"shared/repository"

	"github.com/gin-gonic/gin"
)

func GetAll(c *gin.Context) {
	db := repository.GetDB()
	// Get all applications from database
	var applications []repository.Application
	err := db.Order("created_at DESC").Find(&applications).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get applications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    applications,
	})
}
