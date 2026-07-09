package common

type RequirementOption struct {
	Type    string `json:"type"`    // "language" or "database"
	Key     string `json:"key"`     // "php", "mysql", "mariadb", etc.
	Version string `json:"version"` // "8.2", "8.0", etc.
}

type Requirement struct {
	Name           string              `json:"name"`
	Available      bool                `json:"available"`
	CurrentVersion string              `json:"current_version,omitempty"`
	Options        []RequirementOption `json:"options"`
	Type           string              `json:"type"` // "language" or "database" or "tool"
}
