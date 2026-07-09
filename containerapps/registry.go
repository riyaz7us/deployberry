package containerapps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"containerapps/manifest"
	"shared/globals"
)

const (
	RegistryBaseURL = "https://raw.githubusercontent.com/riyaz7us/deployberry-manifests/refs/heads/master/containerapps"
)

type AppIndexItem struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon"`
	Tags        []string `json:"tags"`
}

func FetchLocalRegistryIndex() ([]AppIndexItem, error) {
	indexPath := filepath.Join(globals.REGISTRY_PATH, "containerapps", "index.json")
	indexfile, err := os.ReadFile(indexPath)
	if err != nil {
		// Fallback: check if it's placed directly in REGISTRY_PATH with a prefix or name
		indexPath = filepath.Join(globals.REGISTRY_PATH, "container_index.json")
		indexfile, err = os.ReadFile(indexPath)
		if err != nil {
			return nil, fmt.Errorf("cannot read local registry index: %w", err)
		}
	}
	var index []AppIndexItem
	if err := json.NewDecoder(bytes.NewReader(indexfile)).Decode(&index); err != nil {
		return nil, err
	}
	return index, nil
}

func FetchLocalManifest(slug string) (*manifest.Manifest, error) {
	manifestPath := filepath.Join(globals.REGISTRY_PATH, "containerapps", fmt.Sprintf("%s.manifest.yaml", slug))
	manifestFile, err := os.ReadFile(manifestPath)
	if err != nil {
		// Fallback to check directly in registry path (e.g. for wordpress-container.manifest.yaml)
		manifestPath = filepath.Join(globals.REGISTRY_PATH, fmt.Sprintf("%s.manifest.yaml", slug))
		manifestFile, err = os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("cannot parse local container manifest: %w", err)
		}
	}
	parsedData, err := manifest.Parse(manifestFile)
	return parsedData, err
}

// FetchRegistryIndex grabs the index.json from GitHub or dev path
func FetchRegistryIndex() ([]AppIndexItem, error) {
	if globals.IsDevelopment() {
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
	if globals.IsDevelopment() {
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

	parsedData, err := manifest.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return parsedData, nil
}
