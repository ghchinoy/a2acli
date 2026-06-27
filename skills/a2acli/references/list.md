# list — List Historical Tasks

Maps to the A2A Protocol's `ListTasks` RPC. Returns a paginated list of tasks the agent has processed. The server must support task history for this to return results.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--limit` | `10` | Maximum number of tasks to return |
| `--page-token` | — | Pagination token from a previous response |

## Usage

```bash
# List the 10 most recent tasks
a2acli list tasks --service-url http://localhost:9001 -n

# Increase the page size
a2acli list tasks --limit 50 --service-url http://localhost:9001 -n

# Paginate using a token from the previous response
a2acli list tasks --page-token <token> --service-url http://localhost:9001 -n
```

## Output Schema

```json
{
  "tasks": [
    {
      "id": "task-abc-123",
      "status": { "state": "TASK_STATE_COMPLETED" }
    }
  ],
  "nextPageToken": "token-for-next-page"
}
```

Use `nextPageToken` as `--page-token` to retrieve subsequent pages.
