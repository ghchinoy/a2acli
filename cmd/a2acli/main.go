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
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
	"github.com/a2aproject/a2a-go/v2/a2acompat/a2av0"
	a2agrpc "github.com/a2aproject/a2a-go/v2/a2agrpc/v1"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/a2acli/internal/oauth"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	serviceURL       string
	skillID          string
	authToken        string
	targetTaskID     string
	refTaskID        string
	contextID        string
	strictMode       bool
	outDir           string
	outFile          string
	instructionFile  string
	disableTUI       bool
	outputMode       string
	requestTimeout   time.Duration
	wait             bool
	immediate        bool
	verbose          bool
	showFull         bool
	discoverExtended bool
	noCache          bool
	transport        string
	protocol         string
	authHeaders      []string
	svcParams        []string

	rootCmd = &cobra.Command{
		Use:   "a2acli",
		Short: "A2A CLI Client",
	}

	// Command group IDs for help organization
	GroupDiscovery = "discovery"
	GroupMessaging = "messaging"
	GroupSystem    = "system"
	GroupServer    = "server"
)

// Standard machine-readable error codes emitted in non-interactive JSON mode.
const (
	ErrCodeUnauthenticated    = "UNAUTHENTICATED"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeTaskFailed         = "TASK_FAILED"
	ErrCodeInvalidArgument    = "INVALID_ARGUMENT"
	ErrCodeNotFound           = "NOT_FOUND"
	ErrCodeInternal           = "INTERNAL_ERROR"
	ErrCodeFailedPrecondition = "FAILED_PRECONDITION"
)

func fatalf(format string, err error, hint string) {
	fatalCode("", format, err, hint)
}

func fatalCode(code string, format string, err error, hint string) {
	if outputMode == "" {
		resolveOutputMode()
	}

	msg := fmt.Sprintf(format+": %v", err)
	if err == nil {
		msg = format
	}

	if code == "" {
		if is401(err) {
			code = ErrCodeUnauthenticated
		} else if isTimeout(err) {
			code = ErrCodeTimeout
		} else {
			code = ErrCodeInternal
		}
	}

	if outputMode == "json" || disableTUI {
		out := map[string]string{
			"error": msg,
			"code":  code,
		}
		if hint != "" {
			out["hint"] = hint
		}
		b, _ := json.Marshal(out)
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
		}
	}
	os.Exit(1)
}

// is401 reports whether an error is an HTTP 401 Unauthorized response.
func is401(err error) bool {
	return err != nil && strings.Contains(err.Error(), "401")
}

// isTimeout reports whether an error indicates a request or context timeout.
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "timeout") || strings.Contains(s, "deadline exceeded")
}

// authHintFromCard generates an actionable authentication hint based on the
// AgentCard's declared security schemes. Called when a protocol call returns 401.
func authHintFromCard(card *a2a.AgentCard) string {
	if card == nil || len(card.SecuritySchemes) == 0 {
		return "Check your --token or --auth flags"
	}
	for name, scheme := range card.SecuritySchemes {
		switch s := scheme.(type) {
		case a2a.OAuth2SecurityScheme:
			_ = s
			// Check if a stored token exists but may be expired.
			if stored, err := oauth.LoadValidToken(context.Background(), serviceURL); err == nil && stored != nil {
				if stored.IsExpired() {
					return fmt.Sprintf("Stored token for %s is expired. Run: a2acli auth login -u %s", serviceURL, serviceURL)
				}
				return fmt.Sprintf("OAuth token present but rejected (%s). Run: a2acli auth login -u %s to re-authenticate", name, serviceURL)
			}
			return fmt.Sprintf("This agent requires OAuth 2.1 authentication (%s). Run: a2acli auth login -u %s", name, serviceURL)
		case a2a.HTTPAuthSecurityScheme:
			return fmt.Sprintf("This agent requires %s authentication (%s). Pass via --token <value>", s.Scheme, name)
		case a2a.APIKeySecurityScheme:
			return fmt.Sprintf("This agent requires an API key (%s). Pass via --auth \"%s: <key>\"", name, s.Name)
		}
	}
	return "Check your --token or --auth flags"
}

// verboseLog writes a diagnostic line to stderr when --verbose is active.
// Output always goes to stderr so it never pollutes --output json stdout.
func verboseLog(format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "[verbose] "+format+"\n", args...)
}

func init() {
	// A2A SDK v0 and v1 packages both register a2a.proto which causes a panic
	// if not ignored.
	_ = os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "ignore")

	rootCmd.AddGroup(
		&cobra.Group{ID: GroupDiscovery, Title: "Discovery & Identity:"},
		&cobra.Group{ID: GroupMessaging, Title: "Messaging & Tasks:"},
		&cobra.Group{ID: GroupServer, Title: "Server & Mocking:"},
		&cobra.Group{ID: GroupSystem, Title: "Client Configuration:"},
	)
	rootCmd.SetHelpFunc(colorizedHelpFunc)
}

type paramInterceptor struct {
	a2aclient.PassthroughInterceptor
	token       string
	authHeaders []string
	svcParams   []string
}

func (i *paramInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, any, error) {
	if i.token != "" || len(i.authHeaders) > 0 || len(i.svcParams) > 0 {
		if req.ServiceParams == nil {
			req.ServiceParams = make(a2aclient.ServiceParams)
		}
		if i.token != "" {
			req.ServiceParams["authorization"] = append(req.ServiceParams["authorization"], "Bearer "+i.token)
		}
		if len(i.authHeaders) > 0 {
			req.ServiceParams["authorization"] = append(req.ServiceParams["authorization"], i.authHeaders...)
		}
		for _, param := range i.svcParams {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				req.ServiceParams[parts[0]] = append(req.ServiceParams[parts[0]], parts[1])
			} else {
				req.ServiceParams[param] = append(req.ServiceParams[param], "")
			}
		}
	}
	return ctx, nil, nil
}

func getResolver() *agentcard.Resolver {
	t := requestTimeout
	if t == 0 {
		t = 30 * time.Second
	}
	verboseLog("resolving agent card from %s (timeout: %s)", serviceURL, t)
	if protocol == "0.3.0" || strings.HasPrefix(protocol, "0.3") {
		return &agentcard.Resolver{
			Client:     &http.Client{Timeout: t},
			CardParser: a2av0.NewAgentCardParser(),
		}
	}
	return &agentcard.Resolver{Client: &http.Client{Timeout: t}}
}

