package repository

import (
	"fmt"
	"log"
	"os"
	"shared/globals"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

var db *gorm.DB

// GORM Models

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"not null" json:"username"`
	Password  string    `gorm:"not null" json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Application struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Path         string    `gorm:"uniqueIndex;not null" json:"path"`
	Domain       string    `gorm:"not null" json:"domain"`
	Provider     string    `gorm:"not null" json:"provider"`
	Title        string    `json:"title"`
	DisplayName  string    `json:"display_name"`
	Version      string    `json:"version"`
	Database     string    `json:"database"`
	DeployMethod string    `json:"deploy_method"`
	Status       string    `gorm:"default:installed" json:"status"`
	Variables    string    `json:"variables"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Backup struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Path      string    `gorm:"not null" json:"path"`
	Type      string    `gorm:"not null" json:"type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatabaseCredential struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Host      string    `gorm:"not null" json:"host"`
	Port      int       `gorm:"not null" json:"port"`
	Database  string    `gorm:"not null" json:"database"`
	Username  string    `gorm:"not null" json:"username"`
	Password  string    `gorm:"not null" json:"password"`
	Type      string    `gorm:"not null" json:"type"`
	AppID     uint      `json:"app_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContainerApp struct {
	ID            uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	ApplicationID uint        `gorm:"uniqueIndex;not null" json:"application_id"`
	Application   Application `gorm:"foreignKey:ApplicationID;constraint:OnDelete:CASCADE" json:"application"`
	ComposeFile   string      `json:"compose_file"`   // Substituted docker-compose.yml written to disk
	HostPort      int         `json:"host_port"`      // Allocated port on host
	ContainerName string      `json:"container_name"` // Main service name/container slug
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type Cron struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Command   string    `gorm:"not null" json:"command"`
	Schedule  string    `gorm:"not null" json:"schedule"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Config struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebServer struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string    `gorm:"uniqueIndex;not null" json:"type"`
	Name      string    `gorm:"not null" json:"name"`
	Version   string    `gorm:"not null" json:"version"`
	Active    bool      `gorm:"default:true" json:"active"`
	Status    string    `gorm:"not null" json:"status"`
	Port      int       `json:"port"`
	SSLPort   int       `json:"ssl_port"`
	ConfigDir string    `json:"config_dir"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebConfig struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Domain          string    `gorm:"uniqueIndex;not null" json:"domain"`
	RootPath        string    `gorm:"not null" json:"root_path"`
	PHPVersion      string    `json:"php_version"`
	ReverseProxyURL string    `json:"reverse_proxy_url"`
	EnableGzip      bool      `gorm:"default:true" json:"enable_gzip"`
	EnableCache     bool      `gorm:"default:true" json:"enable_cache"`
	WebServer       string    `json:"webserver"`
	SSL             bool      `gorm:"default:false" json:"ssl"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Language struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Version   string    `gorm:"not null" json:"version"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatabaseServer struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Type         string    `gorm:"uniqueIndex;not null" json:"type"`
	Version      string    `gorm:"not null" json:"version"`
	Active       bool      `gorm:"default:true" json:"active"`
	Port         int       `json:"port"`
	RootPassword string    `json:"root_password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ensureDBDirectory ensures the database directory exists
func ensureDBDirectory() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(globals.DATA_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}

// initDB initializes the GORM database connection and auto-migrates tables
func initDB() error {
	var err error

	// Ensure database directory exists
	if err := ensureDBDirectory(); err != nil {
		return fmt.Errorf("failed to ensure database directory: %w", err)
	}

	logLevel := logger.Warn
	if globals.IsDevelopment() {
		logLevel = logger.Info
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to SQLite database with GORM using pure Go driver (modernc.org/sqlite)
	db, err = gorm.Open(sqlite.Dialector{
		DSN:        globals.DB_PATH + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)&_pragma=temp_store(memory)&_pragma=busy_timeout(5000)",
		DriverName: "sqlite",
	}, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&User{},
		&Application{},
		&Backup{},
		&DatabaseCredential{},
		&Cron{},
		&WebServer{},
		&WebConfig{},
		&Language{},
		&DatabaseServer{},
		&Config{},
		&ContainerApp{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

var dbMutex sync.Mutex

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if db == nil {
		if err := initDB(); err != nil {
			panic("Failed to initialize database: " + err.Error())
		}
	}
	return db
}

// Close closes the database connection
func Close() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		db = nil
		return sqlDB.Close()
	}
	return nil
}

// InitDatabase initializes the database
func InitDatabase() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	if db == nil {
		if err := initDB(); err != nil {
			return err
		}
	}
	return nil
}
