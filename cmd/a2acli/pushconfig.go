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
	"os"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

// push-config flag vars
var (
	pushAuthScheme      string
	pushAuthCredentials string
	pushConfigID        string
	pushToken           string
	pushPageSize        int
	pushPageToken       string
)

// setupPushConfigCmd builds the `push-config` command group and registers it
// on the root command. Called from main().
func setupPushConfigCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		Use:     "push-config",
		GroupID: GroupMessaging,
		Short:   "Manage task push-notification configurations",
		Long: `Create, list, retrieve, and delete push notification configurations for tasks.

Push notifications allow an A2A server to proactively call a webhook URL when
a task's state changes, rather than requiring the client to poll. The server
must advertise Capabilities.PushNotifications: true in its AgentCard.`,
	}

	// create
	createCmd := &cobra.Command{
		Use:   "create <task-id> <callback-url>",
		Short: "Register a push-notification callback for a task",
		Long: `Create a push notification configuration for a task.
*(Maps to the A2A Protocol's CreateTaskPushNotificationConfig RPC.)*`,
		Example: `  a2acli push-config create task-123 https://myserver.example.com/notify
  a2acli push-config create task-123 https://myserver.example.com/notify \
    --auth-scheme Bearer --auth-credentials mytoken
  a2acli push-config create task-123 https://cb.example.com/notify \
    --id my-config --token validation-token`,
		Args: cobra.ExactArgs(2),
		Run:  runPushConfigCreate,
	}
	createCmd.Flags().StringVar(&pushAuthScheme, "auth-scheme", "", "Auth scheme for the callback endpoint (e.g. Bearer, Basic)")
	createCmd.Flags().StringVar(&pushAuthCredentials, "auth-credentials", "", "Auth credentials for the callback endpoint")
	createCmd.Flags().StringVar(&pushConfigID, "id", "", "Optional client-assigned ID for this push config")
	createCmd.Flags().StringVar(&pushToken, "token", "", "Optional validation token sent with every notification")

	// list
	listCmd := &cobra.Command{
		Use:   "list <task-id>",
		Short: "List push-notification configs for a task",
		Long:  `List all push notification configurations for a task.\n*(Maps to the A2A Protocol's ListTaskPushNotificationConfig RPC.)*`,
		Example: `  a2acli push-config list task-123
  a2acli push-config list task-123 --output json`,
		Args: cobra.ExactArgs(1),
		Run:  runPushConfigList,
	}
	listCmd.Flags().IntVar(&pushPageSize, "page-size", 0, "Maximum number of configs to return")
	listCmd.Flags().StringVar(&pushPageToken, "page-token", "", "Pagination token")

	// get
	getCmd := &cobra.Command{
		Use:   "get <task-id> <config-id>",
		Short: "Retrieve a specific push-notification config",
		Long:  `Retrieve a specific push notification configuration by task ID and config ID.\n*(Maps to the A2A Protocol's GetTaskPushNotificationConfig RPC.)*`,
		Example: `  a2acli push-config get task-123 my-config
  a2acli push-config get task-123 my-config --output json`,
		Args: cobra.ExactArgs(2),
		Run:  runPushConfigGet,
	}

	// delete
	deleteCmd := &cobra.Command{
		Use:     "delete <task-id> <config-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a push-notification config",
		Long:    `Delete a push notification configuration for a task.\n*(Maps to the A2A Protocol's DeleteTaskPushNotificationConfig RPC.)*`,
		Example: `  a2acli push-config delete task-123 my-config`,
		Args:    cobra.ExactArgs(2),
		Run:     runPushConfigDelete,
	}

	pushCmd.AddCommand(createCmd, listCmd, getCmd, deleteCmd)
	return pushCmd
}

