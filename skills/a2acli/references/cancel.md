# cancel — Cancel an Active Task

Maps to the A2A Protocol's `CancelTask` RPC. Requests cancellation of a task that is currently active. Has no effect on tasks already in a terminal state.

## Usage

```bash
a2acli cancel <task_id> --service-url http://localhost:9001 -n
```

## Output

Returns the updated `Task` object with `status.state` set to `TASK_STATE_CANCELED` on success.

```json
{
  "id": "task-abc-123",
  "status": {
    "state": "TASK_STATE_CANCELED"
  }
}
```

If the task is already in a terminal state, the server may return an error or return the task unchanged.
