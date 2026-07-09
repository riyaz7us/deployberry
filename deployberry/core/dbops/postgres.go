package dbops

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"shared/repository"
	"deployberry/utils"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func buildPGDSN(dbName, rootPassword string, port int) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("postgres", rootPassword),
		Host:   fmt.Sprintf("localhost:%d", port),
		Path:   "/" + dbName,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

// Postgres implementation functions

func GetPGConnection() (*sql.DB, error) {
	return GetPGConnectionDB("postgres")
}

func GetPGConnectionDB(dbName string) (*sql.DB, error) {
	rootPassword, port, err := GetPostgresRootCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get Postgres credentials: %v", err)
	}

	dsn := buildPGDSN(dbName, rootPassword, port)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if pingErr := db.Ping(); pingErr != nil {
		db.Close()
		if isAuthError(pingErr) {
			if healedDsn, healErr := healPostgresPassword(dbName, rootPassword, port); healErr == nil {
				return sql.Open("postgres", healedDsn)
			}
		}
		return nil, pingErr
	}

	return db, nil
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "password authentication failed") || strings.Contains(errStr, "28P01")
}

func healPostgresPassword(dbName string, rootPassword string, port int) (string, error) {
	if rootPassword == "" {
		newPass, err := utils.GenerateSecurePassword(16)
		if err != nil {
			return "", err
		}
		rootPassword = newPass

		// Save the new password to DatabaseServer record in DB
		dbg := repository.GetDB()
		var dbServer repository.DatabaseServer
		if err := dbg.Where("type = ?", "postgres").First(&dbServer).Error; err == nil {
			dbServer.RootPassword = rootPassword
			dbg.Save(&dbServer)
		}
	}

	// ALTER USER query using psql under postgres user via sudo (uses UNIX socket peer authentication)
	alterQuery := fmt.Sprintf("ALTER USER postgres PASSWORD '%s';", strings.ReplaceAll(rootPassword, "'", "''"))
	cmd := exec.Command("sudo", "-u", "postgres", "psql", "-c", alterQuery)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	healedDsn := buildPGDSN(dbName, rootPassword, port)

	// Test the new DSN
	testDb, err := sql.Open("postgres", healedDsn)
	if err != nil {
		return "", err
	}
	defer testDb.Close()

	if err := testDb.Ping(); err != nil {
		return "", err
	}

	return healedDsn, nil
}

func GetPostgresRootCredentials() (string, int, error) {
	db := repository.GetDB()
	var dbServer repository.DatabaseServer
	err := db.Where("type = ?", "postgres").First(&dbServer).Error
	if err == nil {
		return dbServer.RootPassword, dbServer.Port, nil
	}

	// Fallback to environment variable or config table
	var passConfig repository.Config
	if db.Where("key = ?", "pgpass").First(&passConfig).Error == nil {
		return passConfig.Value, 5432, nil
	}

	rootPassword := os.Getenv("POSTGRES_PASSWORD")
	if rootPassword == "" {
		rootPassword = os.Getenv("PGPASSWORD")
	}

	return rootPassword, 5432, nil
}

func TestPostgresConnection() error {
	db, err := GetPGConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Ping()
}

// DBProvider Interface Methods for PostgresProvider

func (p *PostgresProvider) IsInstalled() (bool, string, error) {
	pgCmd := exec.Command("psql", "--version")
	pgOutput, pgErr := pgCmd.Output()

	if pgErr == nil {
		if err := TestPostgresConnection(); err != nil {
			return true, string(pgOutput), err
		}
		return true, string(pgOutput), nil
	}

	return false, "", nil
}

func (p *PostgresProvider) GetCredentials() (string, string, error) {
	db := repository.GetDB()
	var userConfig, passConfig repository.Config
	err := db.Where("key = ?", "pguser").First(&userConfig).Error
	if err != nil {
		return "postgres", "", nil
	}
	err = db.Where("key = ?", "pgpass").First(&passConfig).Error
	if err != nil {
		return userConfig.Value, "", nil
	}

	return userConfig.Value, passConfig.Value, nil
}

