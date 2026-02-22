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
	a2agrpc "github.com/a2aproject/a2a-go/a2agrpc/v1"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	serviceURL      string
	skillID         string
	authToken       string
	targetTaskID    string
	refTaskID       string
	outDir          string
	outFile         string
	instructionFile string
	disableTUI      bool
	wait            bool
	transport       string

	rootCmd = &cobra.Command{
		Use:   "a2acli",
		Short: "A2A CLI Client",
	}

	// Command group IDs for help organization
	GroupDiscovery = "discovery"
	GroupMessaging = "messaging"
	GroupSystem    = "system"
)

func fatalf(format string, err error, hint string) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", err)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "Hint: %s\n", hint)
	}
	os.Exit(1)
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: GroupDiscovery, Title: "Discovery & Identity:"},
		&cobra.Group{ID: GroupMessaging, Title: "Messaging & Tasks:"},
		&cobra.Group{ID: GroupSystem, Title: "Client Configuration:"},
	)
	rootCmd.SetHelpFunc(colorizedHelpFunc)
}

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

	// Determine transport
	selectedTransport := a2a.TransportProtocolJSONRPC // Default
	if transport != "" {
		switch strings.ToLower(transport) {
		case "grpc":
			selectedTransport = a2a.TransportProtocolGRPC
		case "jsonrpc":
			selectedTransport = a2a.TransportProtocolJSONRPC
		case "rest", "httpjson":
			selectedTransport = a2a.TransportProtocolHTTPJSON
		default:
			return nil, fmt.Errorf("unsupported transport: %s", transport)
		}
	} else {
		// Dynamic selection based on priority: gRPC > JSON-RPC > HTTP+JSON
		available := make(map[a2a.TransportProtocol]bool)
		for _, iface := range card.SupportedInterfaces {
			available[iface.ProtocolBinding] = true
		}

		if available[a2a.TransportProtocolGRPC] {
			selectedTransport = a2a.TransportProtocolGRPC
		} else if available[a2a.TransportProtocolJSONRPC] {
			selectedTransport = a2a.TransportProtocolJSONRPC
		} else if available[a2a.TransportProtocolHTTPJSON] {
			selectedTransport = a2a.TransportProtocolHTTPJSON
		}
	}

	var transportOpt a2aclient.FactoryOption
	switch selectedTransport {
	case a2a.TransportProtocolGRPC:
		transportOpt = a2agrpc.WithGRPCTransport()
	case a2a.TransportProtocolHTTPJSON:
		transportOpt = a2aclient.WithRESTTransport(httpClient)
	default:
		transportOpt = a2aclient.WithJSONRPCTransport(httpClient)
	}

	if !disableTUI {
		if transport == "" {
			fmt.Printf("Auto-selected transport: %s\n", StyleAccent.Render(string(selectedTransport)))
		} else {
			fmt.Printf("Forcing transport: %s\n", StyleAccent.Render(string(selectedTransport)))
		}
	}

	opts := []a2aclient.FactoryOption{transportOpt}
	if authToken != "" {
		opts = append(opts, a2aclient.WithCallInterceptors(&tokenInterceptor{token: authToken}))
	}
	return a2aclient.NewFromCard(ctx, card, opts...)
}

