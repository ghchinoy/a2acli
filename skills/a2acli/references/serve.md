# serve — Run a Local Mock A2A Agent

Starts an A2A-compliant mock agent on your local machine. Useful for testing client behaviour or developing against a predictable endpoint without a real agent.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `9001` | Listen port |
| `--host` | `127.0.0.1` | Bind address |
| `--echo` | — | Echo mode: returns the user's message as the agent response |

## Usage

```bash
# Start an echo agent on the default port
a2acli serve --echo

# Bind to a different port
a2acli serve --echo --port 8080

# Expose on all interfaces (e.g. for Docker/CI)
a2acli serve --echo --host 0.0.0.0 --port 9001
```

Once running, point any `a2acli` command at it:

```bash
a2acli describe --service-url http://127.0.0.1:9001 --output json
a2acli send "Hello" --service-url http://127.0.0.1:9001 --output json --wait
```
