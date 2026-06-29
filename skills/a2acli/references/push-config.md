# push-config — Push Notification Configs

Manage push notification configurations for A2A tasks. Push notifications allow
the server to proactively POST to a webhook URL when a task's state changes,
rather than requiring the client to poll. The server must advertise
`capabilities.pushNotifications: true` in its AgentCard.

## Subcommands

| Command | Description |
|---|---|
| `push-config create <task-id> <callback-url>` | Register a webhook callback for a task |
| `push-config list <task-id>` | List all push configs for a task |
| `push-config get <task-id> <config-id>` | Retrieve a specific push config |
| `push-config delete <task-id> <config-id>` | Delete a push config |

## Flags (create)

| Flag | Description |
|---|---|
| `--id <id>` | Optional client-assigned config ID (for later get/delete) |
| `--auth-scheme <scheme>` | Auth scheme for the callback endpoint (e.g. `Bearer`) |
| `--auth-credentials <creds>` | Auth credentials for the callback endpoint |
| `--token <token>` | Validation token sent with every notification |

## Usage

```bash
# Register a webhook (minimal)
a2acli push-config create <task-id> https://myserver.example.com/notify

# Register with auth and a stable ID
a2acli push-config create <task-id> https://cb.example.com/notify \
  --id my-config \
  --auth-scheme Bearer --auth-credentials mytoken \
  --token validation-secret

# List all configs for a task
a2acli push-config list <task-id> --output json

# Retrieve a specific config
a2acli push-config get <task-id> my-config

# Remove a config
a2acli push-config delete <task-id> my-config
```

## Output schema (--output json)

```json
{
  "taskId": "task-abc-123",
  "id": "my-config",
  "url": "https://cb.example.com/notify",
  "authentication": {
    "scheme": "Bearer",
    "credentials": "mytoken"
  },
  "token": "validation-secret"
}
```

## Notes

- The task must exist on the server before creating a push config.
- If the AgentCard does not advertise `capabilities.pushNotifications: true`,
  a2acli will emit a warning but still attempt the call.
- The server is responsible for delivering notifications to the callback URL;
  a2acli does not verify delivery.
- Test against: **a2a-simple** (runs locally, `capabilities.pushNotifications: true`).
