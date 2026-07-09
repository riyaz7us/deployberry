package common

type DBVersion struct {
	Version      string `json:"version"`
	Installed    bool   `json:"installed"`
	Active       bool   `json:"active"`
	RootPassword string `json:"root_password,omitempty"`
	Port         int    `json:"port,omitempty"`
}


