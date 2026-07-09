package appinstaller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"deployberry/core/applications/appinstaller/manifest"
	"shared/globals"
)

const (
	// Using raw.githubusercontent.com to get the actual file contents
	RegistryBaseURL = "https://raw.githubusercontent.com/riyaz7us/deployberry-manifests/refs/heads/main"
)

type AppIndexItem struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Tags        []string `json:"tags"`
}

func FetchLocalRegistryIndex() ([]AppIndexItem, error) {
	indexfile, err := os.ReadFile(globals.REGISTRY_PATH + "/index.json")
	if err != nil {
		return nil, fmt.Errorf("Cannot parse local manifest %w", err)
	}
	var index []AppIndexItem
	if err := json.NewDecoder(bytes.NewReader(indexfile)).Decode(&index); err != nil {
		return nil, err
	}
	return index, nil
}

func FetchLocalManifest(slug string) (*manifest.Manifest, error) {
	manifestFile, err := os.ReadFile(fmt.Sprintf("%s/%s.manifest.yaml", globals.REGISTRY_PATH, slug))
	if err != nil {
		return nil, fmt.Errorf("Cannot parse local manifest %w", err)
	}
	parsedData, err := manifest.Parse(manifestFile)
	return parsedData, err
}

// FetchRegistryIndex grabs the index.json from GitHub
func FetchRegistryIndex() ([]AppIndexItem, error) {
	if os.Getenv("ENV") == "dev" {
		return FetchLocalRegistryIndex()
	}
	resp, err := http.Get(RegistryBaseURL + "/index.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	var index []AppIndexItem
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode registry index: %w", err)
	}

	return index, nil
}

// FetchManifest downloads a specific manifest by slug
func FetchManifest(slug string) (*manifest.Manifest, error) {
	if os.Getenv("ENV") == "dev" {
		return FetchLocalManifest(slug)
	}
	url := fmt.Sprintf("%s/%s.manifest.yaml", RegistryBaseURL, slug)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest %s: %w", slug, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to find manifest for %s (status %d)", slug, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest body: %w", err)
	}

	// Uses the parser we built in Step 1
	parsedData, err := manifest.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return parsedData, nil
}
