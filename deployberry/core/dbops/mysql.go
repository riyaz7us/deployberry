package dbops

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"deployberry/utils"
	"shared/globals"
	"shared/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// MySQL implementation functions

func GetDBConnection() (*sql.DB, error) {
	rootPassword, port, err := utils.GetMySQLRootCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get MySQL credentials: %v", err)
	}

	dsn := fmt.Sprintf("root:%s@tcp(localhost:%d)/", rootPassword, port)
	return sql.Open("mysql", dsn)
}

// PerformBackup creates a mysqldump backup of the given database
func PerformBackup(database string) error {
	rootPassword, port, err := utils.GetMySQLRootCredentials()
	if err != nil {
		return fmt.Errorf("failed to get MySQL credentials: %v", err)
	}

	backupDir := filepath.Join(globals.BACKUPS_DIR, "mysql")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	backupFile := fmt.Sprintf("%s/%s_%s.sql", backupDir, database, time.Now().Format("2006-01-02_15-04-05"))

	cmd := exec.Command("mysqldump",
		fmt.Sprintf("--port=%d", port),
		"-u", "root",
		fmt.Sprintf("-p%s", rootPassword),
		database,
		fmt.Sprintf("--result-file=%s", backupFile),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mysqldump failed: %v, output: %s", err, string(out))
	}

	return nil
}

// DBProvider Interface Methods for MySQLProvider

func (m *MySQLProvider) IsInstalled() (bool, string, error) {
	mysqlCmd := exec.Command("mysql", "--version")
	mysqlOutput, mysqlErr := mysqlCmd.Output()

	mariadbCmd := exec.Command("mariadb", "--version")
	mariadbOutput, mariadbErr := mariadbCmd.Output()

	if mysqlErr == nil {
		if err := utils.TestMySQLConnection(); err != nil {
			return true, string(mysqlOutput), err
		}
		return true, string(mysqlOutput), nil
	}
	if mariadbErr == nil {
		return true, string(mariadbOutput), nil
	}

	return false, "", nil
}

func (m *MySQLProvider) GetCredentials() (string, string, error) {
	db := repository.GetDB()
	var userConfig, passConfig repository.Config
	err := db.Where("key = ?", "dbuser").First(&userConfig).Error
	if err != nil {
		return "", "", fmt.Errorf("failed to get username config: %w", err)
	}
	err = db.Where("key = ?", "dbpass").First(&passConfig).Error
	if err != nil {
		return userConfig.Value, "", fmt.Errorf("failed to get password config: %w", err)
	}

	return userConfig.Value, passConfig.Value, nil
}

func (m *MySQLProvider) UpdateCredentials(username, password string) error {
	db := repository.GetDB()

	userConfig := repository.Config{Key: "dbuser", Value: username}
	if err := db.Where(repository.Config{Key: "dbuser"}).Assign(repository.Config{Value: username}).FirstOrCreate(&userConfig).Error; err != nil {
		return err
	}

	passConfig := repository.Config{Key: "dbpass", Value: password}
	return db.Where(repository.Config{Key: "dbpass"}).Assign(repository.Config{Value: password}).FirstOrCreate(&passConfig).Error
}

func (m *MySQLProvider) ListDatabases() ([]string, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, dbName)
	}

	return databases, nil
}

