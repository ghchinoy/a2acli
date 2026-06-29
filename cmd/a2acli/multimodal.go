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
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// Multi-modal input flag vars wired in main() on sendCmd.
var (
	messagePartsJSON string   // --parts: JSON array of part objects
	messageBodyJSON  string   // --json: complete Message as JSON
	attachFiles      []string // --attach: file paths (repeatable)
	dataArgs         []string // --data: JSON values as DataParts (repeatable)
)

// hasMultimodalInput reports whether any non-text multi-modal flag is set.
// Used by the send Args validator so that --json/--parts/--attach/--data
// satisfy the "some input required" check without a positional text arg.
func hasMultimodalInput() bool {
	return messageBodyJSON != "" || messagePartsJSON != "" ||
		len(attachFiles) > 0 || len(dataArgs) > 0
}

// buildMessage constructs an a2a.Message from all active input flags.
// Priority: --json > --parts > (text + --attach + --data combined).
// textArg is the positional message argument (may be empty).
func buildMessage(textArg string) (*a2a.Message, error) {
	// --json: parse a complete Message object directly.
	if messageBodyJSON != "" {
		var msg a2a.Message
		if err := json.Unmarshal([]byte(messageBodyJSON), &msg); err != nil {
			return nil, fmt.Errorf("--json: invalid Message JSON: %w", err)
		}
		if msg.ID == "" {
			msg.ID = a2a.NewMessageID()
		}
		verboseLog("--json: parsed Message with %d part(s)", len(msg.Parts))
		return &msg, nil
	}

	var parts []*a2a.Part

	// --parts: parse a JSON array of part descriptors.
	if messagePartsJSON != "" {
		parsed, err := parsePartsJSON(messagePartsJSON)
		if err != nil {
			return nil, fmt.Errorf("--parts: %w", err)
		}
		verboseLog("--parts: parsed %d part(s)", len(parsed))
		parts = append(parts, parsed...)
	}

	// positional text arg or stdin content.
	if textArg != "" {
		parts = append(parts, a2a.NewTextPart(textArg))
	}

	// --attach: read each file and create a part with auto-detected MIME type.
	for _, path := range attachFiles {
		part, err := fileAttachPart(path)
		if err != nil {
			return nil, fmt.Errorf("--attach %q: %w", path, err)
		}
		verboseLog("--attach: %s (%s, %d bytes)", filepath.Base(path), part.MediaType, len(part.Content.(a2a.Raw)))
		parts = append(parts, part)
	}

	// --data: parse each JSON value and create a DataPart.
	for i, d := range dataArgs {
		var v any
		if err := json.Unmarshal([]byte(d), &v); err != nil {
			return nil, fmt.Errorf("--data[%d]: invalid JSON: %w", i, err)
		}
		parts = append(parts, a2a.NewDataPart(v))
		verboseLog("--data[%d]: DataPart added", i)
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("no message content: provide text, pipe via stdin, or use --json/--parts/--attach/--data")
	}

	return a2a.NewMessage(a2a.MessageRoleUser, parts...), nil
}

// jsonPartDescriptor is the wire shape for a single part in a --parts JSON array.
// Mirrors the a2a Part types: exactly one of Text/Data/Raw/URL should be set.
type jsonPartDescriptor struct {
	Text      *string        `json:"text"`
	Data      any            `json:"data"`
	Raw       *string        `json:"raw"` // base64-encoded bytes
	URL       *string        `json:"url"`
	MediaType string         `json:"mediaType"`
	Metadata  map[string]any `json:"metadata"`
	Filename  string         `json:"filename"`
}

// parsePartsJSON decodes a JSON array of part descriptors into []*a2a.Part.
func parsePartsJSON(raw string) ([]*a2a.Part, error) {
	var descriptors []jsonPartDescriptor
	if err := json.Unmarshal([]byte(raw), &descriptors); err != nil {
		return nil, fmt.Errorf("invalid JSON parts array: %w", err)
	}
	parts := make([]*a2a.Part, 0, len(descriptors))
	for i, d := range descriptors {
		p, err := descriptorToPart(d)
		if err != nil {
			return nil, fmt.Errorf("part[%d]: %w", i, err)
		}
		parts = append(parts, p)
	}
	return parts, nil
}

func descriptorToPart(d jsonPartDescriptor) (*a2a.Part, error) {
	switch {
	case d.Text != nil:
		p := a2a.NewTextPart(*d.Text)
		p.MediaType = d.MediaType
		p.Metadata = d.Metadata
		return p, nil

	case d.Data != nil:
		p := a2a.NewDataPart(d.Data)
		p.MediaType = d.MediaType
		p.Metadata = d.Metadata
		return p, nil

	case d.Raw != nil:
		p := a2a.NewRawPart([]byte(*d.Raw))
		p.MediaType = d.MediaType
		p.Metadata = d.Metadata
		p.Filename = d.Filename
		return p, nil

	case d.URL != nil:
		p := a2a.NewFileURLPart(a2a.URL(*d.URL), d.MediaType)
		p.Metadata = d.Metadata
		p.Filename = d.Filename
		return p, nil

	default:
		return nil, fmt.Errorf("part must have one of: text, data, raw, url")
	}
}

// fileAttachPart reads a file from disk and returns a RawPart with an
// auto-detected (or extension-inferred) MIME type.
func fileAttachPart(path string) (*a2a.Part, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	mediaType := detectMIME(path, data)
	p := a2a.NewRawPart(data)
	p.MediaType = mediaType
	p.Filename = filepath.Base(path)
	return p, nil
}

// detectMIME returns a MIME type for the given filename and content.
// Prefers extension-based detection; falls back to content sniffing.
func detectMIME(filename string, data []byte) string {
	if ext := filepath.Ext(filename); ext != "" {
		if t := mime.TypeByExtension(strings.ToLower(ext)); t != "" {
			// Strip parameters (e.g. "text/plain; charset=utf-8" → "text/plain")
			if idx := strings.Index(t, ";"); idx >= 0 {
				return strings.TrimSpace(t[:idx])
			}
			return t
		}
	}
	return http.DetectContentType(data)
}
