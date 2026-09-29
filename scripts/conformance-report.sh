#!/usr/bin/env bash
set -euo pipefail

# Regenerates docs/CONFORMANCE_REPORT.md from a real `go test -v ./e2e/...` run.
#
# Every status line is DERIVED from that run. They used to be hardcoded, which is
# how the report came to claim "A2A v0.3.0: PASSING" while the log embedded a dozen
# lines below it — in the same file — showed that suite skipping. A suite that did
# not run is reported as NOT RUN, never as PASSING.
#
# Usage: scripts/conformance-report.sh
# Honours A2A_GO_SRC, A2A_SIMPLE_SRC and VERSION from the environment (the Makefile
# passes them through). Exits non-zero if the conformance suite itself failed.

OUT="docs/CONFORMANCE_REPORT.md"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo unknown)}"
A2A_GO_SRC="${A2A_GO_SRC:-../../github/a2a-go}"
A2A_SIMPLE_SRC="${A2A_SIMPLE_SRC:-../../a2a-simple}"

log="$(mktemp)"
trap 'rm -f "$log"' EXIT

echo "Running conformance suite (this starts real SUT servers)..."
test_status=0
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore \
  A2A_GO_SRC="$A2A_GO_SRC" \
  A2A_SIMPLE_SRC="$A2A_SIMPLE_SRC" \
  go test -v ./e2e/... >"$log" 2>&1 || test_status=$?

# status <subtest>... — FAILING if any named subtest failed, NOT RUN if any
# skipped or never reported a verdict, PASSING only if all of them passed.
status() {
  local suite result="PASSING"
  for suite in "$@"; do
    if grep -qE "^[[:space:]]*--- FAIL: TestConformance/${suite}[[:space:]]" "$log"; then
      echo "FAILING"
      return
    fi
    if ! grep -qE "^[[:space:]]*--- PASS: TestConformance/${suite}[[:space:]]" "$log"; then
      result="NOT RUN"
    fi
  done
  echo "$result"
}

# SDK provenance, for reproducing a given report.
sdk_remote="unknown"
sdk_ref="unknown"
if git -C "$A2A_GO_SRC" rev-parse --git-dir >/dev/null 2>&1; then
  sdk_remote="$(git -C "$A2A_GO_SRC" remote get-url origin 2>/dev/null |
    sed -e 's|^ssh://git@github.com/|github.com/|' \
      -e 's|^git@github.com:|github.com/|' \
      -e 's|^https://||' -e 's|\.git$||')"
  sdk_ref="$(git -C "$A2A_GO_SRC" describe --tags --exact-match 2>/dev/null ||
    git -C "$A2A_GO_SRC" rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
fi

{
  echo "# A2A Conformance Report"
  echo
  echo "**Date:** $(date +%Y-%m-%d)"
  echo "**CLI Version:** ${VERSION}"
  echo "**SDK Source:** \`${sdk_remote:-unknown}\`"
  echo "**SDK Ref:** \`${sdk_ref:-unknown}\`"
  echo
  echo "## Conformance Status"
  echo
  echo "- A2A v1.0.0: **$(status JSON-RPC gRPC)** — against a2a-go's \`e2e/tck\` SUT"
  echo "- A2A v0.3.0: **$(status A2A-0.3.0)** — against the vendored \`e2e/fixtures/v0_3_sut\` SUT"
  echo "- A2UI Extension v1.0: **$(status A2UI-Extension-v1.0)** — against a2a-experiments' \`cmd/a2ui\`"
  echo
  echo "\`NOT RUN\` means the suite skipped because its system-under-test or its"
  echo "credentials were unavailable in the environment that generated this report."
  echo "It is not a pass."
  echo
  echo "### Test Results Summary"
  echo
  echo '```text'
  cat "$log"
  echo '```'
  echo
  echo "*(Auto-generated via make conformance-report)*"
} >"$OUT"

echo "Wrote $OUT"
exit "$test_status"