func runDescribe(_ *cobra.Command, _ []string) {
	card, err := agentcard.DefaultResolver.Resolve(context.Background(), serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard: %v", err, "Ensure the A2A server is running at "+serviceURL)
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

func runSend(_ *cobra.Command, args []string) {
	messageText := args[0]

	if instructionFile != "" {
		content, err := os.ReadFile(instructionFile)
		if err != nil {
			fatalf("failed to read instruction file %q", err, "Verify the file path exists and is readable")
		}
		messageText = fmt.Sprintf("%s\n\nSupplemental Instructions:\n%s", messageText, string(content))
	}

	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
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

	if wait {
		b := true
		params.Config = &a2a.SendMessageConfig{
			Blocking: &b,
		}

		if !disableTUI {
			fmt.Printf("Invoking A2A Service (Blocking)...\n\n")
		}

		result, err := client.SendMessage(ctx, params)
		if err != nil {
			fatalf("SendMessage failed", err, "Check service connectivity or skill availability")
		}

		if disableTUI {
			b, err := json.MarshalIndent(result, "", "  ")
			if err == nil {
				fmt.Println(string(b))
			}
			if task, ok := result.(*a2a.Task); ok && (outDir != "" || outFile != "") {
				for i, art := range task.Artifacts {
					_, _ = saveArtifact(outDir, outFile, *art, i)
				}
			}
			return
		}

		if task, ok := result.(*a2a.Task); ok {
			displayTaskResult(task, outDir)
			fmt.Printf("\nTask ID: %s (use --task %s to continue, or --ref %s to reference)\n", task.ID, task.ID, task.ID)
		} else if msg, ok := result.(*a2a.Message); ok {
			fmt.Printf("Received simple message from agent (Task ID: %s)\n", msg.TaskID)
			for _, p := range msg.Parts {
				if tp, ok := p.Content.(a2a.Text); ok {
					fmt.Printf("Agent: %s\n", string(tp))
				}
			}
		} else {
			fmt.Printf("Unknown result type received: %T\n", result)
		}

		return
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

func runWatch(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	if !disableTUI {
		fmt.Printf("Watching Task %s ...\n\n", taskID)
	}

	tid := a2a.TaskID(taskID)

	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		fatalf("failed to retrieve task status", err, "If using an in-memory store, task history is lost on server restart")
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

func runGet(cmd *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	tid := a2a.TaskID(taskID)

	// Default outDir for 'download' command if neither is specified
	if cmd.Name() == "download" || cmd.Name() == "retrieve" {
		if outDir == "" && outFile == "" {
			outDir = "."
		}
	}

	task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
	if err != nil {
		fatalf("failed to retrieve task", err, "Check the task ID or verify the server state")
	}

	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		if outDir != "" || outFile != "" {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}

	// Always display the full result (which handles saving now!)
	displayTaskResult(task, outDir)
}

func runCancel(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()

	card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	tid := a2a.TaskID(taskID)

	task, err := client.CancelTask(ctx, &a2a.CancelTaskRequest{ID: tid})
	if err != nil {
		fatalf("failed to cancel task", err, "Check the task ID or verify the server state")
	}

	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	fmt.Printf("Task %s has been requested to cancel. Current state: %s\n", task.ID, task.Status.State)
}

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/a2acli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&envName, "env", "e", "", "environment name to load from config")
	rootCmd.PersistentFlags().StringVarP(&serviceURL, "service-url", "u", "http://127.0.0.1:9001", "Base URL of the A2A service")
	rootCmd.PersistentFlags().StringVarP(&authToken, "token", "t", "", "Auth token")
	rootCmd.PersistentFlags().StringVarP(&targetTaskID, "task", "k", "", "Existing Task ID to continue (must be non-terminal)")
	rootCmd.PersistentFlags().StringVarP(&refTaskID, "ref", "r", "", "Task ID to reference as context (works for completed tasks)")
	rootCmd.PersistentFlags().BoolVarP(&disableTUI, "no-tui", "n", false, "Disable the Terminal UI (useful for scripting and CI)")
	rootCmd.PersistentFlags().StringVar(&transport, "transport", "", "Force a specific transport protocol (grpc, jsonrpc, rest)")
	rootCmd.Flags().BoolP("version", "V", false, "Print version information")

	if os.Getenv("A2ACLI_NO_TUI") == "true" || os.Getenv("NO_COLOR") != "" {
		disableTUI = true
	}

	var describeCmd = &cobra.Command{
		Use:     "describe",
		GroupID: GroupDiscovery,
		Short:   "Describe the agent card",
		Long: `Retrieve and display the A2A AgentCard for the target service.

The AgentCard contains the agent's identity, description, supported 
interface protocols (e.g., JSON-RPC), and available skills. It also 
lists any security requirements for each skill.`,
		Example: `  a2acli describe
  a2acli describe --service-url http://localhost:9001
  a2acli describe --no-tui --token "my-auth-token"`,
		Run: runDescribe,
	}

	var sendCmd = &cobra.Command{
		Use:     "send [message]",
		GroupID: GroupMessaging,
		Aliases: []string{"invoke", "SendMessage"},
		Short:   "Send a message to an agent (streaming)",
		Long: `Initiate a new task or continue an existing one by sending a message to the agent.

By default, this command uses streaming to provide real-time updates from 
the agent. Use the --wait flag to perform a blocking call instead.

You can save artifacts produced by the task using the --out-dir flag.`,
		Example: `  a2acli send "Write a simple CLI in Go"
  a2acli send "Add error handling to that CLI" --task <taskID>
  a2acli send "Summarize this task" --ref <taskID>
  a2acli send "Generate report" --skill reports --wait --out-dir ./reports`,
		Args: cobra.MinimumNArgs(1),
		Run:  runSend,
	}

	var watchCmd = &cobra.Command{
		Use:     "watch [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"resume", "SubscribeToTask"},
		Short:   "Watch an existing task's streaming updates",
		Long: `Connect to an active task's event stream to receive real-time updates.

This is useful for resuming observation of a long-running task or 
watching a task initiated by another client. If the task is 
already completed, the command will display the final results.`,
		Example: `  a2acli watch <taskID>
  a2acli watch <taskID> --no-tui
  a2acli watch <taskID> --out-dir ./artifacts`,
		Args: cobra.ExactArgs(1),
		Run:  runWatch,
	}

	var getCmd = &cobra.Command{
		Use:     "get [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"status", "GetTask"},
		Short:   "Get the status of a task",
		Long: `Retrieve the current state and results of a task.

Displays the task status (e.g., active, completed, failed) and a 
preview of any artifacts produced. Use the --out-dir flag to 
download artifacts to a directory.`,
		Example: `  a2acli get <taskID>
  a2acli get <taskID> --no-tui
  a2acli get <taskID> --out-dir ./status`,
		Args: cobra.ExactArgs(1),
		Run:  runGet,
	}

	var versionCmd = &cobra.Command{
		Use:     "version",
		GroupID: GroupSystem,
		Short:   "Print the version number of a2acli",
		Run:     runVersion,
	}

	sendCmd.Flags().StringVarP(&skillID, "skill", "s", "", "Skill ID")
	sendCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	sendCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
	sendCmd.Flags().StringVarP(&instructionFile, "instruction-file", "i", "", "Path to a file with supplemental instructions")
	sendCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Block and wait for task completion instead of streaming (maps to A2A Blocking:true)")
	sendCmd.Flags().BoolVar(&wait, "sync", false, "Alias for --wait")

	watchCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	watchCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")

	getCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	getCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")

	var downloadCmd = &cobra.Command{
		Use:     "download [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"retrieve"},
		Short:   "Download artifacts from a task",
		Long: `Download all artifacts produced by a specific task.

This is a convenience command that retrieves the task and saves its 
artifacts to the current directory or a specified output directory.`,
		Example: `  a2acli download <taskID>
  a2acli download <taskID> --out-dir ./results
  a2acli download <taskID> --file output.txt`,
		Args: cobra.ExactArgs(1),
		Run:  runGet, // Reuse runGet which now handles outDir and outFile natively
	}
	downloadCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	downloadCmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")

	var cancelCmd = &cobra.Command{
		Use:     "cancel [taskID]",
		GroupID: GroupMessaging,
		Aliases: []string{"terminate", "CancelTask"},
		Short:   "Cancel an active task",
		Long: `Request cancellation of an active task.

If the task is still running, the agent will attempt to stop its 
execution. This is a best-effort request and the task may 
already have completed or be in a non-cancelable state.`,
		Example: `  a2acli cancel <taskID>
  a2acli cancel <taskID> --no-tui`,
		Args: cobra.ExactArgs(1),
		Run:  runCancel,
	}

	var configCmd = &cobra.Command{
		Use:     "config",
		GroupID: GroupSystem,
		Short:   "View the active configuration",
		Long: `Display the active configuration settings.

Settings are loaded from the default configuration file ($HOME/.config/a2acli/config.yaml) 
and can be overridden by environment variables and command-line flags.`,
		Example: `  a2acli config
  a2acli config --env production
  a2acli config --config ./myconfig.yaml`,
		Run: runConfig,
	}

	rootCmd.Run = func(cmd *cobra.Command, _ []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			runVersion(cmd, nil)
			return
		}
		_ = cmd.Help()
	}

	rootCmd.AddCommand(describeCmd, sendCmd, watchCmd, getCmd, downloadCmd, cancelCmd, configCmd, versionCmd)
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

		if v, ok := msg.Event.(*a2a.TaskArtifactUpdateEvent); ok && (outDir != "" || outFile != "") {
			_, _ = saveArtifact(outDir, outFile, *v.Artifact, 0)
		}
	}
}

func displayTaskResult(task *a2a.Task, outDir string) {
	if disableTUI {
		b, err := json.MarshalIndent(task, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		if outDir != "" || outFile != "" {
			for i, art := range task.Artifacts {
				_, _ = saveArtifact(outDir, outFile, *art, i)
			}
		}
		return
	}

	state := string(task.Status.State)
	var stateStyle lipgloss.Style
	switch task.Status.State {
	case a2a.TaskStateCompleted:
		stateStyle = StylePass
	case a2a.TaskStateFailed, a2a.TaskStateRejected:
		stateStyle = StyleFail
	default:
		stateStyle = StyleWarn
	}

	fmt.Printf("Task Status: [%s]\n", stateStyle.Render(state))

	if len(task.Artifacts) == 0 {
		fmt.Println("No artifacts produced.")
		return
	}

	fmt.Printf("\n%s\n", StyleAccent.Render(fmt.Sprintf("--- %d ARTIFACT(S) AVAILABLE ---", len(task.Artifacts))))

	for i, art := range task.Artifacts {
		fmt.Printf("\nName: %s\n", StyleArtifact.Render(art.Name))
		if art.Description != "" {
			fmt.Printf("Description: %s\n", art.Description)
		}

		truncated := false
		for _, p := range art.Parts {
			if dp, ok := p.Content.(a2a.Data); ok {
				prettyJSON, _ := json.MarshalIndent(dp, "", "  ")
				fmt.Printf("%s\n%s\n", StyleMuted.Render("Data (Preview):"), string(prettyJSON))
			} else if tp, ok := p.Content.(a2a.Text); ok {
				preview := string(tp)
				if len(preview) > 500 {
					preview = preview[:500] + "... (truncated)"
					truncated = true
				}
				fmt.Printf("%s\n%s\n", StyleMuted.Render("Content (Preview):"), preview)
			}
		}

		if outDir != "" || outFile != "" {
			path, err := saveArtifact(outDir, outFile, *art, i)
			if err != nil {
				fmt.Printf("%s %v\n", StyleFail.Render("Error saving artifact:"), err)
			} else {
				fmt.Printf("%s %s\n", StyleAccent.Render(">> Saved to:"), StyleArtifact.Render(path))
			}
		} else if truncated {
			fmt.Printf("%s\n", StyleMuted.Render("(Hint: Use --out-dir <path> or --file <name> to save the full artifact content)"))
		}
	}
	fmt.Printf("\n%s\n", StyleAccent.Render("------------------------------"))
}
func saveArtifact(outDir, outFile string, artifact a2a.Artifact, index int) (string, error) {
	var path string
	if outFile != "" {
		fName := outFile
		if index > 0 {
			ext := filepath.Ext(outFile)
			base := strings.TrimSuffix(outFile, ext)
			fName = fmt.Sprintf("%s_%d%s", base, index, ext)
		}
		if outDir != "" {
			path = filepath.Join(outDir, fName)
		} else {
			path = fName
		}
	} else {
		if outDir == "" {
			outDir = "."
		}
		filename := artifact.Name
		if filename == "" {
			filename = fmt.Sprintf("artifact_%d_%d.txt", time.Now().Unix(), index)
		}
		path = filepath.Join(outDir, filename)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}

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
