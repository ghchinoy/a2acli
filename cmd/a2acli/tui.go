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
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	docStyle    = lipgloss.NewStyle().Margin(1, 2)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))           // Pinkish
	taskIDStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true) // Cyan
	agentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))            // Green
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
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
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
		waitForActivity(m.sub),
	)
}

func waitForActivity(sub <-chan streamMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-sub
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
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}

		// Handle A2A Events
		event := msg.Event

		if event.TaskInfo().TaskID != "" {
			m.taskID = string(event.TaskInfo().TaskID)
		}

		cmds := []tea.Cmd{waitForActivity(m.sub)} // Continue listening

		switch v := event.(type) {
		case *a2a.Message:
			for _, p := range v.Parts {
				if tp, ok := p.(a2a.TextPart); ok {
					m.messages = append(m.messages, agentStyle.Render(fmt.Sprintf("Agent: %s", tp.Text)))
				}
			}
			m.status = "Received Message"

		case *a2a.TaskStatusUpdateEvent:
			m.status = string(v.Status.State)
			statusMsg := ""
			if v.Status.Message != nil && len(v.Status.Message.Parts) > 0 {
				if tp, ok := v.Status.Message.Parts[0].(a2a.TextPart); ok {
					statusMsg = tp.Text
				}
			}
			if statusMsg != "" {
				m.messages = append(m.messages, subtleStyle.Render(fmt.Sprintf("[%s] %s", v.Status.State, statusMsg)))
			}

		case *a2a.TaskArtifactUpdateEvent:
			m.status = "Artifact Received"
			header := fmt.Sprintf("--- ARTIFACT: %s ---", v.Artifact.Name)
			m.messages = append(m.messages, lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(header))

			saveMsg := ""
			if m.outDir != "" {
				path, err := saveArtifact(m.outDir, *v.Artifact)
				if err != nil {
					saveMsg = fmt.Sprintf("Error saving: %v", err)
				} else {
					saveMsg = fmt.Sprintf("Saved to: %s", path)
				}
			}

			// Display preview
			for _, p := range v.Artifact.Parts {
				if dp, ok := p.(a2a.DataPart); ok {
					prettyJSON, _ := json.MarshalIndent(dp.Data, "", "  ")
					preview := string(prettyJSON)
					if len(preview) > 200 {
						preview = preview[:200] + "..."
					}
					m.messages = append(m.messages, fmt.Sprintf("Data: %s", preview))
				} else if tp, ok := p.(a2a.TextPart); ok {
					preview := tp.Text
					if len(preview) > 200 {
						preview = preview[:200] + "..."
					}
					m.messages = append(m.messages, fmt.Sprintf("Text: %s", preview))
				}
			}
			if saveMsg != "" {
				m.messages = append(m.messages, subtleStyle.Render(saveMsg))
			}
		}

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	// Status Bar
	spin := m.spinner.View() + " "
	if m.quitting {
		spin = ""
	}

	statusLine := fmt.Sprintf("%s%s", spin, statusStyle.Render(strings.ToUpper(m.status)))
	if m.taskID != "" {
		statusLine += fmt.Sprintf(" | Task: %s", taskIDStyle.Render(m.taskID))
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
		"%s\n\n%s\n\n(ctrl+c to quit)",
		history,
		statusLine,
	))
}
