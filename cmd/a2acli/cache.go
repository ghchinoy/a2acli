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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

const cardCacheTTL = 10 * time.Minute

type cachedCard struct {
	URL       string         `json:"url"`
	FetchedAt time.Time      `json:"fetchedAt"`
	Card      *a2a.AgentCard `json:"card"`
}

func getCacheDir() (string, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(userCache, "a2acli", "cards")
	return dir, nil
}

func cacheFilePath(targetURL string) (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(targetURL))
	hashStr := hex.EncodeToString(h[:])
	return filepath.Join(dir, hashStr+".json"), nil
}

func loadCachedCard(targetURL string) (*cachedCard, error) {
	path, err := cacheFilePath(targetURL)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cached cachedCard
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	if cached.Card == nil {
		return nil, os.ErrNotExist
	}
	if time.Since(cached.FetchedAt) > cardCacheTTL {
		return nil, os.ErrNotExist
	}
	return &cached, nil
}

func saveCachedCard(targetURL string, card *a2a.AgentCard) error {
	if card == nil {
		return nil
	}
	path, err := cacheFilePath(targetURL)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	cached := cachedCard{
		URL:       targetURL,
		FetchedAt: time.Now(),
		Card:      card,
	}
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
