---
name: a2acli-send
description: Send a message to an Agent-to-Agent (A2A) service to initiate a task or continue a conversation. Use when you need to trigger an action or pass data to another agent.
---

# Initiating a Task with an A2A Agent

The `a2acli send` command sends a message to an A2A service. This either initiates a new task or continues an existing, non-terminal task.

## Important Requirements for Agents

When using this command as an automated agent, **you must ALWAYS include the `--no-tui` (or `-n`) flag** to disable the streaming terminal interface and receive JSON output.

Additionally, to ensure the task finishes executing before you try to process its output, **you must ALWAYS include the `--wait` (or `-w`) flag** to make it a blocking call instead of a streaming one.

## Basic Usage

To send a message to initiate a task:

```bash
a2acli send "Generate a project plan" --service-url <URL> --no-tui --wait
```

### Specifying a Skill

If the agent requires you to target a specific skill, use the `--skill` (`-s`) flag:

```bash
a2acli send "Generate report" --skill reports --service-url <URL> --no-tui --wait
```

### Continuing an Existing Task

To continue an active task, provide the Task ID via `--task` (`-k`):

```bash
a2acli send "Add more details to the intro section" --task <TaskID> --service-url <URL> --no-tui --wait
```

### Referencing a Completed Task

To pass the context of a previous task to a new task, use `--ref` (`-r`):

```bash
a2acli send "Summarize this task" --ref <TaskID> --service-url <URL> --no-tui --wait
```

### Passing Supplemental Instructions

If you have complex instructions or a large file to pass as part of the prompt, save it to a file and reference it:

```bash
a2acli send "Fix the bugs in this code" --instruction-file path/to/file.txt --service-url <URL> --no-tui --wait
```

### Downloading Artifacts

To automatically save artifacts produced by the task to the filesystem:

```bash
a2acli send "Generate image" --out-dir ./output/ --service-url <URL> --no-tui --wait
```

## Parsing the Output

With `--no-tui` and `--wait`, the output will be a JSON object representing the final `Task` state, or a `Message` object. Look at `id` (the task ID) and `status.state` (e.g., "TASK_STATE_COMPLETED", "TASK_STATE_FAILED"). You can use the ID in later commands to fetch the results or continue the conversation.

## When to Use This Skill

- When you need to ask another specialized agent to perform work.
- To execute an A2A task synchronously and get the final result back as JSON.
