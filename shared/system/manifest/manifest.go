package manifest

import (
	"bytes"
	"fmt"
	"shared/system/manifests"
	"text/template"

	"gopkg.in/yaml.v3"
)

// SystemManifest represents the declarative structure for system runtimes and servers.
type SystemManifest struct {
	Name        string `yaml:"name"`
	Type        string   `yaml:"type"` // e.g., "runtime", "database", "server"
	Description string   `yaml:"description,omitempty"`
	Versions    []string `yaml:"versions,omitempty"` // Static list of supported versions

	// Global dependencies required regardless of the action
	Repositories []Repository    `yaml:"repositories,omitempty"`
	Packages     []SystemPackage `yaml:"packages,omitempty"`

	// Standard Lifecycle Actions
	Install    ActionBlock `yaml:"install,omitempty"`
	Uninstall  ActionBlock `yaml:"uninstall,omitempty"`
	Activate   ActionBlock `yaml:"activate,omitempty"`
	Deactivate ActionBlock `yaml:"deactivate,omitempty"`

	// Custom User-Defined Actions
	Commands map[string]ActionBlock `yaml:"commands,omitempty"`
}

type ActionBlock struct {
	Packages []SystemPackage `yaml:"packages,omitempty"`
	Services []Service       `yaml:"services,omitempty"`
	Steps    []ConfigureStep `yaml:"steps,omitempty"`
}

type ConfigureStep struct {
	Label        string   `yaml:"label,omitempty"`
	Type         string   `yaml:"type"` // "cmd", "sql_exec", "file_replace", "file_append", "kill_port", "run_action"
	Command      string   `yaml:"command,omitempty"`
	Args         []string `yaml:"args,omitempty"`
	Env          []string `yaml:"env,omitempty"`
	Executable   string   `yaml:"executable,omitempty"`
	Query        string   `yaml:"query,omitempty"`
	File         string   `yaml:"file,omitempty"`
	Search       string   `yaml:"search,omitempty"`
	Replace      string   `yaml:"replace,omitempty"`
	Block        string   `yaml:"block,omitempty"`
	Port         int      `yaml:"port,omitempty"`
	Action       string   `yaml:"action,omitempty"`
	IgnoreErrors bool     `yaml:"ignore_errors,omitempty"`
}

// Repository describes how to add a software repository for a given OS package manager.
type Repository struct {
	OS  string `yaml:"os"`  // "apt" or "dnf"
	URL string `yaml:"url"` // PPA name or repository URL
}

// SystemPackage describes the packages to install for different package managers.
type SystemPackage struct {
	Name string `yaml:"name"` // Generic internal name
	Apt  string `yaml:"apt,omitempty"`  // Debian/Ubuntu package name
	Dnf  string `yaml:"dnf,omitempty"`  // RHEL/CentOS package name
}

// Service represents a system service to be managed (e.g., via systemctl).
type Service struct {
	// Maps the service name to its desired action (e.g., "php8.3-fpm": "enable_start")
	Name         string `yaml:"name"`
	Apt          string `yaml:"apt,omitempty"`
	Dnf          string `yaml:"dnf,omitempty"`
	Action       string `yaml:"action"`
	IgnoreErrors bool   `yaml:"ignore_errors,omitempty"`
}

// Parse takes raw YAML bytes and returns a typed SystemManifest struct
func Parse(data []byte) (*SystemManifest, error) {
	var m SystemManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse system manifest YAML: %w", err)
	}
	return &m, nil
}

// LoadAndRender reads a manifest file from the embedded filesystem, renders it as a text template using the provided data, and parses it.
func LoadAndRender(name string, data interface{}) (*SystemManifest, error) {
	content, err := manifests.Files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded manifest %s: %w", name, err)
	}

	tmpl, err := template.New("manifest").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render manifest template: %w", err)
	}

	return Parse(buf.Bytes())
}
