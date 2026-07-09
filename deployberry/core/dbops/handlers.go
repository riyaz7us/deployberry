package dbops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// StringListOrString is a custom type that unmarshals either a string list or a comma-separated string
type StringListOrString []string

func (s *StringListOrString) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*s = list
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			*s = nil
			return nil
		}
		parts := strings.Split(str, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		*s = parts
		return nil
	}

	return fmt.Errorf("invalid json type: privileges must be a string array or a comma-separated string")
}

// ProviderMiddleware intercepts engine requests, resolves database provider, and sets context variables
func ProviderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		engine := c.Param("db")
		provider, err := GetProvider(engine)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("db_provider", provider)
		c.Set("db_engine", engine)
		c.Next()
	}
}

func getProvider(c *gin.Context) DBProvider {
	return c.MustGet("db_provider").(DBProvider)
}

func getEngine(c *gin.Context) string {
	return c.MustGet("db_engine").(string)
}

// CheckInstalled checks if the database server package is installed
func CheckInstalled(c *gin.Context) {
	provider := getProvider(c)
	engine := getEngine(c)

	installed, version, connErr := provider.IsInstalled()
	if !installed {
		c.JSON(http.StatusOK, gin.H{"installed": false})
		return
	}

	typeStr := engine
	if engine == "sql" {
		typeStr = "mysql"
	}

	res := gin.H{
		"installed": version,
		"type":      typeStr,
		"connected": connErr == nil,
	}
	if connErr != nil {
		res["error"] = connErr.Error()
	}
	c.JSON(http.StatusOK, res)
}

// GetCredentials retrieves database credentials from config
func GetCredentials(c *gin.Context) {
	provider := getProvider(c)
	engine := getEngine(c)

	username, password, err := provider.GetCredentials()
	if err != nil {
		if engine == "postgres" {
			c.JSON(http.StatusOK, gin.H{"username": "postgres", "password": ""})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"password": password,
	})
}

// UpdateCredentials updates database credentials in config
func UpdateCredentials(c *gin.Context) {
	provider := getProvider(c)
	engine := getEngine(c)

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	if err := provider.UpdateCredentials(req.Username, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	engineLabel := engine
	if engine == "sql" {
		engineLabel = "MySQL"
	} else if engine == "postgres" {
		engineLabel = "PostgreSQL"
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s credentials updated successfully", engineLabel)})
}

// ListDatabases lists databases
func ListDatabases(c *gin.Context) {
	provider := getProvider(c)

	dbs, err := provider.ListDatabases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var results []map[string]interface{}
	for _, dbName := range dbs {
		results = append(results, map[string]interface{}{
			"name": dbName,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "databases": results})
}

// CreateDatabase creates a new database
func CreateDatabase(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		AppName  string `json:"app_name"`
		Name     string `json:"name"`
		Database string `json:"database"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := req.AppName
	if name == "" {
		name = req.Name
	}
	if name == "" {
		name = req.Database
	}

	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_name or name is required"})
		return
	}

	result, err := provider.CreateDatabaseInternal(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  result.Message,
		"database": result.Database,
		"username": result.Username,
	})
}

// DeleteDatabase deletes a database
func DeleteDatabase(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		Database string `json:"database"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbName := req.Database
	if dbName == "" {
		dbName = req.Name
	}

	if dbName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
		return
	}

	if err := provider.DeleteDatabaseInternal(dbName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database deleted successfully", "success": true})
}

// ListUsers lists database users
func ListUsers(c *gin.Context) {
	provider := getProvider(c)
	engine := getEngine(c)

	users, err := provider.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := gin.H{"users": users}
	if engine == "postgres" {
		res["success"] = true
	}
	c.JSON(http.StatusOK, res)
}

// CreateUser creates a new database user
func CreateUser(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		Username string `json:"username"`
		User     string `json:"user"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		username = req.User
	}

	if username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	if err := provider.CreateUserInternal(username, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "success": true})
}

// DeleteUser deletes a database user
func DeleteUser(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		Username string `json:"username"`
		User     string `json:"user"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		username = req.User
	}

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if err := provider.DeleteUserInternal(username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "success": true})
}

// GrantPrivileges grants privileges to a database user
func GrantPrivileges(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		Username   string             `json:"username"`
		User       string             `json:"user"`
		Database   string             `json:"database"`
		Privileges StringListOrString `json:"privileges"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		username = req.User
	}

	if username == "" || req.Database == "" || len(req.Privileges) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User, Database, and Privileges are required"})
		return
	}

	if err := provider.GrantPrivilegesInternal(username, req.Database, req.Privileges); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Privileges granted successfully", "success": true})
}

// RevokePrivileges revokes privileges from a database user
func RevokePrivileges(c *gin.Context) {
	provider := getProvider(c)

	var req struct {
		Username   string             `json:"username"`
		User       string             `json:"user"`
		Database   string             `json:"database"`
		Privileges StringListOrString `json:"privileges"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		username = req.User
	}

	if username == "" || req.Database == "" || len(req.Privileges) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User, Database, and Privileges are required"})
		return
	}

	if err := provider.RevokePrivilegesInternal(username, req.Database, req.Privileges); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Privileges revoked successfully", "success": true})
}

// ExecuteControl runs SQL query or command
func ExecuteControl(c *gin.Context) {
	provider := getProvider(c)

	action := c.Param("action")
	if action == "" {
		if strings.HasSuffix(c.Request.URL.Path, "/query") {
			action = "query"
		} else if strings.HasSuffix(c.Request.URL.Path, "/exec") {
			action = "exec"
		}
	}

	var req struct {
		SQL string `json:"sql"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := provider.ExecuteControl(action, req.SQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
