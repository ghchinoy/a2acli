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
	"encoding/json"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

// TestApiKeyAttachment verifies how the --api-key credential's location and
// parameter name are derived from the agent card's declared security scheme
// (SPEC §12.1 / A2A §4.5.2), defaulting to the X-Api-Key header.
func TestApiKeyAttachment(t *testing.T) {
	tests := []struct {
		name     string
		card     *a2a.AgentCard
		wantLoc  string
		wantName string
	}{
		{"nil card defaults to header X-Api-Key", nil, "header", "X-Api-Key"},
		{"no schemes defaults to header X-Api-Key", &a2a.AgentCard{}, "header", "X-Api-Key"},
		{
			"declared header scheme uses its name",
			&a2a.AgentCard{SecuritySchemes: a2a.NamedSecuritySchemes{
				"apikey": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationHeader, Name: "X-Custom-Key"},
			}},
			"header", "X-Custom-Key",
		},
		{
			"declared cookie scheme",
			&a2a.AgentCard{SecuritySchemes: a2a.NamedSecuritySchemes{
				"apikey": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationCookie, Name: "sid"},
			}},
			"cookie", "sid",
		},
		{
			"declared query scheme",
			&a2a.AgentCard{SecuritySchemes: a2a.NamedSecuritySchemes{
				"apikey": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationQuery, Name: "key"},
			}},
			"query", "key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, name := apiKeyAttachment(tt.card)
			if loc != tt.wantLoc || name != tt.wantName {
				t.Errorf("apiKeyAttachment() = (%q, %q), want (%q, %q)", loc, name, tt.wantLoc, tt.wantName)
			}
		})
	}
}

// TestApiKeyAttachmentDeterministic verifies that a card declaring more than one
// APIKeySecurityScheme yields a stable, documented choice regardless of the map
// iteration order (SecuritySchemes is a Go map). The rule: prefer a scheme named
// by SecurityRequirements (list order, ties broken by sorted scheme name), else
// the first APIKeySecurityScheme by sorted scheme name.
func TestApiKeyAttachmentDeterministic(t *testing.T) {
	t.Run("multiple schemes, no requirements: first by sorted name", func(t *testing.T) {
		card := &a2a.AgentCard{SecuritySchemes: a2a.NamedSecuritySchemes{
			"zeta":  a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationCookie, Name: "zsid"},
			"alpha": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationHeader, Name: "X-Alpha-Key"},
			"mid":   a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationQuery, Name: "mkey"},
		}}
		// Run repeatedly: map order varies but the result must not.
		for i := 0; i < 20; i++ {
			loc, name := apiKeyAttachment(card)
			if loc != "header" || name != "X-Alpha-Key" {
				t.Fatalf("iteration %d: apiKeyAttachment() = (%q, %q), want (header, X-Alpha-Key)", i, loc, name)
			}
		}
	})

	t.Run("requirement-referenced scheme wins over sorted fallback", func(t *testing.T) {
		card := &a2a.AgentCard{
			SecuritySchemes: a2a.NamedSecuritySchemes{
				"alpha": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationHeader, Name: "X-Alpha-Key"},
				"zeta":  a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationCookie, Name: "zsid"},
			},
			// The requirement names "zeta"; it must win even though "alpha" sorts first.
			SecurityRequirements: a2a.SecurityRequirementsOptions{
				a2a.SecurityRequirements{a2a.SecuritySchemeName("zeta"): a2a.SecuritySchemeScopes{}},
			},
		}
		for i := 0; i < 20; i++ {
			loc, name := apiKeyAttachment(card)
			if loc != "cookie" || name != "zsid" {
				t.Fatalf("iteration %d: apiKeyAttachment() = (%q, %q), want (cookie, zsid)", i, loc, name)
			}
		}
	})

	t.Run("multiple api-key schemes in one requirement: sorted tiebreak", func(t *testing.T) {
		card := &a2a.AgentCard{
			SecuritySchemes: a2a.NamedSecuritySchemes{
				"beta":  a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationHeader, Name: "X-Beta-Key"},
				"gamma": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationQuery, Name: "gkey"},
			},
			SecurityRequirements: a2a.SecurityRequirementsOptions{
				a2a.SecurityRequirements{
					a2a.SecuritySchemeName("gamma"): a2a.SecuritySchemeScopes{},
					a2a.SecuritySchemeName("beta"):  a2a.SecuritySchemeScopes{},
				},
			},
		}
		for i := 0; i < 20; i++ {
			loc, name := apiKeyAttachment(card)
			if loc != "header" || name != "X-Beta-Key" {
				t.Fatalf("iteration %d: apiKeyAttachment() = (%q, %q), want (header, X-Beta-Key)", i, loc, name)
			}
		}
	})

	t.Run("requirement names only non-api-key scheme: fall back to sorted api-key", func(t *testing.T) {
		card := &a2a.AgentCard{
			SecuritySchemes: a2a.NamedSecuritySchemes{
				"oauth": a2a.OAuth2SecurityScheme{},
				"zeta":  a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationCookie, Name: "zsid"},
				"alpha": a2a.APIKeySecurityScheme{Location: a2a.APIKeySecuritySchemeLocationHeader, Name: "X-Alpha-Key"},
			},
			SecurityRequirements: a2a.SecurityRequirementsOptions{
				a2a.SecurityRequirements{a2a.SecuritySchemeName("oauth"): a2a.SecuritySchemeScopes{}},
			},
		}
		for i := 0; i < 20; i++ {
			loc, name := apiKeyAttachment(card)
			if loc != "header" || name != "X-Alpha-Key" {
				t.Fatalf("iteration %d: apiKeyAttachment() = (%q, %q), want (header, X-Alpha-Key)", i, loc, name)
			}
		}
	})
}

