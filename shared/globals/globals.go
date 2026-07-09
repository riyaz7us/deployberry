package globals

import (
	"os"
	"path/filepath"
)

var (
	BASE_DIR       string
	DATA_DIR       string
	WEBCONFIGS_DIR string
	TOOLS_DIR      string
	BACKUPS_DIR    string
	LOG_DIR        string
	DB_PATH        string
	REGISTRY_PATH  string
	ENV_PATH       string
)

func init() {
	// Check if running in development mode
	if os.Getenv("ENV") == "dev" {
		// Development: Everything in .dev directory
		baseDir := "../.dev"

		BASE_DIR = baseDir
		DATA_DIR = filepath.Join(baseDir, "data")
		WEBCONFIGS_DIR = filepath.Join(baseDir, "webconfigs")
		TOOLS_DIR = filepath.Join(baseDir, "tools")
		BACKUPS_DIR = filepath.Join(baseDir, "backups")
		LOG_DIR = filepath.Join(baseDir, "logs")
		DB_PATH = filepath.Join(baseDir, "data/deployberry.db")
		REGISTRY_PATH = filepath.Join(baseDir, "../deployberry-manifests")
		ENV_PATH = filepath.Join(baseDir, "env.sh")

		// Create directories if they don't exist
		os.MkdirAll(DATA_DIR, 0755)
		os.MkdirAll(WEBCONFIGS_DIR, 0755)
		os.MkdirAll(TOOLS_DIR, 0755)
		os.MkdirAll(BACKUPS_DIR, 0755)
		os.MkdirAll(LOG_DIR, 0755)

	} else {
		// Production: Standard Linux paths
		BASE_DIR = "/opt/deployberry"
		DATA_DIR = "/var/lib/deployberry/data"
		WEBCONFIGS_DIR = "/var/lib/deployberry/webconfigs"
		TOOLS_DIR = "/var/lib/deployberry/tools"
		BACKUPS_DIR = "/var/lib/deployberry/backups"
		LOG_DIR = "/var/log/deployberry"
		DB_PATH = "/var/lib/deployberry/data/deployberry.db"
		REGISTRY_PATH = ""
		ENV_PATH = "/etc/deployberry/env.sh"
	}
	os.Setenv("DEPLOYBERRY_ENV_PATH", ENV_PATH)

	// Ensure the environment file and its directory exist
	dir := filepath.Dir(ENV_PATH)
	if _, err := os.Stat(ENV_PATH); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
		f, err := os.OpenFile(ENV_PATH, os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString("# DeployBerry Environment File\n")
			f.Close()
		}
	}
}

// Helper to check if in development mode
func IsDevelopment() bool {
	return os.Getenv("ENV") == "dev"
}
