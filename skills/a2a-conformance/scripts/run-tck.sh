#!/bin/sh
# run-tck.sh — Tier 2 official TCK orchestrator
set -e

SERVICE_URL="${1:-http://127.0.0.1:9001}"
TCK_DIR="${2:-/tmp/a2a-tck}"

if [ ! -d "$TCK_DIR" ]; then
  echo "Cloning a2a-tck repository to $TCK_DIR..." >&2
  git clone https://github.com/a2aproject/a2a-tck.git "$TCK_DIR"
fi

cd "$TCK_DIR"

if [ ! -d ".venv" ]; then
  echo "Setting up Python virtualenv in $TCK_DIR..." >&2
  uv venv
  . .venv/bin/activate
  uv pip install -e .
else
  . .venv/bin/activate
fi

echo "Running TCK against $SERVICE_URL..." >&2
python3 ./run_tck.py --sut-host "$SERVICE_URL" --level must "$@"
