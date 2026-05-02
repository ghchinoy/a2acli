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

	"github.com/spf13/cobra"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

var (
	listLimit     int
	listPageToken string
)

func setupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		GroupID: GroupMessaging,
		Short:   "List resources",
	}

	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "List historical tasks from an agent",
		Long: `Query the agent for a list of historical tasks it has processed.
Note: The server must support history for this endpoint to return data.`,
		Example: `  a2acli list tasks --limit 10`,
		Run:     runListTasks,
	}

	tasksCmd.Flags().IntVar(&listLimit, "limit", 10, "Maximum number of tasks to return")
	tasksCmd.Flags().StringVar(&listPageToken, "page-token", "", "Pagination token")

	cmd.AddCommand(tasksCmd)
	return cmd
}

func runListTasks(_ *cobra.Command, _ []string) {
	ctx := context.Background()

	card, err := getResolver().Resolve(ctx, serviceURL)
	if err != nil {
		fatalf("failed to resolve AgentCard", err, "Ensure the A2A server is running at "+serviceURL)
	}

	client, err := createClient(ctx, card)
	if err != nil {
		fatalf("failed to create client", err, "Verify your --token or configuration settings")
	}

	req := &a2a.ListTasksRequest{
		PageSize:  listLimit,
		PageToken: listPageToken,
	}

	resp, err := client.ListTasks(ctx, req)
	if err != nil {
		fatalf("failed to list tasks", err, "Ensure the server supports listing tasks")
	}

	if disableTUI {
		b, err := json.MarshalIndent(resp, "", "  ")
		if err == nil {
			fmt.Println(string(b))
		}
		return
	}

	if len(resp.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("%-36s | %-12s | %s\n", "TASK ID", "STATUS", "CREATED AT")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, t := range resp.Tasks {
		createdAt := ""
		if t.Status.Timestamp != nil {
			createdAt = t.Status.Timestamp.Format("2006-01-02 15:04:05")
		}
		status := string(t.Status.State)
		if status == "" {
			status = "unknown"
		}
		fmt.Printf("%-36s | %-12s | %s\n", t.ID, status, createdAt)
	}

	if resp.NextPageToken != "" {
		fmt.Printf("\nNext Page Token: %s\n", resp.NextPageToken)
	}
}
