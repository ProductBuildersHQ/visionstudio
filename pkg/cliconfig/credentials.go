package cliconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CredentialsFileName holds secrets (VisionStudio Cloud API keys). It is
// kept separate from config.json so that file can be shown, shared, or
// edited without exposing credentials, and is always written 0600.
const CredentialsFileName = "credentials.json"

// CloudCredential is a stored VisionStudio Cloud credential for one cloud
// base URL.
type CloudCredential struct {
	// Token is sent as "Authorization: Bearer <token>": a VisionStudio
	// Cloud API key (vsc_live_...) or a session JWT.
	Token string `json:"token"`
	// Tenant is the default tenant slug used with this URL when none is
	// given explicitly.
	Tenant  string    `json:"tenant,omitempty"`
	SavedAt time.Time `json:"saved_at,omitempty"`
}

// Credentials maps a canonical cloud base URL (scheme://host[/prefix], no
// trailing slash, no /t/{tenant}) to its stored credential.
type Credentials struct {
	Cloud map[string]CloudCredential `json:"cloud,omitempty"`
}

// CredentialsPath returns ~/.productbuildershq/visionstudio/credentials.json.
func CredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home: %w", err)
	}
	return filepath.Join(home, Dir, CredentialsFileName), nil
}

// LoadCredentials reads the default credentials file; a missing file yields
// empty credentials.
func LoadCredentials() (*Credentials, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	return LoadCredentialsFrom(path)
}

// LoadCredentialsFrom reads credentials from path; a missing file yields
// empty credentials.
func LoadCredentialsFrom(path string) (*Credentials, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is the user's own credentials file
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
}

// Save writes the default credentials file.
func (c *Credentials) Save() error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	return c.SaveTo(path)
}

// SaveTo writes credentials to path with owner-only permissions (0600),
// tightening an existing file's mode if it was looser.
func (c *Credentials) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	// WriteFile only applies the mode when creating the file.
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	return nil
}

// Get returns the credential stored for baseURL.
func (c *Credentials) Get(baseURL string) (CloudCredential, bool) {
	cred, ok := c.Cloud[baseURL]
	return cred, ok
}

// Set stores cred for baseURL.
func (c *Credentials) Set(baseURL string, cred CloudCredential) {
	if c.Cloud == nil {
		c.Cloud = map[string]CloudCredential{}
	}
	c.Cloud[baseURL] = cred
}

// Delete removes the credential for baseURL, reporting whether one existed.
func (c *Credentials) Delete(baseURL string) bool {
	if _, ok := c.Cloud[baseURL]; !ok {
		return false
	}
	delete(c.Cloud, baseURL)
	return true
}
