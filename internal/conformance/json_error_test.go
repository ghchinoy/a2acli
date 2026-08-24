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

	// Invoke a failing command (invalid service URL) with --output json.
	// Per SPEC §11.1/§11.4 the machine-readable error envelope is the
	// structured payload and MUST be emitted on stdout; diagnostics stay on
	// stderr.
	cmd := exec.Command(binPath, "discover", "-u", "http://127.0.0.1:1", "--output", "json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected command to fail, but it succeeded")
	}

	stdoutStr := stdout.String()
	t.Logf("Captured stdout: %s", stdoutStr)
	t.Logf("Captured stderr: %s", stderr.String())

	// The error envelope (SPEC Appendix B) is a nested object:
	//   {"error":{"code","message","hint","a2aCode"}}
	var payload struct {
		Error struct {
			Code    string  `json:"code"`
			Message string  `json:"message"`
			Hint    *string `json:"hint"`
			A2ACode any     `json:"a2aCode"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout does not parse as valid JSON error envelope: %v\nRaw stdout: %s", err, stdoutStr)
	}

	if payload.Error.Message == "" {
		t.Error("expected non-empty 'error.message' field in JSON error envelope")
	}
	if payload.Error.Code == "" {
		t.Error("expected non-empty 'error.code' field in JSON error envelope")
	}
	// CLI-local failures MUST use the A2ACLI_ERR_ namespace (SPEC §11.4 / App. D).
	const ns = "A2ACLI_ERR_"
	if len(payload.Error.Code) < len(ns) || payload.Error.Code[:len(ns)] != ns {
		t.Errorf("expected code in %s namespace, got %q", ns, payload.Error.Code)
	}
}
