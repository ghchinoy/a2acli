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

// UI Semantic Tokens based on the Tufte-inspired Design Philosophy.
// These are optimized for both Light and Dark terminal backgrounds using AdaptiveColor.
var (
	// Accent is for primary navigation landmarks (Headers, Group Titles).
	// Uses a deeper Cyan for Light mode and standard Cyan for Dark mode.
	StyleAccent = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "30", // Dark Cyan
		Dark:  "6",  // Cyan
	})

	// Command is for scan-targets (Command names, Flags).
	StyleCommand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "2",  // Green
		Dark:  "10", // Bright Green
	})

	// Muted is for de-emphasized metadata (Defaults, Types, Supplemental info).
	StyleMuted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "244", // Medium Grey
		Dark:  "240", // Dark Grey
	})

	// Pass is for successful A2A Task states (Completed).
	StylePass = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "2",  // Green
		Dark:  "10", // Bright Green
	})

	// Warn is for transient or concerning A2A Task states (Active, Pending).
	// Yellow/Yellow is difficult on light backgrounds; use Orange/Gold instead.
	StyleWarn = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "130", // Brown/Orange
		Dark:  "11",  // Yellow
	})

	// Fail is for terminal error A2A Task states (Failed, Rejected).
	StyleFail = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "1", // Red
		Dark:  "9", // Bright Red
	})

	// ID is for high-value unique identifiers (TaskIDs, SkillIDs).
	StyleID = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "5",  // Magenta
		Dark:  "13", // Bright Magenta
	})

	// Artifact is for data products (Artifact Names, Files).
	StyleArtifact = lipgloss.NewStyle().Bold(true).Underline(true)
)
