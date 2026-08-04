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
	"runtime/debug"

	"github.com/spf13/cobra"
)

// These variables are populated by GoReleaser or the Makefile at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func getVersionInfo() (string, string, string) {
	v, c, d := version, commit, date
	if bi, ok := debug.ReadBuildInfo(); ok {
		if (v == "" || v == "dev") && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			v = bi.Main.Version
		}
		var rev, t string
		var modified bool
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				rev = setting.Value
			case "vcs.time":
				t = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					modified = true
				}
			}
		}
		if rev != "" && modified {
			rev += "-dirty"
		}
		if (c == "" || c == "none") && rev != "" {
			c = rev
		}
		if (d == "" || d == "unknown") && t != "" {
			d = t
		}
	}
	return v, c, d
}

func runVersion(_ *cobra.Command, _ []string) {
	v, c, d := getVersionInfo()
	fmt.Printf("a2acli version %s\n", v)
	fmt.Printf("commit: %s\n", c)
	fmt.Printf("built at: %s\n", d)
}
