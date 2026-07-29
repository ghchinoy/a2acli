#!/bin/sh
# verify-exposure.sh — Automated verification battery for new A2A exposures
set -e

SERVICE_URL="${1:-http://127.0.0.1:9001}"
A2ACLI_BIN="${2:-a2acli}"

if ! command -v "$A2ACLI_BIN" >/dev/null 2>&1; check_local_bin="../bin/a2acli"; then
  if [ -x "$check_local_bin" ]; then
    A2ACLI_BIN="$check_local_bin"
  elif [ -x "./bin/a2acli" ]; then
    A2ACLI_BIN="./bin/a2acli"
  else
    echo "ERROR: a2acli binary not found. Please build or install a2acli." >&2
    exit 1
  fi
fi

echo "==================================================" >&2
echo "Verifying A2A Exposure at $SERVICE_URL" >&2
echo "==================================================" >&2

# 1. Discover Card
echo "[1/5] Fetching Agent Card..." >&2
CARD=$("$A2ACLI_BIN" discover --service-url "$SERVICE_URL" --output json)
echo "$CARD" | grep -q "name" || { echo "FAIL: Agent card missing 'name'"; exit 1; }
echo "✓ Agent Card valid." >&2

# 2. Conformance Smoke Test
echo "[2/5] Running a2acli conformance smoke checks..." >&2
SMOKE=$("$A2ACLI_BIN" conformance --service-url "$SERVICE_URL" --output json 2>/dev/null || true)
if echo "$SMOKE" | grep -q '"passed":[[:space:]]*true'; then
  echo "✓ Conformance smoke checks passed." >&2
else
  echo "FAIL: Conformance smoke check failed." >&2
  echo "$SMOKE" >&2
  exit 1
fi

# 3. Synchronous Blocking Send
echo "[3/5] Testing SendMessage (--wait)..." >&2
SEND=$("$A2ACLI_BIN" send "verification battery message" --service-url "$SERVICE_URL" --wait --output json)
echo "$SEND" | grep -q "id" || { echo "FAIL: SendMessage did not return task ID"; exit 1; }
echo "✓ SendMessage (--wait) passed." >&2

# 4. List Tasks (Optional Capability Check)
echo "[4/5] Testing ListTasks..." >&2
LIST=$("$A2ACLI_BIN" list tasks --service-url "$SERVICE_URL" --limit 5 --output json 2>/dev/null || echo "{}")
if echo "$LIST" | grep -q "tasks"; then
  echo "✓ ListTasks passed." >&2
else
  echo "⚠ ListTasks skipped or auth-gated (optional capability)." >&2
fi

# 5. Summary
echo "==================================================" >&2
echo "SUCCESS: A2A exposure at $SERVICE_URL passed all verification battery checks!" >&2
echo "==================================================" >&2
