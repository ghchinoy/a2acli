# watch — Subscribe to a Task's Event Stream

Maps to the A2A Protocol's `SubscribeToTask` RPC. Streams live status updates and artifacts from an active task. Use when you initiated a task without `--wait` and want to observe it to completion.

## Flags

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory as they arrive |
| `--file` | `-f` | Save artifact to a specific filename |

## Usage

```bash
# Stream updates from a running task
a2acli watch <task_id> --service-url http://localhost:9001 -n

# Stream and save artifacts as they arrive
a2acli watch <task_id> --out-dir ./output/ --service-url http://localhost:9001 -n
```

## Output

With `-n`, emits NDJSON — one JSON object per event line. Each line is a `TaskStatusUpdateEvent` or `TaskArtifactUpdateEvent`. The stream ends when the task reaches a terminal state (`COMPLETED`, `FAILED`, `CANCELED`, `REJECTED`).
