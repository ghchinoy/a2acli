#!/bin/sh
# probe.sh — Tier 1 automated probe runner using a2acli
set -e

SERVICE_URL="${1:-http://127.0.0.1:9001}"
A2ACLI_BIN="${2:-a2acli}"

if ! command -v "$A2ACLI_BIN" >/dev/null 2>&1; check_local_bin="../bin/a2acli"; then
  if [ -x "$check_local_bin" ]; then
    A2ACLI_BIN="$check_local_bin"
  elif [ -x "./bin/a2acli" ]; then
    A2ACLI_BIN="./bin/a2acli"
  else
    echo "{\"error\": \"a2acli binary not found. Please build or install a2acli.\"}"
    exit 1
  fi
fi

echo "Running A2A Conformance Probes against $SERVICE_URL..." >&2

# 1. Discover Card
CARD_JSON=$("$A2ACLI_BIN" discover --service-url "$SERVICE_URL" --output json 2>/dev/null || echo "{}")

# 2. Conformance Smoke Command
SMOKE_JSON=$("$A2ACLI_BIN" conformance --service-url "$SERVICE_URL" --output json 2>/dev/null || echo "{}")

# 3. Round-trip send
SEND_JSON=$("$A2ACLI_BIN" send "conformance probe message" --service-url "$SERVICE_URL" --wait --output json 2>/dev/null || echo "{}")

cat <<EOF
{
  "service_url": "$SERVICE_URL",
  "agent_card": $CARD_JSON,
  "smoke_checks": $SMOKE_JSON,
  "probe_send": $SEND_JSON
}
EOF
