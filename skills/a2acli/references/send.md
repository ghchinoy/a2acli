# send — Send a Message

Maps to the A2A Protocol's `SendMessage` RPC. Initiates a new task or continues an existing non-terminal task.

## Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--wait` / `--sync` | `-w` | false | Block until task completes. **Required for agents.** |
| `--skill` | `-s` | — | Target a specific skill ID on the agent |
| `--out-dir` | `-o` | — | Save artifacts to a directory automatically |
| `--file` | `-f` | — | Save artifact to a specific filename |
| `--instruction-file` | `-i` | — | Path to a file with supplemental instructions |

## Usage

```bash
# Basic: initiate a task and wait for completion
a2acli send "Generate a project plan" \
  --service-url http://localhost:9001 -n --wait

# Target a specific skill
a2acli send "Generate report" --skill reports \
  --service-url http://localhost:9001 -n --wait

# Continue an existing task
a2acli send "Add more detail to section 2" \
  --task <TaskID> --service-url http://localhost:9001 -n --wait

# Reference a completed task as context
a2acli send "Summarize the previous result" \
  --ref <TaskID> --service-url http://localhost:9001 -n --wait

# Pass a large instruction file
a2acli send "Fix the bugs" \
  --instruction-file ./instructions.txt \
  --service-url http://localhost:9001 -n --wait

# Save artifacts to disk
a2acli send "Generate image" \
  --out-dir ./output/ --service-url http://localhost:9001 -n --wait
```

## Output Schema

With `-n --wait`, output is a JSON `Task` object:

```json
{
  "id": "task-abc-123",
  "contextId": "ctx-xyz",
  "status": {
    "state": "TASK_STATE_COMPLETED"
  },
  "artifacts": [
    {
      "name": "result.txt",
      "parts": [{ "text": "..." }]
    }
  ]
}
```

Check `status.state`: `TASK_STATE_COMPLETED` = success, `TASK_STATE_FAILED` = failure. Use `id` in subsequent `get`, `watch`, or `cancel` calls.
