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
	"encoding/json"
	"fmt"
	"testing"
)

// TestExitCodeForError verifies SPEC §11.6: usage errors MUST exit 2, and
// every other failure exits 1 (reserved codes 3/4/5 are not implemented, so
// their conditions report 1 as the spec permits).
func TestExitCodeForError(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{ErrUsage, 2},
		{ErrUnreachable, 1},
		{ErrAuthFailed, 1},
		{ErrTimeout, 1},
		{ErrCardNotFound, 1},
		{ErrCardInvalid, 1},
		{ErrCredentialsMissing, 1},
		{ErrInternal, 1},
		{"", 1},
	}
	for _, tt := range tests {
		if got := exitCodeForError(tt.code); got != tt.want {
			t.Errorf("exitCodeForError(%q) = %d, want %d", tt.code, got, tt.want)
		}
	}
}

// TestClassifyError verifies auto-classification into the A2ACLI_ERR_* namespace.
func TestClassifyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"401 unauthorized", fmt.Errorf("server returned 401 Unauthorized"), ErrAuthFailed},
		{"connection refused", fmt.Errorf("dial tcp 127.0.0.1:9001: connect: connection refused"), ErrUnreachable},
		{"no such host", fmt.Errorf("lookup nope.invalid: no such host"), ErrUnreachable},
		{"tls failure", fmt.Errorf("tls: failed to verify certificate"), ErrUnreachable},
		{"deadline exceeded", fmt.Errorf("context deadline exceeded"), ErrTimeout},
		{"generic", fmt.Errorf("something odd happened"), ErrInternal},
		{"nil", nil, ErrInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyError(tt.err); got != tt.want {
				t.Errorf("classifyError(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

// TestIsUnreachable spot-checks the unreachable detector.
func TestIsUnreachable(t *testing.T) {
	unreachable := []string{
		"dial tcp 127.0.0.1:9001: connect: connection refused",
		"lookup example.invalid: no such host",
		"network is unreachable",
		"x509: certificate signed by unknown authority",
	}
	for _, s := range unreachable {
		if !isUnreachable(fmt.Errorf("%s", s)) {
			t.Errorf("isUnreachable(%q) = false, want true", s)
		}
	}
	reachable := []string{
		"task not found",
		"invalid argument",
	}
	for _, s := range reachable {
		if isUnreachable(fmt.Errorf("%s", s)) {
			t.Errorf("isUnreachable(%q) = true, want false", s)
		}
	}
	if isUnreachable(nil) {
		t.Error("isUnreachable(nil) = true, want false")
	}
}

// TestBuildErrorEnvelopeShape verifies the SPEC §11.4 / Appendix B envelope:
// a nested {"error":{"code","message","hint","a2aCode"}} shape with exact field
// names/casing, hint null when absent, and a2aCode null when unknown.
func TestBuildErrorEnvelopeShape(t *testing.T) {
	env := buildErrorEnvelope(ErrUnreachable, "failed to resolve AgentCard: connection refused", "Ensure the A2A server is running", fmt.Errorf("connection refused"))
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal top: %v", err)
	}
	if len(raw) != 1 {
		t.Fatalf("top-level must have exactly one key %q, got %v", "error", string(b))
	}
	inner, ok := raw["error"]
	if !ok {
		t.Fatalf("missing top-level \"error\" key: %s", string(b))
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(inner, &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	for _, field := range []string{"code", "message", "hint", "a2aCode"} {
		if _, ok := body[field]; !ok {
			t.Errorf("error body missing required field %q: %s", field, string(b))
		}
	}
	// code must be the A2ACLI_ERR_* value we passed.
	var code string
	_ = json.Unmarshal(body["code"], &code)
	if code != ErrUnreachable {
		t.Errorf("code = %q, want %q", code, ErrUnreachable)
	}
}

// TestBuildErrorEnvelopeNullFields verifies hint and a2aCode serialize as JSON
// null when absent (rather than being omitted or empty strings).
func TestBuildErrorEnvelopeNullFields(t *testing.T) {
	env := buildErrorEnvelope(ErrInternal, "boom", "", fmt.Errorf("boom"))
	b, _ := json.Marshal(env)
	var raw struct {
		Error struct {
			Hint    *string `json:"hint"`
			A2ACode any     `json:"a2aCode"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if raw.Error.Hint != nil {
		t.Errorf("hint = %v, want null", *raw.Error.Hint)
	}
	if raw.Error.A2ACode != nil {
		t.Errorf("a2aCode = %v, want null", raw.Error.A2ACode)
	}
}

// TestBuildErrorEnvelopeA2ACode verifies the transport-level code is surfaced
// when discernible (e.g. HTTP 401).
func TestBuildErrorEnvelopeA2ACode(t *testing.T) {
	env := buildErrorEnvelope(ErrAuthFailed, "unauthorized", "pass --token", fmt.Errorf("401 Unauthorized"))
	if env.Error.A2ACode == nil {
		t.Fatal("a2aCode = nil, want 401")
	}
	if got := fmt.Sprintf("%v", env.Error.A2ACode); got != "401" {
		t.Errorf("a2aCode = %v, want 401", env.Error.A2ACode)
	}
}

// TestErrorCodesNamespace verifies all CLI-local codes live in the A2ACLI_ERR_
// namespace (SPEC §11.4 / Appendix D).
func TestErrorCodesNamespace(t *testing.T) {
	codes := []string{
		ErrUsage, ErrCardNotFound, ErrCardInvalid, ErrUnreachable,
		ErrCredentialsMissing, ErrAuthFailed, ErrTimeout, ErrInternal,
	}
	for _, c := range codes {
		if len(c) < len("A2ACLI_ERR_") || c[:len("A2ACLI_ERR_")] != "A2ACLI_ERR_" {
			t.Errorf("code %q is not in the A2ACLI_ERR_ namespace", c)
		}
	}
}
