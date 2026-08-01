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

package oauth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ghchinoy/a2acli/internal/oauth"
)

func TestRefreshAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.FormValue("grant_type") != "refresh_token" {
			t.Errorf("expected grant_type refresh_token, got %s", r.FormValue("grant_type"))
		}
		if r.FormValue("refresh_token") != "test_refresh_token" {
			t.Errorf("expected refresh_token test_refresh_token, got %s", r.FormValue("refresh_token"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new_access_token",
			"refresh_token": "new_refresh_token",
			"expires_in":    3600,
			"scope":         "agent:invoke",
		})
	}))
	defer ts.Close()

	ctx := context.Background()
	resp, err := oauth.RefreshAccessToken(ctx, ts.URL, "test_refresh_token")
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if resp.AccessToken != "new_access_token" {
		t.Errorf("expected access_token new_access_token, got %s", resp.AccessToken)
	}
	if resp.RefreshToken != "new_refresh_token" {
		t.Errorf("expected refresh_token new_refresh_token, got %s", resp.RefreshToken)
	}
}

func TestLoadValidToken_ExpiredWithRefresh(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "refreshed_access_token",
			"refresh_token": "refreshed_refresh_token",
			"expires_in":    1800,
			"scope":         "agent:invoke",
		})
	}))
	defer ts.Close()

	serviceURL := "http://test-service.local:9999"

	expiredToken := &oauth.StoredToken{
		AccessToken:  "old_access_token",
		RefreshToken: "old_refresh_token",
		ExpiresAt:    time.Now().Add(-10 * time.Minute),
		Scope:        "agent:invoke",
		TokenURL:     ts.URL,
	}

	if err := oauth.SaveToken(serviceURL, expiredToken); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}
	defer func() { _ = oauth.DeleteToken(serviceURL) }()

	ctx := context.Background()
	loaded, err := oauth.LoadValidToken(ctx, serviceURL)
	if err != nil {
		t.Fatalf("LoadValidToken failed: %v", err)
	}

	if loaded.AccessToken != "refreshed_access_token" {
		t.Errorf("expected refreshed access token, got %s", loaded.AccessToken)
	}
	if loaded.IsExpired() {
		t.Error("expected loaded token to be valid (not expired)")
	}
}
