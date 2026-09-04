package e2e_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTier1CLIContract locks the headline Tier-1 CLI contract behaviors
// (SPEC §11) end-to-end through the built binary:
//   - usage errors (unknown command / unknown flag) exit 2
//   - default `send -o json` emits a SINGLE json document (blocking)
//   - `send --stream -o json` emits JSONL (multiple independently-parseable lines)
//
// It is self-contained: the system-under-test is a2acli's own `serve --echo`
// mock, so it needs no external SDK sources.
func TestTier1CLIContract(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "a2acli")
	build := exec.Command("go", "build", "-o", bin, "../cmd/a2acli")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build a2acli: %v\nOutput:\n%s", err, out)
	}

	runCLI := func(args ...string) *exec.Cmd {
		cmd := exec.Command(bin, args...)
		cmd.Env = append(os.Environ(), "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore")
		return cmd
	}

	// exitCode extracts the process exit code from a completed command run.
	exitCode := func(err error) int {
		if err == nil {
			return 0
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		return -1
	}

	t.Run("UsageErrorsExit2", func(t *testing.T) {
		cases := []struct {
			name string
			args []string
		}{
			{"UnknownCommand", []string{"boguscmd"}},
			{"UnknownFlag", []string{"discover", "--totally-bogus-flag"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				out, err := runCLI(tc.args...).CombinedOutput()
				if got := exitCode(err); got != 2 {
					t.Errorf("expected exit 2 for %v, got %d\nOutput:\n%s", tc.args, got, out)
				}
			})
		}

		// Sanity: --version and --help are not usage errors (exit 0).
		if _, err := runCLI("--version").CombinedOutput(); exitCode(err) != 0 {
			t.Errorf("expected exit 0 for --version, got %d", exitCode(err))
		}
	})

	// Start the self-contained echo mock for the send-shape assertions.
	const port = 9033
	sutURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	serve := runCLI("serve", "--echo", "--port", fmt.Sprintf("%d", port))
	var serveOut bytes.Buffer
	serve.Stdout = &serveOut
	serve.Stderr = &serveOut
	if err := serve.Start(); err != nil {
		t.Fatalf("failed to start serve --echo: %v", err)
	}
	defer func() { _ = serve.Process.Kill() }()

	if err := waitForServer(sutURL+"/", 10*time.Second); err != nil {
		t.Fatalf("echo server failed to start. Logs:\n%s", serveOut.String())
	}

	t.Run("DefaultSendIsSingleJSONDoc", func(t *testing.T) {
		out, err := runCLI("send", "hello", "--output", "json", "-u", sutURL).CombinedOutput()
		if err != nil {
			t.Fatalf("send (default) failed: %v\nOutput:\n%s", err, out)
		}
		// A single JSON document: the whole output must parse as exactly one
		// top-level value. Trailing data (a second doc / JSONL) makes
		// json.Unmarshal fail with "invalid character after top-level value".
		var doc map[string]any
		if err := json.Unmarshal(out, &doc); err != nil {
			t.Fatalf("default send did not emit a single JSON document: %v\nOutput:\n%s", err, out)
		}
		// SPEC §11.3 (L421) / Appendix B (L567): `send` MUST emit the App-B
		// SendMessageResponse wrapper — the terminal object under "task" (a task
		// was created) or "message" — never a bare Task/Message.
		if _, hasTask := doc["task"]; !hasTask {
			if _, hasMsg := doc["message"]; !hasMsg {
				t.Fatalf("send -o json must use the App-B SendMessageResponse wrapper (top-level \"task\" or \"message\"); got keys %v\nOutput:\n%s", keysOf(doc), out)
			}
		}
	})

	t.Run("StreamSendIsJSONL", func(t *testing.T) {
		out, err := runCLI("send", "hello", "--stream", "--output", "json", "-u", sutURL).CombinedOutput()
		if err != nil {
			t.Fatalf("send --stream failed: %v\nOutput:\n%s", err, out)
		}
		lines := nonEmptyLines(out)
		if len(lines) < 2 {
			t.Fatalf("expected JSONL (>=2 lines) from --stream, got %d line(s)\nOutput:\n%s", len(lines), out)
		}
		for i, line := range lines {
			var v any
			if err := json.Unmarshal([]byte(line), &v); err != nil {
				t.Errorf("JSONL line %d is not independently parseable: %v\nLine: %s", i, err, line)
			}
		}
		// The whole blob must NOT parse as a single document — that would mean
		// the stream collapsed back to blocking behavior.
		var single any
		if err := json.Unmarshal(out, &single); err == nil {
			t.Errorf("--stream output parsed as a single JSON document; expected JSONL\nOutput:\n%s", out)
		}
	})
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func nonEmptyLines(b []byte) []string {
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}
