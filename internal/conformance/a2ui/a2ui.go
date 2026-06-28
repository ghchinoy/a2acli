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

// Package a2ui provides renderer-agnostic, wire-level conformance validation for
// the A2UI A2A extension (v1.0). A2UI rides inside A2A DataParts, so the byte
// contract can be fully verified with an A2A client plus JSON-Schema validation —
// no UI renderer required.
//
// Schemas are vendored from a2ui-project/a2ui specification/v1_0/json/ and
// embedded at build time for reproducible, offline validation.
package a2ui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/ghchinoy/a2acli/internal/conformance"
)

// ExtensionURI is the canonical URI for the A2UI v1.0 A2A extension, as
// advertised in AgentCapabilities.extensions and the X-A2A-Extensions header.
const ExtensionURI = "https://a2ui.org/a2a-extension/a2ui/v1.0"

// MIMEType is the v1.0 media type for A2UI DataParts.
const MIMEType = "application/a2ui+json"

// LegacyMIMEType is the pre-v1.0 media type. Detecting it produces a clear
// "you're on the old format" diagnostic rather than an opaque schema failure.
const LegacyMIMEType = "application/json+a2ui"

// Version is the required envelope version string for every A2UI v1.0 message.
const Version = "v1.0"

// serverToClientListURL is the schema URL the captured DataPart `data` array is
// validated against.
const serverToClientListURL = "https://a2ui.org/specification/v1_0/json/server_to_client_list.json"

// catalogAliasURL satisfies the upstream cross-file ref: server_to_client.json
// references "catalog.json#/$defs/{anyComponent,surfaceProperties}", which
// resolves to this URL relative to its $id base. The defs actually live in the
// basic catalog (specification/v1_0/catalogs/basic/catalog.json, vendored as
// catalog.json), whose own $id is the deeper catalogs/basic path — so we
// register that document under this alias URL.
const catalogAliasURL = "https://a2ui.org/specification/v1_0/catalog.json"

//go:embed schemas/*.json
var schemaFS embed.FS

// NewSchemaSet builds the compiled A2UI v1.0 schema set from the vendored,
// embedded schemas. Each schema is registered under its own $id; the catalog
// definition is additionally aliased to satisfy the catalog.json ref.
func NewSchemaSet() (*conformance.SchemaSet, error) {
	entries, err := fs.ReadDir(schemaFS, "schemas")
	if err != nil {
		return nil, fmt.Errorf("read embedded schemas: %w", err)
	}

	var raws []conformance.RawSchema
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := schemaFS.ReadFile("schemas/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read schema %s: %w", e.Name(), err)
		}
		id, err := schemaID(data)
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", e.Name(), err)
		}
		raw := conformance.RawSchema{URL: id, Data: data}
		if e.Name() == "catalog.json" {
			// The basic catalog's own $id is the catalogs/basic/ path; alias it
			// to the bare catalog.json URL that server_to_client.json refs.
			raw.Aliases = []string{catalogAliasURL}
		}
		raws = append(raws, raw)
	}

	return conformance.NewSchemaSet(raws)
}

// schemaID extracts the $id from a JSON Schema document.
func schemaID(data []byte) (string, error) {
	var doc struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return "", fmt.Errorf("parse $id: %w", err)
	}
	if doc.ID == "" {
		return "", fmt.Errorf("missing $id")
	}
	return doc.ID, nil
}

// CapturedPart is the minimal view of an A2A DataPart needed for A2UI
// validation, decoupled from the SDK Part type so the engine stays testable with
// plain fixtures.
type CapturedPart struct {
	// MediaType is the SDK Part.MediaType field (JSON "mediaType").
	MediaType string
	// Metadata is the SDK Part.Metadata map (JSON "metadata").
	Metadata map[string]any
	// Data is the decoded structured payload (SDK Data.Value).
	Data any
}

// mimeOf returns the effective A2UI MIME for a part, checking the v1.0 location
// (metadata.mimeType) first, then falling back to the SDK MediaType field. The
// second return reports where it was found ("metadata" or "mediaType").
func (p CapturedPart) mimeOf() (string, string) {
	if p.Metadata != nil {
		if v, ok := p.Metadata["mimeType"].(string); ok && v != "" {
			return v, "metadata"
		}
	}
	return p.MediaType, "mediaType"
}

