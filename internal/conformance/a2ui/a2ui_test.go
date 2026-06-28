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

package a2ui

import (
	"encoding/json"
	"testing"
)

// conformantList is a minimal, hand-built A2UI v1.0 server->client message list:
// a createSurface (required: surfaceId, catalogId) followed by an updateDataModel.
const conformantList = `[
  {"version":"v1.0","createSurface":{"surfaceId":"surface_1","catalogId":"io.a2ui.basic/v1.0"}},
  {"version":"v1.0","updateDataModel":{"surfaceId":"surface_1","value":{"greeting":"hello"}}}
]`

func mustDecode(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return v
}

func TestSchemaSetCompiles(t *testing.T) {
	if _, err := NewSchemaSet(); err != nil {
		t.Fatalf("NewSchemaSet failed: %v", err)
	}
}

func TestConformantPart(t *testing.T) {
	schemas, err := NewSchemaSet()
	if err != nil {
		t.Fatalf("NewSchemaSet: %v", err)
	}
	part := CapturedPart{
		Metadata: map[string]any{"mimeType": MIMEType},
		Data:     mustDecode(t, conformantList),
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if !report.Passed {
		for _, r := range report.Results {
			if !r.Passed && !r.Skipped {
				t.Errorf("unexpected FAIL: %s — %s", r.Name, r.Message)
			}
		}
	}
}

func TestLegacyMIMEFlagged(t *testing.T) {
	schemas, _ := NewSchemaSet()
	// Mirrors the current apex sample server: legacy MIME on MediaType.
	part := CapturedPart{
		MediaType: LegacyMIMEType,
		Data:      mustDecode(t, conformantList),
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if report.Passed {
		t.Fatal("expected legacy MIME to fail conformance, but report passed")
	}
	found := false
	for _, r := range report.Results {
		if !r.Passed && contains(r.Message, "legacy MIME") {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'legacy MIME' failure message")
	}
}

func TestWrongMIMELocationFlagged(t *testing.T) {
	schemas, _ := NewSchemaSet()
	// Correct v1.0 value, but on the SDK MediaType field instead of metadata.mimeType.
	part := CapturedPart{
		MediaType: MIMEType,
		Data:      mustDecode(t, conformantList),
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if report.Passed {
		t.Fatal("expected wrong MIME location to fail, but report passed")
	}
	found := false
	for _, r := range report.Results {
		if !r.Passed && contains(r.Message, "metadata.mimeType") {
			found = true
		}
	}
	if !found {
		t.Error("expected a 'metadata.mimeType' location failure")
	}
}

func TestNonArrayPayloadFlagged(t *testing.T) {
	schemas, _ := NewSchemaSet()
	// data is a single object, not an array (a common v0.9 mistake).
	part := CapturedPart{
		Metadata: map[string]any{"mimeType": MIMEType},
		Data:     mustDecode(t, `{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"c"}}`),
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if report.Passed {
		t.Fatal("expected non-array payload to fail, but report passed")
	}
}

func TestWrongVersionFlagged(t *testing.T) {
	schemas, _ := NewSchemaSet()
	part := CapturedPart{
		Metadata: map[string]any{"mimeType": MIMEType},
		Data:     mustDecode(t, `[{"version":"v0.9","createSurface":{"surfaceId":"s","catalogId":"c"}}]`),
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if report.Passed {
		t.Fatal("expected v0.9 envelope to fail, but report passed")
	}
}

func TestNoA2UIParts(t *testing.T) {
	schemas, _ := NewSchemaSet()
	part := CapturedPart{
		MediaType: "text/plain",
		Data:      "just text",
	}
	report := ValidateParts(schemas, []CapturedPart{part})
	if report.Passed {
		t.Fatal("expected report to fail when no A2UI parts present")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
