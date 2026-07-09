package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest represents the root of a .manifest.yaml file
type Manifest struct {
	Name           string          `yaml:"name"`
	DisplayName    string          `yaml:"display_name"`
	Description    string          `yaml:"description"`
	Version        string          `yaml:"version"`
	Tags           []string        `yaml:"tags"`
	Runtime        []Runtime       `yaml:"runtime"`
	SystemPackages []SystemPackage `yaml:"system_packages"`
	Tools          []Tool          `yaml:"tools"`
	Deploy         Deploy          `yaml:"deploy"`
	Database       Database        `yaml:"database"`
	InMemory       []InMemoryDB    `yaml:"in_memory,omitempty"`
	Webserver      Webserver       `yaml:"webserver"`
	Process        Process         `yaml:"process"`
	Variables      []Variable      `yaml:"variables"`
	Hooks          Hooks           `yaml:"hooks"`
	Install        []Step          `yaml:"install"`
	Commands       map[string]Cmd  `yaml:"commands"`
	Update         []Step          `yaml:"update"`
	Delete         DeleteConfig    `yaml:"delete"`
	EditableFiles  []string        `yaml:"editable_files"`
}

type Runtime struct {
	Language string `yaml:"language"`
	Version  string `yaml:"version"`
}

type SystemPackage struct {
	Name string `yaml:"name"`
	Apt  string `yaml:"apt"`
	Dnf  string `yaml:"dnf"`
}

type Tool struct {
	Name     string `yaml:"name"`
	Url      string `yaml:"url"`
	Run      string `yaml:"run"`
	Dest     string `yaml:"dest"`
	Verify   string `yaml:"verify"`
	Npm      string `yaml:"npm"`
	Pip      string `yaml:"pip"`
	Composer string `yaml:"composer"`
}

type Deploy struct {
	Required    *bool      `yaml:"required,omitempty"`    // whether deployment is mandatory
	Default     string     `yaml:"default"`               // git|upload|manual - default method
	Git         *GitDeploy `yaml:"git,omitempty"`         // git deployment config
	Upload      bool       `yaml:"upload,omitempty"`      // enable upload method
	Manual      bool       `yaml:"manual,omitempty"`      // enable manual path method
	Destination string     `yaml:"destination,omitempty"` // optional subdirectory
}

type GitDeploy struct {
	Source   string `yaml:"source"`             // git URL
	Branch   string `yaml:"branch"`             // optional branch
	Blobless bool   `yaml:"blobless,omitempty"` // optional — whether to use a blobless clone
}

type Database struct {
	Required bool             `yaml:"required"`
	Engines  []DatabaseEngine `yaml:"engines"`
}

type DatabaseEngine struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type InMemoryDB struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Webserver struct {
	Mode string `yaml:"mode"` // php-fpm, proxy, static
	Root string `yaml:"root"`
	Port int    `yaml:"port"`
}

type Process struct {
	Manager string `yaml:"manager"` // pm2, systemd, none
	Start   string `yaml:"start"`
	Name    string `yaml:"name"`
}

type Variable struct {
	Key      string   `yaml:"key"`
	Prompt   string   `yaml:"prompt"`
	Default  string   `yaml:"default"`
	Required bool     `yaml:"required"`
	Secret   bool     `yaml:"secret"`
	Options  []string `yaml:"options,omitempty"`
	Type     string   `yaml:"type,omitempty"`   // text, number, select
	Helper   string   `yaml:"helper,omitempty"` // Help text for UI
}

type Hooks struct {
	PreInstall  string `yaml:"pre_install"`
	PostInstall string `yaml:"post_install"`
	PreUpdate   string `yaml:"pre_update"`
	PostUpdate  string `yaml:"post_update"`
	PreDelete   string `yaml:"pre_delete"`
	PostDelete  string `yaml:"post_delete"`
}

type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

type Cmd struct {
	Label string   `yaml:"label"`
	Run   string   `yaml:"run"`
	Args  []string `yaml:"args,omitempty"`
}

type DeleteConfig struct {
	DropDatabase          bool   `yaml:"drop_database"`
	RemoveWebserverConfig bool   `yaml:"remove_webserver_config"`
	RemoveFiles           bool   `yaml:"remove_files"`
	PreDelete             string `yaml:"pre_delete"`
}

func (s *SystemPackage) ResolveForOS(os string) string {
	switch os {
	case "apt":
		if s.Apt != "" {
			return s.Apt
		}
	case "dnf":
		if s.Dnf != "" {
			return s.Dnf
		}
	}
	return s.Name // Fallback to the simple name if OS-specific isn't defined
}

// Parse takes raw YAML bytes and returns a typed Manifest struct
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}
	return &m, nil
}
