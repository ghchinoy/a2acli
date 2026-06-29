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
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/ghchinoy/a2acli/internal/oauth"
	"github.com/spf13/cobra"
)

var (
	authClientID     string
	authClientSecret string
)

// setupAuthCmd builds the `auth` command group.
func setupAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:     "auth",
		GroupID: GroupSystem,
		Short:   "Manage OAuth 2.1 authentication for A2A services",
		Long: `Obtain, inspect, and revoke OAuth 2.1 tokens for A2A services that
require authentication. Tokens are stored in ~/.config/a2acli/tokens/ (0600)
and used automatically by send/discover when present.

Client identity: a2acli uses the CIMD pattern — its client_id is the URL of its
own metadata document (` + oauth.CIMDURL + `).`,
	}

	// login
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Obtain an OAuth 2.1 token for a service",
		Long: `Drive an OAuth 2.1 auth-code + PKCE flow against the service's
advertised authorization server. Opens a browser for the user to authenticate.

The service's AgentCard must advertise an OAuth2SecurityScheme with an
authorization URL and token endpoint. Use --client-id/--client-secret for
non-interactive client credentials flow instead.

Callback server binds to ` + oauth.RedirectURI + ` (pre-registered in the
mithlond consent SPA). Port 8080 must be free.`,
		Example: `  a2acli auth login --service-url https://eldamo.mithlond.com
  a2acli auth login -u https://eldamo.mithlond.com --client-id myid --client-secret mysecret`,
		Run: runAuthLogin,
	}
	loginCmd.Flags().StringVar(&authClientID, "client-id", "", "Client ID for client credentials flow (non-interactive)")
	loginCmd.Flags().StringVar(&authClientSecret, "client-secret", "", "Client secret for client credentials flow")

	// status
	statusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Show stored token status for a service",
		Example: `  a2acli auth status --service-url https://eldamo.mithlond.com`,
		Run:     runAuthStatus,
	}

	// logout
	logoutCmd := &cobra.Command{
		Use:     "logout",
		Short:   "Delete the stored token for a service",
		Example: `  a2acli auth logout --service-url https://eldamo.mithlond.com`,
		Run:     runAuthLogout,
	}

	// token (print raw JWT — for scripting, equivalent to make token)
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Print the stored access token (for scripting)",
		Long: `Print the raw JWT access token for the service. Equivalent to
'make token' in the service's own repo, but driven from a2acli's token store.
Exits non-zero if no valid token is stored.`,
		Example: `  a2acli auth token --service-url https://eldamo.mithlond.com
  TOKEN=$(a2acli auth token -u https://eldamo.mithlond.com)`,
		Run: runAuthToken,
	}

	authCmd.AddCommand(loginCmd, statusCmd, logoutCmd, tokenCmd)
	return authCmd
}

// oauthSchemeFromCard extracts the first OAuth2SecurityScheme from an AgentCard.
// Returns nil if none is present.
func oauthSchemeFromCard(card *a2a.AgentCard) *a2a.OAuth2SecurityScheme {
	for _, scheme := range card.SecuritySchemes {
		if s, ok := scheme.(a2a.OAuth2SecurityScheme); ok {
			return &s
		}
	}
	return nil
}

// authURLsFromScheme extracts (authorizationURL, tokenURL) from an OAuth2 scheme.
func authURLsFromScheme(s *a2a.OAuth2SecurityScheme) (authURL, tokenURL string) {
	if s == nil {
		return "", ""
	}
	switch f := s.Flows.(type) {
	case a2a.AuthorizationCodeOAuthFlow:
		return f.AuthorizationURL, f.TokenURL
	case a2a.ClientCredentialsOAuthFlow:
		return "", f.TokenURL
	case a2a.DeviceCodeOAuthFlow:
		return f.DeviceAuthorizationURL, f.TokenURL
	}
	return "", ""
}

