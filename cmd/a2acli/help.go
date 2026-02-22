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
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// colorizedHelpFunc wraps Cobra's default help with semantic coloring.
func colorizedHelpFunc(cmd *cobra.Command, _ []string) {
	// If TUI is disabled or NO_COLOR is set, use default help
	if disableTUI {
		fmt.Print(cmd.UsageString())
		return
	}

	var output strings.Builder

	// Include Long description first (Unstyled per Tufte guidelines)
	if cmd.Long != "" {
		output.WriteString(cmd.Long)
		output.WriteString("\n\n")
	} else if cmd.Short != "" {
		output.WriteString(cmd.Short)
		output.WriteString("\n\n")
	}

	// Add the usage string which contains commands, flags, etc.
	output.WriteString(cmd.UsageString())

	// Apply semantic coloring
	result := colorizeHelpOutput(output.String())
	fmt.Print(result)
}

// colorizeHelpOutput applies semantic colors to help text via regex.
func colorizeHelpOutput(help string) string {
	// Match group header lines (e.g., "Messaging & Tasks:")
	groupHeaderRE := regexp.MustCompile(`(?m)^([A-Z][A-Za-z &]+:)\s*$`)
	result := groupHeaderRE.ReplaceAllStringFunc(help, func(match string) string {
		return StyleAccent.Render(match)
	})

	// Match section headers (Usage:, Flags:, Examples:, Aliases:, Available Commands:)
	sectionHeaderRE := regexp.MustCompile(`(?m)^(Usage|Flags|Examples|Aliases|Available Commands|Global Flags):`)
	result = sectionHeaderRE.ReplaceAllStringFunc(result, func(match string) string {
		return StyleAccent.Render(match)
	})

	// Match command lines: "  command   Description text"
	cmdLineRE := regexp.MustCompile(`(?m)^(  )([a-z][a-z0-9]*(?:-[a-z0-9]+)*)(\s{2,})(.*)$`)
	result = cmdLineRE.ReplaceAllStringFunc(result, func(match string) string {
		parts := cmdLineRE.FindStringSubmatch(match)
		if len(parts) != 5 {
			return match
		}
		indent := parts[1]
		cmdName := parts[2]
		spacing := parts[3]
		description := parts[4]

		return indent + StyleCommand.Render(cmdName) + spacing + description
	})

	// Match flag lines: "  -f, --file string   Description"
	flagLineRE := regexp.MustCompile(`(?m)^(\s+)(-\w,\s+--[\w-]+|--[\w-]+)(\s+)(string|int|duration|bool)?(\s*.*)$`)
	result = flagLineRE.ReplaceAllStringFunc(result, func(match string) string {
		parts := flagLineRE.FindStringSubmatch(match)
		if len(parts) < 6 {
			return match
		}
		indent := parts[1]
		flags := parts[2]
		spacing := parts[3]
		typeStr := parts[4]
		desc := parts[5]

		if typeStr != "" {
			return indent + StyleCommand.Render(flags) + spacing + StyleMuted.Render(typeStr) + desc
		}
		return indent + StyleCommand.Render(flags) + spacing + desc
	})

	// Mute default values (default "...")
	defaultRE := regexp.MustCompile(`(\(default[^)]*\))`)
	result = defaultRE.ReplaceAllStringFunc(result, func(match string) string {
		return StyleMuted.Render(match)
	})

	return result
}
