package processes

import (
	"shared/shell"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProcessConfig struct {
	Name   string   `json:"name"`
	Script string   `json:"script"`
	Args   []string `json:"args"`
}

func pm2Command(command string, args ...string) error {
	var cmdParts []string
	cmdParts = append(cmdParts, "pm2", command)
	for _, arg := range args {
		cmdParts = append(cmdParts, shell.EscapeShellArg(arg))
	}
	fullCmd := strings.Join(cmdParts, " ")
	res := shell.ExecuteAsAppUser(fullCmd, "")
	return res.Error
}

func Pm2Installed(c *gin.Context) {
	res := shell.ExecuteAsAppUser("pm2 -h", "")
	if res.Error != nil {
		c.JSON(200, gin.H{"success": false, "error": res.Error.Error(), "output": res.Output, "exitCode": res.ExitCode})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2List(c *gin.Context) {
	res := shell.ExecuteAsAppUser("pm2 jlist", "")
	if res.Error != nil {
		c.JSON(500, gin.H{"error": res.Error.Error(), "output": res.Output})
		return
	}
	c.Header("Content-Type", "application/json")
	c.String(200, res.Output)
}

func Pm2StartAll(c *gin.Context) {
	if err := pm2Command("start", "all"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2StopAll(c *gin.Context) {
	if err := pm2Command("stop", "all"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2RestartAll(c *gin.Context) {
	if err := pm2Command("restart", "all"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func runPm2Action(action string, config ProcessConfig) error {
	var cmdArgs []string
	if action == "start" {
		if config.Name != "" {
			cmdArgs = append(cmdArgs, "--name", config.Name)
		}
		cmdArgs = append(cmdArgs, config.Script)
		if len(config.Args) > 0 {
			cmdArgs = append(cmdArgs, config.Args...)
		}
	} else {
		// For stop/restart/delete, use Name if available, otherwise Script
		target := config.Name
		if target == "" {
			target = config.Script
		}
		cmdArgs = append(cmdArgs, target)
	}
	return pm2Command(action, cmdArgs...)
}

func Pm2ProcessStart(c *gin.Context) {
	var config ProcessConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := runPm2Action("start", config); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2ProcessStop(c *gin.Context) {
	var config ProcessConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := runPm2Action("stop", config); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2ProcessRestart(c *gin.Context) {
	var config ProcessConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := runPm2Action("restart", config); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2ProcessSave(c *gin.Context) {
	if err := pm2Command("save"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2ProcessResurrect(c *gin.Context) {
	if err := pm2Command("resurrect"); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func Pm2ProcessDelete(c *gin.Context) {
	var config ProcessConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := runPm2Action("delete", config); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}


