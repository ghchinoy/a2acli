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
	"testing"

	"github.com/spf13/cobra"
)

// resetGlobalFlags points the package-level rootCmd at a fresh command with the
// full global flag set registered, and clears the variables those flags bind to.
// It exists so alias-resolution tests can parse argument vectors in isolation:
// the *FlagChanged helpers and resolveAliasEnv read rootCmd, so the tests must
// drive that same command. Tests run sequentially (no t.Parallel), so mutating
// the shared rootCmd between them is safe.
func resetGlobalFlags(t *testing.T) *cobra.Command {
	t.Helper()
	serviceURL, protocol, contextID, targetTaskID = "", "", "", ""
	authToken, bearerToken, apiKey = "", "", ""
	outputMode = ""
	disableTUI = false
	transports = nil

	rootCmd = &cobra.Command{Use: "a2acli", SilenceErrors: true, SilenceUsage: true}
	rootCmd.Run = func(*cobra.Command, []string) {}
	addGlobalFlags(rootCmd)
	return rootCmd
}

// TestFlagAliasesResolveToSameVar verifies each canonical alias (Roadmap A1)
// sets the exact same underlying variable as its legacy spelling.
func TestFlagAliasesResolveToSameVar(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want func() (got, want string)
	}{
		{"agent-card -> serviceURL", []string{"--agent-card", "http://x:1"},
			func() (string, string) { return serviceURL, "http://x:1" }},
		{"-a -> serviceURL", []string{"-a", "http://x:2"},
			func() (string, string) { return serviceURL, "http://x:2" }},
		{"endpoint -> serviceURL", []string{"--endpoint", "http://x:3"},
			func() (string, string) { return serviceURL, "http://x:3" }},
		{"service-url legacy still works", []string{"--service-url", "http://x:4"},
			func() (string, string) { return serviceURL, "http://x:4" }},
		{"context-id -> contextID", []string{"--context-id", "ctx-1"},
			func() (string, string) { return contextID, "ctx-1" }},
		{"context legacy still works", []string{"--context", "ctx-2"},
			func() (string, string) { return contextID, "ctx-2" }},
		{"task-id -> targetTaskID", []string{"--task-id", "task-1"},
			func() (string, string) { return targetTaskID, "task-1" }},
		{"task/-k legacy still works", []string{"-k", "task-2"},
			func() (string, string) { return targetTaskID, "task-2" }},
		{"a2a-version -> protocol", []string{"--a2a-version", "0.3.0"},
			func() (string, string) { return protocol, "0.3.0" }},
		{"protocol/-p legacy still works", []string{"-p", "0.3.0"},
			func() (string, string) { return protocol, "0.3.0" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := resetGlobalFlags(t)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags(%v): %v", tt.args, err)
			}
			got, want := tt.want()
			if got != want {
				t.Errorf("after %v: got %q, want %q", tt.args, got, want)
			}
		})
	}
}

// TestFlagAliasLastWins verifies that when a setting is supplied under two
// spellings on one command line, the value is applied left-to-right and the last
// occurrence wins — never a silent double-apply (front-loaded A1 requirement).
func TestFlagAliasLastWins(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"service-url then agent-card: agent-card wins", []string{"--service-url", "http://a", "--agent-card", "http://b"}, "http://b"},
		{"agent-card then service-url: service-url wins", []string{"--agent-card", "http://b", "--service-url", "http://a"}, "http://a"},
		{"endpoint then agent-card: agent-card wins", []string{"--endpoint", "http://e", "--agent-card", "http://c"}, "http://c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := resetGlobalFlags(t)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags(%v): %v", tt.args, err)
			}
			if serviceURL != tt.want {
				t.Errorf("after %v: serviceURL = %q, want %q", tt.args, serviceURL, tt.want)
			}
			if !serviceURLFlagChanged() {
				t.Errorf("serviceURLFlagChanged() = false, want true after explicit flag")
			}
		})
	}
}

// TestResolveAliasEnvPrecedence verifies the service-URL env aliases and their
// precedence relative to an explicit flag (flag > env), and the order among the
// env aliases themselves (A2ACLI_AGENT_CARD > A2ACLI_ENDPOINT > A2ACLI_SERVICE_URL).
func TestResolveAliasEnvPrecedence(t *testing.T) {
	t.Run("env fills serviceURL when no flag given", func(t *testing.T) {
		resetGlobalFlags(t)
		t.Setenv("A2ACLI_AGENT_CARD", "http://env-card")
		resolveAliasEnv()
		if serviceURL != "http://env-card" {
			t.Errorf("serviceURL = %q, want http://env-card", serviceURL)
		}
	})

	t.Run("explicit flag overrides env", func(t *testing.T) {
		cmd := resetGlobalFlags(t)
		t.Setenv("A2ACLI_AGENT_CARD", "http://env-card")
		if err := cmd.ParseFlags([]string{"--endpoint", "http://flag"}); err != nil {
			t.Fatal(err)
		}
		resolveAliasEnv()
		if serviceURL != "http://flag" {
			t.Errorf("serviceURL = %q, want http://flag (flag must beat env)", serviceURL)
		}
	})

	t.Run("agent-card env beats endpoint and service-url env", func(t *testing.T) {
		resetGlobalFlags(t)
		t.Setenv("A2ACLI_AGENT_CARD", "http://card")
		t.Setenv("A2ACLI_ENDPOINT", "http://endpoint")
		t.Setenv("A2ACLI_SERVICE_URL", "http://svc")
		resolveAliasEnv()
		if serviceURL != "http://card" {
			t.Errorf("serviceURL = %q, want http://card", serviceURL)
		}
	})
}

// TestAsyncAliasBindsImmediate verifies the send command's --async / --immediate
// wiring (Roadmap A1): both spellings bind to the single `immediate` variable, so
// either sets it and neither double-applies. It mirrors the exact registration in
// main() (both BoolVar into &immediate) on a throwaway command, since sendCmd is
// constructed inside main() and is not reachable from tests.
func TestAsyncAliasBindsImmediate(t *testing.T) {
	newSend := func() *cobra.Command {
		c := &cobra.Command{Use: "send"}
		c.Flags().BoolVar(&immediate, "immediate", false, "")
		c.Flags().BoolVar(&immediate, "async", false, "")
		return c
	}

	t.Run("--async sets immediate", func(t *testing.T) {
		immediate = false
		if err := newSend().ParseFlags([]string{"--async"}); err != nil {
			t.Fatal(err)
		}
		if !immediate {
			t.Error("--async did not set the immediate variable")
		}
	})

	t.Run("--immediate still sets immediate", func(t *testing.T) {
		immediate = false
		if err := newSend().ParseFlags([]string{"--immediate"}); err != nil {
			t.Fatal(err)
		}
		if !immediate {
			t.Error("--immediate did not set the immediate variable")
		}
	})

	immediate = false
}
