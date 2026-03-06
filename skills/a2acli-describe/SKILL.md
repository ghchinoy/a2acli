---
name: a2acli-describe
description: Discover and describe the capabilities, identity, and required security schemes of an A2A service. Use when you need to know what skills an agent supports.
---

# Inspecting an A2A Agent's Capabilities

The `a2acli describe` command fetches the AgentCard from an A2A server. It allows you to inspect the agent's identity, registered skills, and security requirements.

## Important Requirements for Agents

When using this command as an automated agent, **you must ALWAYS include the `--no-tui` (or `-n`) flag**. This disables the interactive terminal UI and returns raw JSON, which is necessary for you to parse the output.

## Basic Usage

To fetch the AgentCard of a service running at a specific URL:

```bash
a2acli describe --service-url <URL> --no-tui
```

If the agent requires authentication to even be described, provide a token:

```bash
a2acli describe --service-url <URL> --token "<TOKEN>" --no-tui
```

## Parsing the Output

The output will be a JSON object containing the AgentCard. Look for the `skills` array to understand what the agent can do. Each skill has an `id`, `name`, `description`, and potentially `securityRequirements`.

Example output:

```json
{
  "name": "Example Agent",
  "description": "A helpful assistant",
  "skills": [
    {
      "id": "reports",
      "name": "Generate Report",
      "description": "Generates a PDF report",
      "securityRequirements": []
    }
  ]
}
```

## When to Use This Skill

- Before initiating a task, to ensure the target agent actually supports the skill you want to use.
- To discover what transports (gRPC, JSON-RPC, REST) the agent advertises.
- To check if a skill requires an authorization token.
