package dbops

import (
	"fmt"
)

// DatabaseResult holds the result of database creation operations
type DatabaseResult struct {
	Database string
	Username string
	Password string
	Message  string
	Port     int
}

// DBProvider defines the contract for any database engine
type DBProvider interface {
	IsInstalled() (bool, string, error)
	GetCredentials() (string, string, error)
	UpdateCredentials(username, password string) error
	ListDatabases() ([]string, error)
	CreateDatabaseInternal(appName string) (*DatabaseResult, error)
	DeleteDatabaseInternal(databaseName string) error
	ListUsers() (interface{}, error)
	CreateUserInternal(username, password string) error
	DeleteUserInternal(username string) error
	GrantPrivilegesInternal(username, database string, privileges []string) error
	RevokePrivilegesInternal(username, database string, privileges []string) error
	ExecuteControl(action, sqlQuery string) (interface{}, error)
}

// GetProvider returns the provider for the requested engine using simple switch case
func GetProvider(engine string) (DBProvider, error) {
	// Map 'sql' to 'mysql' for frontend backward compatibility
	if engine == "sql" {
		engine = "mysql"
	}

	switch engine {
	case "mysql", "mariadb":
		return &MySQLProvider{}, nil
	case "postgres":
		return &PostgresProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported database engine: %s", engine)
	}
}

// MySQLProvider implements DBProvider interface
type MySQLProvider struct{}

// PostgresProvider implements DBProvider interface
type PostgresProvider struct{}
