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

package e2e_test

import (
	"context"
	"iter"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
)

// TestCredentialFlagsOnTheWire proves that the canonical Tier-1 credential flags
// (SPEC §12.1 / §7.2) and their environment equivalents attach the correct
// transport credential on the wire:
//   - --bearer / A2ACLI_BEARER   -> Authorization: Bearer <token>
//   - --api-key / A2ACLI_API_KEY -> X-Api-Key: <key> (default header, §12.1)
//   - an explicit flag overrides the environment variable
//   - the CLI never prints the secret value (redaction, §12.1)
//
// The system-under-test is an in-process A2A JSON-RPC server whose Agent Card
// points back at itself, wrapped in a header-capturing middleware so the message
// send request's headers are observed exactly as sent.
func TestCredentialFlagsOnTheWire(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "a2acli")
	build := exec.Command("go", "build", "-o", bin, "../cmd/a2acli")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build a2acli: %v\nOutput:\n%s", err, out)
	}

	cap := &headerCapture{byPath: map[string]http.Header{}}
	srvURL, closeSrv := newCapturingServer(t, cap)
	defer closeSrv()

	runCLI := func(env []string, args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore")
		cmd.Env = append(cmd.Env, env...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	sendPath := "/" // JSON-RPC message send lands on the root handler.

	t.Run("BearerFlag", func(t *testing.T) {
		cap.reset()
		out, err := runCLI(nil, "send", "hi", "-o", "json", "-u", srvURL, "--bearer", "secret-bearer-123")
		if err != nil {
			t.Fatalf("send failed: %v\n%s", err, out)
		}
		if got := cap.get(sendPath).Get("Authorization"); got != "Bearer secret-bearer-123" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer secret-bearer-123")
		}
		if strings.Contains(out, "secret-bearer-123") {
			t.Errorf("secret leaked into CLI output:\n%s", out)
		}
	})

	t.Run("APIKeyFlagDefaultHeader", func(t *testing.T) {
		cap.reset()
		out, err := runCLI(nil, "send", "hi", "-o", "json", "-u", srvURL, "--api-key", "secret-key-abc")
		if err != nil {
			t.Fatalf("send failed: %v\n%s", err, out)
		}
		if got := cap.get(sendPath).Get("X-Api-Key"); got != "secret-key-abc" {
			t.Errorf("X-Api-Key header = %q, want %q", got, "secret-key-abc")
		}
		if strings.Contains(out, "secret-key-abc") {
			t.Errorf("secret leaked into CLI output:\n%s", out)
		}
	})

	t.Run("BearerEnv", func(t *testing.T) {
		cap.reset()
		out, err := runCLI([]string{"A2ACLI_BEARER=env-bearer-xyz"}, "send", "hi", "-o", "json", "-u", srvURL)
		if err != nil {
			t.Fatalf("send failed: %v\n%s", err, out)
		}
		if got := cap.get(sendPath).Get("Authorization"); got != "Bearer env-bearer-xyz" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer env-bearer-xyz")
		}
	})

	t.Run("APIKeyEnv", func(t *testing.T) {
		cap.reset()
		out, err := runCLI([]string{"A2ACLI_API_KEY=env-key-987"}, "send", "hi", "-o", "json", "-u", srvURL)
		if err != nil {
			t.Fatalf("send failed: %v\n%s", err, out)
		}
		if got := cap.get(sendPath).Get("X-Api-Key"); got != "env-key-987" {
			t.Errorf("X-Api-Key header = %q, want %q", got, "env-key-987")
		}
	})

	t.Run("FlagOverridesEnv", func(t *testing.T) {
		cap.reset()
		out, err := runCLI([]string{"A2ACLI_BEARER=from-env"}, "send", "hi", "-o", "json", "-u", srvURL, "--bearer", "from-flag")
		if err != nil {
			t.Fatalf("send failed: %v\n%s", err, out)
		}
		if got := cap.get(sendPath).Get("Authorization"); got != "Bearer from-flag" {
			t.Errorf("explicit flag must override env: Authorization = %q, want %q", got, "Bearer from-flag")
		}
	})
}

// headerCapture records the request headers observed per path.
type headerCapture struct {
	mu     sync.Mutex
	byPath map[string]http.Header
}

func (h *headerCapture) record(path string, hdr http.Header) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byPath[path] = hdr.Clone()
}

func (h *headerCapture) get(path string) http.Header {
	h.mu.Lock()
	defer h.mu.Unlock()
	if v, ok := h.byPath[path]; ok {
		return v
	}
	return http.Header{}
}

func (h *headerCapture) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byPath = map[string]http.Header{}
}

// newCapturingServer starts an in-process A2A JSON-RPC server whose Agent Card
// points back at its own URL, wrapped so every inbound request's headers are
// captured. It returns the server URL and a close function.
func newCapturingServer(t *testing.T, cap *headerCapture) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	selfURL := "http://" + listener.Addr().String()

	card := &a2a.AgentCard{
		Name:         "capturing-mock-agent",
		Description:  "Header-capturing echo agent for auth flag tests",
		Version:      "1.0.0",
		Capabilities: a2a.AgentCapabilities{Streaming: true},
		SupportedInterfaces: []*a2a.AgentInterface{
			a2a.NewAgentInterface(selfURL, a2a.TransportProtocolJSONRPC),
		},
	}

	handler := a2asrv.NewHandler(&echoExec{})
	mux := http.NewServeMux()
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))
	mux.Handle("/", a2asrv.NewJSONRPCHandler(handler))

	capturing := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.record(r.URL.Path, r.Header)
		mux.ServeHTTP(w, r)
	})

	srv := &http.Server{Handler: capturing}
	go func() { _ = srv.Serve(listener) }()
	return selfURL, func() { _ = srv.Close() }
}

// echoExec is a minimal executor that completes a task echoing the input text.
type echoExec struct{}

func (e *echoExec) Execute(_ context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		if execCtx.StoredTask == nil {
			if !yield(a2a.NewSubmittedTask(execCtx, execCtx.Message), nil) {
				return
			}
		}
		evt := a2a.NewArtifactEvent(execCtx, a2a.NewTextPart("ok"))
		evt.LastChunk = true
		if !yield(evt, nil) {
			return
		}
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, nil), nil)
	}
}

func (e *echoExec) Cancel(_ context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCanceled, nil), nil)
	}
}

