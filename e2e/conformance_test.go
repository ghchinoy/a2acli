package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
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

func TestConformance(t *testing.T) {
	cmdBuild := exec.Command("go", "build", "-o", "../bin/a2acli", "../cmd/a2acli")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to build a2acli: %v\nOutput:\n%s", err, string(out))
	}

	a2aGoSrc := os.Getenv("A2A_GO_SRC")
	if a2aGoSrc == "" {
		// Default to relative path for standard local development
		a2aGoSrc = "../../github/a2a-go"
	}

	sutDir := a2aGoSrc + "/e2e/tck"
	if _, err := os.Stat(sutDir); os.IsNotExist(err) {
		t.Fatalf("\n\n❌ REQUIRED DEPENDENCY MISSING ❌\n\nThe a2a-go SDK source code was not found at:\n'%s'\n\nRunning e2e conformance tests requires the a2a-go source to be checked out locally so the TCK SUT server can be spun up.\n\nPlease clone 'https://github.com/a2aproject/a2a-go' or provide the correct path using:\n\n    make test-e2e A2A_GO_SRC=/path/to/a2a-go\n\n", a2aGoSrc)
	}

	sutCmd := exec.Command("go", "run", "sut.go", "sut_agent_executor.go")
	sutCmd.Dir = sutDir
	var sutOut bytes.Buffer
	sutCmd.Stdout = &sutOut
	sutCmd.Stderr = &sutOut
	if err := sutCmd.Start(); err != nil {
		t.Fatalf("failed to start SUT: %v", err)
	}

	defer func() {
		if sutCmd.Process != nil {
			_ = sutCmd.Process.Kill()
		}
	}()

	sutURL := "http://127.0.0.1:9999"
	if err := waitForServer(sutURL+"/agent", 5*time.Second); err != nil {
		t.Fatalf("Server failed to start. Logs:\n%s", sutOut.String())
	}

	cliPath := "../bin/a2acli"

	t.Run("Describe", func(t *testing.T) {
		cmd := exec.Command(cliPath, "describe", "--no-tui", "-u", sutURL)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("describe failed: %v\nOutput: %s", err, out)
		}

		var card map[string]any
		if err := json.Unmarshal(out, &card); err != nil {
			t.Fatalf("failed to parse JSON from describe: %v\nOutput: %s", err, out)
		}

		if name, _ := card["name"].(string); name != "TCK Core Agent" {
			t.Errorf("expected name 'TCK Core Agent', got %v", name)
		}
	})

	t.Run("Send", func(t *testing.T) {
		cmd := exec.Command(cliPath, "send", "hello", "--no-tui", "-u", sutURL)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("invoke failed: %v\nOutput: %s", err, out)
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")

		var states []string
		for _, line := range lines {
			if strings.HasPrefix(line, "Task ID:") {
				continue
			}
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}

			if statusBlock, ok := event["status"].(map[string]any); ok {
				if state, ok := statusBlock["state"].(string); ok {
					states = append(states, state)
				}
			}
		}

		if len(states) < 3 {
			t.Fatalf("Expected at least 3 state updates, got %d. States: %v\nOutput:\n%s", len(states), states, out)
		}

		expected := []string{"SUBMITTED", "WORKING", "COMPLETED"}

		tail := states[len(states)-3:]
		for i, exp := range expected {
			if tail[i] != exp {
				t.Errorf("Expected state[%d] = %s, got %s", i, exp, tail[i])
			}
		}
	})
	t.Run("SendWait", func(t *testing.T) {
		cmd := exec.Command(cliPath, "send", "hello", "--no-tui", "--wait", "-u", sutURL)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("send --wait failed: %v\nOutput: %s", err, out)
		}

		var task map[string]any
		if err := json.Unmarshal(out, &task); err != nil {
			t.Fatalf("failed to parse JSON from describe: %v\nOutput: %s", err, out)
		}

		if statusBlock, ok := task["status"].(map[string]any); ok {
			if state, ok := statusBlock["state"].(string); ok {
				if state != "COMPLETED" {
					t.Fatalf("Expected state COMPLETED, got %s", state)
				}
			} else {
				t.Fatalf("No state string found in status block")
			}
		} else {
			t.Fatalf("No status block found in JSON payload")
		}
	})
}
