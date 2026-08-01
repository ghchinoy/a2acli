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

package conformance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type pluginManifest struct {
	Schema      string            `json:"$schema"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	License     string            `json:"license"`
	Keywords    []string          `json:"keywords"`
	Author      map[string]string `json:"author"`
}

func TestPluginManifest(t *testing.T) {
	// Root plugin.json location relative to internal/conformance
	rootPath := filepath.Join("..", "..", "plugin.json")
	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("failed to read root plugin.json: %v", err)
	}

	var manifest pluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("plugin.json is not valid JSON: %v", err)
	}

	expectedSchema := "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json"
	if manifest.Schema != expectedSchema {
		t.Errorf("expected $schema %q, got %q", expectedSchema, manifest.Schema)
	}

	if manifest.Name != "a2acli" {
		t.Errorf("expected name %q, got %q", "a2acli", manifest.Name)
	}

	validName := regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
	if !validName.MatchString(manifest.Name) {
		t.Errorf("plugin name %q violates Agent Plugins naming constraints", manifest.Name)
	}

	if manifest.Description == "" {
		t.Error("plugin.json description is empty")
	}

	if manifest.License != "Apache-2.0" {
		t.Errorf("expected license Apache-2.0, got %q", manifest.License)
	}
}

func TestAgentSkillsLayout(t *testing.T) {
	skillsDir := filepath.Join("..", "..", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("failed to read skills directory: %v", err)
	}

	validSkillName := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

	skillCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		skillPath := filepath.Join(skillsDir, skillName, "SKILL.md")

		data, err := os.ReadFile(skillPath)
		if err != nil {
			t.Errorf("skill directory %s missing SKILL.md: %v", skillName, err)
			continue
		}

		skillCount++
		if !validSkillName.MatchString(skillName) {
			t.Errorf("skill directory name %q violates Agent Skills naming rules", skillName)
		}

		content := string(data)
		if !strings.HasPrefix(content, "---") {
			t.Errorf("skill %s/SKILL.md missing frontmatter start '---'", skillName)
		}
		if !strings.Contains(content, "name: "+skillName) {
			t.Errorf("skill %s/SKILL.md frontmatter name does not match directory name", skillName)
		}
	}

	if skillCount == 0 {
		t.Error("no skills found in skills/ directory")
	}
}
