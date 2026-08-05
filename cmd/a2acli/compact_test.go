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
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestRenderCompactBlock(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	now := time.Now()
	status := &a2a.TaskStatus{
		State:     a2a.TaskStateCompleted,
		Timestamp: &now,
	}
	artifacts := []*a2a.Artifact{
		{
			ID:          "art-1",
			Name:        "report.txt",
			Description: "Report preview",
			Parts:       []*a2a.Part{a2a.NewTextPart("Summary content here")},
		},
	}
	history := []string{
		"[user] Write report",
		"[agent] Done writing report",
	}

	renderCompactBlock("task-100", "ctx-200", status, artifacts, history)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Task:     task-100") {
		t.Errorf("expected Task in compact block, got:\n%s", out)
	}
	if !strings.Contains(out, "Context:  ctx-200") {
		t.Errorf("expected Context in compact block, got:\n%s", out)
	}
	if !strings.Contains(out, "Status:   completed") {
		t.Errorf("expected Status completed in compact block, got:\n%s", out)
	}
	if !strings.Contains(out, "[report.txt] Summary content here") {
		t.Errorf("expected Artifact preview in compact block, got:\n%s", out)
	}
	if !strings.Contains(out, "[user] Write report") {
		t.Errorf("expected History in compact block, got:\n%s", out)
	}
}

func TestRunCompactStream(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stream := make(chan streamMsg, 2)
	now := time.Now()

	statusEvt := &a2a.TaskStatusUpdateEvent{
		Status: a2a.TaskStatus{
			State:     a2a.TaskStateCompleted,
			Timestamp: &now,
		},
		TaskID:    "task-999",
		ContextID: "ctx-999",
	}
	stream <- streamMsg{Event: statusEvt}
	close(stream)

	summary, err := runCompact(stream, "")

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runCompact returned unexpected error: %v", err)
	}
	if summary.taskID != "task-999" || summary.contextID != "ctx-999" {
		t.Errorf("unexpected summary: %+v", summary)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "Task:     task-999") {
		t.Errorf("expected Task in output, got:\n%s", out)
	}
}
