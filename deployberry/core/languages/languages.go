package languages

import (
	"fmt"
	"net/http"
	"shared/globals"
	"shared/repository"
	"shared/shell"
	"shared/system/installer"
	"shared/system/manifest"
	"strings"

	"github.com/gin-gonic/gin"
)

// Generic Request Body
type LanguageRequest struct {
	Version    string   `json:"version"`
	Extensions []string `json:"extensions"` // Used mainly for PHP
}

func parseRequest(c *gin.Context) LanguageRequest {
	var req LanguageRequest
	// Try parsing JSON body first
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback to query parameters
		req.Version = c.Query("version")
		exts := c.Query("extensions")
		if exts != "" {
			req.Extensions = strings.Split(exts, ",")
		}
	}
	return req
}

func ListVersions(c *gin.Context) {
	langName := c.Param("lang")
	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", langName), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Manifest not found: " + err.Error()})
		return
	}

	available := m.Versions
	installedSet := map[string]bool{}
	db := repository.GetDB()

	var langs []repository.Language
	db.Where("name = ?", langName).Find(&langs)

	var active string
	for _, l := range langs {
		installedSet[l.Version] = true
		if l.Active {
			active = l.Version
		}
	}

	type Entry struct {
		Version   string `json:"version"`
		Installed bool   `json:"installed"`
		Active    bool   `json:"active"`
	}
	var entries []Entry
	for _, v := range available {
		entries = append(entries, Entry{
			Version:   v,
			Installed: installedSet[v],
			Active:    v == active,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": entries})
}

func GetCurrentVersion(c *gin.Context) {
	langName := c.Param("lang")
	db := repository.GetDB()
	var lang repository.Language
	if err := db.Where("name = ? AND active = ?", langName, true).First(&lang).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "version": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "version": lang.Version})
}

// InstallLanguage acts as the universal installer for node, php, python, etc.
func InstallLanguage(c *gin.Context) {
	lang := c.Param("lang") // Matches /languages/:lang/install
	req := parseRequest(c)

	if req.Version == "" {
		m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", lang), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": fmt.Sprintf("Failed to load %s manifest to determine latest version", lang),
				"error":   err.Error(),
			})
			return
		}
		if len(m.Versions) > 0 {
			req.Version = m.Versions[0]
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No versions found in %s manifest", lang)})
			return
		}
	}

	// Dynamic data to pass into the manifest text/template parser
	data := map[string]interface{}{
		"Version":    req.Version,
		"Extensions": req.Extensions,
		"EnvPath":    globals.ENV_PATH,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", lang), data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to load %s manifest", lang),
			"error":   err.Error(),
		})
		return
	}

	runner := installer.NewRunner()
	steps, err := runner.RunAction(m, "install")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to install %s %s", lang, req.Version),
			"error":   err.Error(),
			"steps":   steps,
		})
		return
	}

	// Persist to database
	db := repository.GetDB()
	langRecord := repository.Language{
		Name:    lang,
		Version: req.Version,
	}
	_ = db.Where("name = ? AND version = ?", lang, req.Version).FirstOrCreate(&langRecord).Error
	langRecord.Active = false // Newly installed versions are not automatically active in DB unless specified
	db.Save(&langRecord)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%s %s installed successfully", lang, req.Version),
		"steps":   steps,
	})
}

// UninstallLanguage universally removes a language
func UninstallLanguage(c *gin.Context) {
	lang := c.Param("lang")
	req := parseRequest(c)

	if req.Version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Version parameter is required"})
		return
	}

	data := map[string]interface{}{
		"Version":    req.Version,
		"Extensions": req.Extensions,
		"EnvPath":    globals.ENV_PATH,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", lang), data)
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
	db.Where("name = ? AND version = ?", lang, req.Version).Delete(&repository.Language{})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Uninstalled %s %s", lang, req.Version), "steps": steps})
}

// ActivateLanguage marks a language version as active and runs the activate action
func ActivateLanguage(c *gin.Context) {
	lang := c.Param("lang")
	req := parseRequest(c)

	if req.Version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Version parameter is required"})
		return
	}

	data := map[string]interface{}{
		"Version": req.Version,
		"EnvPath": globals.ENV_PATH,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", lang), data)
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

	// Update DB to reflect active version
	db := repository.GetDB()
	db.Model(&repository.Language{}).Where("name = ?", lang).Update("active", false)
	db.Model(&repository.Language{}).Where("name = ? AND version = ?", lang, req.Version).Update("active", true)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Activated %s %s", lang, req.Version), "steps": steps})
}

// DeactivateLanguage marks a language version as inactive
func DeactivateLanguage(c *gin.Context) {
	lang := c.Param("lang")
	req := parseRequest(c)

	if req.Version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Version parameter is required"})
		return
	}

	data := map[string]interface{}{
		"Version": req.Version,
		"EnvPath": globals.ENV_PATH,
	}

	m, err := manifest.LoadAndRender(fmt.Sprintf("%s.yaml", lang), data)
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

	db := repository.GetDB()
	db.Model(&repository.Language{}).Where("name = ? AND version = ?", lang, req.Version).Update("active", false)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("Deactivated %s %s", lang, req.Version), "steps": steps})
}

func checkPHP() (string, bool) {
	if out, ok := shell.GetCommandVersion("php -v"); ok {
		return strings.Split(out, "\n")[0], true
	}
	return "", false
}

func checkGo() (string, bool) {
	if out, ok := shell.GetCommandVersion("go version"); ok {
		return out, true
	}
	return "", false
}

func checkNode() (string, bool) {
	return shell.GetCommandVersion("node -v")
}

func checkPython() (string, bool) {
	if ensurePyenvInstalled() == nil {
		if out, err := pyenvCommandWithEnv("version-name"); err == nil {
			return strings.TrimSpace(out), true
		}
	}
	if out, ok := shell.GetCommandVersion("python3 --version"); ok {
		return out, true
	}
	return shell.GetCommandVersion("python --version")
}

func HardCheckAll(c *gin.Context) {
	langParam := c.DefaultQuery("lang", "")
	var languagesToCheck []string
	if langParam == "" {
		languagesToCheck = []string{"php", "node", "python", "golang"}
	} else {
		languagesToCheck = strings.Split(langParam, ",")
	}

	data := map[string]any{}
	for _, lang := range languagesToCheck {
		langClean := strings.TrimSpace(strings.ToLower(lang))
		var version string
		var ok bool
		switch langClean {
		case "php":
			version, ok = checkPHP()
		case "node":
			version, ok = checkNode()
		case "python":
			version, ok = checkPython()
		case "golang":
			version, ok = checkGo()
		}
		if ok && version != "" {
			data[langClean] = version
		} else {
			data[langClean] = ""
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func CheckAll(c *gin.Context) {
	data, err := CheckAllService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
