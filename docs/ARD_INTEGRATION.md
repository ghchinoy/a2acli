# ARD Integration Design

How a2acli incorporates the [Agentic Resource Discovery (ARD) spec](https://github.com/ards-project/ard-spec)
so that **discovery** and **interaction** live in one command-line tool.

> Tracked as epic **a2ac-alu** in the bd issue tracker.

## Goal

a2acli today answers *"talk to the agent at this URL."* ARD lets it answer the
prior question — *"which agent should I talk to?"* — and then act on the answer
without switching tools or mental models.

```
ARD search / catalog  →  resolve agent card  →  discover / send / subscribe
```

## North Star: a smart target resolver, not a command silo

The design avoids bolting on a parallel `catalog`/`search` world that users must
context-switch into. Instead, **every command that needs an agent already takes a
target — we make the target smarter.** Discovery and use merge because the thing
you pass to `discover`/`send`/`subscribe` gets richer.

### Target grammar

| Target form | Resolution | Example |
|---|---|---|
| Raw URL | direct (current behavior) | `http://localhost:9001` |
| Bare host | try `/.well-known/agent-card.json`, then `ai-catalog.json` | `acme.com` |
| Local file | read card or catalog from disk | `./card.json` |
| ARD URN | resolve via configured catalog/registry | `urn:air:acme.com:agent:assistant` |
| `name@catalog` | resolve from a named catalog | `assistant@acme` |
| Bare name | search configured registries (may disambiguate) | `translator` |

So these are the *same* command — the resolver figures out the rest:

```bash
a2acli send "Bonjour" http://localhost:9001     # direct
a2acli send "Bonjour" translator                 # discovered
```

The resolver is **protocol-neutral internally**: each catalog entry carries an
IANA media type, and invocation is dispatched by type. A2A entries are invocable;
non-A2A entries (MCP servers, skills) are resolvable and displayable but clearly
marked "not invocable here" — see [Naming & Scope](#naming--scope).

## Configuration: catalogs & registries as first-class, alongside environments

Extends the existing XDG config (`~/.config/a2acli/config.yaml`):

```yaml
default_env: local
envs:
  local: { service_url: "http://127.0.0.1:9001" }
  prod:  { service_url: "https://agent.example.com", token: "..." }
  corp:  { registry: "https://registry.internal.corp/api/v1/" }  # env as a discovery source

# discovery sources
catalogs:
  acme: "https://acme.com/.well-known/ai-catalog.json"
registries:
  public: "https://finder.nlweb.ai/search"
  corp:   "https://registry.internal.corp/api/v1/"
default_registry: corp
```

`assistant@acme` resolves through the `acme` catalog; bare `a2acli search "..."`
hits `default_registry`. The existing precedence chain (flags > env > config >
defaults) extends unchanged.

## New verbs (discovery surfaces that feed the existing verbs)

```
a2acli catalog fetch|list|validate|add|remove|pull <...>
a2acli search "<query>" [--registry NAME] [--filter k=v] [--federation auto|referrals|none]
a2acli explore [--registry NAME] [--facet field]
a2acli agents                       # locally-known (cached) agents
```

These produce targets that the existing `discover`/`send`/`subscribe`/etc.
consume. They are surfaces, not silos.

## Three interaction tiers from the same primitives

**1. Scripting / agents (composable):**
```bash
a2acli search "translation agent" --output json \
  | jq -r '.results[0].identifier' \
  | xargs -I{} a2acli send "Bonjour" {} --wait --output json
```

**2. Power users (one-shot resolve-and-act):**
```bash
a2acli send "Bonjour" --search "translation agent" --top
```
Gated behind explicit `--search`/`--top` so it never silently guesses.

**3. Humans (interactive — the TUI becomes a discovery browser):**
```bash
a2acli search "translation agent"
# TUI: ranked results → arrow-select → view card → "Send a message?" → streams inline
```
This is the daily-driver experience. It gives the chat/REPL idea (**a2ac-srx**) a
clear entry point: the browser handles discover→select; the REPL handles the
converse loop after selection.

## Bridging discovery → repeated use: local cache

```bash
a2acli catalog pull acme    # cache acme's entries locally (XDG cache)
a2acli agents               # list known agents, fast & offline
a2acli send "hi" assistant  # resolves from cache instantly
```
Turns a2acli from "a tool I point at a URL" into "a tool that knows my agents"
(mirrors agntcy `.a2aagents/` and spec-works' install model).

## Conformance (a2acli's signature mode, applied to ARD)

```bash
a2acli conformance catalog <url|file>   # validate manifest vs ARD JSON Schema + semantic rules
a2acli conformance registry <url>       # probe /search, /explore, /agents for spec compliance
```
Mirrors how a2acli already tests A2A v0.3/v1.0. ARD ships formal schemas
(CDDL / JSON Schema / OpenAPI) and a Python conformance tool to target for parity.

## Self-contained test topology

Because a2acli already has `serve --echo`, it can stand up the entire ARD loop
with no external dependencies:

```
a2acli serve --echo     --port 9001   # the A2A agent under test
a2acli serve --catalog  manifest.json # static ai-catalog.json → points at :9001
a2acli serve --registry index.json    # mock registry indexing the catalog

# end-to-end:
a2acli search "echo agent" --registry http://localhost:9010   # find
a2acli discover <resolved-url>                                  # resolve
a2acli send "ping" <resolved-url> --wait                        # invoke
```

One binary, three roles, full discover→resolve→invoke fixture.

## Risk management

ARD is **v0.9 Draft** — IANA media types are pending and the wire format may
change. Therefore:

- Build the **stable foundation first**: resolver core + static catalog
  (well-known URI is registered; manifests are plain JSON).
- **Isolate the Draft registry REST API** behind an interface so a spec revision
  is a localized change.
- Treat `trustManifest`/SPIFFE verification as out of scope initially (the spec
  itself decouples trust from discovery).
- ARD support is net-new HTTP/JSON code in a2acli (the a2a-go SDK does not provide
  it), unlike the SDK-backed A2A calls.

## Naming & Scope

ARD catalogs list more than A2A agents (MCP servers, skills, datasets, nested
catalogs). Default behavior: **filter to A2A entries; show others with a clear
"not invocable here" note.**

This is the pressure point behind a possible future rename: the moment a2acli
resolves and *uses* non-A2A resources, "a2acli" is the wrong name (e.g. `agents`).
To keep that option cheap, the resolver and catalog core are **protocol-neutral**
— entries carry a `type`, invocation dispatches by type — so an eventual rename is
a packaging change, not a rearchitecture. No rename is proposed now.

## Phasing & task map

| Phase | bd task | Notes |
|---|---|---|
| 1 | **a2ac-alu.1** Target resolver core | Keystone; everything depends on it |
| 2 | **a2ac-alu.2** Static catalog support | Stable layer; blocks config + conformance |
| 2 | **a2ac-alu.3** Catalog/registry config + URN resolution | Ties into named environments |
| 2 | **a2ac-alu.4** Registry search client | Draft API — isolate behind interface |
| 3 | **a2ac-alu.5** Local cache / known-agents | discovery → repeated use |
| 3 | **a2ac-alu.6** TUI discovery browser | composes with a2ac-srx (REPL) |
| 3 | **a2ac-alu.7** ARD conformance testing | signature mode |
| 4 | **a2ac-alu.8** serve mock catalog + registry | self-contained test topology |

Dependency keystone: **a2ac-alu.1** blocks .2/.3/.4/.5/.6; **a2ac-alu.2** blocks
.3/.5/.7/.8; **a2ac-alu.4** blocks .6/.7/.8. Build the resolver + static catalog
first.

## Related

- [docs/COMPARISON.md](COMPARISON.md) — a2acli vs the community CLIs and proposals
- Issue [a2aproject/A2A#1929](https://github.com/a2aproject/A2A/issues/1929) — official CLI; catalogs explicitly deferred there
- spec-works/a2a-ask — prior art for ad-hoc catalog operations
