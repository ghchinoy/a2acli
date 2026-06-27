# describe — Fetch an Agent's AgentCard

Fetches the AgentCard from the A2A discovery endpoint (`/.well-known/agent.json`). Use before sending tasks to verify the agent's skills, transport capabilities, and auth requirements.

## Usage

```bash
# Basic discovery
a2acli discover --service-url http://localhost:9001 --output json

# With auth (if the agent requires a token to expose its card)
a2acli discover --service-url http://localhost:9001 --token "<TOKEN>" --output json
```

## Output Schema

```json
{
  "name": "Example Agent",
  "description": "A helpful assistant",
  "url": "http://localhost:9001",
  "skills": [
    {
      "id": "reports",
      "name": "Generate Report",
      "description": "Generates a PDF report"
    }
  ],
  "defaultInputModes": ["text/plain"],
  "defaultOutputModes": ["text/plain"]
}
```

Use the `skills[].id` values as the `--skill` argument to `send`.
Use the transport fields to understand whether `grpc`, `jsonrpc`, or `rest` is available.
