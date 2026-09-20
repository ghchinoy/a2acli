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
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// TestResolverAcceptsSecuritySchemeShapes locks in http/bearer parity with
// OFFICIAL: OURS's real getResolver() must accept OpenAPI-style AgentCard
// security schemes in the DEFAULT config (no --protocol), and must not regress
// on proto-wrapper-form or no-scheme cards.
//
// Before the fix, getResolver() only wired the v0 compat parser when
// --protocol 0.3.x was passed; the default path used DefaultCardParser, which
// rejects OpenAPI-style schemes ("unknown security scheme type ..."). The fix
// always uses the compat parser, mirroring OFFICIAL
// (a2a-cli/internal/cli/client.go). Since the compat parser is a strict
// superset of DefaultCardParser, all three card shapes below must parse under
// BOTH the default and --protocol 0.3.x configurations.
func TestResolverAcceptsSecuritySchemeShapes(t *testing.T) {
	const (
		wellKnownPath = "/.well-known/agent-card.json"
		agentName     = "Scheme Parity Agent"
	)

	cases := []struct {
		name string
		// cardBody is the JSON served at the well-known path.
		cardBody string
		// wantScheme, when non-empty, must be present in the resolved
		// card's SecuritySchemes as an HTTPAuthSecurityScheme.
		wantSchemeName string
		wantHTTPScheme string // e.g. "bearer"; "" to skip HTTP-scheme assertions
		wantSchemes    int
	}{
		{
			// The §6-Q3 alignment card: OpenAPI-style http/bearer. This is
			// the shape OURS previously rejected without --protocol 0.3.x.
			name: "openapi_http_bearer",
			cardBody: `{
				"name": "` + agentName + `",
				"version": "1.0.0",
				"protocolVersion": "0.3",
				"securitySchemes": {
					"bearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
				}
			}`,
			wantSchemeName: "bearerAuth",
			wantHTTPScheme: "bearer",
			wantSchemes:    1,
		},
		{
			// Proto-wrapper form — the shape DefaultCardParser handled.
			// Must not regress under the always-compat parser.
			name: "proto_wrapper_http",
			cardBody: `{
				"name": "` + agentName + `",
				"version": "1.0.0",
				"protocolVersion": "0.3",
				"securitySchemes": {
					"bearerAuth": {"httpAuthSecurityScheme": {"scheme": "bearer", "bearerFormat": "JWT"}}
				}
			}`,
			wantSchemeName: "bearerAuth",
			wantHTTPScheme: "bearer",
			wantSchemes:    1,
		},
		{
			// No security schemes at all — must still parse.
			name: "no_schemes",
			cardBody: `{
				"name": "` + agentName + `",
				"version": "1.0.0",
				"protocolVersion": "0.3"
			}`,
			wantSchemes: 0,
		},
	}

	// getResolver() reads these package-level globals; restore afterwards.
	origProtocol, origTimeout := protocol, requestTimeout
	t.Cleanup(func() {
		protocol = origProtocol
		requestTimeout = origTimeout
	})
	requestTimeout = 10 * time.Second

	// Both configurations getResolver() can produce. The default (no
	// --protocol) path is the one the brief requires to accept OpenAPI-style
	// cards; the 0.3.x path must keep working too.
	protocolCases := []struct {
		name     string
		protocol string
	}{
		{name: "default_no_protocol", protocol: ""},
		{name: "protocol_0_3", protocol: "0.3"},
	}

	for _, pc := range protocolCases {
		t.Run(pc.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path != wellKnownPath {
							http.NotFound(w, r)
							return
						}
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(tc.cardBody))
					}))
					defer srv.Close()

					protocol = pc.protocol
					card, err := getResolver().Resolve(context.Background(), srv.URL)
					if err != nil {
						t.Fatalf("getResolver().Resolve(%q) failed: %v", srv.URL, err)
					}
					if card.Name != agentName {
						t.Errorf("card name = %q, want %q", card.Name, agentName)
					}
					if got := len(card.SecuritySchemes); got != tc.wantSchemes {
						t.Errorf("got %d security schemes, want %d: %+v", got, tc.wantSchemes, card.SecuritySchemes)
					}
					if tc.wantSchemeName == "" {
						return
					}
					scheme, ok := card.SecuritySchemes[a2a.SecuritySchemeName(tc.wantSchemeName)]
					if !ok {
						t.Fatalf("security scheme %q not present; got %+v", tc.wantSchemeName, card.SecuritySchemes)
					}
					httpScheme, ok := scheme.(a2a.HTTPAuthSecurityScheme)
					if !ok {
						t.Fatalf("scheme %q has type %T, want a2a.HTTPAuthSecurityScheme", tc.wantSchemeName, scheme)
					}
					if httpScheme.Scheme != tc.wantHTTPScheme {
						t.Errorf("scheme %q = %q, want %q", tc.wantSchemeName, httpScheme.Scheme, tc.wantHTTPScheme)
					}
				})
			}
		})
	}
}