func (m *MySQLProvider) CreateDatabaseInternal(appName string) (*DatabaseResult, error) {
	if strings.TrimSpace(appName) == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	dbName := utils.UtilSanitizeName(appName, "db_")
	userName := dbName
	password, err := utils.GenerateSecurePassword(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database password: %w", err)
	}

	db, err := GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check if database already exists
	var exists int
	checkQuery := "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?"
	if err := db.QueryRow(checkQuery, dbName).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check database existence: %v", err)
	}
	if exists > 0 {
		return nil, fmt.Errorf("database already exists")
	}

	// Check if user already exists
	var userExists int
	checkUserQuery := "SELECT COUNT(*) FROM mysql.user WHERE User = ?"
	if err := db.QueryRow(checkUserQuery, userName).Scan(&userExists); err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}
	if userExists > 0 {
		return nil, fmt.Errorf("database user already exists")
	}

	// Create database
	createDBQuery := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	if _, err := db.Exec(createDBQuery); err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	// Preemptively drop user if exists and flush privileges to resolve any stale memory caches (prevents Error 1396)
	db.Exec(fmt.Sprintf("DROP USER IF EXISTS `%s`@'%%'", userName))
	db.Exec("FLUSH PRIVILEGES")

	// Create user and grant privileges
	escapedPassword := strings.ReplaceAll(password, "'", "''")
	createUserQuery := fmt.Sprintf("CREATE USER `%s`@'%%' IDENTIFIED BY '%s'", userName, escapedPassword)
	if _, err := db.Exec(createUserQuery); err != nil {
		// Try to clean up database if user creation fails
		db.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
		return nil, fmt.Errorf("failed to create database user: %v", err)
	}

	grantQuery := fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO `%s`@'%%'", dbName, userName)
	if _, err := db.Exec(grantQuery); err != nil {
		// Clean up if grant fails
		db.Exec(fmt.Sprintf("DROP USER `%s`@'%%'", userName))
		db.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
		return nil, fmt.Errorf("failed to grant privileges: %v", err)
	}

	// Flush privileges
	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		return nil, fmt.Errorf("database and user created but failed to flush privileges: %v", err)
	}

	// Save credentials to control panel database
	dbg := repository.GetDB()
	dbCred := repository.DatabaseCredential{
		Host:     "localhost",
		Port:     3306,
		Database: dbName,
		Username: userName,
		Password: password,
		Type:     "mysql",
	}
	dbg.Create(&dbCred)

	_, port, _ := utils.GetMySQLRootCredentials()
	if port == 0 {
		port = 3306
	}

	message := fmt.Sprintf("%s database and user created successfully", appName)

	return &DatabaseResult{
		Database: dbName,
		Username: userName,
		Password: password,
		Message:  message,
		Port:     port,
	}, nil
}

func (m *MySQLProvider) DeleteDatabaseInternal(databaseName string) error {
	if databaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if strings.ContainsAny(databaseName, "`; ") {
		return fmt.Errorf("invalid characters in database name")
	}

	db, err := GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	query := fmt.Sprintf("DROP DATABASE `%s`", databaseName)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to drop database: %v", err)
	}

	// Remove credentials from database
	repository.GetDB().Where("database = ?", databaseName).Delete(&repository.DatabaseCredential{})

	return nil
}

type GrantInfo struct {
	Database   string   `json:"database"`
	Privileges []string `json:"privileges"`
}

type UserInfo struct {
	Username string      `json:"username"`
	Host     string      `json:"host"`
	Grants   []GrantInfo `json:"grants"`
}

func parseGrantString(grantStr string) *GrantInfo {
	grantStr = strings.ReplaceAll(grantStr, "  ", " ")
	onIdx := strings.Index(grantStr, " ON ")
	toIdx := strings.Index(grantStr, " TO ")
	if onIdx == -1 || toIdx == -1 || onIdx >= toIdx {
		return nil
	}

	privPart := grantStr[len("GRANT "):onIdx]
	targetPart := grantStr[onIdx+len(" ON ") : toIdx]

	dbName := targetPart
	dotIdx := strings.Index(targetPart, ".")
	if dotIdx != -1 {
		dbName = targetPart[:dotIdx]
	}
	dbName = strings.ReplaceAll(dbName, "`", "")

	var privileges []string
	if strings.Contains(strings.ToUpper(privPart), "ALL PRIVILEGES") {
		privileges = []string{"ALL"}
	} else {
		parts := strings.Split(privPart, ",")
		for _, p := range parts {
			priv := strings.TrimSpace(p)
			if priv != "" {
				privileges = append(privileges, priv)
			}
		}
	}

	return &GrantInfo{
		Database:   dbName,
		Privileges: privileges,
	}
}