// newState generates a random CSRF state token.
func newState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func runAuthLogin(_ *cobra.Command, _ []string) {
	ctx := context.Background()

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	scheme := oauthSchemeFromCard(card)
	if scheme == nil {
		fatalf("no OAuth2 security scheme in AgentCard", fmt.Errorf("serviceURL=%s", serviceURL),
			"Run 'a2acli discover' to inspect the card's security schemes")
	}

	authURL, tokenURL := authURLsFromScheme(scheme)
	if tokenURL == "" {
		fatalf("no token URL in OAuth2 scheme", fmt.Errorf("flows=%T", scheme.Flows),
			"The AgentCard's OAuth2SecurityScheme must include a token endpoint")
	}

	verboseLog("oauth2: authURL=%s tokenURL=%s", authURL, tokenURL)

	// Client credentials flow (non-interactive).
	if authClientID != "" {
		runClientCredentials(ctx, tokenURL)
		return
	}

	// Auth-code + PKCE flow requires an authorization URL.
	if authURL == "" {
		fatalf("no authorization URL in OAuth2 scheme", fmt.Errorf("flows=%T", scheme.Flows),
			"Use --client-id/--client-secret for client credentials flow, or check the AgentCard")
	}

	pkce, err := oauth.NewChallenge()
	if err != nil {
		fatalf("failed to generate PKCE challenge", err, "")
	}

	state, err := newState()
	if err != nil {
		fatalf("failed to generate state", err, "")
	}

	// Start callback server — fails immediately if port 8080 is in use (a2ac-38z.2).
	ch, stop, err := oauth.StartCallbackServer(ctx, state)
	if err != nil {
		fatalf("failed to start callback server", err,
			"Free port 8080 and retry. The OAuth callback must listen on "+oauth.RedirectURI)
	}
	defer stop()

	loginURL := oauth.AuthURL(authURL, state, pkce)
	verboseLog("opening browser: %s", loginURL)

	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If the browser does not open, visit:\n  %s\n\n", loginURL)

	if err := oauth.OpenBrowser(loginURL); err != nil {
		verboseLog("browser open failed: %v (user must navigate manually)", err)
	}

	fmt.Printf("Waiting for callback on %s ...\n", oauth.RedirectURI)

	select {
	case result := <-ch:
		if result.Err != nil {
			fatalf("authentication failed", result.Err, "Check the browser for error details")
		}
		verboseLog("received code, exchanging for token")
		tok, err := oauth.ExchangeCode(ctx, tokenURL, result.Code, pkce)
		if err != nil {
			fatalf("token exchange failed", err, "Check the token endpoint or your credentials")
		}
		stored := &oauth.StoredToken{
			AccessToken:  tok.AccessToken,
			RefreshToken: tok.RefreshToken,
			Scope:        tok.Scope,
			TokenURL:     tokenURL,
		}
		if tok.ExpiresIn > 0 {
			stored.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		}
		if err := oauth.SaveToken(serviceURL, stored); err != nil {
			fatalf("failed to save token", err, "Check ~/.config/a2acli/tokens/ permissions")
		}
		fmt.Printf("Authenticated. Token stored for %s\n", serviceURL)
		if !stored.ExpiresAt.IsZero() {
			fmt.Printf("Expires: %s\n", stored.ExpiresAt.Format(time.RFC3339))
		}

	case <-time.After(5 * time.Minute):
		fatalf("authentication timed out", fmt.Errorf("no callback after 5 minutes"),
			"Restart 'a2acli auth login' and complete the browser flow within 5 minutes")
	}
}

func runClientCredentials(ctx context.Context, tokenURL string) {
	body := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		authClientID, authClientSecret)
	_ = body // exchange via ExchangeCode variant — TODO: implement client_credentials exchange
	fmt.Printf("client credentials flow not yet implemented — use --token $(make token) for now\n")
	_ = ctx
}

func runAuthStatus(_ *cobra.Command, _ []string) {
	tok, err := oauth.LoadToken(serviceURL)
	if err != nil {
		fatalf("failed to load token", err, "")
	}
	if tok == nil {
		fmt.Printf("No token stored for %s\n", serviceURL)
		fmt.Printf("Hint: run 'a2acli auth login --service-url %s'\n", serviceURL)
		return
	}

	if disableTUI {
		b, _ := json.MarshalIndent(map[string]any{
			"service_url":   serviceURL,
			"has_token":     tok.AccessToken != "",
			"expires_at":    tok.ExpiresAt,
			"expired":       tok.IsExpired(),
			"has_refresh":   tok.RefreshToken != "",
			"scope":         tok.Scope,
		}, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("Token for %s:\n", serviceURL)
	if tok.IsExpired() {
		fmt.Printf("  Status:  EXPIRED\n")
	} else if tok.ExpiresAt.IsZero() {
		fmt.Printf("  Status:  valid (no expiry)\n")
	} else {
		fmt.Printf("  Status:  valid (expires %s)\n", tok.ExpiresAt.Format(time.RFC3339))
	}
	fmt.Printf("  Scope:   %s\n", tok.Scope)
	fmt.Printf("  Refresh: %v\n", tok.RefreshToken != "")
}

func runAuthLogout(_ *cobra.Command, _ []string) {
	if err := oauth.DeleteToken(serviceURL); err != nil {
		fatalf("failed to delete token", err, "")
	}
	fmt.Printf("Token deleted for %s\n", serviceURL)
}

func runAuthToken(_ *cobra.Command, _ []string) {
	tok, err := oauth.LoadToken(serviceURL)
	if err != nil {
		fatalf("failed to load token", err, "")
	}
	if tok == nil || tok.AccessToken == "" {
		fatalf("no token stored", fmt.Errorf("serviceURL=%s", serviceURL),
			"Run 'a2acli auth login --service-url "+serviceURL+"'")
	}
	if tok.IsExpired() {
		fatalf("token is expired", fmt.Errorf("expiredAt=%s", tok.ExpiresAt),
			"Run 'a2acli auth login --service-url "+serviceURL+"' to refresh")
	}
	fmt.Print(tok.AccessToken)
}
