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

// Package oauth provides PKCE utilities and the local callback server for
// a2acli's OAuth 2.1 auth-code flow.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// CallbackAddr is the pre-registered redirect URI base. The mithlond consent SPA's
// metadata.json registers http://127.0.0.1:8080/callback — the port is FIXED.
// a2acli must bind exactly here; port randomisation is not possible without a new
// registration.
const CallbackAddr = "127.0.0.1:8080"
const CallbackPath = "/callback"
const RedirectURI = "http://" + CallbackAddr + CallbackPath

// CIMDURL is a2acli's own Client Instance Metadata Document URL. Used as client_id.
const CIMDURL = "https://ghchinoy.github.io/a2acli/metadata.json"

// Challenge holds a PKCE S256 code_verifier / code_challenge pair.
type Challenge struct {
	Verifier  string
	Challenge string
}

// NewChallenge generates a fresh PKCE S256 code_verifier and derives the
// code_challenge (BASE64URL(SHA256(verifier))).
func NewChallenge() (Challenge, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return Challenge{}, fmt.Errorf("generate verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return Challenge{Verifier: verifier, Challenge: challenge}, nil
}

// AuthURL constructs the authorization URL for the auth-code + PKCE flow.
func AuthURL(authEndpoint, state string, pkce Challenge) string {
	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {CIMDURL},
		"redirect_uri":          {RedirectURI},
		"state":                 {state},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {"S256"},
	}
	return authEndpoint + "?" + v.Encode()
}

// OpenBrowser opens the given URL in the system default browser.
func OpenBrowser(u string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{u}
	case "linux":
		cmd, args = "xdg-open", []string{u}
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", u}
	default:
		return fmt.Errorf("unsupported OS %q — open %s manually", runtime.GOOS, u)
	}
	return exec.Command(cmd, args...).Start()
}

// CallbackResult is the outcome of the local callback server.
type CallbackResult struct {
	Code  string
	State string
	Err   error
}

// StartCallbackServer starts a local HTTP server on CallbackAddr and waits for
// the OAuth callback. Returns a channel that yields exactly one CallbackResult.
//
// It returns an error immediately if the port cannot be bound — this is the
// "port 8080 in use" failure path documented in a2ac-38z.2.
func StartCallbackServer(ctx context.Context, expectedState string) (<-chan CallbackResult, func(), error) {
	ln, err := net.Listen("tcp", CallbackAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot bind %s: %w\nHint: free port 8080 and retry (the OAuth callback server must listen on that exact address)", CallbackAddr, err)
	}

	ch := make(chan CallbackResult, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc(CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state := q.Get("state")
		code := q.Get("code")
		errParam := q.Get("error")

		if errParam != "" {
			ch <- CallbackResult{Err: fmt.Errorf("auth server error: %s — %s", errParam, q.Get("error_description"))}
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s</p><p>You may close this tab.</p></body></html>", errParam)
			return
		}
		if state != expectedState {
			ch <- CallbackResult{Err: fmt.Errorf("state mismatch: got %q, expected %q (possible CSRF)", state, expectedState)}
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>State mismatch.</p></body></html>")
			return
		}
		ch <- CallbackResult{Code: code, State: state}
		fmt.Fprintf(w, "<html><body><h2>Authentication successful!</h2><p>You may close this tab and return to a2acli.</p></body></html>")
	})

	go func() {
		_ = srv.Serve(ln)
	}()

	stop := func() { _ = srv.Shutdown(context.Background()) }
	return ch, stop, nil
}

// TokenResponse is the minimal shape of an OAuth 2.1 token endpoint response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// ExchangeCode exchanges an authorization code for tokens at tokenEndpoint.
func ExchangeCode(ctx context.Context, tokenEndpoint, code string, pkce Challenge) (*TokenResponse, error) {
	body := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {RedirectURI},
		"client_id":     {CIMDURL},
		"code_verifier": {pkce.Verifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint,
		strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
			Desc  string `json:"error_description"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, fmt.Errorf("token endpoint returned %d: %s — %s", resp.StatusCode, e.Error, e.Desc)
	}

	var tok TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	return &tok, nil
}

// StoredToken is the on-disk token record keyed by service URL host.
type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	TokenURL     string    `json:"token_url"`
}

// IsExpired reports whether the token has expired (with a 30s buffer).
func (t *StoredToken) IsExpired() bool {
	return !t.ExpiresAt.IsZero() && time.Now().After(t.ExpiresAt.Add(-30*time.Second))
}
