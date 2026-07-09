package files

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup) {
	filemanager := r.Group("/filemanager")
	files := r.Group("/files")
	git := r.Group("/git")
	{
		// File Manager routes
		files.GET("", ListFiles)
		files.POST("/upload", UploadFile)
		files.POST("/delete", DeleteFile)
		files.POST("/mkdir", CreateDirectory)
		files.POST("/rename", RenameFile)
		filemanager.POST("/download", DownloadFile)
		filemanager.POST("/read", ReadFile)
		filemanager.POST("/write", WriteFile)

		// Git Management routes
		git.POST("/operation", ExecuteGitOperation)
		git.GET("/status", GetGitStatus)
		git.GET("/ssh-key", GetSSHKey)
	}
}
