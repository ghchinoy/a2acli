package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

	t.Run("JSON-RPC", func(t *testing.T) {
		sutCmd, sutURL, _ := runSUT(t, sutDir, "http")
		defer func() { _ = sutCmd.Process.Kill() }()

		t.Run("Describe", func(t *testing.T) {
			cmd := exec.Command(cliPath, "describe", "--no-tui", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("describe failed: %v\nOutput: %s", err, out)
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
			cmd := exec.Command(cliPath, "send", "hello", "--no-tui", "--wait", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait failed: %v\nOutput: %s", err, out)
			}
			var task map[string]any
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			status := task["status"].(map[string]any)
			if status["state"] != "COMPLETED" {
				t.Errorf("expected COMPLETED, got %v", status["state"])
			}
		})
	})

	t.Run("gRPC", func(t *testing.T) {
		sutCmd, sutURL, _ := runSUT(t, sutDir, "grpc")
		defer func() { _ = sutCmd.Process.Kill() }()

		t.Run("SendWait", func(t *testing.T) {
			// This should auto-select gRPC because the SUT only advertises gRPC in this mode
			cmd := exec.Command(cliPath, "send", "hello", "--no-tui", "--wait", "-u", sutURL)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait (gRPC) failed: %v\nOutput: %s", err, out)
			}
			var task map[string]any
			if err := json.Unmarshal(out, &task); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}
			status := task["status"].(map[string]any)
			if status["state"] != "COMPLETED" {
				t.Errorf("expected COMPLETED, got %v", status["state"])
			}
		})

		t.Run("ForcegRPC", func(t *testing.T) {
			cmd := exec.Command(cliPath, "send", "hello", "--no-tui", "--wait", "-u", sutURL, "--transport", "grpc")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("send --wait --transport grpc failed: %v\nOutput: %s", err, out)
			}
		})
	})
}
