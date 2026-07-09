package files

import (
	"net/http"
	"shared/files"

	"github.com/gin-gonic/gin"
)

// ExecuteGitOperation handles various git operations for a given path
func ExecuteGitOperation(c *gin.Context) {
	var op files.GitOperation
	if err := c.ShouldBindJSON(&op); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result files.GitResult

	switch op.Action {
	case "pull":
		result = files.GitPull(op)
	case "force_pull":
		result = files.GitForcePull(op)
	case "stash":
		result = files.GitStash(op)
	case "apply_stash":
		result = files.GitApplyStash(op)
	case "reset_hard":
		result = files.GitResetHard(op)
	case "status":
		result = files.GitStatusCmd(op)
	case "log":
		result = files.GitLog(op)
	case "fetch":
		result = files.GitFetch(op)
	case "branch":
		result = files.GitBranch(op)
	case "checkout":
		result = files.GitCheckout(op)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Unknown git action"})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusInternalServerError, result)
	}
}

// GetGitStatus returns comprehensive git status for a path
func GetGitStatus(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	status, err := files.GetGitRepoStatus(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetSSHKey handles SSH key retrieval and generation
func GetSSHKey(c *gin.Context) {
	result := files.GetOrGenerateSSHKey()

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusInternalServerError, result)
	}
}
