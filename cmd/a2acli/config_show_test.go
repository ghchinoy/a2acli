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

package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it
// wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

// TestConfigShowRedactsCredentials verifies `config show` never emits credential
// material — token, bearer, and api-key values MUST be redacted regardless of
// output mode (SPEC §12.1, not defeasible).
func TestConfigShowRedactsCredentials(t *testing.T) {
	const secretTok = "super-secret-token-value"
	const secretBearer = "super-secret-bearer-value"
	const secretKey = "super-secret-api-key-value"

	for _, jsonMode := range []bool{false, true} {
		name := "table"
		if jsonMode {
			name = "json"
		}
		t.Run(name, func(t *testing.T) {
			resetGlobalFlags(t)
			authToken = secretTok
			bearerToken = secretBearer
			apiKey = secretKey
			serviceURL = "http://host:9001"
			disableTUI = jsonMode

			out := captureStdout(t, func() { runConfigShow(nil, nil) })

			for _, secret := range []string{secretTok, secretBearer, secretKey} {
				if strings.Contains(out, secret) {
					t.Errorf("output leaked a credential value %q:\n%s", secret, out)
				}
			}
			// The redaction marker appears literally (<redacted>) in the table
			// output and unicode-escaped by Go's JSON encoder; both contain the
			// substring "redacted".
			if !strings.Contains(out, "redacted") {
				t.Errorf("expected redaction marker in output:\n%s", out)
			}
			// Non-credential setting is shown in the clear.
			if !strings.Contains(out, "http://host:9001") {
				t.Errorf("expected service-url shown in clear:\n%s", out)
			}
		})
	}

	authToken, bearerToken, apiKey = "", "", ""
}

// TestConfigShowJSONShape verifies the JSON output is an array of settingView
// rows and that unset credentials render as (unset), not empty leaks.
func TestConfigShowJSONShape(t *testing.T) {
	resetGlobalFlags(t)
	authToken, bearerToken, apiKey = "", "", ""
	disableTUI = true

	out := captureStdout(t, func() { runConfigShow(nil, nil) })

	var rows []settingView
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("output is not a settingView array: %v\n%s", err, out)
	}
	byName := map[string]settingView{}
	for _, r := range rows {
		byName[r.Name] = r
	}
	for _, cred := range []string{"token", "bearer", "api-key"} {
		if byName[cred].Value != "(unset)" {
			t.Errorf("%s value = %q, want (unset)", cred, byName[cred].Value)
		}
	}
	if byName["service-url"].Source != "default" {
		t.Errorf("service-url source = %q, want default", byName["service-url"].Source)
	}
}

// TestResolveSourcePrecedence verifies source resolution follows
// flag > env > config-file > default (SPEC §8.3 / CONFIG_002).
func TestResolveSourcePrecedence(t *testing.T) {
	t.Run("default when nothing set", func(t *testing.T) {
		resetGlobalFlags(t)
		if s := resolveSource([]string{"service-url"}, []string{"A2ACLI_SERVICE_URL"}, ""); s != "default" {
			t.Errorf("source = %q, want default", s)
		}
	})

	t.Run("env beats default", func(t *testing.T) {
		resetGlobalFlags(t)
		t.Setenv("A2ACLI_SERVICE_URL", "http://env")
		if s := resolveSource([]string{"service-url"}, []string{"A2ACLI_SERVICE_URL"}, ""); s != "env" {
			t.Errorf("source = %q, want env", s)
		}
	})

	t.Run("flag beats env", func(t *testing.T) {
		cmd := resetGlobalFlags(t)
		t.Setenv("A2ACLI_SERVICE_URL", "http://env")
		if err := cmd.ParseFlags([]string{"--service-url", "http://flag"}); err != nil {
			t.Fatal(err)
		}
		if s := resolveSource([]string{"service-url"}, []string{"A2ACLI_SERVICE_URL"}, ""); s != "flag" {
			t.Errorf("source = %q, want flag", s)
		}
	})

	t.Run("alias flag counts as flag source", func(t *testing.T) {
		cmd := resetGlobalFlags(t)
		if err := cmd.ParseFlags([]string{"--agent-card", "http://flag"}); err != nil {
			t.Fatal(err)
		}
		if s := resolveSource(serviceURLSpellings, nil, ""); s != "flag" {
			t.Errorf("source = %q, want flag", s)
		}
	})
}