func resolveAgentCard(ctx context.Context, targetURL string) (*a2a.AgentCard, error) {
	if !noCache {
		if cached, err := loadCachedCard(targetURL); err == nil && cached != nil {
			verboseLog("using cached AgentCard for %s (fetched %s ago)", targetURL, time.Since(cached.FetchedAt).Round(time.Second))
			return cached.Card, nil
		}
	}

	var opts []agentcard.ResolveOption
	if authToken != "" {
		opts = append(opts, agentcard.WithRequestHeader("Authorization", "Bearer "+authToken))
	}
	for _, h := range authHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			opts = append(opts, agentcard.WithRequestHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
		} else if len(parts) == 1 && strings.HasPrefix(strings.ToLower(parts[0]), "bearer ") {
			opts = append(opts, agentcard.WithRequestHeader("Authorization", strings.TrimSpace(parts[0])))
		}
	}

	card, err := getResolver().Resolve(ctx, targetURL, opts...)
	if err != nil {
		return nil, err
	}

	if err := saveCachedCard(targetURL, card); err != nil {
		verboseLog("failed to save AgentCard to disk cache: %v", err)
	}

	return card, nil
}

func createClient(ctx context.Context, card *a2a.AgentCard) (*a2aclient.Client, error) {
	httpClient := &http.Client{Timeout: 15 * time.Minute}

	// Determine transport
	selectedTransport := a2a.TransportProtocolJSONRPC // Default
	if transport != "" {
		switch strings.ToLower(transport) {
		case "grpc":
			selectedTransport = a2a.TransportProtocolGRPC
		case "jsonrpc":
			selectedTransport = a2a.TransportProtocolJSONRPC
		case "rest", "httpjson":
			selectedTransport = a2a.TransportProtocolHTTPJSON
		default:
			return nil, fmt.Errorf("unsupported transport: %s", transport)
		}
	} else {
		// Dynamic selection based on priority: gRPC > JSON-RPC > HTTP+JSON
		available := make(map[a2a.TransportProtocol]bool)
		for _, iface := range card.SupportedInterfaces {
			available[iface.ProtocolBinding] = true
		}

		if available[a2a.TransportProtocolGRPC] {
			selectedTransport = a2a.TransportProtocolGRPC
		} else if available[a2a.TransportProtocolJSONRPC] {
			selectedTransport = a2a.TransportProtocolJSONRPC
		} else if available[a2a.TransportProtocolHTTPJSON] {
			selectedTransport = a2a.TransportProtocolHTTPJSON
		}
	}

	var transportOpt a2aclient.FactoryOption
	switch selectedTransport {
	case a2a.TransportProtocolGRPC:
		if protocol == "0.3.0" || strings.HasPrefix(protocol, "0.3") {
			return nil, fmt.Errorf("A2A 0.3.0 gRPC transport is not supported in this CLI build to prevent protobuf conflicts")
		}
		transportOpt = a2agrpc.WithGRPCTransport()
	case a2a.TransportProtocolHTTPJSON:
		if protocol == "0.3.0" || strings.HasPrefix(protocol, "0.3") {
			return nil, fmt.Errorf("A2A 0.3.0 does not support REST transport in this CLI")
		}
		transportOpt = a2aclient.WithRESTTransport(httpClient)
	default:
		if protocol == "0.3.0" || strings.HasPrefix(protocol, "0.3") {
			transportOpt = a2aclient.WithCompatTransport("0.3.0", a2a.TransportProtocolJSONRPC, a2av0.NewJSONRPCTransportFactory(a2av0.JSONRPCTransportConfig{Client: httpClient}))
		} else {
			transportOpt = a2aclient.WithJSONRPCTransport(httpClient)
		}
	}

	if transport == "" {
		verboseLog("auto-selected transport: %s", selectedTransport)
		if outputMode == "tui" && verbose {
			fmt.Printf("Auto-selected transport: %s\n", StyleAccent.Render(string(selectedTransport)))
		}
	} else {
		verboseLog("forcing transport: %s", selectedTransport)
		if outputMode == "tui" && verbose {
			fmt.Printf("Forcing transport: %s\n", StyleAccent.Render(string(selectedTransport)))
		}
	}

	// Auto-use stored OAuth token when no explicit --token is given.
	resolvedToken := authToken
	if resolvedToken == "" {
		if stored, err := oauth.LoadValidToken(ctx, serviceURL); err == nil && stored != nil && !stored.IsExpired() {
			resolvedToken = stored.AccessToken
			verboseLog("using stored OAuth token for %s (expires %s)", serviceURL, stored.ExpiresAt.Format("15:04:05"))
		}
	}

	opts := []a2aclient.FactoryOption{transportOpt}
	if resolvedToken != "" || len(authHeaders) > 0 || len(svcParams) > 0 {
		opts = append(opts, a2aclient.WithCallInterceptors(&paramInterceptor{
			token:       resolvedToken,
			authHeaders: authHeaders,
			svcParams:   svcParams,
		}))
	}
	return a2aclient.NewFromCard(ctx, card, opts...)
}

