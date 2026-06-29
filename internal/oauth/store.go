// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// tokenDir returns the XDG-compliant directory for token storage.
// Tokens live at ~/.config/a2acli/tokens/<host>.json with 0600 permissions.
func tokenDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	dir := filepath.Join(base, "a2acli", "tokens")
	return dir, os.MkdirAll(dir, 0700)
}

// tokenKey derives a filesystem-safe key from a service URL (uses the host).
func tokenKey(serviceURL string) (string, error) {
	u, err := url.Parse(serviceURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	// Sanitise: replace any remaining non-alphanum with '_'
	safe := strings.NewReplacer(".", "_", ":", "_", "/", "_").Replace(host)
	return safe + ".json", nil
}

// SaveToken persists a StoredToken for the given service URL.
func SaveToken(serviceURL string, tok *StoredToken) error {
	dir, err := tokenDir()
	if err != nil {
		return err
	}
	key, err := tokenKey(serviceURL)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, key)
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

// LoadToken retrieves the stored token for a service URL.
// Returns nil, nil if no token is stored.
func LoadToken(serviceURL string) (*StoredToken, error) {
	dir, err := tokenDir()
	if err != nil {
		return nil, err
	}
	key, err := tokenKey(serviceURL)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(dir, key))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var tok StoredToken
	if err := json.Unmarshal(b, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// DeleteToken removes the stored token for a service URL.
func DeleteToken(serviceURL string) error {
	dir, err := tokenDir()
	if err != nil {
		return err
	}
	key, err := tokenKey(serviceURL)
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(dir, key))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