func runPushConfigCreate(_ *cobra.Command, args []string) {
	taskID, callbackURL := args[0], args[1]
	ctx := context.Background()
	verboseLog("push-config create: task=%s url=%s scheme=%q", taskID, callbackURL, pushAuthScheme)

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}
	if !card.Capabilities.PushNotifications {
		fmt.Fprintf(os.Stderr, "Hint: This agent's AgentCard does not advertise PushNotifications support.\n")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	cfg := a2a.PushConfig{
		TaskID: a2a.TaskID(taskID),
		ID:     pushConfigID,
		URL:    callbackURL,
		Token:  pushToken,
	}
	if pushAuthScheme != "" {
		cfg.Auth = &a2a.PushAuthInfo{
			Scheme:      pushAuthScheme,
			Credentials: pushAuthCredentials,
		}
	}

	result, err := client.CreateTaskPushConfig(ctx, &cfg)
	if err != nil {
		fatalf("CreateTaskPushConfig failed", err, "Ensure the task exists and the server supports push notifications")
	}
	verboseLog("push-config created: id=%s", result.ID)

	if disableTUI {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Printf("Push config created for task %s\n", result.TaskID)
	fmt.Printf("  Config ID: %s\n", result.ID)
	fmt.Printf("  URL:       %s\n", result.URL)
}

func runPushConfigList(_ *cobra.Command, args []string) {
	taskID := args[0]
	ctx := context.Background()
	verboseLog("push-config list: task=%s pageSize=%d", taskID, pushPageSize)

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	req := &a2a.ListTaskPushConfigRequest{
		TaskID:    a2a.TaskID(taskID),
		PageSize:  pushPageSize,
		PageToken: pushPageToken,
	}
	result, err := client.ListTaskPushConfigs(ctx, req)
	if err != nil {
		fatalf("ListTaskPushConfigs failed", err, "Ensure the task exists and the server supports push notifications")
	}
	verboseLog("push-config list: returned %d configs", len(result))

	if disableTUI {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
		return
	}
	if len(result) == 0 {
		fmt.Printf("No push configs for task %s\n", taskID)
		return
	}
	fmt.Printf("Push configs for task %s (%d):\n", taskID, len(result))
	for _, cfg := range result {
		fmt.Printf("  - %s  url=%s\n", cfg.ID, cfg.URL)
	}
}

func runPushConfigGet(_ *cobra.Command, args []string) {
	taskID, configID := args[0], args[1]
	ctx := context.Background()
	verboseLog("push-config get: task=%s config=%s", taskID, configID)

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	result, err := client.GetTaskPushConfig(ctx, &a2a.GetTaskPushConfigRequest{
		TaskID: a2a.TaskID(taskID),
		ID:     configID,
	})
	if err != nil {
		fatalf("GetTaskPushConfig failed", err, "Check the task ID and config ID")
	}
	verboseLog("push-config get: url=%s", result.URL)

	if disableTUI {
		b, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Printf("Push config %s (task %s):\n", result.ID, result.TaskID)
	fmt.Printf("  URL:   %s\n", result.URL)
	if result.Auth != nil {
		fmt.Printf("  Auth:  %s\n", result.Auth.Scheme)
	}
	if result.Token != "" {
		fmt.Printf("  Token: %s\n", result.Token)
	}
}

func runPushConfigDelete(_ *cobra.Command, args []string) {
	taskID, configID := args[0], args[1]
	ctx := context.Background()
	verboseLog("push-config delete: task=%s config=%s", taskID, configID)

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Check --service-url or A2ACLI_SERVICE_URL")
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	err = client.DeleteTaskPushConfig(ctx, &a2a.DeleteTaskPushConfigRequest{
		TaskID: a2a.TaskID(taskID),
		ID:     configID,
	})
	if err != nil {
		fatalf("DeleteTaskPushConfig failed", err, "Check the task ID and config ID")
	}
	verboseLog("push-config deleted")

	if disableTUI {
		fmt.Printf("{\"deleted\": true, \"taskId\": %q, \"configId\": %q}\n", taskID, configID)
		return
	}
	fmt.Printf("Deleted push config %s from task %s\n", configID, taskID)
}