// isTTY reports whether stdout is an interactive terminal.
// Used to decide whether to render the Bubble Tea TUI.
func isTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// isStdinPiped reports whether stdin is being piped (not a terminal).
// Used to decide whether 'send' can read its message from stdin.
// This is intentionally separate from isTTY: when running
//
//	echo "msg" | a2acli send --env mithlond --wait
//
// stdout is still a terminal (isTTY returns true) but stdin is a pipe.
func isStdinPiped() bool {
	return !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// resolveOutputMode determines the effective output mode from flags and env vars.
// Priority: --output flag > -n/--no-tui > A2ACLI_NO_TUI env > NO_COLOR env > no-TTY > default (tui)
func resolveOutputMode() {
	switch outputMode {
	case "tui", "text", "json", "compact":
		// explicit --output value is valid; honour it even in a non-TTY context
		// (the user knows what they asked for)
	case "":
		// not explicitly set — derive from signals, most-specific first
		if disableTUI {
			outputMode = "json"
		} else if os.Getenv("A2ACLI_NO_TUI") == "true" {
			outputMode = "json"
		} else if os.Getenv("NO_COLOR") != "" {
			outputMode = "text"
		} else if os.Getenv("CI") != "" {
			// Standard CI env var: degrade to text so streaming still works
			outputMode = "text"
		} else if !isTTY() {
			// No interactive terminal: degrade to text so streaming events
			// print line-by-line rather than aborting with "could not open a new TTY"
			outputMode = "text"
			verboseLog("no TTY detected — degrading from tui to text mode")
		} else {
			outputMode = "tui"
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid --output value %q (must be tui, text, json, or compact)\n", outputMode)
		os.Exit(1)
	}
	// Sync disableTUI for any existing code that checks it directly.
	// text mode and tui mode both leave disableTUI=false; only json sets it true.
	disableTUI = (outputMode == "json")
	// Also honour A2ACLI_VERBOSE env var.
	if os.Getenv("A2ACLI_VERBOSE") == "true" {
		verbose = true
	}
	verboseLog("output mode: %s, protocol: %s, transport: %q, timeout: %s",
		outputMode, protocol, transport, requestTimeout)
}

// runText prints a human-readable stream of events to stdout without the Bubble Tea TUI.
// Used when --output text is set.
func runText(stream chan streamMsg, outDir string) (streamSummary, error) {
	var summary streamSummary
	for msg := range stream {
		if msg.Err != nil {
			return summary, msg.Err
		}
		summary.events++
		if msg.Event != nil {
			info := msg.Event.TaskInfo()
			if info.TaskID != "" {
				summary.taskID = string(info.TaskID)
			}
			if info.ContextID != "" {
				summary.contextID = info.ContextID
			}
		}

		switch e := msg.Event.(type) {
		case *a2a.TaskStatusUpdateEvent:
			verboseLog("event: TaskStatusUpdate state=%s", e.Status.State)
			fmt.Printf("Status: %s\n", e.Status.State)
		case *a2a.TaskArtifactUpdateEvent:
			verboseLog("event: TaskArtifactUpdate artifact=%q append=%v lastChunk=%v",
				e.Artifact.Name, e.Append, e.LastChunk)
			fmt.Printf("Artifact: %s\n", e.Artifact.Name)
			for _, p := range e.Artifact.Parts {
				if tp, ok := p.Content.(a2a.Text); ok {
					fmt.Println(string(tp))
				}
			}
			if outDir != "" || outFile != "" {
				_, _ = saveArtifact(outDir, outFile, *e.Artifact, 0)
			}
		}
	}

	printContinuationFooter(summary.taskID, summary.contextID)
	return summary, nil
}

func fetchExtendedCard(ctx context.Context, card *a2a.AgentCard) *a2a.AgentCard {
	verboseLog("requesting extended AgentCard")
	if !card.Capabilities.ExtendedAgentCard {
		fatalf("extended card not supported", fmt.Errorf("serviceURL=%s", serviceURL),
			"The AgentCard does not advertise ExtendedAgentCard: true in its capabilities.")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	extCard, err := client.GetExtendedAgentCard(ctx, &a2a.GetExtendedAgentCardRequest{})
	if err != nil {
		hint := "Verify your credentials or server availability"
		if is401(err) {
			hint = "Requesting the extended card requires authentication. Run 'a2acli auth login' or pass --token"
		}
		fatalf("GetExtendedAgentCard failed", err, hint)
	}
	verboseLog("resolved extended AgentCard: name=%q version=%q skills=%d",
		extCard.Name, extCard.Version, len(extCard.Skills))
	return extCard
}

func printSecurityScheme(name string, scheme a2a.SecurityScheme) {
	switch s := scheme.(type) {
	case a2a.HTTPAuthSecurityScheme:
		label := s.Scheme
		if s.BearerFormat != "" {
			label += " (" + s.BearerFormat + ")"
		}
		fmt.Printf("  %s: http/%s\n", name, label)
		fmt.Printf("    Hint: pass via --token <value>\n")
		verboseLog("security scheme %q: http scheme=%s bearerFormat=%s", name, s.Scheme, s.BearerFormat)
	case a2a.OAuth2SecurityScheme:
		fmt.Printf("  %s: oauth2\n", name)
		if s.Oauth2MetadataURL != "" {
			fmt.Printf("    Metadata URL: %s\n", s.Oauth2MetadataURL)
		}
		switch f := s.Flows.(type) {
		case a2a.AuthorizationCodeOAuthFlow:
			fmt.Printf("    Flow:         authorization_code\n")
			if f.TokenURL != "" {
				fmt.Printf("    Token URL:    %s\n", f.TokenURL)
			}
			if f.AuthorizationURL != "" {
				fmt.Printf("    Auth URL:     %s\n", f.AuthorizationURL)
			}
		case a2a.ClientCredentialsOAuthFlow:
			fmt.Printf("    Flow:         client_credentials\n")
			if f.TokenURL != "" {
				fmt.Printf("    Token URL:    %s\n", f.TokenURL)
			}
		case a2a.DeviceCodeOAuthFlow:
			fmt.Printf("    Flow:         device_code\n")
			if f.TokenURL != "" {
				fmt.Printf("    Token URL:    %s\n", f.TokenURL)
			}
		}
		fmt.Printf("    Hint: run 'a2acli auth login -u %s' or pass via --token <jwt>\n", serviceURL)
		verboseLog("security scheme %q: oauth2 metadataURL=%s flows=%T", name, s.Oauth2MetadataURL, s.Flows)
	case a2a.APIKeySecurityScheme:
		fmt.Printf("  %s: apiKey in %s (header: %s)\n", name, s.Location, s.Name)
		fmt.Printf("    Hint: pass via --auth \"%s: <key>\"\n", s.Name)
		verboseLog("security scheme %q: apiKey location=%s name=%s", name, s.Location, s.Name)
	case a2a.OpenIDConnectSecurityScheme:
		fmt.Printf("  %s: openIdConnect\n", name)
		if s.OpenIDConnectURL != "" {
			fmt.Printf("    Discovery: %s\n", s.OpenIDConnectURL)
		}
		verboseLog("security scheme %q: openIdConnect url=%s", name, s.OpenIDConnectURL)
	case a2a.MutualTLSSecurityScheme:
		fmt.Printf("  %s: mutualTLS\n", name)
		verboseLog("security scheme %q: mutualTLS", name)
	default:
		fmt.Printf("  %s: (unrecognised scheme type %T)\n", name, scheme)
	}
}

func printSkills(skills []a2a.AgentSkill) {
	if len(skills) == 0 {
		return
	}
	fmt.Printf("\nSkills:\n")
	for _, s := range skills {
		fmt.Printf("  - [%s] %s\n", s.ID, s.Name)
		if s.Description != "" {
			fmt.Printf("    Description: %s\n", s.Description)
		}
		if len(s.SecurityRequirements) > 0 {
			for _, req := range s.SecurityRequirements {
				for schemeName, scopes := range req {
					if len(scopes) > 0 {
						fmt.Printf("    Security: %s [scopes: %s]\n", schemeName, strings.Join(scopes, ", "))
					} else {
						fmt.Printf("    Security: %s\n", schemeName)
					}
				}
			}
		}
	}
}

func runDescribe(_ *cobra.Command, args []string) {
	ctx := context.Background()

	if len(args) > 0 && args[0] != "" {
		serviceURL = args[0]
	}

	card, err := resolveAgentCard(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Ensure the A2A server is running at "+serviceURL+" or specify --service-url / -u")
	}
	verboseLog("resolved AgentCard: name=%q version=%q skills=%d interfaces=%d",
		card.Name, card.Version, len(card.Skills), len(card.SupportedInterfaces))

	if discoverExtended {
		card = fetchExtendedCard(ctx, card)
	}

	if disableTUI {
		b, err := json.MarshalIndent(card, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Agent: %s\n", card.Name)
	if card.Version != "" {
		fmt.Printf("Version: %s\n", card.Version)
	}
	if card.Description != "" {
		fmt.Printf("Description: %s\n", card.Description)
	}

	var formats []string
	seenFormats := make(map[string]bool)
	for _, iface := range card.SupportedInterfaces {
		b := string(iface.ProtocolBinding)
		if b != "" && !seenFormats[b] {
			seenFormats[b] = true
			formats = append(formats, b)
		}
	}
	if len(formats) > 0 {
		fmt.Printf("Supported Bindings: %s\n", strings.Join(formats, ", "))
	}

	fmt.Printf("Capabilities: [Streaming: %v]\n", card.Capabilities.Streaming)

	if len(card.SecuritySchemes) > 0 {
		fmt.Printf("\nSecurity Schemes:\n")
		for name, scheme := range card.SecuritySchemes {
			printSecurityScheme(string(name), scheme)
		}
	}

	printSkills(card.Skills)
}

func handleSendImmediate(ctx context.Context, client *a2aclient.Client, params *a2a.SendMessageRequest, card *a2a.AgentCard) {
	params.Config = &a2a.SendMessageConfig{ReturnImmediately: true}
	result, err := client.SendMessage(ctx, params)
	if err != nil {
		hint := "Check service connectivity or skill availability"
		if is401(err) {
			hint = authHintFromCard(card)
		}
		fatalf("SendMessage failed", err, hint)
	}
	if outputMode == "json" {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
	} else if task, ok := result.(*a2a.Task); ok {
		fmt.Printf("Task submitted: %s\n", task.ID)
		if task.ContextID != "" {
			fmt.Printf("Context ID:     %s\n", task.ContextID)
		}
		fmt.Printf("\nUse 'a2acli subscribe %s' to follow progress\n", task.ID)
		if task.ContextID != "" {
			fmt.Printf("Use 'a2acli send --context %s \"...\"' to continue this conversation\n", task.ContextID)
		}
	}
}

func handleSendWait(ctx context.Context, client *a2aclient.Client, params *a2a.SendMessageRequest, card *a2a.AgentCard) {
	params.Config = &a2a.SendMessageConfig{ReturnImmediately: false}
	if outputMode == "tui" {
		fmt.Printf("Invoking A2A Service (Blocking)...\n\n")
	}
	result, err := client.SendMessage(ctx, params)
	if err != nil {
		hint := "Check service connectivity or skill availability"
		if is401(err) {
			hint = authHintFromCard(card)
		}
		fatalf("SendMessage failed", err, hint)
	}
	if outputMode == "json" {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
		if task, ok := result.(*a2a.Task); ok && (outDir != "" || outFile != "") {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}
	if task, ok := result.(*a2a.Task); ok {
		displayTaskResult(task, outDir)
		printContinuationFooter(string(task.ID), task.ContextID)
	} else if msg, ok := result.(*a2a.Message); ok {
		fmt.Printf("Received simple message from agent (Task ID: %s)\n", msg.TaskID)
		for _, p := range msg.Parts {
			if tp, ok := p.Content.(a2a.Text); ok {
				fmt.Printf("Agent: %s\n", string(tp))
			}
		}
		printContinuationFooter(string(msg.TaskID), msg.ContextID)
	} else {
		fmt.Printf("Unknown result type received: %T\n", result)
	}
}

func runSend(_ *cobra.Command, args []string) {
	if err := validateOutDir(outDir); err != nil {
		fatalCode(ErrCodeInvalidArgument, "invalid --out-dir / -d argument", err, "Use -o or --output to set output format (tui/text/json)")
	}

	var messageText string
	if len(args) == 0 {
		// No positional arg — read from stdin (only reachable when stdin is not a TTY)
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fatalf("failed to read message from stdin", err, "Ensure stdin is readable or provide the message as an argument")
		}
		messageText = strings.TrimRight(string(data), "\r\n")
		verboseLog("read message from stdin: %d bytes", len(messageText))
	} else {
		messageText = args[0]
	}

	if instructionFile != "" {
		content, err := os.ReadFile(instructionFile)
		if err != nil {
			fatalf("failed to read instruction file %q", err, "Verify the file path exists and is readable")
		}
		messageText = fmt.Sprintf("%s\n\nSupplemental Instructions:\n%s", messageText, string(content))
	}

	ctx := context.Background()

	card, err := resolveAgentCard(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	msg, err := buildMessage(messageText)
	if err != nil {
		fatalf("failed to build message", err, "Check --json/--parts/--attach/--data flags for valid input")
	}

	if targetTaskID != "" {
		msg.TaskID = a2a.TaskID(targetTaskID)
		verboseLog("continuing task: %s", targetTaskID)

		// Pre-flight GetTask check to validate task state and auto-populate context ID
		t, getErr := client.GetTask(ctx, &a2a.GetTaskRequest{ID: a2a.TaskID(targetTaskID)})
		if getErr == nil && t != nil {
			if contextID == "" && t.ContextID != "" {
				contextID = t.ContextID
			}
			if checkErr := checkTaskContinuable(t); checkErr != nil {
				if strictMode {
					fatalCode(ErrCodeFailedPrecondition, "cannot continue task", checkErr,
						fmt.Sprintf("Use --context %s to continue the conversation thread", t.ContextID))
				} else {
					fmt.Fprintf(os.Stderr, "Warning: %v. Use --context %s to continue the conversation thread.\n\n", checkErr, t.ContextID)
				}
			}
		}

		if !disableTUI {
			fmt.Printf("Continuing Task: %s\n", targetTaskID)
		}
	}

	if contextID != "" {
		msg.ContextID = contextID
		verboseLog("setting context ID: %s", contextID)
		if !disableTUI {
			fmt.Printf("Context ID: %s\n", contextID)
		}
	}

	if refTaskID != "" {
		msg.ReferenceTasks = []a2a.TaskID{a2a.TaskID(refTaskID)}
		verboseLog("referencing task: %s", refTaskID)
		if !disableTUI {
			fmt.Printf("Referencing Task: %s\n", refTaskID)
		}
	}

	params := &a2a.SendMessageRequest{
		Message: msg,
	}
	if skillID != "" {
		params.Metadata = map[string]any{"skillId": skillID}
		verboseLog("targeting skill: %s", skillID)
	}
	verboseLog("sending message: text_len=%d task=%q context=%q ref=%q immediate=%v wait=%v",
		len(messageText), targetTaskID, contextID, refTaskID, immediate, wait)

	// --immediate: fire-and-forget — return the task ID without waiting or streaming
	if immediate {
		handleSendImmediate(ctx, client, params, card)
		return
	}

	// --wait: blocking call
	if wait {
		handleSendWait(ctx, client, params, card)
		return
	}

	// Default: streaming
	if outputMode == "tui" {
		fmt.Printf("Invoking A2A Service (Streaming)...\n\n")
	}

	stream := make(chan streamMsg)
	go func() {
		defer close(stream)
		for event, err := range client.SendStreamingMessage(ctx, params) {
			stream <- streamMsg{Event: event, Err: err}
			if err != nil {
				return
			}
		}
	}()

	var summary streamSummary
	var renderErr error

	switch outputMode {
	case "json":
		summary, renderErr = runRaw(stream, outDir)
	case "text":
		summary, renderErr = runText(stream, outDir)
	case "compact":
		summary, renderErr = runCompact(stream, outDir)
	default:
		summary, renderErr = runTUI(stream)
	}

	if renderErr != nil {
		fatalCode(ErrCodeFailedPrecondition, "streaming failed", renderErr, "Ensure the service is accessible and the task is active")
	}

	if summary.events == 0 {
		fatalCode(ErrCodeFailedPrecondition, "agent returned no events", fmt.Errorf("stream closed with 0 events received"),
			"The agent did not emit any status or artifact events. Verify that the task is active or use --context for multi-turn conversations.")
	}
}

func runWatch(_ *cobra.Command, args []string) {
	if err := validateOutDir(outDir); err != nil {
		fatalCode(ErrCodeInvalidArgument, "invalid --out-dir / -d argument", err, "Use -o or --output to set output format (tui/text/json)")
	}

	taskID := args[0]
	ctx := context.Background()

	card, err := resolveAgentCard(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	if outputMode == "tui" {
		fmt.Printf("Subscribing to Task %s ...\n\n", taskID)
	}

	tid := a2a.TaskID(taskID)

	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		fatalf("failed to retrieve task status", err, "If using an in-memory store, task history is lost on server restart")
	}

	if task.Status.State == a2a.TaskStateCompleted || task.Status.State == a2a.TaskStateFailed || task.Status.State == a2a.TaskStateRejected {
		displayTaskResult(task, outDir)
		return
	}

	if outputMode == "tui" {
		fmt.Println("Task is active. Connecting to stream...")
	}

	stream := make(chan streamMsg)
	go func() {
		defer close(stream)
		for event, err := range client.SubscribeToTask(ctx, &a2a.SubscribeToTaskRequest{ID: tid}) {
			stream <- streamMsg{Event: event, Err: err}
			if err != nil {
				return
			}
		}
	}()

	switch outputMode {
	case "json":
		runRaw(stream, outDir)
	case "text":
		runText(stream, outDir)
	default:
		runTUI(stream)
	}
}

func runGet(cmd *cobra.Command, args []string) {
	if err := validateOutDir(outDir); err != nil {
		fatalCode(ErrCodeInvalidArgument, "invalid --out-dir / -d argument", err, "Use -o or --output to set output format (tui/text/json)")
	}

	taskID := args[0]
	ctx := context.Background()
	verboseLog("GetTask: %s", taskID)

	card, err := resolveAgentCard(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	tid := a2a.TaskID(taskID)

	// Default outDir for 'download' command if neither is specified
	if cmd.Name() == "download" || cmd.Name() == "retrieve" {
		if outDir == "" && outFile == "" {
			outDir = "."
		}
	}

	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		hint := "Check the task ID or verify the server state"
		if is401(err) {
			hint = authHintFromCard(card)
		}
		fatalf("failed to retrieve task", err, hint)
	}
	verboseLog("GetTask response: state=%s artifacts=%d", task.Status.State, len(task.Artifacts))

	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		if outDir != "" || outFile != "" {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}

	// Always display the full result (which handles saving now!)
	displayTaskResult(task, outDir)
}

func runCancel(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()
	verboseLog("CancelTask: %s", taskID)

	card, err := resolveAgentCard(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	tid := a2a.TaskID(taskID)

	task, err := client.CancelTask(ctx, &a2a.CancelTaskRequest{ID: tid})
	if err != nil {
		hint := "Check the task ID or verify the server state"
		if is401(err) {
			hint = authHintFromCard(card)
		}
		fatalf("failed to cancel task", err, hint)
	}
	verboseLog("CancelTask response: state=%s", task.Status.State)

	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Task %s has been requested to cancel. Current state: %s\n", task.ID, task.Status.State)
}

func main() {
	cobra.OnInitialize(initConfig, resolveOutputMode)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/a2acli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&envName, "env", "e", "", "environment name to load from config")
	rootCmd.PersistentFlags().StringVarP(&serviceURL, "service-url", "u", "http://127.0.0.1:9001", "Base URL of the A2A service")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Auth token")
	rootCmd.PersistentFlags().StringSliceVar(&authHeaders, "auth", nil, "Authorization headers to send (e.g. 'Bearer ...')")
	rootCmd.PersistentFlags().StringSliceVar(&svcParams, "svc-param", nil, "Service parameters to send (e.g. 'key=value')")
	rootCmd.PersistentFlags().StringVarP(&targetTaskID, "task", "k", "", "Existing Task ID to continue (for active tasks)")
	rootCmd.PersistentFlags().StringVar(&contextID, "context", "", "Context ID for multi-turn conversation thread")
	rootCmd.PersistentFlags().StringVarP(&refTaskID, "ref", "r", "", "Task ID to reference for cross-task artifact chaining (does not continue conversation)")
	rootCmd.PersistentFlags().BoolVar(&strictMode, "strict", false, "Fail fast on warnings (e.g. continuing terminal tasks)")
	rootCmd.PersistentFlags().BoolVar(&noCache, "no-cache", false, "Bypass agent card disk cache and fetch fresh")
	rootCmd.PersistentFlags().BoolVarP(&disableTUI, "no-tui", "n", false, "Disable the Terminal UI — alias for --output json (backwards compat)")
	rootCmd.PersistentFlags().StringVarP(&outputMode, "output", "o", "", "Output mode: tui (default), text (plain, no animations), json (NDJSON for scripting)")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "timeout", 0, "Request timeout, e.g. 30s, 2m (0 = no timeout)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print diagnostic info to stderr (also: A2ACLI_VERBOSE=true)")
	rootCmd.PersistentFlags().StringVar(&transport, "transport", "", "Force a specific transport protocol (grpc, jsonrpc, rest)")
	rootCmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "1.0.0", "A2A protocol version (1.0.0 or 0.3.0)")
	rootCmd.Flags().BoolP("version", "V", false, "Print version information")

	var describeCmd = &cobra.Command{
		Use:     "discover [url]",
		Aliases: []string{"describe"},
		GroupID: GroupDiscovery,
		Short:   "Discover and display the agent card",
		Long: `Retrieve and display the A2A AgentCard for the target service.

The AgentCard contains the agent's identity, description, supported 
interface protocols (e.g., JSON-RPC), and available skills. It also 
lists any security requirements for each skill.

The agent URL can be passed as an optional positional argument or via --service-url / -u.

Use --extended to fetch the richer, authenticated AgentCard via the
GetExtendedAgentCard protocol RPC.

'describe' is accepted as a backwards-compatible alias.`,
		Example: `  a2acli discover
  a2acli discover http://localhost:9001
  a2acli discover --extended
  a2acli discover --service-url http://localhost:9001 --extended`,
		Args: cobra.MaximumNArgs(1),
		Run:  runDescribe,
	}
	describeCmd.Flags().BoolVar(&discoverExtended, "extended", false, "Fetch the authenticated extended AgentCard")

	var sendCmd = &cobra.Command{
		Use:     "send [message]",
		GroupID: GroupMessaging,
		Aliases: []string{"invoke", "SendMessage"},
		Short:   "Send a message to an agent (streaming)",
		Long: `Initiate a new task or continue an existing one by sending a message to the agent.

By default, this command uses streaming to provide real-time updates from 
the agent. Use the --wait flag to perform a blocking call instead.

You can save artifacts produced by the task using the --out-dir flag.`,
		Example: `  a2acli send "Write a simple CLI in Go"
  a2acli send "Add error handling to that CLI" --context <contextID>
  a2acli send "Summarize this report" --skill summarize --ref <taskID>
  a2acli send "Generate report" --skill reports --wait --out-dir ./reports`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 && !isStdinPiped() && !hasMultimodalInput() {
				return fmt.Errorf("message text required: provide as argument, pipe via stdin, or use --json/--parts/--attach/--data")
			}
			if len(args) > 1 {
				return fmt.Errorf("accepts at most 1 arg, received %d", len(args))
			}
			return nil
		},
		Run: runSend,
	}

	var watchCmd = &cobra.Command{
		Use:     "subscribe [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"watch", "resume", "SubscribeToTask"},
		Short:   "Subscribe to an active task's streaming updates",
		Long: `Connect to an active task's event stream to receive real-time updates.

This is useful for resuming observation of a long-running task or 
watching a task initiated by another client. If the task is 
already completed, the command will display the final results.

'watch' is accepted as a backwards-compatible alias.`,
		Example: `  a2acli subscribe <taskID>
  a2acli subscribe <taskID> --output json
  a2acli subscribe <taskID> --out-dir ./artifacts`,
		Args: cobra.ExactArgs(1),
		Run:  runWatch,
	}

	var getCmd = &cobra.Command{
		Use:     "get [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"status", "GetTask"},
		Short:   "Get the status of a task",
		Long: `Retrieve the current state and results of a task.

Displays the task status (e.g., active, completed, failed) and a 
preview of any artifacts produced. Use the --out-dir flag to 
download artifacts to a directory.`,
		Example: `  a2acli get <taskID>
  a2acli get <taskID> --no-tui
  a2acli get <taskID> --out-dir ./status`,
		Args: cobra.ExactArgs(1),
		Run:  runGet,
	}

	var versionCmd = &cobra.Command{
		Use:     "version",
		GroupID: GroupSystem,
		Short:   "Print the version number of a2acli",
		Args:    cobra.NoArgs,
		Run:     runVersion,
	}

	sendCmd.Flags().StringVarP(&skillID, "skill", "s", "", "Skill ID")
	sendCmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	sendCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
	sendCmd.Flags().StringVarP(&instructionFile, "instruction-file", "i", "", "Path to a file with supplemental instructions")
	sendCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Block and wait for task completion instead of streaming (maps to A2A Blocking:true)")
	sendCmd.Flags().BoolVar(&wait, "sync", false, "Alias for --wait")
	sendCmd.Flags().BoolVar(&immediate, "immediate", false, "Fire-and-forget: submit task and return ID immediately without waiting or streaming")
	sendCmd.Flags().BoolVar(&showFull, "full", false, "Show complete artifact content without truncating (default preview is 500 chars)")
	sendCmd.Flags().StringVar(&messagePartsJSON, "parts", "", "Message parts as a JSON array, e.g. '[{\"text\":\"hello\"},{\"data\":{\"k\":\"v\"}}]'")
	sendCmd.Flags().StringVar(&messageBodyJSON, "json", "", "Complete Message as a JSON object (overrides text arg and other input flags)")
	sendCmd.Flags().StringArrayVar(&attachFiles, "attach", nil, "Attach a file as a message part (repeatable; MIME type auto-detected)")
	sendCmd.Flags().StringArrayVar(&dataArgs, "data", nil, "Add a JSON value as a DataPart (repeatable)")

	watchCmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	watchCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")

	getCmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	getCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
	getCmd.Flags().BoolVar(&showFull, "full", false, "Show complete artifact content without truncating")

	var downloadCmd = &cobra.Command{
		Use:     "download [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"retrieve"},
		Short:   "Download artifacts from a task",
		Long: `Download all artifacts produced by a specific task.

This is a convenience command that retrieves the task and saves its 
artifacts to the current directory or a specified output directory.`,
		Example: `  a2acli download <taskID>
  a2acli download <taskID> --out-dir ./results
  a2acli download <taskID> --file output.txt`,
		Args: cobra.ExactArgs(1),
		Run:  runGet, // Reuse runGet which now handles outDir and outFile natively
	}
	downloadCmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	downloadCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
	downloadCmd.Flags().BoolVar(&showFull, "full", false, "Show complete artifact content without truncating")

	var cancelCmd = &cobra.Command{
		Use:     "cancel [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"terminate", "CancelTask"},
		Short:   "Cancel an active task",
		Long: `Request cancellation of an active task.

If the task is still running, the agent will attempt to stop its 
execution. This is a best-effort request and the task may 
already have completed or be in a non-cancelable state.`,
		Example: `  a2acli cancel <taskID>
  a2acli cancel <taskID> --no-tui`,
		Args: cobra.ExactArgs(1),
		Run:  runCancel,
	}

	rootCmd.Run = func(cmd *cobra.Command, _ []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			runVersion(cmd, nil)
			return
		}
		_ = cmd.Help()
	}

	rootCmd.AddCommand(describeCmd, sendCmd, watchCmd, getCmd, downloadCmd, cancelCmd, setupConfigCmd(), versionCmd, setupServeCmd(), setupListCmd(), setupPushConfigCmd(), setupConformanceCmd(), setupA2UICmd(), setupAuthCmd())
	if err := rootCmd.Execute(); err != nil {
		fatalCode(ErrCodeInvalidArgument, "command execution failed", err, "")
	}
}

type streamSummary struct {
	taskID    string
	contextID string
	events    int
}

func checkTaskContinuable(task *a2a.Task) error {
	if task == nil || task.Status.State == "" {
		return nil
	}
	if task.Status.State.Terminal() {
		return fmt.Errorf("task %s is in terminal state %s and cannot be continued", task.ID, task.Status.State)
	}
	return nil
}

func printContinuationFooter(taskID, contextID string) {
	if disableTUI || outputMode == "json" {
		return
	}
	if taskID == "" && contextID == "" {
		return
	}
	fmt.Println()
	if taskID != "" {
		fmt.Printf("Task ID:    %s\n", StyleID.Render(taskID))
	}
	if contextID != "" {
		fmt.Printf("Context ID: %s\n", StyleID.Render(contextID))
	}
	if contextID != "" {
		fmt.Printf("\nContinue this conversation:\n  a2acli send --context %s \"your next message\"\n", contextID)
	} else if taskID != "" {
		fmt.Printf("\nContinue this task:\n  a2acli send --task %s \"your next message\"\n", taskID)
	}
}

func runTUI(stream chan streamMsg) (streamSummary, error) {
	p := tea.NewProgram(initialModel(stream, outDir))
	finalModel, err := p.Run()
	if err != nil {
		return streamSummary{}, err
	}

	m, ok := finalModel.(model)
	if !ok {
		return streamSummary{}, fmt.Errorf("unexpected TUI model type")
	}

	if m.err != nil {
		return streamSummary{taskID: m.taskID, contextID: m.contextID, events: m.eventCount}, m.err
	}

	summary := streamSummary{
		taskID:    m.taskID,
		contextID: m.contextID,
		events:    m.eventCount,
	}

	printContinuationFooter(summary.taskID, summary.contextID)
	return summary, nil
}

func runRaw(stream chan streamMsg, outDir string) (streamSummary, error) {
	var summary streamSummary
	for msg := range stream {
		if msg.Err != nil {
			fmt.Fprintf(os.Stderr, "{\"error\": %q}\n", msg.Err.Error())
			return summary, msg.Err
		}
		summary.events++
		if msg.Event != nil {
			info := msg.Event.TaskInfo()
			if info.TaskID != "" {
				summary.taskID = string(info.TaskID)
			}
			if info.ContextID != "" {
				summary.contextID = info.ContextID
			}
		}

		switch e := msg.Event.(type) {
		case *a2a.TaskStatusUpdateEvent:
			verboseLog("event: TaskStatusUpdate state=%s", e.Status.State)
		case *a2a.TaskArtifactUpdateEvent:
			verboseLog("event: TaskArtifactUpdate artifact=%q append=%v lastChunk=%v",
				e.Artifact.Name, e.Append, e.LastChunk)
			if outDir != "" || outFile != "" {
				_, _ = saveArtifact(outDir, outFile, *e.Artifact, 0)
			}
		}

		b, err := json.Marshal(msg.Event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "{\"error\": \"failed to encode event to json\"}\n")
			continue
		}
		fmt.Println(string(b))
	}

	return summary, nil
}

func displayTaskResult(task *a2a.Task, outDir string) {
	if outputMode == "compact" {
		var hist []string
		if task.History != nil {
			for _, m := range task.History {
				role := strings.ToLower(string(m.Role))
				role = strings.TrimPrefix(role, "role_")
				for _, p := range m.Parts {
					if tp, ok := p.Content.(a2a.Text); ok {
						hist = append(hist, fmt.Sprintf("[%s] %s", role, string(tp)))
					}
				}
			}
		}
		renderCompactBlock(string(task.ID), task.ContextID, &task.Status, task.Artifacts, hist)
		if outDir != "" || outFile != "" {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}
	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		if outDir != "" || outFile != "" {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}

	state := string(task.Status.State)
	var stateStyle lipgloss.Style
	switch task.Status.State {
	case a2a.TaskStateCompleted:
		stateStyle = StylePass
	case a2a.TaskStateFailed, a2a.TaskStateRejected:
		stateStyle = StyleFail
	default:
		stateStyle = StyleWarn
	}

	fmt.Printf("Task Status: [%s]\n", stateStyle.Render(state))

	if len(task.Artifacts) == 0 {
		fmt.Println("No artifacts produced.")
		return
	}

	fmt.Printf("\n%s\n", StyleAccent.Render(fmt.Sprintf("--- %d ARTIFACT(S) AVAILABLE ---", len(task.Artifacts))))

	for i, art := range task.Artifacts {
		fmt.Printf("\nName: %s\n", StyleArtifact.Render(art.Name))
		if art.Description != "" {
			fmt.Printf("Description: %s\n", art.Description)
		}

		truncated := false
		for _, p := range art.Parts {
			switch v := p.Content.(type) {
			case a2a.Data:
				prettyJSON, _ := json.MarshalIndent(v.Value, "", "  ")
				fmt.Printf("%s\n%s\n", StyleMuted.Render("Data (Preview):"), string(prettyJSON))
			case a2a.Text:
				content := string(v)
				if !showFull && len(content) > 500 {
					fmt.Printf("%s\n%s\n", StyleMuted.Render("Content (Preview):"), content[:500]+"... (truncated)")
					truncated = true
				} else {
					label := "Content:"
					if !showFull {
						label = "Content (Preview):"
					}
					fmt.Printf("%s\n%s\n", StyleMuted.Render(label), content)
				}
			case a2a.Raw:
				mediaType := p.MediaType
				if mediaType == "" {
					mediaType = "application/octet-stream"
				}
				fname := p.Filename
				if fname == "" {
					fname = art.Name + mimeToExt(mediaType)
				}
				fmt.Printf("%s %s (%d bytes)\n",
					StyleMuted.Render("Binary:"), StyleArtifact.Render(fname), len(v))
				truncated = true
			case a2a.URL:
				fmt.Printf("%s %s\n", StyleMuted.Render("URL:"), StyleArtifact.Render(string(v)))
				if p.MediaType != "" {
					fmt.Printf("%s %s\n", StyleMuted.Render("Type:"), p.MediaType)
				}
				truncated = true
			}
		}

		if outDir != "" || outFile != "" {
			path, err := saveArtifact(outDir, outFile, *art, i)
			if err != nil {
				fmt.Printf("%s %v\n", StyleFail.Render("Error saving artifact:"), err)
			} else if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
				// URL fallback — download failed or --out-dir not set
				fmt.Printf("%s %s\n", StyleMuted.Render("URL (use --out-dir to download):"), StyleArtifact.Render(path))
			} else {
				fmt.Printf("%s %s\n", StyleAccent.Render(">> Saved to:"), StyleArtifact.Render(path))
			}
		} else if truncated {
			fmt.Printf("%s\n", StyleMuted.Render("(Hint: Use --full to show complete content, or --out-dir <path> to save binary/URL artifacts)"))
		}
	}
	fmt.Printf("\n%s\n", StyleAccent.Render("------------------------------"))
}

// mimeToExt returns a file extension for a MIME type, e.g. "audio/mpeg" → ".mp3".
// Falls back to ".bin" for unknown binary types.
func mimeToExt(mediaType string) string {
	// Strip parameters ("audio/mpeg; charset=..." → "audio/mpeg")
	if idx := strings.Index(mediaType, ";"); idx >= 0 {
		mediaType = strings.TrimSpace(mediaType[:idx])
	}
	switch strings.ToLower(mediaType) {
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "audio/flac":
		return ".flac"
	case "audio/aac":
		return ".aac"
	case "audio/mp4":
		return ".m4a"
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "application/pdf":
		return ".pdf"
	case "application/json":
		return ".json"
	case "text/plain":
		return ".txt"
	case "text/html":
		return ".html"
	case "text/csv":
		return ".csv"
	case "application/zip":
		return ".zip"
	}
	// Try stdlib mime package for anything else.
	exts, _ := mime.ExtensionsByType(mediaType)
	if len(exts) > 0 {
		return exts[0]
	}
	return ".bin"
}

// downloadURL fetches content from a URL, forwarding auth headers if set.
// On failure it returns the URL string and a nil error so callers can
// surface the URL as a fallback rather than treating it as an error.
func downloadURL(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	// Forward auth if configured — note: pre-signed GCS URLs reject an
	// Authorization header, so only add it for non-GCS hosts.
	if authToken != "" && !strings.Contains(rawURL, "storage.googleapis.com") {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	buf := make([]byte, 0, resp.ContentLength)
	tmp := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf, nil
}

func validateOutDir(dir string) error {
	if dir == "" {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(dir))
	if lower == "json" || lower == "text" || lower == "tui" || lower == "ndjson" {
		return fmt.Errorf("invalid directory name %q: did you mean -o or --output %s?", dir, lower)
	}
	return nil
}

func saveArtifact(outDir, outFile string, artifact a2a.Artifact, index int) (string, error) {
	// Determine the base path (before we know the extension from content type).
	basePath := func(ext string) string {
		if outFile != "" {
			fName := outFile
			if index > 0 {
				e := filepath.Ext(outFile)
				base := strings.TrimSuffix(outFile, e)
				fName = fmt.Sprintf("%s_%d%s", base, index, e)
			}
			if outDir != "" {
				return filepath.Join(outDir, fName)
			}
			return fName
		}
		dir := outDir
		if dir == "" {
			dir = "."
		}
		name := artifact.Name
		if name == "" {
			name = fmt.Sprintf("artifact_%d_%d", time.Now().Unix(), index)
		}
		// Append ext if not already present.
		if ext != "" && !strings.HasSuffix(strings.ToLower(name), strings.ToLower(ext)) {
			name += ext
		}
		return filepath.Join(dir, name)
	}

	var (
		path         string
		contentBytes []byte
		urlFallback  string // set when URL download was requested but --out-dir not given
	)

	for _, p := range artifact.Parts {
		switch v := p.Content.(type) {
		case a2a.Text:
			contentBytes = []byte(string(v))
			path = basePath("")

		case a2a.Data:
			prettyJSON, _ := json.MarshalIndent(v.Value, "", "  ")
			contentBytes = prettyJSON
			ext := ".json"
			if p.MediaType != "" {
				ext = mimeToExt(p.MediaType)
			}
			path = basePath(ext)

		case a2a.Raw:
			contentBytes = []byte(v)
			ext := ".bin"
			if p.Filename != "" {
				if e := filepath.Ext(p.Filename); e != "" {
					ext = e
				}
			}
			if p.MediaType != "" {
				ext = mimeToExt(p.MediaType)
			}
			verboseLog("saveArtifact: Raw part %d bytes mediaType=%q ext=%s", len(contentBytes), p.MediaType, ext)
			path = basePath(ext)

		case a2a.URL:
			rawURL := string(v)
			verboseLog("saveArtifact: URL part %s mediaType=%q", rawURL, p.MediaType)
			if outDir != "" || outFile != "" {
				// Attempt download.
				data, err := downloadURL(rawURL)
				if err != nil {
					verboseLog("saveArtifact: URL download failed: %v — printing URL instead", err)
					urlFallback = rawURL
				} else {
					contentBytes = data
					ext := ".bin"
					if p.MediaType != "" {
						ext = mimeToExt(p.MediaType)
					} else if p.Filename != "" {
						if e := filepath.Ext(p.Filename); e != "" {
							ext = e
						}
					}
					path = basePath(ext)
				}
			} else {
				urlFallback = rawURL
			}
		}
	}

	// If we only have a URL fallback (no --out-dir, or download failed), return
	// it as a "path" so callers can surface it.
	if urlFallback != "" && contentBytes == nil {
		return urlFallback, nil
	}

	if len(contentBytes) == 0 || path == "" {
		return "", fmt.Errorf("no saveable content in artifact")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, contentBytes, 0644); err != nil {
		return "", err
	}
	return path, nil
}
