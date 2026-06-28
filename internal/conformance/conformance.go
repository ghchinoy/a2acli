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

// Package conformance provides a shared, schema-driven conformance engine used
// by a2acli's conformance surfaces: base A2A smoke checks, the A2UI extension
// validator (internal/conformance/a2ui), and — in the future — ARD catalog and
// registry validation. All surfaces emit the same Result shape so they compose
// and report identically.
package conformance

// Result is the outcome of a single conformance check. It mirrors the shape
// used by cmd/a2acli's live smoke-check command so all conformance output is
// uniform across surfaces.
type Result struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
}

// Pass returns a passing Result.
func Pass(name, msg string) Result { return Result{Name: name, Passed: true, Message: msg} }

// Fail returns a failing Result.
func Fail(name, msg string) Result { return Result{Name: name, Passed: false, Message: msg} }

// Skip returns a skipped Result.
func Skip(name, msg string) Result { return Result{Name: name, Skipped: true, Message: msg} }

// Report is an ordered collection of Results plus an overall pass flag.
type Report struct {
	Results []Result `json:"results"`
	Passed  bool     `json:"passed"`
}

// NewReport builds a Report from results, computing Passed as "no non-skipped
// result failed".
func NewReport(results []Result) Report {
	passed := true
	for _, r := range results {
		if !r.Passed && !r.Skipped {
			passed = false
			break
		}
	}
	return Report{Results: results, Passed: passed}
}
