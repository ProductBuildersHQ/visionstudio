package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Organization represents the workspace/organization configuration for VisionStudio.
// This is the top-level entity that contains organization-wide V2MOMs and settings.
type Organization struct {
	// ID is a unique identifier for the organization.
	ID string `json:"id"`

	// Name is the display name of the organization.
	Name string `json:"name"`

	// Description provides additional context about the organization.
	Description string `json:"description,omitempty"`

	// V2MOMPath is the path to the directory containing organization-level V2MOMs.
	// If empty, defaults to ~/.visionstudio/v2moms/
	V2MOMPath string `json:"v2momPath,omitempty"`

	// FiscalYearStart defines when the fiscal year begins (e.g., "01-01" for Jan 1).
	FiscalYearStart string `json:"fiscalYearStart,omitempty"`

	// CreatedAt is when the organization was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the organization was last modified.
	UpdatedAt time.Time `json:"updatedAt"`
}

// organizationConfigPath returns the path to organization.json
func organizationConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "organization.json"), nil
}

// defaultV2MOMPath returns the default path for organization V2MOMs
func defaultV2MOMPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "v2moms"), nil
}

// LoadOrganization loads the organization configuration from disk.
// Returns nil if no organization is configured yet.
func LoadOrganization() (*Organization, error) {
	path, err := organizationConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return nil if no organization configured
			return nil, nil
		}
		return nil, fmt.Errorf("read organization config: %w", err)
	}

	var org Organization
	if err := json.Unmarshal(data, &org); err != nil {
		return nil, fmt.Errorf("parse organization config: %w", err)
	}

	return &org, nil
}

// SaveOrganization saves the organization configuration to disk.
func SaveOrganization(org *Organization) error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	// Update timestamp
	org.UpdatedAt = time.Now()

	// Set default V2MOM path if not specified
	if org.V2MOMPath == "" {
		defaultPath, err := defaultV2MOMPath()
		if err != nil {
			return fmt.Errorf("get default v2mom path: %w", err)
		}
		org.V2MOMPath = defaultPath
	}

	// Ensure V2MOM directory exists
	if err := os.MkdirAll(org.V2MOMPath, 0755); err != nil {
		return fmt.Errorf("create v2mom directory: %w", err)
	}

	path, err := organizationConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(org, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal organization config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write organization config: %w", err)
	}

	return nil
}

// CreateOrganization creates a new organization with the given name.
func CreateOrganization(name string) (*Organization, error) {
	// Check if organization already exists
	existing, err := LoadOrganization()
	if err != nil {
		return nil, fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("organization already exists: %s", existing.Name)
	}

	org := &Organization{
		ID:        generateOrgID(name),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := SaveOrganization(org); err != nil {
		return nil, fmt.Errorf("save organization: %w", err)
	}

	return org, nil
}

// UpdateOrganization updates an existing organization.
func UpdateOrganization(org *Organization) error {
	existing, err := LoadOrganization()
	if err != nil {
		return fmt.Errorf("load existing: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("no organization exists")
	}

	// Preserve creation time and ID
	org.ID = existing.ID
	org.CreatedAt = existing.CreatedAt

	return SaveOrganization(org)
}

// DeleteOrganization removes the organization configuration.
// This does not delete V2MOM files or project associations.
func DeleteOrganization() error {
	path, err := organizationConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove organization config: %w", err)
	}

	return nil
}

// GetOrganizationV2MOMPath returns the path to organization V2MOMs directory.
func GetOrganizationV2MOMPath() (string, error) {
	org, err := LoadOrganization()
	if err != nil {
		return "", err
	}

	if org == nil {
		// Return default path even if org not configured
		return defaultV2MOMPath()
	}

	if org.V2MOMPath != "" {
		return org.V2MOMPath, nil
	}

	return defaultV2MOMPath()
}

// generateOrgID creates a URL-safe ID from the organization name.
func generateOrgID(name string) string {
	// Simple kebab-case conversion
	id := ""
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			id += string(r)
		} else if r >= 'A' && r <= 'Z' {
			id += string(r + 32) // lowercase
		} else if r >= '0' && r <= '9' {
			id += string(r)
		} else if r == ' ' || r == '-' || r == '_' {
			if len(id) > 0 && id[len(id)-1] != '-' {
				id += "-"
			}
		}
	}
	// Trim trailing dash
	if len(id) > 0 && id[len(id)-1] == '-' {
		id = id[:len(id)-1]
	}
	return id
}
