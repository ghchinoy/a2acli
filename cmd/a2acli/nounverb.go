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
	"time"

	"github.com/spf13/cobra"
)

// Roadmap B — noun-verb command grammar.
//
// OFFICIAL (a2aproject/a2a-cli) uses a noun-verb grammar (`a2a task get <id>`,
// `a2a card get`) while OURS historically exposed flat verbs (`get`, `discover`,
// …). This file adds `card` and `task` parent commands whose subcommands are
// ADDITIVE aliases of the existing flat verbs: they reuse the exact same run
// functions and register the identical flag sets. The flat verbs remain
// registered and unchanged, so every existing invocation keeps working
// byte-for-byte; scripts written for OFFICIAL's grammar now also resolve.
//
// A single *cobra.Command may only have one parent, so each noun-verb path is a
// thin child command that shares the flat verb's run function and flag-adder.
// Global/credential flags (--agent-card, --endpoint, --context-id, --task-id,
// --a2a-version, --bearer, --api-key, --transport, -o/--output, …) are persistent
// on the root command and are therefore inherited by these nested subcommands
// automatically.

// addGetFlags registers the flag set shared by the flat `get` verb and the
// `task get` noun-verb alias so the two paths accept identical flags.
//
// Roadmap A4: --wait turns the one-shot read into a poll loop (SPEC §9.3/§10.3),
// --poll-interval spaces the polls (default 2s per SPEC §9.3 RECOMMENDED; overall
// budget is --timeout), and --history requests up to n history messages.
func addGetFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	cmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
	cmd.Flags().BoolVar(&showFull, "full", false, "Show complete artifact content without truncating")
	cmd.Flags().BoolVar(&getWait, "wait", false, "Poll until the task reaches a terminal or interrupted (input/auth-required) state (SPEC §9.3)")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", 2*time.Second, "Interval between polls while --wait is set; overall budget is --timeout")
	cmd.Flags().IntVar(&historyLen, "history", 0, "Include up to n task history messages (maps to A2A historyLength)")
}

// addSubscribeFlags registers the flag set shared by the flat `subscribe` verb
// and the `task subscribe` noun-verb alias.
func addSubscribeFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&outDir, "out-dir", "d", "", "Directory to save artifacts to")
	cmd.Flags().StringVarP(&outFile, "file", "f", "", "Specific filename to save the artifact to")
}

// setupNounVerbCommands builds the additive `card` and `task` parent commands
// (Roadmap B) and returns them for registration on the root command. Their
// subcommands alias the existing flat verbs, sharing the same run functions and
// flag sets; the flat verbs themselves are untouched.
func setupNounVerbCommands() []*cobra.Command {
	return []*cobra.Command{newCardCmd(), newTaskCmd()}
}

// newCardCmd builds the `card` noun command. Its only verb, `card get`, aliases
// the flat `discover` verb (fetch and display the agent card).
func newCardCmd() *cobra.Command {
	cardCmd := &cobra.Command{
		Use:     "card",
		GroupID: GroupDiscovery,
		Short:   "Agent card operations",
		Long: `Noun-verb grammar for AgentCard operations.

'card get' is an alias of the flat 'discover' verb: it retrieves and displays
the A2A AgentCard for the target service. The flat 'discover'/'describe' verb
remains available and unchanged.`,
		Example: `  a2acli card get
  a2acli card get http://localhost:9001
  a2acli card get --extended`,
	}

	cardGet := &cobra.Command{
		Use:   "get [url]",
		Short: "Discover and display the agent card (alias of 'discover')",
		Long: `Retrieve and display the A2A AgentCard for the target service.

This is a noun-verb alias of the flat 'discover' verb and behaves identically:
same flags, same output, same exit codes.

The agent URL can be passed as an optional positional argument or via
--service-url / -u (or the canonical --agent-card / --endpoint aliases).`,
		Example: `  a2acli card get
  a2acli card get http://localhost:9001
  a2acli card get --extended`,
		Args: cobra.MaximumNArgs(1),
		Run:  runDescribe,
	}
	cardGet.Flags().BoolVar(&discoverExtended, "extended", false, "Fetch the authenticated extended AgentCard")

	cardCmd.AddCommand(cardGet)
	return cardCmd
}

// newTaskCmd builds the `task` noun command. Its verbs alias the flat task
// verbs: get, cancel, list, subscribe, and push-config.
func newTaskCmd() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:     "task",
		GroupID: GroupMessaging,
		Short:   "Task operations",
		Long: `Noun-verb grammar for A2A task operations.

Each subcommand is an alias of an existing flat verb and behaves identically
(same flags, same output, same exit codes):

  task get         alias of 'get'          (task status/details)
  task cancel      alias of 'cancel'       (cancel a task)
  task list        alias of 'list tasks'   (list historical tasks)
  task subscribe   alias of 'subscribe'    (stream task events)
  task push-config alias of 'push-config'  (manage push-notification configs)

The flat verbs remain available and unchanged.`,
		Example: `  a2acli task get <taskID>
  a2acli task cancel <taskID>
  a2acli task list --limit 10
  a2acli task subscribe <taskID>
  a2acli task push-config list <taskID>`,
	}

	taskGet := &cobra.Command{
		Use:   "get [taskID]",
		Short: "Get the status of a task (alias of 'get')",
		Long: `Retrieve the current state and results of a task.

This is a noun-verb alias of the flat 'get' verb and behaves identically.`,
		Example: `  a2acli task get <taskID>
  a2acli task get <taskID> --wait
  a2acli task get <taskID> --history 10 --out-dir ./status`,
		Args: cobra.ExactArgs(1),
		Run:  runGet,
	}
	addGetFlags(taskGet)

	taskCancel := &cobra.Command{
		Use:   "cancel [taskID]",
		Short: "Cancel an active task (alias of 'cancel')",
		Long: `Request cancellation of an active task.

This is a noun-verb alias of the flat 'cancel' verb and behaves identically.`,
		Example: `  a2acli task cancel <taskID>
  a2acli task cancel <taskID> --no-tui`,
		Args: cobra.ExactArgs(1),
		Run:  runCancel,
	}

	taskList := &cobra.Command{
		Use:   "list",
		Short: "List historical tasks from an agent (alias of 'list tasks')",
		Long: `Query the agent for a list of historical tasks it has processed.
Note: The server must support history for this endpoint to return data.

This is a noun-verb alias of the flat 'list tasks' verb and behaves identically.`,
		Example: `  a2acli task list --limit 10
  a2acli task list --status completed
  a2acli task list --context ctx-123`,
		Args: cobra.NoArgs,
		Run:  runListTasks,
	}
	addListTasksFlags(taskList)

	taskSubscribe := &cobra.Command{
		Use:   "subscribe [taskID]",
		Short: "Subscribe to an active task's streaming updates (alias of 'subscribe')",
		Long: `Connect to an active task's event stream to receive real-time updates.

This is a noun-verb alias of the flat 'subscribe' verb and behaves identically.`,
		Example: `  a2acli task subscribe <taskID>
  a2acli task subscribe <taskID> --output json
  a2acli task subscribe <taskID> --out-dir ./artifacts`,
		Args: cobra.ExactArgs(1),
		Run:  runWatch,
	}
	addSubscribeFlags(taskSubscribe)

	// task push-config: reuse the flat push-config group builder with no cobra
	// group ID (the task parent does not register the root command groups).
	taskPushConfig := newPushConfigCmd("")

	taskCmd.AddCommand(taskGet, taskCancel, taskList, taskSubscribe, taskPushConfig)
	return taskCmd
}
