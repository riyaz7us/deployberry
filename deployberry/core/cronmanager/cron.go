package cronmanager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"deployberry/core/dbops"
	"deployberry/utils"
	cronengine "shared/cronmanager"
	"shared/globals"
	"shared/repository"

	"github.com/gin-gonic/gin"
)

func init() {
	// Inject the PerformBackup logic into the cron engine to avoid circular dependencies
	cronengine.RegisterBackupFunc(dbops.PerformBackup)
}

func GetBackupCrons(c *gin.Context) {
	db := repository.GetDB()
	database := c.Query("database")
	if database == "" {
		c.JSON(400, gin.H{"error": "database parameter is required"})
		return
	}

	var cronJobs []repository.Cron
	err := db.Where("name = ?", database).Find(&cronJobs).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var cronStrings []string
	for _, cronJob := range cronJobs {
		if cronJob.Schedule != "" {
			cronStrings = append(cronStrings, cronJob.Schedule)
		}
	}

	c.JSON(200, gin.H{"success": true, "crons": cronStrings})
}

func CreateBackupCron(c *gin.Context) {
	var req struct {
		Database   string `json:"database"`
		CronString string `json:"cron_string"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Database == "" || req.CronString == "" {
		c.JSON(400, gin.H{"error": "database and cron_string are required"})
		return
	}

	db := repository.GetDB()
	var existingCron repository.Cron
	err := db.Where("name = ? AND schedule = ?", req.Database, req.CronString).First(&existingCron).Error
	if err == nil {
		c.JSON(400, gin.H{"error": "this cron schedule already exists"})
		return
	}

	cronJob := &repository.Cron{
		Name:     req.Database,
		Command:  fmt.Sprintf("backup-%s", req.Database),
		Schedule: req.CronString,
		Active:   true,
	}

	err = db.Create(cronJob).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = cronengine.AddCronJob(req.Database, req.CronString)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to schedule cron: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": fmt.Sprintf("Backup schedule added for database '%s'", req.Database),
	})
}

func DeleteBackupCron(c *gin.Context) {
	database := c.Query("database")
	cronString := c.Query("cron")

	if database == "" || cronString == "" {
		c.JSON(400, gin.H{"error": "database and cron parameters are required"})
		return
	}

	db := repository.GetDB()
	err := db.Where("name = ? AND schedule = ?", database, cronString).Delete(&repository.Cron{}).Error
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	cronengine.RemoveCronJob(database, cronString)

	c.JSON(200, gin.H{
		"success": true,
		"message": "Backup schedule removed successfully",
	})
}

func GetBackupsList(c *gin.Context) {
	database := c.Query("database")
	if database == "" {
		c.JSON(400, gin.H{"error": "database parameter is required"})
		return
	}

	backupDir := filepath.Join(globals.BACKUPS_DIR, "mysql")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		c.JSON(200, gin.H{"success": true, "backups": []interface{}{}})
		return
	}

	filesList, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to read backup directory: %v", err)})
		return
	}

	type BackupFileInfo struct {
		Name string    `json:"name"`
		Path string    `json:"path"`
		Size int64     `json:"size"`
		Date time.Time `json:"date"`
	}

	var backups []BackupFileInfo
	prefix := database + "_"
	for _, file := range filesList {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".sql") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			filePath := filepath.Join(backupDir, name)
			backups = append(backups, BackupFileInfo{
				Name: name,
				Path: filePath,
				Size: info.Size(),
				Date: info.ModTime(),
			})
		}
	}

	c.JSON(200, gin.H{"success": true, "backups": backups})
}

func TriggerBackup(c *gin.Context) {
	var req struct {
		Database string `json:"database"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Database == "" {
		c.JSON(400, gin.H{"error": "database is required"})
		return
	}

	err := dbops.PerformBackup(req.Database)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Backup created successfully"})
}

func RestoreBackupHandler(c *gin.Context) {
	var req struct {
		Database   string `json:"database"`
		BackupFile string `json:"backup_file"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Database == "" || req.BackupFile == "" {
		c.JSON(400, gin.H{"error": "database and backup_file are required"})
		return
	}

	cleanPath := filepath.Clean(req.BackupFile)
	expectedPrefix := filepath.Clean(filepath.Join(globals.BACKUPS_DIR, "mysql"))
	if !strings.HasPrefix(cleanPath, expectedPrefix) {
		c.JSON(400, gin.H{"error": "invalid backup file path"})
		return
	}

	rootPassword, port, err := utils.GetMySQLRootCredentials()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get MySQL credentials: %v", err)})
		return
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to open backup file: %v", err)})
		return
	}
	defer file.Close()

	cmd := exec.Command("mysql",
		fmt.Sprintf("--port=%d", port),
		"-u", "root",
		fmt.Sprintf("-p%s", rootPassword),
		req.Database,
	)
	cmd.Stdin = file

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to restore backup: %v, stderr: %s", err, stderr.String())})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Backup restored successfully"})
}

func DeleteBackupFile(c *gin.Context) {
	var req struct {
		Database   string `json:"database"`
		BackupFile string `json:"backup_file"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if req.Database == "" || req.BackupFile == "" {
		c.JSON(400, gin.H{"error": "database and backup_file are required"})
		return
	}

	cleanPath := filepath.Clean(req.BackupFile)
	expectedPrefix := filepath.Clean(filepath.Join(globals.BACKUPS_DIR, "mysql"))
	if !strings.HasPrefix(cleanPath, expectedPrefix) {
		c.JSON(400, gin.H{"error": "invalid backup file path"})
		return
	}

	if err := os.Remove(cleanPath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to delete backup file: %v", err)})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Backup file deleted successfully"})
}
