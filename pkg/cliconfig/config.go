package cliconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	Dir      = ".productbuildershq/visionstudio"
	FileName = "config.json"
)

type Defaults struct {
	Workflow string `json:"workflow,omitempty"`
}

// Cloud holds the local client's connection to VisionStudio Cloud
// (INIT-VISIONSTUDIO-002). Per the open-core split (TRD T5a), the cloud
// product itself — tenant entities, auth, serving — lives in the private
// visionstudio-cloud repo; this config is only the local client's record
// of which registered repo syncs to which tenant.
type Cloud struct {
	BaseURL string `json:"base_url,omitempty"`
	// TenantAssignments maps a registered repository ID to the cloud
	// tenant slug its data syncs to. A repo not present here is not
	// synced (opt-in, per-project, one machine can serve many tenants).
	TenantAssignments map[string]string `json:"tenant_assignments,omitempty"`
	// TenantRemotes maps a tenant slug to its Dolt remote URL. Sync today
	// is whole-database push (dogfood-only per the RMI-205 spike — a
	// synced-entity projection is required before any external tenant);
	// TenantAssignments records per-project intent ahead of that.
	TenantRemotes map[string]string `json:"tenant_remotes,omitempty"`
}

type Config struct {
	DSN      string   `json:"dsn,omitempty"`
	Defaults Defaults `json:"defaults,omitempty"`
	Cloud    Cloud    `json:"cloud,omitempty"`
}

// AssignTenant records that repoID's synced data belongs to tenant slug.
func (c *Config) AssignTenant(repoID, slug string) {
	if c.Cloud.TenantAssignments == nil {
		c.Cloud.TenantAssignments = map[string]string{}
	}
	c.Cloud.TenantAssignments[repoID] = slug
}

// UnassignTenant removes repoID's tenant assignment. Reports whether one
// existed.
func (c *Config) UnassignTenant(repoID string) bool {
	if _, ok := c.Cloud.TenantAssignments[repoID]; !ok {
		return false
	}
	delete(c.Cloud.TenantAssignments, repoID)
	return true
}

// TenantFor returns the tenant slug assigned to repoID, if any.
func (c *Config) TenantFor(repoID string) (string, bool) {
	slug, ok := c.Cloud.TenantAssignments[repoID]
	return slug, ok
}

// SetTenantRemote records the Dolt remote URL for a tenant slug.
func (c *Config) SetTenantRemote(slug, url string) {
	if c.Cloud.TenantRemotes == nil {
		c.Cloud.TenantRemotes = map[string]string{}
	}
	c.Cloud.TenantRemotes[slug] = url
}

// TenantRemote returns the Dolt remote URL configured for a tenant slug.
func (c *Config) TenantRemote(slug string) (string, bool) {
	url, ok := c.Cloud.TenantRemotes[slug]
	return url, ok
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, Dir, FileName), nil
}

func Load() (*Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return &Config{}, nil
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return c.SaveTo(path)
}

func (c *Config) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