// IsA2UIPart reports whether a captured part looks like an A2UI part by either
// the v1.0 or legacy MIME, in either location.
func (p CapturedPart) IsA2UIPart() bool {
	mime, _ := p.mimeOf()
	return mime == MIMEType || mime == LegacyMIMEType
}

// ValidateParts runs the Phase A wire-level conformance checks over the captured
// DataParts and returns an ordered conformance report.
func ValidateParts(schemas *conformance.SchemaSet, parts []CapturedPart) conformance.Report {
	var results []conformance.Result

	// Filter to candidate A2UI parts.
	var a2uiParts []CapturedPart
	for _, p := range parts {
		if p.IsA2UIPart() {
			a2uiParts = append(a2uiParts, p)
		}
	}

	if len(a2uiParts) == 0 {
		results = append(results, conformance.Fail("A2UI DataParts present",
			"no DataPart with an A2UI MIME type was found in the response"))
		return conformance.NewReport(results)
	}
	results = append(results, conformance.Pass("A2UI DataParts present",
		fmt.Sprintf("found %d candidate A2UI DataPart(s)", len(a2uiParts))))

	for i, p := range a2uiParts {
		idx := i + 1

		// Check 1: MIME type and location.
		mime, loc := p.mimeOf()
		switch {
		case mime == LegacyMIMEType:
			results = append(results, conformance.Fail(
				fmt.Sprintf("DataPart %d: MIME type", idx),
				fmt.Sprintf("uses legacy MIME %q; v1.0 requires %q", LegacyMIMEType, MIMEType)))
		case mime != MIMEType:
			results = append(results, conformance.Fail(
				fmt.Sprintf("DataPart %d: MIME type", idx),
				fmt.Sprintf("unexpected MIME %q; v1.0 requires %q", mime, MIMEType)))
		case loc != "metadata":
			// Correct value but wrong location: v1.0 expects metadata.mimeType.
			results = append(results, conformance.Fail(
				fmt.Sprintf("DataPart %d: MIME location", idx),
				fmt.Sprintf("MIME %q found on part.mediaType; v1.0 expects it at metadata.mimeType", mime)))
		default:
			results = append(results, conformance.Pass(
				fmt.Sprintf("DataPart %d: MIME type", idx),
				fmt.Sprintf("%s at metadata.mimeType", MIMEType)))
		}

		// Check 2: data is an array of messages.
		arr, ok := p.Data.([]any)
		if !ok {
			results = append(results, conformance.Fail(
				fmt.Sprintf("DataPart %d: payload shape", idx),
				"data must be a JSON array of messages (v1.0); got a non-array value"))
			continue
		}
		results = append(results, conformance.Pass(
			fmt.Sprintf("DataPart %d: payload shape", idx),
			fmt.Sprintf("array of %d message(s)", len(arr))))

		// Check 3: each message envelope declares version v1.0.
		versionOK := true
		for j, m := range arr {
			obj, ok := m.(map[string]any)
			if !ok {
				results = append(results, conformance.Fail(
					fmt.Sprintf("DataPart %d msg %d: envelope", idx, j+1),
					"message is not a JSON object"))
				versionOK = false
				continue
			}
			if v, _ := obj["version"].(string); v != Version {
				results = append(results, conformance.Fail(
					fmt.Sprintf("DataPart %d msg %d: version", idx, j+1),
					fmt.Sprintf("version is %q; expected %q", v, Version)))
				versionOK = false
			}
		}
		if versionOK {
			results = append(results, conformance.Pass(
				fmt.Sprintf("DataPart %d: envelope versions", idx),
				fmt.Sprintf("all %d message(s) declare version %s", len(arr), Version)))
		}

		// Check 4: validate the whole list against the official v1.0 schema.
		// Non-transactional: report per-DataPart, continue on failure.
		if err := schemas.Validate(serverToClientListURL, p.Data); err != nil {
			results = append(results, conformance.Fail(
				fmt.Sprintf("DataPart %d: schema validation", idx),
				fmt.Sprintf("failed server_to_client_list schema: %v", err)))
		} else {
			results = append(results, conformance.Pass(
				fmt.Sprintf("DataPart %d: schema validation", idx),
				"validates against server_to_client_list.json (v1.0)"))
		}
	}

	return conformance.NewReport(results)
}
