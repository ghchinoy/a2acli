# Vendored A2UI v1.0 JSON Schemas

These schemas are vendored verbatim from the official A2UI specification for
reproducible, offline conformance validation. Do not edit them by hand.

- **Source:** https://github.com/a2ui-project/a2ui
- **Path:** `specification/v1_0/json/` (and `specification/v1_0/catalogs/basic/catalog.json`)
- **Upstream commit:** `55d8a8aae08e` (main)
- **Vendored:** 2026-06-28

## Files

| File | Upstream location | Notes |
|------|-------------------|-------|
| `server_to_client.json` | `specification/v1_0/json/` | message envelope (oneOf 6 types) |
| `server_to_client_list.json` | `specification/v1_0/json/` | array of server→client messages (primary validation target) |
| `client_to_server.json` | `specification/v1_0/json/` | client→server message envelope |
| `client_to_server_list.json` | `specification/v1_0/json/` | array of client→server messages |
| `catalog.json` | `specification/v1_0/catalogs/basic/` | basic catalog; defines `anyComponent`, `surfaceProperties` |
| `catalog_definition.json` | `specification/v1_0/json/` | catalog definition schema |
| `common_types.json` | `specification/v1_0/json/` | shared `$defs` (CallId, FunctionCall, …) |
| `client_data_model.json` | `specification/v1_0/json/` | data-model sync schema |
| `client_capabilities.json` | `specification/v1_0/json/` | client capabilities |
| `server_capabilities.json` | `specification/v1_0/json/` | server capabilities |

## Ref-resolution note

`server_to_client.json` references `catalog.json#/$defs/{anyComponent,surfaceProperties}`,
which resolves (relative to its `$id` base) to
`https://a2ui.org/specification/v1_0/catalog.json`. The defs actually live in the
**basic catalog** (`catalogs/basic/catalog.json`), whose own `$id` is the deeper
`catalogs/basic/` path. The validator (`internal/conformance/a2ui/a2ui.go`)
registers that document under the bare `catalog.json` alias URL to satisfy the
ref without modifying any schema bytes.

## Updating

Re-fetch from the upstream commit, replace the files, update the commit hash
above, and run `go test ./internal/...` to confirm the schema set still compiles
and the fixtures still validate.
