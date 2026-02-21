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

// Package main provides the entry point for the A2A CLI.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/a2aproject/a2a-go/a2aclient/agentcard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	serviceURL      string
	skillID         string
	authToken       string
	targetTaskID    string
	refTaskID       string
	outDir          string
	instructionFile string
	disableTUI      bool
)

type tokenInterceptor struct {
	a2aclient.PassthroughInterceptor
	token string
}

func (i *tokenInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, any, error) {
	if i.token != "" {
		if req.ServiceParams == nil {
			req.ServiceParams = make(a2aclient.ServiceParams)
		}
		req.ServiceParams["authorization"] = []string{"Bearer " + i.token}
	}
	return ctx, nil, nil
}

func createClient(ctx context.Context, card *a2a.AgentCard) (*a2aclient.Client, error) {
	httpClient := &http.Client{Timeout: 15 * time.Minute}
	opts := []a2aclient.FactoryOption{a2aclient.WithJSONRPCTransport(httpClient)}
	if authToken != "" {
		opts = append(opts, a2aclient.WithCallInterceptors(&tokenInterceptor{token: authToken}))
	}
	return a2aclient.NewFromCard(ctx, card, opts...)
}

func runDescribe(_ *cobra.Command, _ []string) {
	card, err := agentcard.DefaultResolver.Resolve(context.Background(), serviceURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if disableTUI {
		b, err := json.MarshalIndent(card, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Agent: %s\n", card.Name)
	if card.Description != "" {
		fmt.Printf("Description: %s\n", card.Description)
	}

	var formats []string
	seenFormats := make(map[string]bool)
	for _, iface := range card.SupportedInterfaces {
		b := string(iface.ProtocolBinding)
		if b != "" && !seenFormats[b] {
			seenFormats[b] = true
			formats = append(formats, b)
		}
	}
	formatStr := strings.Join(formats, ", ")

	if formatStr != "" {
		fmt.Printf("Supported Bindings: %s\n", formatStr)
	}

	fmt.Printf("Capabilities: [Streaming: %v]\n", card.Capabilities.Streaming)
	fmt.Printf("\nSkills:\n")
	for _, s := range card.Skills {
		fmt.Printf("  - [%s] %s\n", s.ID, s.Name)
		if s.Description != "" {
			fmt.Printf("    Description: %s\n", s.Description)
		}
		if len(s.SecurityRequirements) > 0 {
			var schemes []string
			for _, req := range s.SecurityRequirements {
				for name := range req {
					schemes = append(schemes, string(name))
				}
			}
			fmt.Printf("    Security: %s\n", strings.Join(schemes, ", "))
		}
	}
}

func runInvoke(_ *cobra.Command, args []string) {
	messageText := args[0]

	if instructionFile != "" {
		content, err := os.ReadFile(instructionFile)
		if err != nil {
			log.Fatalf("Error reading instruction file: %v", err)
		}
		messageText = fmt.Sprintf("%s\n\nSupplemental Instructions:\n%s", messageText, string(content))
	}

	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	client, err := createClient(ctx, card)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(messageText))
	if targetTaskID != "" {
		msg.TaskID = a2a.TaskID(targetTaskID)
		if !disableTUI {
			fmt.Printf("Continuing Task: %s\n", targetTaskID)
		}
	}
	if refTaskID != "" {
		msg.ReferenceTasks = []a2a.TaskID{a2a.TaskID(refTaskID)}
		if !disableTUI {
			fmt.Printf("Referencing Task: %s\n", refTaskID)
		}
	}

	params := &a2a.SendMessageRequest{
		Message: msg,
	}
	if skillID != "" {
		params.Metadata = map[string]any{"skillId": skillID}
	}

	if !disableTUI {
		fmt.Printf("Invoking A2A Service (Streaming)...\n\n")
	}

	stream := make(chan streamMsg)
	go func() {
		defer close(stream)
		for event, err := range client.SendStreamingMessage(ctx, params) {
			stream <- streamMsg{Event: event, Err: err}
			if err != nil {
				return
			}
		}
	}()

	if disableTUI {
		runRaw(stream, outDir)
	} else {
		runTUI(stream)
	}
}

func runResume(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	client, err := createClient(ctx, card)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if !disableTUI {
		fmt.Printf("Resuming Task %s ...\n\n", taskID)
	}

	tid := a2a.TaskID(taskID)

	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		errMsg := err.Error()
		if len(errMsg) > 0 {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Hint: If you are using the default in-memory store, restarting the server wipes all tasks.")
			os.Exit(1)
		}
		log.Fatalf("Error retrieving task status: %v", err)
	}

	if task.Status.State == a2a.TaskStateCompleted || task.Status.State == a2a.TaskStateFailed || task.Status.State == a2a.TaskStateRejected {
		displayTaskResult(task, outDir)
		return
	}

	if !disableTUI {
		fmt.Println("Task is active. Connecting to stream...")
	}

	stream := make(chan streamMsg)
	go func() {
		defer close(stream)
		for event, err := range client.SubscribeToTask(ctx, &a2a.SubscribeToTaskRequest{ID: tid}) {
			stream <- streamMsg{Event: event, Err: err}
			if err != nil {
				return
			}
		}
	}()

	if disableTUI {
		runRaw(stream, outDir)
	} else {
		runTUI(stream)
	}
}

func runStatus(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	client, err := createClient(ctx, card)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	tid := a2a.TaskID(taskID)
	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		log.Fatalf("Error retrieving task: %v", err)
	}

	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Task ID: %s\n", task.ID)
	fmt.Printf("Status:  %s\n", task.Status.State)
	if task.Status.Message != nil {
		for _, p := range task.Status.Message.Parts {
			if tp, ok := p.Content.(a2a.Text); ok {
				fmt.Printf("Message: %s\n", string(tp))
			}
		}
	}
	fmt.Printf("Artifacts: %d\n", len(task.Artifacts))

	if len(task.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		for k, v := range task.Metadata {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "a2acli",
		Short: "A2A CLI Client",
	}

	rootCmd.PersistentFlags().StringVarP(&serviceURL, "service-url", "u", "http://127.0.0.1:9001", "Base URL of the A2A service")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Auth token")
	rootCmd.PersistentFlags().StringVarP(&targetTaskID, "task", "k", "", "Existing Task ID to continue (must be non-terminal)")
	rootCmd.PersistentFlags().StringVarP(&refTaskID, "ref", "r", "", "Task ID to reference as context (works for completed tasks)")
	rootCmd.PersistentFlags().BoolVar(&disableTUI, "no-tui", false, "Disable the Terminal UI (useful for scripting and CI)")

	if os.Getenv("A2ACLI_NO_TUI") == "true" || os.Getenv("NO_COLOR") != "" {
		disableTUI = true
	}

	var describeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the agent card",
		Run:   runDescribe,
	}

	var invokeCmd = &cobra.Command{
		Use:   "invoke [message]",
		Short: "Invoke a skill (streaming)",
		Args:  cobra.MinimumNArgs(1),
		Run:   runInvoke,
	}

	var resumeCmd = &cobra.Command{
		Use:   "resume [taskID]",
		Short: "Resume listening to an existing task",
		Args:  cobra.ExactArgs(1),
		Run:   runResume,
	}

	var statusCmd = &cobra.Command{
		Use:   "status [taskID]",
		Short: "Get the status of a task",
		Args:  cobra.ExactArgs(1),
		Run:   runStatus,
	}

	invokeCmd.Flags().StringVarP(&skillID, "skill", "s", "", "Skill ID")
	invokeCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	invokeCmd.Flags().StringVarP(&instructionFile, "instruction-file", "f", "", "Path to a file with supplemental instructions")

	resumeCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")

	rootCmd.AddCommand(describeCmd, invokeCmd, resumeCmd, statusCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(stream chan streamMsg) {
	p := tea.NewProgram(initialModel(stream, outDir))
	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}

	if m, ok := finalModel.(model); ok && m.taskID != "" {
		fmt.Printf("\nTask ID: %s (use --task %s to continue, or --ref %s to reference)\n", m.taskID, m.taskID, m.taskID)
	}
}

func runRaw(stream chan streamMsg, outDir string) {
	for msg := range stream {
		if msg.Err != nil {
			fmt.Fprintf(os.Stderr, "{\"error\": %q}\n", msg.Err.Error())
			os.Exit(1)
		}

		b, err := json.Marshal(msg.Event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "{\"error\": \"failed to encode event to json\"}\n")
			continue
		}
		fmt.Println(string(b))

		if v, ok := msg.Event.(*a2a.TaskArtifactUpdateEvent); ok && outDir != "" {
			_, _ = saveArtifact(outDir, *v.Artifact)
		}
	}
}

