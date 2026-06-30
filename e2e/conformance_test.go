package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for server at %s", url)
}

func runSUT(t *testing.T, sutDir string, mode string) (*exec.Cmd, string, *bytes.Buffer) {
	sutCmd := exec.Command("go", "run", "sut.go", "sut_agent_executor.go", "-mode", mode)
	sutCmd.Dir = sutDir
	var sutOut bytes.Buffer
	sutCmd.Stdout = &sutOut
	sutCmd.Stderr = &sutOut
	if err := sutCmd.Start(); err != nil {
		t.Fatalf("failed to start SUT (%s): %v", mode, err)
	}

	sutURL := "http://127.0.0.1:9999"
	if err := waitForServer(sutURL+"/agent", 10*time.Second); err != nil {
		_ = sutCmd.Process.Kill()
		t.Fatalf("Server (%s) failed to start. Logs:\n%s", mode, sutOut.String())
	}
	return sutCmd, sutURL, &sutOut
}

func TestConformance(t *testing.T) {
	cmdBuild := exec.Command("go", "build", "-o", "../bin/a2acli", "../cmd/a2acli")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to build a2acli: %v\nOutput:\n%s", err, string(out))
	}

	a2aGoSrc := os.Getenv("A2A_GO_SRC")
	if a2aGoSrc == "" {
		a2aGoSrc = "../../github/a2a-go"
	}

	sutDir := a2aGoSrc + "/e2e/tck"
	if _, err := os.Stat(sutDir); os.IsNotExist(err) {
		t.Fatalf("a2a-go SDK source not found at %s", a2aGoSrc)
	}

	cliPath := "../bin/a2acli"

	runCLI := func(args ...string) *exec.Cmd {
		cmd := exec.Command(cliPath, args...)
		cmd.Env = append(os.Environ(), "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore")
		return cmd
	}

	t.Run("JSON-RPC", func(t *testing.T) {
		sutCmd, sutURL, _ := runSUT(t, sutDir, "http")
		defer func() { _ = sutCmd.Process.Kill() }()

		t.Run("Describe", func(t *testing.T) {
			cmd := runCLI("discover", "--output", "json", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("discover failed: %v\nOutput: %s", err, out)
			}
			var card map[string]any
			if err := json.Unmarshal(out, &card); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			if name, _ := card["name"].(string); name != "TCK Core Agent" {
				t.Errorf("expected TCK Core Agent, got %v", name)
			}
		})

		t.Run("SendWait", func(t *testing.T) {
			cmd := runCLI("send", "hello", "--no-tui", "--wait", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait failed: %v\nOutput: %s", err, out)
			}
			var task map[string]any
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			status := task["status"].(map[string]any)
			if status["state"] != "TASK_STATE_COMPLETED" {
				t.Errorf("expected TASK_STATE_COMPLETED, got %v", status["state"])
			}
		})

		t.Run("SendStdin", func(t *testing.T) {
			// Pipe message via stdin — no positional arg.
			cmd := exec.Command(cliPath, "send", "--output", "json", "--wait", "-u", sutURL)
			cmd.Env = append(os.Environ(), "GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore")
			cmd.Stdin = bytes.NewBufferString("hello from stdin")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send via stdin failed: %v\nOutput: %s", err, out)
			}
			var task map[string]any
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse JSON from stdin send: %v\nOutput: %s", err, out)
			}
			status := task["status"].(map[string]any)
			if status["state"] != "TASK_STATE_COMPLETED" {
				t.Errorf("expected TASK_STATE_COMPLETED, got %v", status["state"])
			}
		})

		t.Run("ConformanceSmoke", func(t *testing.T) {
			// Run the conformance command in JSON mode and assert all checks passed.
			cmd := runCLI("conformance", "--output", "json", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("conformance command failed: %v\nOutput: %s", err, out)
			}
			var result struct {
				Passed  bool `json:"passed"`
				Results []struct {
					Name    string `json:"name"`
					Passed  bool   `json:"passed"`
					Skipped bool   `json:"skipped"`
					Message string `json:"message"`
				} `json:"results"`
			}
			if err := json.Unmarshal(out, &result); err != nil {
				t.Fatalf("failed to parse conformance JSON: %v\nOutput: %s", err, out)
			}
			if !result.Passed {
				for _, r := range result.Results {
					if !r.Passed && !r.Skipped {
						t.Errorf("conformance check %q failed: %s", r.Name, r.Message)
					}
				}
			}
		})
	})

	t.Run("gRPC", func(t *testing.T) {
		sutCmd, sutURL, _ := runSUT(t, sutDir, "grpc")
		defer func() { _ = sutCmd.Process.Kill() }()

		t.Run("SendWait", func(t *testing.T) {
			// This should auto-select gRPC because the SUT only advertises gRPC in this mode
			cmd := runCLI("send", "hello", "--no-tui", "--wait", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait (gRPC) failed: %v\nOutput: %s", err, out)
			}
			var task map[string]any
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			status := task["status"].(map[string]any)
			if status["state"] != "TASK_STATE_COMPLETED" {
				t.Errorf("expected TASK_STATE_COMPLETED, got %v", status["state"])
			}
		})

		t.Run("ForcegRPC", func(t *testing.T) {
			cmd := runCLI("send", "hello", "--no-tui", "--wait", "-u", sutURL, "--transport", "grpc")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait --transport grpc failed: %v\nOutput: %s", err, out)
			}
		})
	})

	t.Run("A2A-0.3.0", func(t *testing.T) {
		// Run the 0.3.0 compat server from the SDK
		compatSutDir := a2aGoSrc + "/e2e/compat/v0_3"
		if _, err := os.Stat(compatSutDir); os.IsNotExist(err) {
			t.Skipf("0.3.0 compat SUT not found at %s", compatSutDir)
		}

		sutCmd := exec.Command("go", "run", "main.go", "server")
		sutCmd.Dir = compatSutDir
		var sutOut bytes.Buffer
		sutCmd.Stdout = &sutOut
		sutCmd.Stderr = &sutOut
		if err := sutCmd.Start(); err != nil {
			t.Fatalf("failed to start 0.3.0 SUT: %v", err)
		}
		defer func() { _ = sutCmd.Process.Kill() }()

		// The server prints the port to stdout. We need to capture it.
		// Wait a bit for it to start and print.
		time.Sleep(2 * time.Second)
		portStr := sutOut.String()
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			t.Fatalf("failed to parse 0.3.0 SUT port from %q: %v", portStr, err)
		}
		if port == 0 {
			t.Fatalf("failed to capture 0.3.0 SUT port. Output:\n%s", portStr)
		}

		sutURL := fmt.Sprintf("http://127.0.0.1:%d", port)

		t.Run("Describe", func(t *testing.T) {
			cmd := runCLI("discover", "--output", "json", "-u", sutURL, "--protocol", "0.3.0")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("discover 0.3.0 failed: %v\nOutput: %s", err, out)
			}
			var card map[string]any
			if err := json.Unmarshal(out, &card); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			if name, _ := card["name"].(string); name != "Compat Test Agent" {
				t.Errorf("expected Compat Test Agent, got %v", name)
			}
		})

		t.Run("SendWait", func(t *testing.T) {
			cmd := runCLI("send", "ping", "--no-tui", "--wait", "-u", sutURL, "--protocol", "0.3.0")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait 0.3.0 failed: %v\nOutput: %s", err, out)
			}
			// 0.3.0 server in this mode returns a Message directly if non-blocking
			// but SendMessage in a2acli --wait should handle it.
			// Actually the compat server responds with a Message.
			var result map[string]any
			if err := json.Unmarshal(out, &result); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			// Check if it's a message or task
			if _, ok := result["messageId"]; !ok {
				t.Errorf("expected Message response, got: %v", result)
			}
		})
	})

	t.Run("A2UI-Extension-v1.0", func(t *testing.T) {
		// Run the apex A2UI sample server.
		// Respect APEX_SRC env var if set (for CI), else check sibling checkouts.
		apexSutDir := os.Getenv("APEX_SRC")
		if apexSutDir == "" {
			apexSutDir = "../../../apex"
		}
		apexSutDir = filepath.Join(apexSutDir, "tools/sample_servers/a2a_a2ui")

		if _, err := os.Stat(apexSutDir); os.IsNotExist(err) {
			// Try absolute path under home
			home, _ := os.UserHomeDir()
			apexSutDir = filepath.Join(home, "projects/apex/tools/sample_servers/a2a_a2ui")
			if _, err := os.Stat(apexSutDir); os.IsNotExist(err) {
				t.Skipf("apex A2UI sample server not found at %s", apexSutDir)
			}
		}

		sutCmd := exec.Command("go", "run", ".")
		sutCmd.Dir = apexSutDir
		var sutOut bytes.Buffer
		sutCmd.Stdout = &sutOut
		sutCmd.Stderr = &sutOut
		if err := sutCmd.Start(); err != nil {
			t.Fatalf("failed to start apex A2UI SUT: %v", err)
		}
		defer func() { _ = sutCmd.Process.Kill() }()

		sutURL := "http://127.0.0.1:9002"
		if err := waitForServer(sutURL+"/a2a", 10*time.Second); err != nil {
			t.Fatalf("apex A2UI SUT failed to start. Logs:\n%s", sutOut.String())
		}

		t.Run("Validate", func(t *testing.T) {
			// apex-c4x is resolved: the server now emits fully conformant A2UI v1.0.
			// Assert all checks pass and the command exits 0.
			cmd := runCLI("a2ui", "validate", "--output", "json", "-u", sutURL, "--probe", "show me a showcase card about cats")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("a2ui validate failed unexpectedly: %v\nOutput: %s", err, out)
			}

			var report struct {
				Passed  bool `json:"passed"`
				Results []struct {
					Name    string `json:"name"`
					Passed  bool   `json:"passed"`
					Skipped bool   `json:"skipped"`
					Message string `json:"message"`
				} `json:"results"`
			}
			if err := json.Unmarshal(out, &report); err != nil {
				t.Fatalf("failed to parse validation JSON: %v\nOutput: %s", err, out)
			}

			if !report.Passed {
				for _, r := range report.Results {
					if !r.Passed && !r.Skipped {
						t.Errorf("unexpected FAIL: %s — %s", r.Name, r.Message)
					}
				}
			}
		})
	})

	// A2A-Simple multi-transport echo: a deterministic sister-agent fixture
	// (a2a-simple `grpc-echo`) that serves JSON-RPC, REST/HTTP+JSON, and gRPC
	// from a single process. Unlike the TCK SUT (one transport per mode), this
	// validates a2acli's transport selection and the three bindings against the
	// same agent simultaneously. See docs/TEST_AGENTS.md (a2ac-k9i).
	t.Run("A2A-Simple-MultiTransport", func(t *testing.T) {
		simpleSrc := os.Getenv("A2A_SIMPLE_SRC")
		if simpleSrc == "" {
			simpleSrc = "../../a2a-simple"
		}
		if _, err := os.Stat(simpleSrc); os.IsNotExist(err) {
			home, _ := os.UserHomeDir()
			simpleSrc = filepath.Join(home, "projects/a2a-simple")
			if _, err := os.Stat(simpleSrc); os.IsNotExist(err) {
				t.Skipf("a2a-simple source not found (set A2A_SIMPLE_SRC)")
			}
		}

		// Build the fixture to a temp binary and exec it directly. Unlike `go run`,
		// this lets Process.Kill() actually stop the server (go run orphans its
		// child), preventing port leaks across local reruns.
		echoBin := filepath.Join(t.TempDir(), "grpc-echo")
		buildEcho := exec.Command("go", "build", "-o", echoBin, "./cmd/grpc-echo")
		buildEcho.Dir = simpleSrc
		if out, err := buildEcho.CombinedOutput(); err != nil {
			t.Fatalf("failed to build a2a-simple grpc-echo: %v\nOutput: %s", err, out)
		}

		// Use dedicated ports to avoid colliding with other fixtures (apex uses 9002).
		const httpPort, grpcPort = 9014, 9015
		sutCmd := exec.Command(echoBin,
			"-port", fmt.Sprintf("%d", httpPort),
			"-grpc-port", fmt.Sprintf("%d", grpcPort))
		sutCmd.Dir = simpleSrc
		var sutOut bytes.Buffer
		sutCmd.Stdout = &sutOut
		sutCmd.Stderr = &sutOut
		if err := sutCmd.Start(); err != nil {
			t.Fatalf("failed to start a2a-simple grpc-echo: %v", err)
		}
		defer func() { _ = sutCmd.Process.Kill() }()

		sutURL := fmt.Sprintf("http://127.0.0.1:%d", httpPort)
		if err := waitForServer(sutURL+"/.well-known/agent-card.json", 20*time.Second); err != nil {
			t.Fatalf("grpc-echo failed to start. Logs:\n%s", sutOut.String())
		}

		t.Run("Discover", func(t *testing.T) {
			cmd := runCLI("discover", "--output", "json", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("discover failed: %v\nOutput: %s", err, out)
			}
			var card struct {
				SupportedInterfaces []struct {
					ProtocolBinding string `json:"protocolBinding"`
				} `json:"supportedInterfaces"`
			}
			if err := json.Unmarshal(out, &card); err != nil {
				t.Fatalf("failed to parse card JSON: %v\nOutput: %s", err, out)
			}
			got := map[string]bool{}
			for _, i := range card.SupportedInterfaces {
				got[i.ProtocolBinding] = true
			}
			for _, want := range []string{"JSONRPC", "HTTP+JSON", "GRPC"} {
				if !got[want] {
					t.Errorf("expected card to advertise %s transport; got %v", want, got)
				}
			}
		})

		// assertEchoCompleted runs `send` and asserts the task completed with the
		// deterministic echo artifact (part-0-text).
		assertEchoCompleted := func(t *testing.T, transport string) {
			t.Helper()
			args := []string{"send", "hello-" + transport, "--output", "json", "--wait", "-u", sutURL}
			if transport != "" {
				args = append(args, "--transport", transport)
			}
			cmd := runCLI(args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send (%s) failed: %v\nOutput: %s", transport, err, out)
			}
			// --wait emits a single final task object.
			var task struct {
				Status struct {
					State string `json:"state"`
				} `json:"status"`
				Artifacts []struct {
					Name string `json:"name"`
				} `json:"artifacts"`
			}
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse task JSON (%s): %v\nOutput: %s", transport, err, out)
			}
			if task.Status.State != "TASK_STATE_COMPLETED" {
				t.Errorf("[%s] expected TASK_STATE_COMPLETED, got %q", transport, task.Status.State)
			}
			found := false
			for _, a := range task.Artifacts {
				if a.Name == "part-0-text" {
					found = true
				}
			}
			if !found {
				t.Errorf("[%s] expected artifact 'part-0-text', got %+v", transport, task.Artifacts)
			}
		}

		t.Run("JSONRPC", func(t *testing.T) { assertEchoCompleted(t, "jsonrpc") })
		t.Run("REST", func(t *testing.T) { assertEchoCompleted(t, "rest") })
		t.Run("gRPC", func(t *testing.T) { assertEchoCompleted(t, "grpc") })
	})
}
