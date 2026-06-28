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

package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SchemaSet is a reusable, compiled collection of JSON Schemas keyed by the URL
// that other schemas reference them under. It wraps santhosh-tekuri/jsonschema
// (draft 2020-12) and is the generic validation primitive shared by all
// conformance surfaces.
type SchemaSet struct {
	compiler *jsonschema.Compiler
	compiled map[string]*jsonschema.Schema
}

// RawSchema is a vendored schema document plus the URL(s) it should be
// registered under. Aliases let us satisfy upstream refs whose target filename
// differs from the actual vendored file (e.g. server_to_client.json refs
// "catalog.json" but the file is "catalog_definition.json").
type RawSchema struct {
	URL     string   // canonical registration URL (usually the schema's $id)
	Aliases []string // additional URLs that resolve to this same document
	Data    []byte   // the raw JSON Schema bytes
}

// NewSchemaSet compiles the provided schemas into a reusable set. Each schema is
// registered under its URL and any aliases so cross-file $refs resolve offline.
func NewSchemaSet(schemas []RawSchema) (*SchemaSet, error) {
	c := jsonschema.NewCompiler()

	register := func(url string, data []byte) error {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("parse schema %s: %w", url, err)
		}
		if err := c.AddResource(url, doc); err != nil {
			return fmt.Errorf("add schema resource %s: %w", url, err)
		}
		return nil
	}

	for _, s := range schemas {
		if err := register(s.URL, s.Data); err != nil {
			return nil, err
		}
		for _, alias := range s.Aliases {
			// The embedded $id takes precedence over the registration URL in
			// jsonschema/v6, so a raw re-registration under a different URL is
			// ignored. Rewrite $id to the alias URL so the document is indexed
			// where dangling cross-file refs expect to find it.
			aliased, err := rewriteSchemaID(s.Data, alias)
			if err != nil {
				return nil, fmt.Errorf("alias %s: %w", alias, err)
			}
			if err := register(alias, aliased); err != nil {
				return nil, err
			}
		}
	}

	return &SchemaSet{compiler: c, compiled: make(map[string]*jsonschema.Schema)}, nil
}

// rewriteSchemaID returns a copy of the schema JSON with its top-level $id set to
// newID, so the document can be registered (and resolved) under an alias URL.
func rewriteSchemaID(data []byte, newID string) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse for $id rewrite: %w", err)
	}
	doc["$id"] = newID
	return json.Marshal(doc)
}

// Validate checks an already-decoded JSON value (map[string]any, []any, etc.)
// against the schema registered at schemaURL. A nil error means valid.
func (s *SchemaSet) Validate(schemaURL string, value any) error {
	sch, ok := s.compiled[schemaURL]
	if !ok {
		var err error
		sch, err = s.compiler.Compile(schemaURL)
		if err != nil {
			return fmt.Errorf("compile %s: %w", schemaURL, err)
		}
		s.compiled[schemaURL] = sch
	}
	return sch.Validate(value)
}
