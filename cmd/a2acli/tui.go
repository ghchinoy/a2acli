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

// Package main provides the entry point for the A2A CLI.
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type streamMsg struct {
	Event a2a.Event
	Err   error
}

type model struct {
	sub      <-chan streamMsg
	messages []string
	spinner  spinner.Model
	status   string
	taskID   string
	quitting bool
	err      error
	outDir   string
	width    int
}

type eventMsg streamMsg
type errMsg error
type doneMsg struct{}

func initialModel(sub <-chan streamMsg, outDir string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = StyleAccent
	return model{
		sub:      sub,
		spinner:  s,
		status:   "Initializing...",
		messages: []string{},
		outDir:   outDir,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.waitForActivity(),
	)
}

func (m model) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.sub
		if !ok {
			return doneMsg{}
		}
		return eventMsg(msg)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case doneMsg:
		m.quitting = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case eventMsg:
		return m.handleEvent(msg)
	}

	return m, nil
}

func (m model) handleEvent(msg eventMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		return m, tea.Quit
	}

	// Handle A2A Events
	event := msg.Event

	if event.TaskInfo().TaskID != "" {
		m.taskID = string(event.TaskInfo().TaskID)
	}

	cmds := []tea.Cmd{m.waitForActivity()} // Continue listening

	switch v := event.(type) {
	case *a2a.Message:
		for _, p := range v.Parts {
			if tp, ok := p.Content.(a2a.Text); ok {
				m.messages = append(m.messages, fmt.Sprintf("%s %s", StyleCommand.Render("Agent:"), string(tp)))
			}
		}
		m.status = "Received Message"

	case *a2a.TaskStatusUpdateEvent:
		m.handleStatusUpdate(v)

	case *a2a.TaskArtifactUpdateEvent:
		m.handleArtifactUpdate(v)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleStatusUpdate(v *a2a.TaskStatusUpdateEvent) {
	m.status = string(v.Status.State)
	statusMsg := ""
	if v.Status.Message != nil && len(v.Status.Message.Parts) > 0 {
		if tp, ok := v.Status.Message.Parts[0].Content.(a2a.Text); ok {
			statusMsg = string(tp)
		}
	}

	var stateStyle lipgloss.Style
	switch v.Status.State {
	case a2a.TaskStateCompleted:
		stateStyle = StylePass
	case a2a.TaskStateFailed, a2a.TaskStateRejected:
		stateStyle = StyleFail
	default:
		stateStyle = StyleWarn
	}

	if statusMsg != "" {
		m.messages = append(m.messages, fmt.Sprintf("[%s] %s", stateStyle.Render(string(v.Status.State)), StyleMuted.Render(statusMsg)))
	}
}

func (m *model) handleArtifactUpdate(v *a2a.TaskArtifactUpdateEvent) {
	m.status = "Artifact Received"
	m.messages = append(m.messages, StyleArtifact.Render(fmt.Sprintf("ARTIFACT: %s", v.Artifact.Name)))

	saveMsg := ""
	if m.outDir != "" || outFile != "" {
		path, err := saveArtifact(m.outDir, outFile, *v.Artifact, 0)
		if err != nil {
			saveMsg = fmt.Sprintf("Error saving: %v", err)
		} else {
			saveMsg = fmt.Sprintf("Saved to: %s", path)
		}
	}

	// Display preview
	for _, p := range v.Artifact.Parts {
		if dp, ok := p.Content.(a2a.Data); ok {
			prettyJSON, _ := json.MarshalIndent(dp, "", "  ")
			preview := string(prettyJSON)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			m.messages = append(m.messages, fmt.Sprintf("%s\n%s", StyleMuted.Render("Data (Preview):"), preview))
		} else if tp, ok := p.Content.(a2a.Text); ok {
			preview := string(tp)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			m.messages = append(m.messages, fmt.Sprintf("%s\n%s", StyleMuted.Render("Content (Preview):"), preview))
		}
	}
	if saveMsg != "" {
		m.messages = append(m.messages, StyleAccent.Render(saveMsg))
	}
}

func (m model) View() string {
	if m.err != nil {
		return StyleFail.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	// Status Bar
	spin := m.spinner.View() + " "
	if m.quitting {
		spin = ""
	}

	state := strings.ToUpper(m.status)
	statusLine := fmt.Sprintf("%s%s", spin, StyleAccent.Render(state))
	if m.taskID != "" {
		statusLine += fmt.Sprintf(" | Task: %s", StyleID.Render(m.taskID))
	}

	// History
	history := ""
	start := 0
	if len(m.messages) > 15 {
		start = len(m.messages) - 15
	}
	history = strings.Join(m.messages[start:], "\n")

	// Adjust width to account for margins
	width := m.width - 4
	if width < 0 {
		width = 0
	}

	return docStyle.Width(width).Render(fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		history,
		statusLine,
		StyleMuted.Render("(ctrl+c to quit)"),
	))
}
