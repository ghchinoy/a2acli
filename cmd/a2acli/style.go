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
	"github.com/charmbracelet/lipgloss"
)

// UI Semantic Tokens based on the Tufte-inspired Design Philosophy and Ayu theme.
// These are aligned with the 'beads' project for cross-tool visual consistency.
var (
	// Accent is for primary navigation landmarks (Headers, Group Titles).
	StyleAccent = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#399ee6", // ayu light bright blue
		Dark:  "#59c2ff", // ayu dark bright blue
	})

	// Command is for scan-targets (Command names, Flags).
	// Uses subtle foreground colors to prioritize data-ink.
	StyleCommand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#5c6166", // ayu light command grey
		Dark:  "#bfbdb6", // ayu dark command grey
	})

	// Muted is for de-emphasized metadata (Defaults, Types, Supplemental info).
	StyleMuted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#828c99", // ayu light muted
		Dark:  "#6c7680", // ayu dark muted
	})

	// Pass is for successful A2A Task states (Completed).
	StylePass = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#86b300", // ayu light bright green
		Dark:  "#c2d94c", // ayu dark bright green
	})

	// Warn is for transient or concerning A2A Task states (Active, Pending).
	StyleWarn = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#f2ae49", // ayu light bright yellow
		Dark:  "#ffb454", // ayu dark bright yellow
	})

	// Fail is for terminal error A2A Task states (Failed, Rejected).
	StyleFail = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#f07171", // ayu light bright red
		Dark:  "#f07178", // ayu dark bright red
	})

	// ID is for high-value unique identifiers (TaskIDs, SkillIDs).
	// Using Ayu Cyan for identifiers helps them stand out from standard text.
	StyleID = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#46ba94", // ayu light cyan
		Dark:  "#95e6cb", // ayu dark cyan
	})

	// Artifact is for data products (Artifact Names, Files).
	StyleArtifact = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#ffffff",
	})
)
