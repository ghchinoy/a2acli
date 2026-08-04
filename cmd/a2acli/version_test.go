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
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	origV, origC, origD := version, commit, date
	defer func() {
		version, commit, date = origV, origC, origD
	}()

	version = "1.9.0"
	commit = "abc1234"
	date = "2026-08-04T00:00:00Z"

	v, c, d := getVersionInfo()
	if v != "1.9.0" || c != "abc1234" || d != "2026-08-04T00:00:00Z" {
		t.Errorf("expected buildtime variables to take precedence, got v=%q, c=%q, d=%q", v, c, d)
	}

	// Test fallback path when build variables are default/empty
	version = "dev"
	commit = "none"
	date = "unknown"

	v, c, d = getVersionInfo()
	if v == "" || c == "" || d == "" {
		t.Errorf("getVersionInfo() returned empty values: v=%q, c=%q, d=%q", v, c, d)
	}
}
