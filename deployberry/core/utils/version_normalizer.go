package utils

import (
	"regexp"
	"strings"

	"github.com/blang/semver/v4"
)

// extractVersionNumber extracts version number from various format strings
// This handles the prefix stripping that semver library doesn't support
func extractVersionNumber(input string) string {
	if input == "" {
		return ""
	}

	input = strings.TrimSpace(input)

	// Language patterns
	// PHP: "PHP 8.2.29" or "PHP 8.2"
	if m := regexp.MustCompile(`PHP\s+([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// Node.js: "v18.19.0" or "18.19.0"
	if m := regexp.MustCompile(`^v?([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// Python: "Python 3.11.2" or "Python 3.11"
	if m := regexp.MustCompile(`Python\s+([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// Go: "go version go1.21.6" or "go1.21.6"
	if m := regexp.MustCompile(`go version go([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// Database patterns
	// MySQL: "mysql  Ver 8.0.36" or "mysqld  Ver 8.0.36"
	if m := regexp.MustCompile(`(?:mysql|mysqld)\s+Ver\s+([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// MariaDB: "mariadb  Ver 10.11.6"
	if m := regexp.MustCompile(`mariadb\s+Ver\s+([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// PostgreSQL: "PostgreSQL 16.2"
	if m := regexp.MustCompile(`PostgreSQL\s+([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// MongoDB: "db version v7.0.8" or "mongod --version" output
	if m := regexp.MustCompile(`db version v?([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// Redis: "Redis server v=7.2.4"
	if m := regexp.MustCompile(`Redis server v=([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	// SQLite: "3.44.2" (already clean)
	if m := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)*)`).FindStringSubmatch(input); len(m) > 1 {
		return m[1]
	}

	return ""
}

// normalizeVersion extracts clean semantic version (x.x.x) from various version string formats
// Uses semver library for validation and normalization
// IMPORTANT: Only use this for OS/system detection, not for our curated version lists
func normalizeVersion(input string) string {
	// Extract version number from language-specific format
	versionNum := extractVersionNumber(input)
	if versionNum == "" {
		return ""
	}

	// Parse with semver library for validation and normalization
	version, err := semver.Parse(versionNum)
	if err != nil {
		// If parsing fails, try to normalize partial versions (e.g., "8.2" -> "8.2.0")
		parts := strings.Split(versionNum, ".")
		for len(parts) < 3 {
			parts = append(parts, "0")
		}
		normalized := strings.Join(parts[:3], ".")
		
		version, err = semver.Parse(normalized)
		if err != nil {
			return "" // Invalid version format
		}
	}

	// Return canonical semantic version string
	return version.String()
}

// normalizeForOSSystemDetection is specifically for OS/system version detection
// This handles unpredictable OS version formats
func normalizeForOSSystemDetection(input string) string {
	return normalizeVersion(input)
}

