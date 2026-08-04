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
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestCheckTaskContinuable(t *testing.T) {
	tests := []struct {
		name    string
		state   a2a.TaskState
		wantErr bool
	}{
		{"empty state", "", false},
		{"submitted state", a2a.TaskStateSubmitted, false},
		{"working state", a2a.TaskStateWorking, false},
		{"completed state", a2a.TaskStateCompleted, true},
		{"canceled state", a2a.TaskStateCanceled, true},
		{"failed state", a2a.TaskStateFailed, true},
		{"rejected state", a2a.TaskStateRejected, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &a2a.Task{
				ID: "task-123",
				Status: a2a.TaskStatus{
					State: tt.state,
				},
			}
			err := checkTaskContinuable(task)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkTaskContinuable() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), string(tt.state)) {
				t.Errorf("expected error message to contain state %q, got: %v", tt.state, err)
			}
		})
	}
}

func TestPrintContinuationFooter(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Save and restore flags
	origDisableTUI := disableTUI
	origOutputMode := outputMode
	defer func() {
		disableTUI = origDisableTUI
		outputMode = origOutputMode
		os.Stdout = oldStdout
	}()

	disableTUI = false
	outputMode = "text"

	printContinuationFooter("task-abc", "ctx-xyz")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Task ID:    task-abc") {
		t.Errorf("expected Task ID in footer, got:\n%s", output)
	}
	if !strings.Contains(output, "Context ID: ctx-xyz") {
		t.Errorf("expected Context ID in footer, got:\n%s", output)
	}
	if !strings.Contains(output, "a2acli send --context ctx-xyz") {
		t.Errorf("expected continuation hint with --context, got:\n%s", output)
	}
}

func TestBuildMessageWithContext(t *testing.T) {
	msg, err := buildMessage("Hello AI")
	if err != nil {
		t.Fatalf("buildMessage failed: %v", err)
	}
	if len(msg.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(msg.Parts))
	}
	if tp, ok := msg.Parts[0].Content.(a2a.Text); !ok || string(tp) != "Hello AI" {
		t.Errorf("unexpected part content: %v", msg.Parts[0].Content)
	}
}

func TestValidateOutDir(t *testing.T) {
	invalidDirs := []string{"json", "JSON", "text", "TEXT", "tui", "TUI", "ndjson"}
	for _, dir := range invalidDirs {
		err := validateOutDir(dir)
		if err == nil {
			t.Errorf("expected error for validateOutDir(%q), got nil", dir)
		} else if !strings.Contains(err.Error(), "did you mean -o or --output") {
			t.Errorf("expected error message to contain 'did you mean -o or --output', got: %v", err)
		}
	}

	validDirs := []string{"", "./artifacts", "/tmp/out", "reports"}
	for _, dir := range validDirs {
		if err := validateOutDir(dir); err != nil {
			t.Errorf("unexpected error for validateOutDir(%q): %v", dir, err)
		}
	}
}

func TestListStatusFormatting(t *testing.T) {
	states := []struct {
		input a2a.TaskState
		want  string
	}{
		{a2a.TaskStateCompleted, "COMPLETED"},
		{a2a.TaskStateWorking, "WORKING"},
		{a2a.TaskStateFailed, "FAILED"},
		{a2a.TaskStateCanceled, "CANCELED"},
		{a2a.TaskStateRejected, "REJECTED"},
		{a2a.TaskStateSubmitted, "SUBMITTED"},
	}

	for _, tt := range states {
		got := strings.TrimPrefix(string(tt.input), "TASK_STATE_")
		if got != tt.want {
			t.Errorf("TrimPrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
