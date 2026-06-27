# get — Retrieve Task Status and Artifacts

Maps to the A2A Protocol's `GetTask` RPC. Retrieves the current state and any artifacts of a task by its ID.

## Flags

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Save artifacts to a directory |
| `--file` | `-f` | Save artifact to a specific filename (index appended for multiple) |

## Usage

```bash
# Check task status
a2acli get <task_id> --service-url http://localhost:9001 --output json

# Retrieve and save artifacts
a2acli get <task_id> --out-dir ./output/ --service-url http://localhost:9001 --output json

# Save to a specific file
a2acli get <task_id> --file result.txt --service-url http://localhost:9001 --output json
```

## Output Schema

```json
{
  "id": "task-abc-123",
  "status": {
    "state": "TASK_STATE_COMPLETED"
  },
  "artifacts": [
    {
      "name": "report.pdf",
      "parts": [{ "text": "..." }]
    }
  ]
}
```

Possible `status.state` values: `TASK_STATE_ACTIVE`, `TASK_STATE_COMPLETED`, `TASK_STATE_FAILED`, `TASK_STATE_REJECTED`, `TASK_STATE_CANCELED`.
