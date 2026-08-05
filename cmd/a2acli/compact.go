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
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// runCompact accumulates stream events and renders a static, single-frame compact summary block.
// Used when --output compact is set.
func runCompact(stream chan streamMsg, outDir string) (streamSummary, error) {
	var summary streamSummary
	var lastStatus *a2a.TaskStatus
	var artifacts []*a2a.Artifact
	var history []string

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
		case *a2a.Message:
			for _, p := range e.Parts {
				if tp, ok := p.Content.(a2a.Text); ok {
					role := strings.ToLower(string(e.Role))
					role = strings.TrimPrefix(role, "role_")
					history = append(history, fmt.Sprintf("[%s] %s", role, string(tp)))
				}
			}
		case *a2a.TaskStatusUpdateEvent:
			verboseLog("event: TaskStatusUpdate state=%s", e.Status.State)
			lastStatus = &e.Status
			if e.Status.Message != nil {
				for _, p := range e.Status.Message.Parts {
					if tp, ok := p.Content.(a2a.Text); ok {
						role := strings.ToLower(string(e.Status.Message.Role))
						role = strings.TrimPrefix(role, "role_")
						if role == "" {
							role = "agent"
						}
						history = append(history, fmt.Sprintf("[%s] %s", role, string(tp)))
					}
				}
			}
		case *a2a.TaskArtifactUpdateEvent:
			verboseLog("event: TaskArtifactUpdate artifact=%q append=%v lastChunk=%v",
				e.Artifact.Name, e.Append, e.LastChunk)
			artifacts = append(artifacts, e.Artifact)
			if outDir != "" || outFile != "" {
				_, _ = saveArtifact(outDir, outFile, *e.Artifact, len(artifacts)-1)
			}
		}
	}

	renderCompactBlock(summary.taskID, summary.contextID, lastStatus, artifacts, history)
	return summary, nil
}

func renderCompactBlock(taskID, contextID string, status *a2a.TaskStatus, artifacts []*a2a.Artifact, history []string) {
	if taskID != "" {
		fmt.Printf("Task:     %s\n", StyleID.Render(taskID))
	}
	if contextID != "" {
		fmt.Printf("Context:  %s\n", StyleID.Render(contextID))
	}
	if status != nil {
		tsStr := ""
		if status.Timestamp != nil {
			tsStr = fmt.Sprintf(" (%s)", status.Timestamp.Format("2006-01-02T15:04:05Z"))
		}
		stateStr := strings.TrimPrefix(string(status.State), "TASK_STATE_")
		fmt.Printf("Status:   %s%s\n", strings.ToLower(stateStr), tsStr)

		if status.Message != nil {
			for _, p := range status.Message.Parts {
				if tp, ok := p.Content.(a2a.Text); ok && string(tp) != "" {
					fmt.Printf("  %s\n", string(tp))
				}
			}
		}
	}

	if len(artifacts) > 0 {
		fmt.Println("\nArtifacts:")
		for _, art := range artifacts {
			name := art.Name
			if name == "" {
				name = string(art.ID)
			}
			preview := ""
			for _, p := range art.Parts {
				if tp, ok := p.Content.(a2a.Text); ok {
					preview = string(tp)
					if len(preview) > 100 {
						preview = preview[:100] + "..."
					}
					break
				}
			}
			if preview != "" {
				fmt.Printf("  [%s] %s\n", StyleArtifact.Render(name), preview)
			} else {
				fmt.Printf("  [%s]\n", StyleArtifact.Render(name))
			}
		}
	}

	if len(history) > 0 {
		fmt.Println("\nHistory:")
		for _, h := range history {
			fmt.Printf("  %s\n", h)
		}
	}
}
