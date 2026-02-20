package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
)

type tokenInterceptor struct {
	a2aclient.PassthroughInterceptor
	token string
}

func (i *tokenInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, error) {
	if i.token != "" {
		req.Meta["Authorization"] = []string{"Bearer " + i.token}
	}
	return ctx, nil
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

	var describeCmd = &cobra.Command{
		Use:   "describe",
		Short: "Describe the agent card",
		Run: func(cmd *cobra.Command, args []string) {
			card, err := agentcard.DefaultResolver.Resolve(context.Background(), serviceURL)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			fmt.Printf("Agent: %s\n", card.Name)
			if card.Description != "" {
				fmt.Printf("Description: %s\n", card.Description)
			}
			fmt.Printf("Capabilities: [Streaming: %v]\n", card.Capabilities.Streaming)
			fmt.Printf("\nSkills:\n")
			for _, s := range card.Skills {
				fmt.Printf("  - [%s] %s\n", s.ID, s.Name)
				if s.Description != "" {
					fmt.Printf("    Description: %s\n", s.Description)
				}
				if len(s.Security) > 0 {
					var schemes []string
					for _, req := range s.Security {
						for name := range req {
							schemes = append(schemes, string(name))
						}
					}
					fmt.Printf("    Security: %s\n", strings.Join(schemes, ", "))
				}
			}
		},
	}

	var invokeCmd = &cobra.Command{
		Use:   "invoke [message]",
		Short: "Invoke a skill (streaming)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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

			// 2. Create Client
			httpClient := &http.Client{Timeout: 15 * time.Minute}
			opts := []a2aclient.FactoryOption{a2aclient.WithJSONRPCTransport(httpClient)}
			if authToken != "" {
				opts = append(opts, a2aclient.WithInterceptors(&tokenInterceptor{token: authToken}))
			}

			client, err := a2aclient.NewFromCard(ctx, card, opts...)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			// 3. Prepare Message
			msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.TextPart{Text: messageText})
			if targetTaskID != "" {
				msg.TaskID = a2a.TaskID(targetTaskID)
				fmt.Printf("Continuing Task: %s\n", targetTaskID)
			}
			if refTaskID != "" {
				msg.ReferenceTasks = []a2a.TaskID{a2a.TaskID(refTaskID)}
				fmt.Printf("Referencing Task: %s\n", refTaskID)
			}

			params := &a2a.MessageSendParams{
				Message: msg,
			}
			if skillID != "" {
				params.Metadata = map[string]any{"skillId": skillID}
			}

			fmt.Printf("Invoking A2A Service (Streaming)...\n\n")

			// Adapter: Convert Iterator to Channel for Bubble Tea
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

			runTUI(stream)
		},
	}

	var resumeCmd = &cobra.Command{
		Use:   "resume [taskID]",
		Short: "Resume listening to an existing task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := args[0]
			ctx := context.Background()

			card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			httpClient := &http.Client{Timeout: 15 * time.Minute}
			opts := []a2aclient.FactoryOption{a2aclient.WithJSONRPCTransport(httpClient)}
			if authToken != "" {
				opts = append(opts, a2aclient.WithInterceptors(&tokenInterceptor{token: authToken}))
			}

			client, err := a2aclient.NewFromCard(ctx, card, opts...)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			fmt.Printf("Resuming Task %s ...\n\n", taskID)

			tid := a2a.TaskID(taskID)
			
			// Check status first
			task, err := client.GetTask(ctx, &a2a.TaskQueryParams{ID: tid})
			if err != nil {
				errMsg := err.Error()
				// Simple string check for "not found" as the SDK error might be wrapped
				if len(errMsg) > 0 { // Check if it looks like a "not found" error
					// Just print the error for now, but add a hint
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

			// If active, stream updates
			fmt.Println("Task is active. Connecting to stream...")
			stream := make(chan streamMsg)
			go func() {
				defer close(stream)
				for event, err := range client.ResubscribeToTask(ctx, &a2a.TaskIDParams{ID: tid}) {
					stream <- streamMsg{Event: event, Err: err}
					if err != nil {
						return
					}
				}
			}()

			runTUI(stream)
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status [taskID]",
		Short: "Get the status of a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := args[0]
			ctx := context.Background()

			card, err := agentcard.DefaultResolver.Resolve(ctx, serviceURL)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			httpClient := &http.Client{Timeout: 15 * time.Minute}
			opts := []a2aclient.FactoryOption{a2aclient.WithJSONRPCTransport(httpClient)}
			if authToken != "" {
				opts = append(opts, a2aclient.WithInterceptors(&tokenInterceptor{token: authToken}))
			}

			client, err := a2aclient.NewFromCard(ctx, card, opts...)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			tid := a2a.TaskID(taskID)
			task, err := client.GetTask(ctx, &a2a.TaskQueryParams{ID: tid})
			if err != nil {
				log.Fatalf("Error retrieving task: %v", err)
			}

			fmt.Printf("Task ID: %s\n", task.ID)
			fmt.Printf("Status:  %s\n", task.Status.State)
			if task.Status.Message != nil {
				for _, p := range task.Status.Message.Parts {
					if tp, ok := p.(a2a.TextPart); ok {
						fmt.Printf("Message: %s\n", tp.Text)
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
		},
	}

	invokeCmd.Flags().StringVarP(&skillID, "skill", "s", "", "Skill ID")
	invokeCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")
	invokeCmd.Flags().StringVarP(&instructionFile, "instruction-file", "f", "", "Path to a file with supplemental instructions")
	
	resumeCmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Directory to save artifacts to")

	rootCmd.AddCommand(describeCmd, invokeCmd, resumeCmd, statusCmd)
	rootCmd.Execute()
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

func displayTaskResult(task *a2a.Task, outDir string) {
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
		// Preview content
		for _, p := range art.Parts {
			if dp, ok := p.(a2a.DataPart); ok {
				prettyJSON, _ := json.MarshalIndent(dp.Data, "", "  ")
				fmt.Printf("Data (Preview):\n%s\n", string(prettyJSON))
			} else if tp, ok := p.(a2a.TextPart); ok {
				preview := tp.Text
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
