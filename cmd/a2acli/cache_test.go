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
	"os"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestCacheSaveAndLoad(t *testing.T) {
	tmpCache, err := os.MkdirTemp("", "a2acli-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp cache dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpCache) }()

	t.Setenv("XDG_CACHE_HOME", tmpCache)

	targetURL := "http://127.0.0.1:9999"
	card := &a2a.AgentCard{
		Name:        "Test Agent",
		Version:     "1.0.0",
		Description: "Cache test agent",
	}

	// Save to cache
	if err := saveCachedCard(targetURL, card); err != nil {
		t.Fatalf("saveCachedCard failed: %v", err)
	}

	// Load from cache
	loaded, err := loadCachedCard(targetURL)
	if err != nil {
		t.Fatalf("loadCachedCard failed: %v", err)
	}

	if loaded.Card == nil || loaded.Card.Name != "Test Agent" {
		t.Errorf("unexpected cached card: %+v", loaded.Card)
	}
}

func TestCacheExpiration(t *testing.T) {
	tmpCache, err := os.MkdirTemp("", "a2acli-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp cache dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpCache) }()

	t.Setenv("XDG_CACHE_HOME", tmpCache)

	targetURL := "http://127.0.0.1:9998"
	card := &a2a.AgentCard{
		Name: "Expired Agent",
	}

	// Save card with old timestamp directly
	path, err := cacheFilePath(targetURL)
	if err != nil {
		t.Fatalf("cacheFilePath failed: %v", err)
	}

	oldCached := cachedCard{
		URL:       targetURL,
		FetchedAt: time.Now().Add(-15 * time.Minute), // expired
		Card:      card,
	}

	if err := saveCachedCard(targetURL, card); err != nil {
		t.Fatalf("saveCachedCard failed: %v", err)
	}

	// Overwrite with old timestamp
	_ = saveCachedCard(targetURL, card)
	oldData := []byte(`{"url":"` + targetURL + `","fetchedAt":"2026-08-01T00:00:00Z","card":{"name":"Expired Agent"}}`)
	_ = os.WriteFile(path, oldData, 0600)

	_ = oldCached

	// Attempt load
	_, err = loadCachedCard(targetURL)
	if err == nil {
		t.Errorf("expected error for expired cache entry, got nil")
	}
}
