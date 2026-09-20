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
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestResolverPathBehavior locks in the a2a-go v2.5.0 agent-card resolver
// path-handling contract that OURS depends on (a2a-go #413/#424).
//
// v2.4.0 ALWAYS appended /.well-known/agent-card.json to the base URL.
// v2.5.0 changed this: a base URL with a NON-ROOT path is fetched directly
// as the complete card URL, while a base URL with no path or a root path
// still gets /.well-known/agent-card.json appended.
//
// OURS passes the raw user-supplied service URL to Resolve with no explicit
// WithPath option (see resolveAgentCard in main.go), so this default routing
// is exactly what OURS relies on. The test proves OURS still discovers the
// card under both resolver configurations getResolver() can produce.
func TestResolverPathBehavior(t *testing.T) {
	const (
		wellKnownPath = "/.well-known/agent-card.json"
		nonRootPath   = "/agents/geo/card"
		cardBody      = `{"name":"Path Behavior Agent","version":"1.0.0","protocolVersion":"0.3"}`
		wantName      = "Path Behavior Agent"
	)

	var mu sync.Mutex
	var requested []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requested = append(requested, r.URL.Path)
		mu.Unlock()

		switch r.URL.Path {
		case wellKnownPath, nonRootPath:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(cardBody))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	lastRequested := func() string {
		mu.Lock()
		defer mu.Unlock()
		if len(requested) == 0 {
			return ""
		}
		return requested[len(requested)-1]
	}

	// getResolver() reads these package-level globals; restore them afterwards
	// so the test does not leak state into other tests.
	origProtocol, origTimeout := protocol, requestTimeout
	t.Cleanup(func() {
		protocol = origProtocol
		requestTimeout = origTimeout
	})
	requestTimeout = 10 * time.Second

	// Both resolver configurations getResolver() can hand back: the default
	// parser (no --protocol) and the v0.3 compat parser (--protocol 0.3.x).
	// Path routing is parser-independent, but wiring through getResolver()
	// proves OURS's real configuration routes correctly in both cases.
	resolverCases := []struct {
		name     string
		protocol string
	}{
		{name: "default_parser", protocol: ""},
		{name: "compat_parser", protocol: "0.3"},
	}

	for _, rc := range resolverCases {
		t.Run(rc.name, func(t *testing.T) {
			// Root base URL: /.well-known/agent-card.json must still be
			// appended (behavior unchanged from v2.4.0).
			t.Run("root_url_appends_well_known", func(t *testing.T) {
				protocol = rc.protocol
				card, err := getResolver().Resolve(context.Background(), srv.URL)
				if err != nil {
					t.Fatalf("Resolve(%q) failed: %v", srv.URL, err)
				}
				if card.Name != wantName {
					t.Errorf("card name = %q, want %q", card.Name, wantName)
				}
				if got := lastRequested(); got != wellKnownPath {
					t.Errorf("root URL fetched %q, want the well-known path %q", got, wellKnownPath)
				}
			})

			// Non-root path base URL: v2.5.0 fetches it DIRECTLY and must
			// NOT append the well-known path. This is the #413/#424 change
			// with regression potential — the point of the exercise.
			t.Run("non_root_url_fetched_directly", func(t *testing.T) {
				protocol = rc.protocol
				target := srv.URL + nonRootPath
				card, err := getResolver().Resolve(context.Background(), target)
				if err != nil {
					t.Fatalf("Resolve(%q) failed: %v", target, err)
				}
				if card.Name != wantName {
					t.Errorf("card name = %q, want %q", card.Name, wantName)
				}
				if got := lastRequested(); got != nonRootPath {
					t.Errorf("non-root URL fetched %q, want the URL fetched directly at %q "+
						"(a2a-go #413/#424: v2.5.0 no longer appends %q for non-root paths)",
						got, nonRootPath, wellKnownPath)
				}
			})
		})
	}
}