func (m *MySQLProvider) ListUsers() (interface{}, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT User, Host FROM mysql.user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var username, host string
		if err := rows.Scan(&username, &host); err != nil {
			return nil, err
		}

		if username == "mysql.session" || username == "mysql.sys" || username == "mysql.infoschema" {
			continue
		}

		grantsRows, err := db.Query(fmt.Sprintf("SHOW GRANTS FOR `%s`@`%s`", username, host))
		var grants []GrantInfo
		if err == nil {
			for grantsRows.Next() {
				var grantStr string
				if err := grantsRows.Scan(&grantStr); err == nil {
					grant := parseGrantString(grantStr)
					if grant != nil {
						grants = append(grants, *grant)
					}
				}
			}
			grantsRows.Close()
		}

		users = append(users, UserInfo{
			Username: username,
			Host:     host,
			Grants:   grants,
		})
	}
	return users, nil
}

func (m *MySQLProvider) CreateUserInternal(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}
	if strings.ContainsAny(username, "`; ") {
		return fmt.Errorf("invalid characters in username")
	}

	db, err := GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check if user already exists
	var userExists int
	checkUserQuery := "SELECT COUNT(*) FROM mysql.user WHERE User = ?"
	if err := db.QueryRow(checkUserQuery, username).Scan(&userExists); err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if userExists > 0 {
		return fmt.Errorf("user already exists")
	}

	// Create user
	escapedPassword := strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf("CREATE USER `%s`@'%%' IDENTIFIED BY '%s'", username, escapedPassword)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (m *MySQLProvider) DeleteUserInternal(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if strings.ContainsAny(username, "`; ") {
		return fmt.Errorf("invalid characters in username")
	}

	db, err := GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	query := fmt.Sprintf("DROP USER `%s`@'%%'", username)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to drop user: %v", err)
	}

	return nil
}

func (m *MySQLProvider) GrantPrivilegesInternal(username, database string, privileges []string) error {
	if username == "" || database == "" || len(privileges) == 0 {
		return fmt.Errorf("username, database, and privileges are required")
	}
	if strings.ContainsAny(username, "`; ") || strings.ContainsAny(database, "`; ") {
		return fmt.Errorf("invalid characters in input")
	}

	db, err := GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	privsStr := strings.Join(privileges, ", ")
	query := fmt.Sprintf("GRANT %s ON `%s`.* TO `%s`@'%%'", privsStr, database, username)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant privileges: %v", err)
	}

	// Flush privileges
	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		return fmt.Errorf("failed to flush privileges: %v", err)
	}

	return nil
}

func (m *MySQLProvider) RevokePrivilegesInternal(username, database string, privileges []string) error {
	if username == "" || database == "" || len(privileges) == 0 {
		return fmt.Errorf("username, database, and privileges are required")
	}
	if strings.ContainsAny(username, "`; ") || strings.ContainsAny(database, "`; ") {
		return fmt.Errorf("invalid characters in input")
	}

	db, err := GetDBConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	privsStr := strings.Join(privileges, ", ")
	query := fmt.Sprintf("REVOKE %s ON `%s`.* FROM `%s`@'%%'", privsStr, database, username)
	if _, err := db.Exec(query); err != nil {
		return err
	}

	// Flush privileges
	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		return err
	}

	return nil
}

func (m *MySQLProvider) ExecuteControl(action, sqlQuery string) (interface{}, error) {
	db, err := GetDBConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	switch action {
	case "query":
		rows, err := db.Query(sqlQuery)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			results = append(results, row)
		}
		return gin.H{
			"columns": columns,
			"rows":    results,
		}, nil

	case "exec":
		result, err := db.Exec(sqlQuery)
		if err != nil {
			return nil, err
		}
		affected, _ := result.RowsAffected()
		return gin.H{
			"affected": affected,
			"message":  "Command executed successfully",
		}, nil

	default:
		return nil, fmt.Errorf("invalid action. Use 'query' or 'exec'")
	}
}