func saveArtifact(outDir string, artifact a2a.Artifact) (string, error) {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", err
	}

	filename := artifact.Name
	if filename == "" {
		filename = fmt.Sprintf("artifact_%d.txt", time.Now().Unix())
	}
	path := filepath.Join(outDir, filename)

	var contentBytes []byte
	for _, p := range artifact.Parts {
		if dp, ok := p.Content.(a2a.Data); ok {
			prettyJSON, _ := json.MarshalIndent(dp, "", "  ")
			contentBytes = prettyJSON
		} else if tp, ok := p.Content.(a2a.Text); ok {
			contentBytes = []byte(string(tp))
		}
	}

	if err := os.WriteFile(path, contentBytes, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func displayTaskResult(task *a2a.Task, outDir string) {
	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		if outDir != "" {
			for _, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, *art)
			}
		}
		return
	}

	fmt.Printf("Task Status: [%s]\n", task.Status.State)

	if len(task.Artifacts) == 0 {
		fmt.Println("No artifacts produced.")
		return
	}

	fmt.Printf("\n--- %d ARTIFACT(S) AVAILABLE ---\n", len(task.Artifacts))

	for _, art := range task.Artifacts {
		fmt.Printf("\nName: %s\n", art.Name)
		fmt.Printf("Description: %s\n", art.Description)

		truncated := false
		for _, p := range art.Parts {
			if dp, ok := p.Content.(a2a.Data); ok {
				prettyJSON, _ := json.MarshalIndent(dp, "", "  ")
				fmt.Printf("Data (Preview):\n%s\n", string(prettyJSON))
			} else if tp, ok := p.Content.(a2a.Text); ok {
				preview := string(tp)
				if len(preview) > 500 {
					preview = preview[:500] + "... (truncated)"
					truncated = true
				}
				fmt.Printf("Content (Preview):\n%s\n", preview)
			}
		}

		if outDir != "" {
			path, err := saveArtifact(outDir, *art)
			if err != nil {
				fmt.Printf("Error saving artifact: %v\n", err)
			} else {
				fmt.Printf(">> Saved to: %s\n", path)
			}
		} else if truncated {
			fmt.Println("(Hint: Use --out-dir <path> to save the full artifact content)")
		}
	}
	fmt.Println("\n------------------------------")
}
