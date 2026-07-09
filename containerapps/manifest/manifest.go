package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Manifest represents the structure of a containerapp manifest
type Manifest struct {
	Name            string            `yaml:"name"`
	DisplayName     string            `yaml:"display_name"`
	Description     string            `yaml:"description"`
	Version         string            `yaml:"version"`
	Icon            string            `yaml:"icon"`
	Tags            []string          `yaml:"tags,omitempty"`
	ComposeTemplate string            `yaml:"compose_template"` // Raw docker-compose template string
	Variables       []Variable        `yaml:"variables,omitempty"`
	Install         []Step            `yaml:"install,omitempty"`
	Commands        map[string]Cmd    `yaml:"commands,omitempty"`
	Update          []Step            `yaml:"update,omitempty"`
	Webserver       WebserverConfig   `yaml:"webserver,omitempty"`
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

type Step struct {
	Name      string `yaml:"name"`
	Run       string `yaml:"run"`
	Service   string `yaml:"service,omitempty"`     // Specific service to run in (defaults to primary)
	RunOnHost bool   `yaml:"run_on_host,omitempty"` // Whether step runs on host instead of container
}

type Cmd struct {
	Label   string   `yaml:"label"`
	Run     string   `yaml:"run"`
	Args    []string `yaml:"args,omitempty"`
	Service string   `yaml:"service,omitempty"` // Specific service to run in (defaults to primary)
}

type WebserverConfig struct {
	Mode string `yaml:"mode"` // proxy, static, etc.
	Port int    `yaml:"port"` // port inside container to proxy to
}

// Parse takes raw YAML bytes and returns a typed Manifest struct
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse container manifest YAML: %w", err)
	}
	return &m, nil
}
