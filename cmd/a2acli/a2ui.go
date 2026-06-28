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
	"os"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/a2acli/internal/conformance"
	"github.com/ghchinoy/a2acli/internal/conformance/a2ui"
	"github.com/spf13/cobra"
)

var a2uiProbeMessage string

// setupA2UICmd builds the `a2ui` command group (A2UI extension conformance).
func setupA2UICmd() *cobra.Command {
	a2uiCmd := &cobra.Command{
		Use:     "a2ui",
		GroupID: GroupSystem,
		Short:   "A2UI extension conformance validation",
		Long: `Validate that an A2A server emits a conformant A2UI v1.0 extension stream.

A2UI is transport-decoupled and rides inside A2A DataParts, so a2acli can verify
the byte-level contract against the official A2UI v1.0 JSON Schemas without a UI
renderer. Schemas are vendored and embedded for reproducible, offline validation.`,
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a live server's A2UI v1.0 conformance",
		Long: `Send a probe message, capture the returned A2UI DataParts, and validate them:

  1. AgentCard declares the A2UI v1.0 extension (and its capability params)
  2. DataParts use metadata.mimeType == application/a2ui+json
  3. Each DataPart's data is an array of v1.0 message envelopes
  4. Each message list validates against server_to_client_list.json (v1.0)

Exit code is non-zero if any check fails.`,
		Example: `  a2acli a2ui validate --service-url http://localhost:9002
  a2acli a2ui validate -u http://localhost:9002 --probe "show me the showcase card"
  a2acli a2ui validate -u http://localhost:9002 --output json`,
		Run: runA2UIValidate,
	}
	validateCmd.Flags().StringVar(&a2uiProbeMessage, "probe", "render a UI", "Message text to elicit an A2UI response")

	a2uiCmd.AddCommand(validateCmd)
	return a2uiCmd
}

func runA2UIValidate(_ *cobra.Command, _ []string) {
	ctx := context.Background()
	var results []conformance.Result

	schemas, err := a2ui.NewSchemaSet()
	if err != nil {
		fatalf("failed to load A2UI schemas", err, "This is a build problem; please file an issue")
	}

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}
	verboseLog("resolved card %q; checking for A2UI extension", card.Name)

	// ── Check 1: AgentCard declares the A2UI v1.0 extension ─────────────────
	var ext *a2a.AgentExtension
	for i := range card.Capabilities.Extensions {
		if card.Capabilities.Extensions[i].URI == a2ui.ExtensionURI {
			ext = &card.Capabilities.Extensions[i]
			break
		}
	}
	if ext == nil {
		results = append(results, conformance.Fail("AgentCard A2UI extension",
			fmt.Sprintf("no extension with uri=%s in capabilities.extensions", a2ui.ExtensionURI)))
		// Without the extension declared, still attempt the stream — some servers
		// emit A2UI without advertising. Continue.
	} else {
		msg := fmt.Sprintf("extension declared (required=%v)", ext.Required)
		if ext.Params != nil {
			if cats, ok := ext.Params["supportedCatalogIds"]; ok {
				msg += fmt.Sprintf("; supportedCatalogIds=%v", cats)
			}
			if inline, ok := ext.Params["acceptsInlineCatalogs"]; ok {
				msg += fmt.Sprintf("; acceptsInlineCatalogs=%v", inline)
			}
		}
		results = append(results, conformance.Pass("AgentCard A2UI extension", msg))
		verboseLog("A2UI extension params: %+v", ext.Params)
	}

	// ── Send probe message with the extension activated ─────────────────────
	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(a2uiProbeMessage))
	// Activate the extension via the standard header, carried as a service param
	// through the existing interceptor plumbing.
	svcParams = append(svcParams, "X-A2A-Extensions="+a2ui.ExtensionURI)
	verboseLog("activating extension via X-A2A-Extensions=%s", a2ui.ExtensionURI)

	req := &a2a.SendMessageRequest{
		Message: msg,
		Config:  &a2a.SendMessageConfig{ReturnImmediately: false},
	}

	result, err := client.SendMessage(ctx, req)
	if err != nil {
		results = append(results, conformance.Fail("Probe send", fmt.Sprintf("SendMessage failed: %v", err)))
		emitA2UIReport(conformance.NewReport(results))
		os.Exit(1)
	}

	// ── Collect DataParts from the response ─────────────────────────────────
	parts := collectParts(result)
	verboseLog("collected %d part(s) from response", len(parts))

	captured := make([]a2ui.CapturedPart, 0, len(parts))
	for _, p := range parts {
		cp := a2ui.CapturedPart{
			MediaType: p.MediaType,
			Metadata:  p.Metadata,
		}
		if dp, ok := p.Content.(a2a.Data); ok {
			cp.Data = dp.Value
		}
		captured = append(captured, cp)
	}

	// ── Run the wire-level validation engine ────────────────────────────────
	report := a2ui.ValidateParts(schemas, captured)
	results = append(results, report.Results...)

	final := conformance.NewReport(results)
	emitA2UIReport(final)
	if !final.Passed {
		os.Exit(1)
	}
}

// collectParts extracts all content parts from a SendMessage result (Task or Message).
func collectParts(result a2a.SendMessageResult) []*a2a.Part {
	var parts []*a2a.Part
	switch r := result.(type) {
	case *a2a.Task:
		for _, art := range r.Artifacts {
			parts = append(parts, art.Parts...)
		}
		// Some servers carry parts in status messages too.
		if r.Status.Message != nil {
			parts = append(parts, r.Status.Message.Parts...)
		}
	case *a2a.Message:
		parts = append(parts, r.Parts...)
	}
	return parts
}

func emitA2UIReport(report conformance.Report) {
	if disableTUI {
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
		return
	}

	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#c2d94c")).Bold(true)
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f07178")).Bold(true)
	skipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bfbdb6"))

	fmt.Printf("\nA2UI v1.0 Extension Conformance — %s\n\n", serviceURL)
	for _, r := range report.Results {
		var label string
		switch {
		case r.Skipped:
			label = skipStyle.Render("SKIP")
		case r.Passed:
			label = passStyle.Render("PASS")
		default:
			label = failStyle.Render("FAIL")
		}
		fmt.Printf("  [%s] %s\n", label, r.Name)
		if r.Message != "" {
			fmt.Printf("       %s\n", r.Message)
		}
	}
	fmt.Println()
	if report.Passed {
		fmt.Println(passStyle.Render("A2UI v1.0 conformance: PASS"))
	} else {
		fmt.Println(failStyle.Render("A2UI v1.0 conformance: FAIL"))
	}
}