// TestBearerTokenPrecedence verifies that when both --bearer and the legacy
// --token are supplied, --bearer wins and exactly one Authorization value is
// emitted (never two, which would be undefined on the wire).
func TestBearerTokenPrecedence(t *testing.T) {
	i := &paramInterceptor{token: "legacy", bearer: "canonical"}
	req := &a2aclient.Request{}
	if _, _, err := i.Before(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	got := req.ServiceParams["authorization"]
	if len(got) != 1 || got[0] != "Bearer canonical" {
		t.Errorf("authorization = %v, want [Bearer canonical]", got)
	}
}

// TestParamInterceptorAttachesCredentials verifies the interceptor places bearer
// and API-key credentials into the correct ServiceParams / URL location.
func TestParamInterceptorAttachesCredentials(t *testing.T) {
	t.Run("bearer", func(t *testing.T) {
		i := &paramInterceptor{bearer: "tok"}
		req := &a2aclient.Request{}
		if _, _, err := i.Before(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if got := req.ServiceParams["authorization"]; len(got) != 1 || got[0] != "Bearer tok" {
			t.Errorf("authorization = %v, want [Bearer tok]", got)
		}
	})

	t.Run("api-key header default", func(t *testing.T) {
		i := &paramInterceptor{apiKey: "k", apiKeyLoc: "header", apiKeyName: "X-Api-Key"}
		req := &a2aclient.Request{}
		if _, _, err := i.Before(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if got := req.ServiceParams["X-Api-Key"]; len(got) != 1 || got[0] != "k" {
			t.Errorf("X-Api-Key = %v, want [k]", got)
		}
	})

	t.Run("api-key cookie", func(t *testing.T) {
		i := &paramInterceptor{apiKey: "k", apiKeyLoc: "cookie", apiKeyName: "sid"}
		req := &a2aclient.Request{}
		if _, _, err := i.Before(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if got := req.ServiceParams["Cookie"]; len(got) != 1 || got[0] != "sid=k" {
			t.Errorf("Cookie = %v, want [sid=k]", got)
		}
	})

	t.Run("api-key query", func(t *testing.T) {
		i := &paramInterceptor{apiKey: "k", apiKeyLoc: "query", apiKeyName: "key"}
		req := &a2aclient.Request{BaseURL: "http://host/rpc"}
		if _, _, err := i.Before(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if req.BaseURL != "http://host/rpc?key=k" {
			t.Errorf("BaseURL = %q, want %q", req.BaseURL, "http://host/rpc?key=k")
		}
	})

	t.Run("no credentials leaves request untouched", func(t *testing.T) {
		i := &paramInterceptor{}
		req := &a2aclient.Request{}
		if _, _, err := i.Before(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if req.ServiceParams != nil {
			t.Errorf("ServiceParams = %v, want nil", req.ServiceParams)
		}
	})
}

// TestResolveCredentialsEnv verifies env fallback with flag precedence.
func TestResolveCredentialsEnv(t *testing.T) {
	t.Run("env fills empty flags", func(t *testing.T) {
		bearerToken, apiKey = "", ""
		t.Setenv("A2ACLI_BEARER", "env-b")
		t.Setenv("A2ACLI_API_KEY", "env-k")
		resolveCredentials()
		if bearerToken != "env-b" || apiKey != "env-k" {
			t.Errorf("got bearer=%q api-key=%q, want env-b / env-k", bearerToken, apiKey)
		}
	})

	t.Run("explicit flag overrides env", func(t *testing.T) {
		bearerToken, apiKey = "flag-b", "flag-k"
		t.Setenv("A2ACLI_BEARER", "env-b")
		t.Setenv("A2ACLI_API_KEY", "env-k")
		resolveCredentials()
		if bearerToken != "flag-b" || apiKey != "flag-k" {
			t.Errorf("got bearer=%q api-key=%q, want flag-b / flag-k", bearerToken, apiKey)
		}
	})

	// Reset globals so other tests are unaffected.
	bearerToken, apiKey = "", ""
}

// TestSendJSONWrapper verifies that a send result serializes with the App-B
// SendMessageResponse wrapper (SPEC §11.3 / Appendix B): a Task under "task",
// a Message under "message".
func TestSendJSONWrapper(t *testing.T) {
	tests := []struct {
		name    string
		result  a2a.SendMessageResult
		wantKey string
	}{
		{"task result", &a2a.Task{ID: "t1"}, "task"},
		{"message result", &a2a.Message{ID: "m1"}, "message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := marshalSendResult(tt.result)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(b, &doc); err != nil {
				t.Fatalf("not a single JSON document: %v\n%s", err, b)
			}
			keys := make([]string, 0, len(doc))
			for k := range doc {
				keys = append(keys, k)
			}
			if len(doc) != 1 {
				t.Fatalf("expected exactly one top-level key, got %v", keys)
			}
			if _, ok := doc[tt.wantKey]; !ok {
				t.Errorf("top-level key = %v, want %q", keys, tt.wantKey)
			}
		})
	}
}
