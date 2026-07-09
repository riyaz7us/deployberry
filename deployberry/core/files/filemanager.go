package files

import (
	"fmt"
	"os"
	"path/filepath"
	"shared/files"

	"github.com/gin-gonic/gin"
)

func ListFiles(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	validatedPath, err := files.ValidatePathAccess(path, false)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	list, err := files.ListFiles(validatedPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, list)
}

func UploadFile(c *gin.Context) {
	destination := c.PostForm("destination")
	if destination == "" {
		c.JSON(400, gin.H{"error": "Destination path is required"})
		return
	}

	validatedDest, err := files.ValidatePathAccess(destination, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.SaveUploadedFile(file, validatedDest); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	files.EnsureCorrectFileOwnership(validatedDest)
	c.JSON(200, gin.H{"message": "File uploaded successfully"})
}

func DeleteFile(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" {
		c.JSON(400, gin.H{"error": "Path is required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	if err := files.DeleteFile(validatedPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "File deleted successfully"})
}

func CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" {
		c.JSON(400, gin.H{"error": "Path is required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	if err := files.CreateDirectory(validatedPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Directory created successfully"})
}

func RenameFile(c *gin.Context) {
	var req struct {
		Path     string `json:"path"`
		NewName  string `json:"new_name"`
		NewPath  string `json:"new_path"`
		NewOwner string `json:"new_owner"`
		NewGroup string `json:"new_group"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" || req.NewName == "" || req.NewPath == "" {
		c.JSON(400, gin.H{"error": "Path, new name, and new path are required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	targetPath := filepath.Join(req.NewPath, req.NewName)
	_, err = files.ValidatePathAccess(targetPath, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	if err := files.RenameFile(validatedPath, req.NewName, req.NewPath, req.NewOwner, req.NewGroup); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "File renamed successfully"})
}

func DownloadFile(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" {
		c.JSON(400, gin.H{"error": "Path is required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, false)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(validatedPath)))
	c.File(validatedPath)
}

func ReadFile(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" {
		c.JSON(400, gin.H{"error": "Path is required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, false)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	content, err := os.ReadFile(validatedPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read file: " + err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(200, string(content))
}

func WriteFile(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Path == "" {
		c.JSON(400, gin.H{"error": "Path is required"})
		return
	}

	validatedPath, err := files.ValidatePathAccess(req.Path, true)
	if err != nil {
		c.JSON(403, gin.H{"error": err.Error()})
		return
	}

	err = os.WriteFile(validatedPath, []byte(req.Content), 0644)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to write file: " + err.Error()})
		return
	}

	files.EnsureCorrectFileOwnership(validatedPath)
	c.JSON(200, gin.H{"message": "File saved successfully"})
}
