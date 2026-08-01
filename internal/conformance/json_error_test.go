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
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestStructuredJSONErrorOutput(t *testing.T) {
	// Build a temporary binary to run CLI execution test
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "a2acli_test_bin")

	buildCmd := exec.Command("go", "build", "-o", binPath, "../../cmd/a2acli")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build test binary: %v\nOutput: %s", err, string(out))
	}

	// Invoke a failing command (invalid service URL) with --output json
	cmd := exec.Command(binPath, "discover", "-u", "http://127.0.0.1:1", "--output", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected command to fail, but it succeeded")
	}

	stderrStr := stderr.String()
	t.Logf("Captured stderr: %s", stderrStr)

	var errPayload map[string]string
	if err := json.Unmarshal(stderr.Bytes(), &errPayload); err != nil {
		t.Fatalf("stderr does not parse as valid JSON: %v\nRaw stderr: %s", err, stderrStr)
	}

	if errPayload["error"] == "" {
		t.Error("expected 'error' field in JSON error payload")
	}
	if errPayload["code"] == "" {
		t.Error("expected 'code' field in JSON error payload")
	}
}
