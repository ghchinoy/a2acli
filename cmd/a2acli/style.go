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
// These are optimized for both Light and Dark terminal backgrounds using Ayu hex codes.
var (
	// Accent is for primary navigation landmarks (Headers, Group Titles).
	// Uses Ayu Blue.
	StyleAccent = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#3199e1",
		Dark:  "#53bdfa",
	})

	// Command is for scan-targets (Command names, Flags).
	// Uses Ayu Bright Green.
	StyleCommand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#86b300",
		Dark:  "#c2d94c",
	})

	// Muted is for de-emphasized metadata (Defaults, Types, Supplemental info).
	// Uses Ayu Foreground/Grey.
	StyleMuted = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#6c7680",
		Dark:  "#b3b1ad",
	})

	// Pass is for successful A2A Task states (Completed).
	// Uses Ayu Green.
	StylePass = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#99bf4d",
		Dark:  "#91b362",
	})

	// Warn is for transient or concerning A2A Task states (Active, Pending).
	// Uses Ayu Yellow.
	StyleWarn = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#eca944",
		Dark:  "#f9af4f",
	})

	// Fail is for terminal error A2A Task states (Failed, Rejected).
	// Uses Ayu Red.
	StyleFail = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#ea6c6d",
		Dark:  "#ea6c73",
	})

	// ID is for high-value unique identifiers (TaskIDs, SkillIDs).
	// Uses Ayu Cyan.
	StyleID = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#46ba94",
		Dark:  "#90e1c6",
	})

	// Artifact is for data products (Artifact Names, Files).
	// Uses Bold Contrast.
	StyleArtifact = lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#ffffff",
	})
)