func (p *PostgresProvider) UpdateCredentials(username, password string) error {
	db := repository.GetDB()

	userConfig := repository.Config{Key: "pguser", Value: username}
	if err := db.Where(repository.Config{Key: "pguser"}).Assign(repository.Config{Value: username}).FirstOrCreate(&userConfig).Error; err != nil {
		return err
	}

	passConfig := repository.Config{Key: "pgpass", Value: password}
	return db.Where(repository.Config{Key: "pgpass"}).Assign(repository.Config{Value: password}).FirstOrCreate(&passConfig).Error
}

func (p *PostgresProvider) ListDatabases() ([]string, error) {
	db, err := GetPGConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false")
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

func (p *PostgresProvider) CreateDatabaseInternal(appName string) (*DatabaseResult, error) {
	if strings.TrimSpace(appName) == "" {
		return nil, fmt.Errorf("app_name is required")
	}

	dbName := utils.UtilSanitizeName(appName, "db_")
	userName := dbName
	password, err := utils.GenerateSecurePassword(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database password: %w", err)
	}

	db, err := GetPGConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Check if database already exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	if err := db.QueryRow(checkQuery, dbName).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to check database existence: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("database already exists")
	}

	// Check if user already exists
	var userExists bool
	checkUserQuery := "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)"
	if err := db.QueryRow(checkUserQuery, userName).Scan(&userExists); err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}
	if userExists {
		return nil, fmt.Errorf("database user already exists")
	}

	// Create user (role with login)
	if strings.ContainsAny(userName, "\";' ") {
		return nil, fmt.Errorf("invalid username format")
	}
	escapedPassword := strings.ReplaceAll(password, "'", "''")
	createUserQuery := fmt.Sprintf("CREATE ROLE \"%s\" WITH LOGIN PASSWORD '%s'", userName, escapedPassword)
	if _, err := db.Exec(createUserQuery); err != nil {
		return nil, fmt.Errorf("failed to create database user: %v", err)
	}

	// Allow the user to create their own schemas/databases (helpful for Odoo)
	alterUserQuery := fmt.Sprintf("ALTER ROLE \"%s\" CREATEDB", userName)
	if _, err := db.Exec(alterUserQuery); err != nil {
		db.Exec(fmt.Sprintf("DROP ROLE IF EXISTS \"%s\"", userName))
		return nil, fmt.Errorf("failed to alter user roles: %v", err)
	}

	// Create database owned by user
	if strings.ContainsAny(dbName, "\";' ") {
		db.Exec(fmt.Sprintf("DROP ROLE IF EXISTS \"%s\"", userName))
		return nil, fmt.Errorf("invalid database name format")
	}
	createDBQuery := fmt.Sprintf("CREATE DATABASE \"%s\" OWNER \"%s\"", dbName, userName)
	if _, err := db.Exec(createDBQuery); err != nil {
		db.Exec(fmt.Sprintf("DROP ROLE IF EXISTS \"%s\"", userName))
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	// Save credentials to control panel database
	dbg := repository.GetDB()
	dbCred := repository.DatabaseCredential{
		Host:     "localhost",
		Port:     5432,
		Database: dbName,
		Username: userName,
		Password: password,
		Type:     "postgres",
	}
	dbg.Create(&dbCred)

	_, port, _ := GetPostgresRootCredentials()
	if port == 0 {
		port = 5432
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

func (p *PostgresProvider) DeleteDatabaseInternal(databaseName string) error {
	if databaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if strings.ContainsAny(databaseName, "\";' ") {
		return fmt.Errorf("invalid characters in database name")
	}

	db, err := GetPGConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Terminate other active sessions to avoid locking errors during DROP
	terminateQuery := fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		  AND pid <> pg_backend_pid()`, databaseName)
	db.Exec(terminateQuery)

	query := fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", databaseName)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to drop database: %v", err)
	}

	// Remove credentials from database
	repository.GetDB().Where("database = ? AND type = ?", databaseName, "postgres").Delete(&repository.DatabaseCredential{})

	return nil
}

func (p *PostgresProvider) ListUsers() (interface{}, error) {
	db, err := GetPGConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT rolname FROM pg_roles WHERE rolcanlogin = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		users = append(users, map[string]interface{}{
			"username": username,
		})
	}

	return users, nil
}

func (p *PostgresProvider) CreateUserInternal(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}
	if strings.ContainsAny(username, "\";' ") {
		return fmt.Errorf("invalid characters in username")
	}

	db, err := GetPGConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	var userExists bool
	checkUserQuery := "SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname = $1)"
	if err := db.QueryRow(checkUserQuery, username).Scan(&userExists); err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if userExists {
		return fmt.Errorf("user already exists")
	}

	escapedPassword := strings.ReplaceAll(password, "'", "''")
	query := fmt.Sprintf("CREATE ROLE \"%s\" WITH LOGIN PASSWORD '%s'", username, escapedPassword)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (p *PostgresProvider) DeleteUserInternal(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if strings.ContainsAny(username, "\";' ") {
		return fmt.Errorf("invalid characters in username")
	}

	db, err := GetPGConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Reassign/drop owned schemas or tables to prevent dependent object errors
	db.Exec(fmt.Sprintf("REASSIGN OWNED BY \"%s\" TO postgres", username))
	db.Exec(fmt.Sprintf("DROP OWNED BY \"%s\"", username))

	query := fmt.Sprintf("DROP ROLE IF EXISTS \"%s\"", username)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to drop user: %v", err)
	}

	return nil
}

func (p *PostgresProvider) GrantPrivilegesInternal(username, database string, privileges []string) error {
	if username == "" || database == "" || len(privileges) == 0 {
		return fmt.Errorf("username, database, and privileges are required")
	}
	if strings.ContainsAny(username, "\";' ") || strings.ContainsAny(database, "\";' ") {
		return fmt.Errorf("invalid characters in input")
	}

	db, err := GetPGConnectionDB(database)
	if err != nil {
		return fmt.Errorf("failed to connect to database %s: %v", database, err)
	}
	defer db.Close()

	// Grant connect privileges
	if _, err := db.Exec(fmt.Sprintf("GRANT CONNECT ON DATABASE \"%s\" TO \"%s\"", database, username)); err != nil {
		return fmt.Errorf("failed to grant CONNECT on database: %v", err)
	}

	var privs []string
	for _, pr := range privileges {
		up := strings.ToUpper(strings.TrimSpace(pr))
		if up == "ALL PRIVILEGES" {
			privs = append(privs, "ALL")
		} else {
			privs = append(privs, up)
		}
	}
	privsStr := strings.Join(privs, ", ")

	// Grant schema privileges
	query := fmt.Sprintf("GRANT %s ON SCHEMA public TO \"%s\"", privsStr, username)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to grant schema privileges: %v", err)
	}

	// Grant default privileges for future objects
	defaultPrivQuery := fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT %s ON TABLES TO \"%s\"", privsStr, username)
	db.Exec(defaultPrivQuery)

	return nil
}

func (p *PostgresProvider) RevokePrivilegesInternal(username, database string, privileges []string) error {
	if username == "" || database == "" || len(privileges) == 0 {
		return fmt.Errorf("username, database, and privileges are required")
	}
	if strings.ContainsAny(username, "\";' ") || strings.ContainsAny(database, "\";' ") {
		return fmt.Errorf("invalid characters in input")
	}

	db, err := GetPGConnectionDB(database)
	if err != nil {
		return err
	}
	defer db.Close()

	var privs []string
	for _, pr := range privileges {
		up := strings.ToUpper(strings.TrimSpace(pr))
		if up == "ALL PRIVILEGES" {
			privs = append(privs, "ALL")
		} else {
			privs = append(privs, up)
		}
	}
	privsStr := strings.Join(privs, ", ")

	query := fmt.Sprintf("REVOKE %s ON SCHEMA public FROM \"%s\"", privsStr, username)
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}

func (p *PostgresProvider) ExecuteControl(action, sqlQuery string) (interface{}, error) {
	db, err := GetPGConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if strings.ToLower(action) == "query" {
		rows, err := db.Query(sqlQuery)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		var results []map[string]interface{}
		for rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				return nil, err
			}

			rowMap := make(map[string]interface{})
			for i, colName := range cols {
				val := columns[i]
				b, ok := val.([]byte)
				if ok {
					rowMap[colName] = string(b)
				} else {
					rowMap[colName] = val
				}
			}
			results = append(results, rowMap)
		}

		return gin.H{"results": results}, nil
	} else {
		res, err := db.Exec(sqlQuery)
		if err != nil {
			return nil, err
		}
		rowsAffected, _ := res.RowsAffected()
		return gin.H{"message": "Command executed successfully", "rows_affected": rowsAffected}, nil
	}
}
