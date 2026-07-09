package databases

import (
	"deployberry/core/databases/common"
	"deployberry/utils"
	"fmt"
	"net/http"
	"shared/repository"
	"shared/shell"
	"shared/system/installer"
	"shared/system/manifest"
	"strings"

	"github.com/gin-gonic/gin"
)

type DatabaseRequest struct {
	Version  string `json:"version"`
	Password string `json:"password"`
}

func parseRequest(c *gin.Context) DatabaseRequest {
	var req DatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Version = c.Query("version")
		req.Password = c.Query("password")
	}
	return req
}

func ListVersions(c *gin.Context) {
	dbName := c.Param("db")
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	db := repository.GetDB()
	var dbServer repository.DatabaseServer
	installed := false
	if err := db.Where("type = ?", dbName).First(&dbServer).Error; err == nil {
		installed = true
	}

	var entries []common.DBVersion
	for _, v := range m.Versions {
		isThisVersion := installed && dbServer.Version == v
		entry := common.DBVersion{
			Version:   v,
			Installed: isThisVersion,
			Active:    isThisVersion && dbServer.Active,
		}
		if isThisVersion {
			entry.RootPassword = dbServer.RootPassword
			entry.Port = dbServer.Port
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": entries})
}

func GetCurrentVersion(c *gin.Context) {
	dbName := c.Param("db")
	db := repository.GetDB()
	var dbServer repository.DatabaseServer
	if err := db.Where("type = ? AND active = ?", dbName, true).First(&dbServer).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "version": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "version": dbServer.Version})
}

func InstallDatabase(c *gin.Context) {
	dbName := c.Param("db")
	req := parseRequest(c)

	if req.Password == "" {
		if pass, err := utils.GenerateSecurePassword(16); err == nil {
			req.Password = pass
		}
	}

	if req.Version == "" {
		m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": fmt.Sprintf("Failed to load %s manifest to determine latest version", dbName),
				"error":   err.Error(),
			})
			return
		}
		if len(m.Versions) > 0 {
			req.Version = m.Versions[0]
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No versions found in %s manifest", dbName)})
			return
		}
	}

	data := map[string]interface{}{
		"Version":  req.Version,
		"Password": req.Password,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "install")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	// Persist to DB
	db := repository.GetDB()
	dbRecord := repository.DatabaseServer{
		Type: dbName,
	}
	_ = db.Where("type = ?", dbName).FirstOrCreate(&dbRecord).Error
	dbRecord.Version = req.Version
	if req.Password != "" {
		dbRecord.RootPassword = req.Password
	}
	// Assign standard ports
	switch dbName {
	case "mysql", "mariadb":
		dbRecord.Port = 3306
	case "postgres":
		dbRecord.Port = 5432
	case "redis":
		dbRecord.Port = 6379
	case "mongodb":
		dbRecord.Port = 27017
	}
	dbRecord.Active = true // Active after successful install
	db.Save(&dbRecord)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s installed", dbName), "steps": steps})
}

func UninstallDatabase(c *gin.Context) {
	dbName := c.Param("db")

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "uninstall")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	db := repository.GetDB()
	db.Where("type = ?", dbName).Delete(&repository.DatabaseServer{})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("%s uninstalled", dbName), "steps": steps})
}

func ConfigurePassword(c *gin.Context) {
	dbName := c.Param("db")
	req := parseRequest(c)

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password parameter is required"})
		return
	}

	data := map[string]interface{}{
		"Password": req.Password,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "configure_password")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	// Persist to DB
	db := repository.GetDB()
	db.Model(&repository.DatabaseServer{}).Where("type = ?", dbName).Update("root_password", req.Password)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Password configured for %s", dbName), "steps": steps})
}

func ActivateDatabase(c *gin.Context) {
	dbName := c.Param("db")
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "activate")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	// Update DB to active
	db := repository.GetDB()
	db.Model(&repository.DatabaseServer{}).Where("type = ?", dbName).Update("active", true)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Activated %s", dbName), "steps": steps})
}

func DeactivateDatabase(c *gin.Context) {
	dbName := c.Param("db")
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", dbName), map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "deactivate")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "steps": steps})
		return
	}

	// Update DB to inactive
	db := repository.GetDB()
	db.Model(&repository.DatabaseServer{}).Where("type = ?", dbName).Update("active", false)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Deactivated %s", dbName), "steps": steps})
}

func checkMySQL() (string, bool) {
	if out, ok := shell.GetCommandVersion("mysql --version"); ok {
		if !strings.Contains(strings.ToLower(out), "mariadb") {
			return out, true
		}
	}
	return "", false
}

func checkMariaDB() (string, bool) {
	if out, ok := shell.GetCommandVersion("mariadb --version"); ok {
		return out, true
	}
	if out, ok := shell.GetCommandVersion("mysql --version"); ok {
		if strings.Contains(strings.ToLower(out), "mariadb") {
			return out, true
		}
	}
	return "", false
}

func HardCheckAll(c *gin.Context) {
	dbParam := c.DefaultQuery("db", "")
	var databasesToCheck []string
	if dbParam == "" {
		databasesToCheck = []string{"mysql", "postgres", "redis", "mongodb", "sqlite", "mariadb"}
	} else {
		databasesToCheck = strings.Split(dbParam, ",")
	}

	data := map[string]any{}
	for _, db := range databasesToCheck {
		dbClean := strings.TrimSpace(strings.ToLower(db))
		var version string
		var ok bool
		switch dbClean {
		case "mysql":
			version, ok = checkMySQL()
			if ok {
				data["mysql_type"] = "mysql"
			}
		case "mariadb":
			version, ok = checkMariaDB()
		case "postgres":
			version, ok = shell.GetCommandVersion("psql --version")
		case "redis":
			version, ok = shell.GetCommandVersion("redis-server --version")
		case "mongodb":
			version, ok = shell.GetCommandVersion("mongod --version")
			if ok {
				version = strings.Split(version, "\n")[0]
			}
		case "sqlite":
			version, ok = shell.GetCommandVersion("sqlite3 --version")
		}
		if ok && version != "" {
			data[dbClean] = version
		} else {
			data[dbClean] = ""
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func CheckAll(c *gin.Context) {
	db := repository.GetDB()
	var databaseServers []repository.DatabaseServer
	err := db.Find(&databaseServers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve installed database servers",
		})
		return
	}

	data := gin.H{
		"mysql":    "",
		"mariadb":  "",
		"redis":    "",
		"postgres": "",
		"sqlite":   "",
		"mongodb":  "",
	}

	for _, server := range databaseServers {
		if server.Active && server.Version != "" {
			data[server.Type] = server.Version
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
