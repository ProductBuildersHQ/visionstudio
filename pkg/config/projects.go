package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TrackedProject represents a project tracked by VisionStudio
type TrackedProject struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Profile string `json:"profile"`
}

// ProjectsConfig holds the list of tracked projects
type ProjectsConfig struct {
	Projects []TrackedProject `json:"projects"`
}

// configDir returns the VisionStudio config directory
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".visionstudio"), nil
}

// configPath returns the path to projects.json
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "projects.json"), nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadProjects loads the projects configuration from disk
func LoadProjects() (*ProjectsConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return empty config if file doesn't exist
			return &ProjectsConfig{Projects: []TrackedProject{}}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg ProjectsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// SaveProjects saves the projects configuration to disk
func SaveProjects(cfg *ProjectsConfig) error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// AddProject adds a new project to the configuration
func AddProject(name, path, profile string) error {
	cfg, err := LoadProjects()
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	// Check for duplicate name
	for _, p := range cfg.Projects {
		if p.Name == name {
			return fmt.Errorf("project %q already exists", name)
		}
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Verify path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("verify path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	cfg.Projects = append(cfg.Projects, TrackedProject{
		Name:    name,
		Path:    absPath,
		Profile: profile,
	})

	if err := SaveProjects(cfg); err != nil {
		return fmt.Errorf("save projects: %w", err)
	}

	return nil
}

// RemoveProject removes a project from the configuration (does not delete files)
func RemoveProject(name string) error {
	cfg, err := LoadProjects()
	if err != nil {
		return fmt.Errorf("load projects: %w", err)
	}

	found := false
	filtered := make([]TrackedProject, 0, len(cfg.Projects))
	for _, p := range cfg.Projects {
		if p.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, p)
	}

	if !found {
		return fmt.Errorf("project %q not found", name)
	}

	cfg.Projects = filtered

	if err := SaveProjects(cfg); err != nil {
		return fmt.Errorf("save projects: %w", err)
	}

	return nil
}

// GetProject returns a tracked project by name
func GetProject(name string) (*TrackedProject, error) {
	cfg, err := LoadProjects()
	if err != nil {
		return nil, fmt.Errorf("load projects: %w", err)
	}

	for _, p := range cfg.Projects {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("project %q not found", name)
}
