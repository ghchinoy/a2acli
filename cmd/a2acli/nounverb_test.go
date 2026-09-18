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
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// nounVerbTestRoot builds a throwaway root command wired exactly as main() wires
// it for the noun-verb layer: the full global (persistent) flag set plus the
// additive `card`/`task` parents. It lets tests verify the Roadmap B paths
// resolve and carry the right flags without invoking main().
func nounVerbTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "a2acli", SilenceErrors: true, SilenceUsage: true}
	root.AddGroup(
		&cobra.Group{ID: GroupDiscovery, Title: "Discovery & Identity:"},
		&cobra.Group{ID: GroupMessaging, Title: "Messaging & Tasks:"},
	)
	addGlobalFlags(root)
	root.AddCommand(setupNounVerbCommands()...)
	return root
}

// flagNames returns the sorted set of local flag names registered on cmd.
func flagNames(cmd *cobra.Command) []string {
	var names []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		names = append(names, f.Name)
	})
	sort.Strings(names)
	return names
}

// TestNounVerbPathsResolve verifies each Roadmap B noun-verb path resolves to a
// real, wired command (never "unknown command") and reaches the intended leaf.
func TestNounVerbPathsResolve(t *testing.T) {
	root := nounVerbTestRoot()

	tests := []struct {
		name     string
		args     []string
		wantLeaf string // expected leaf command Name()
		runnable bool   // whether the leaf must itself be runnable
	}{
		{"card get", []string{"card", "get"}, "get", true},
		{"task get", []string{"task", "get"}, "get", true},
		{"task cancel", []string{"task", "cancel"}, "cancel", true},
		{"task list", []string{"task", "list"}, "list", true},
		{"task subscribe", []string{"task", "subscribe"}, "subscribe", true},
		{"task push-config", []string{"task", "push-config"}, "push-config", false},
		{"task push-config create", []string{"task", "push-config", "create"}, "create", true},
		{"task push-config list", []string{"task", "push-config", "list"}, "list", true},
		{"task push-config get", []string{"task", "push-config", "get"}, "get", true},
		{"task push-config delete", []string{"task", "push-config", "delete"}, "delete", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _, err := root.Find(tt.args)
			if err != nil {
				t.Fatalf("Find(%v): unexpected error %v", tt.args, err)
			}
			if c.Name() != tt.wantLeaf {
				t.Fatalf("Find(%v): resolved to %q, want leaf %q (path %q)",
					tt.args, c.Name(), tt.wantLeaf, c.CommandPath())
			}
			// A fully-resolved path's CommandPath must equal the requested tokens;
			// an "unknown command" would stop short at the parent instead.
			wantPath := "a2acli " + strings.Join(tt.args, " ")
			if c.CommandPath() != wantPath {
				t.Fatalf("CommandPath = %q, want %q (path stopped short = unknown command)",
					c.CommandPath(), wantPath)
			}
			if tt.runnable && !c.Runnable() {
				t.Fatalf("%q resolved but is not runnable (no Run wired)", tt.name)
			}
		})
	}
}

// TestNounVerbUnknownVerbNotWired is the negative control: an unknown verb under
// a real parent must NOT resolve to a leaf of that name — it stops at the parent.
func TestNounVerbUnknownVerbNotWired(t *testing.T) {
	root := nounVerbTestRoot()
	c, _, err := root.Find([]string{"task", "bogus"})
	if err != nil {
		t.Fatalf("Find: unexpected error %v", err)
	}
	if c.Name() == "bogus" {
		t.Fatalf("unexpected: 'task bogus' resolved to a wired command")
	}
	if c.Name() != "task" {
		t.Fatalf("'task bogus' resolved to %q, want to stop at parent 'task'", c.Name())
	}
}

// TestNounVerbSharesFlatFlags verifies each noun-verb leaf carries exactly the
// same local flag set as its flat verb. Both are built from the same flag-adder
// helpers, so comparing against a reference command constructed with the same
// helper proves the two paths accept identical flags.
func TestNounVerbSharesFlatFlags(t *testing.T) {
	root := nounVerbTestRoot()

	leaf := func(args ...string) *cobra.Command {
		c, _, err := root.Find(args)
		if err != nil || c.CommandPath() != "a2acli "+strings.Join(args, " ") {
			t.Fatalf("could not resolve %v", args)
		}
		return c
	}

	// task get <-> addGetFlags
	refGet := &cobra.Command{Use: "ref"}
	addGetFlags(refGet)
	assertSameFlags(t, "task get", flagNames(refGet), flagNames(leaf("task", "get")))

	// task subscribe <-> addSubscribeFlags
	refSub := &cobra.Command{Use: "ref"}
	addSubscribeFlags(refSub)
	assertSameFlags(t, "task subscribe", flagNames(refSub), flagNames(leaf("task", "subscribe")))

	// task list <-> addListTasksFlags
	refList := &cobra.Command{Use: "ref"}
	addListTasksFlags(refList)
	assertSameFlags(t, "task list", flagNames(refList), flagNames(leaf("task", "list")))

	// card get <-> --extended (matches flat 'discover')
	assertSameFlags(t, "card get", []string{"extended"}, flagNames(leaf("card", "get")))

	// task push-config subcommands <-> newPushConfigCmd reference
	ref := newPushConfigCmd("")
	for _, sub := range []string{"create", "list", "get", "delete"} {
		var want *cobra.Command
		for _, c := range ref.Commands() {
			if c.Name() == sub {
				want = c
			}
		}
		if want == nil {
			t.Fatalf("reference push-config missing subcommand %q", sub)
		}
		assertSameFlags(t, "task push-config "+sub, flagNames(want), flagNames(leaf("task", "push-config", sub)))
	}
}

// TestNounVerbInheritsGlobalFlags confirms the #39 canonical alias flags and the
// core credential/transport flags are reachable from a noun-verb leaf (they are
// persistent on root, so nested subcommands inherit them).
func TestNounVerbInheritsGlobalFlags(t *testing.T) {
	root := nounVerbTestRoot()
	c, _, err := root.Find([]string{"task", "get"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"agent-card", "endpoint", "context-id", "task-id", "a2a-version",
		"bearer", "api-key", "transport", "output",
	}
	inherited := c.InheritedFlags()
	for _, name := range want {
		if inherited.Lookup(name) == nil {
			t.Errorf("noun-verb leaf 'task get' does not inherit global flag --%s", name)
		}
	}
}

func assertSameFlags(t *testing.T, label string, want, got []string) {
	t.Helper()
	if strings.Join(want, ",") != strings.Join(got, ",") {
		t.Errorf("%s flag set = %v, want %v (must match the flat verb)", label, got, want)
	}
}
