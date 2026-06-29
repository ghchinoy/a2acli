# auth — OAuth 2.1 Authentication

Manage OAuth 2.1 tokens for A2A agents that require authentication. Tokens are
stored at `~/.config/a2acli/tokens/` (0600 permissions) and used automatically
by all commands — no `--token` flag needed after `auth login`.

## Subcommands

| Command | Description |
|---|---|
| `auth login` | Obtain a token via auth-code + PKCE (interactive, opens browser) |
| `auth status` | Show stored token validity, expiry, and scope |
| `auth logout` | Delete the stored token for a service |
| `auth token` | Print the raw JWT access token (for scripting) |

## The auth workflow

```bash
# 1. One-time interactive login
a2acli auth login --service-url https://agent.example.com
# → browser opens → user signs in → token stored

# 2. All subsequent commands auto-use the stored token
a2acli send "hello" --service-url https://agent.example.com --output json --wait
a2acli conformance --service-url https://agent.example.com --output json

# 3. Check token validity
a2acli auth status --service-url https://agent.example.com

# 4. Logout when done
a2acli auth logout --service-url https://agent.example.com
```

## Token storage

Tokens are keyed by the service URL **host** and stored as JSON files at
`~/.config/a2acli/tokens/<host>.json`. Two different hostnames for the same
backend (e.g. `candir.mithlond.com` and `cano.mithlond.com`) have separate
token slots — run `auth login` for each hostname you want to use.

## For AI coding agents (non-interactive use)

`auth login` requires a browser and cannot run non-interactively. For automated
agent contexts (CI, coding agents), the pattern is:

```bash
# Human runs this once, interactively
a2acli auth login --service-url https://agent.example.com

# Agent retrieves the stored token for use in scripts
TOKEN=$(a2acli auth token --service-url https://agent.example.com)
a2acli send "do work" --service-url https://agent.example.com \
  --token "$TOKEN" --output json --wait
```

If no valid token is stored and `auth login` cannot be run (non-interactive),
the agent should fail with a clear message rather than attempting to authenticate
itself. Coordinate with the human operator to pre-authenticate.

## With named environments

```bash
# Login using an environment name
a2acli auth login --env mithlond

# All commands with --env mithlond then auto-authenticate
a2acli send "name star silver quenya" --env mithlond --skill name-generate --output json --wait
```

## How it works (PKCE + CIMD)

a2acli uses the **OAuth 2.1 auth-code + PKCE** flow:

1. Reads the AgentCard's `OAuth2SecurityScheme` to find the authorization and token endpoints.
2. Generates a PKCE `code_verifier` / `code_challenge` (S256).
3. Starts a local callback server on `http://127.0.0.1:8080/callback`.
4. Opens the browser to the authorization URL with the PKCE challenge.
5. User signs in; browser redirects to the local callback with an authorization code.
6. a2acli exchanges the code + verifier for an access token.
7. Token is stored at `~/.config/a2acli/tokens/<host>.json`.

a2acli identifies itself to the consent SPA using a **CIMD** (Client Instance
Metadata Document) at `https://ghchinoy.github.io/a2acli/metadata.json`. The
consent SPA fetches this URL at runtime to display the client name — no
pre-registration is required.

**Port 8080 must be free** when running `auth login`. If it is in use,
a2acli exits immediately with a clear error and Hint.
