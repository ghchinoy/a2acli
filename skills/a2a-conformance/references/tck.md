# Tier 2: Technology Compatibility Kit (`a2aproject/a2a-tck`)

Tier 2 executes the official A2A Technology Compatibility Kit against the System Under Test (SUT).

---

## 1. Overview of the TCK

The TCK is a pytest-based compliance suite published by the A2A Project:
- Repository: `https://github.com/a2aproject/a2a-tck`
- Driver script: `./run_tck.py`
- Execution categories: `mandatory`, `capabilities`
- Output artifacts: `reports/compatibility.json`, `reports/compatibility.html`, `reports/junitreport.xml`

---

## 2. Prerequisites & Setup

The TCK requires Python 3.11+ and `uv`:

```bash
# Clone a2a-tck if not already present
git clone https://github.com/a2aproject/a2a-tck.git /tmp/a2a-tck
cd /tmp/a2a-tck

# Create virtual environment and install dependencies
uv venv
source .venv/bin/activate
uv pip install -e .
```

---

## 3. Execution Commands

### Run Full Conformance Suite
```bash
./run_tck.py --sut-host http://127.0.0.1:9001
```

### Run Specific Requirement Level (RFC 2119)
```bash
# Run MUST requirements (hard failures block compatibility)
./run_tck.py --sut-host http://127.0.0.1:9001 --level must

# Run SHOULD requirements (expected failures/warnings)
./run_tck.py --sut-host http://127.0.0.1:9001 --level should

# Run MAY requirements (optional capabilities)
./run_tck.py --sut-host http://127.0.0.1:9001 --level may
```

### Run Specific Transports
```bash
./run_tck.py --sut-host http://127.0.0.1:9001 --transport jsonrpc
./run_tck.py --sut-host http://127.0.0.1:9001 --transport grpc
./run_tck.py --sut-host http://127.0.0.1:9001 --transport http_json
```

---

## 4. Orchestration via Go SUT Helper

If testing a Go implementation using `a2a-go`, you can also run the TCK via `orchestrate_tck.py` inside `$A2A_GO_SRC/e2e/tck/`:

```bash
cd $A2A_GO_SRC/e2e/tck
./run_tck.sh --path /tmp/a2a-tck
```

---

## 5. Interpreting Results

Parse `reports/compatibility.json`:
- `summary.passed`: count of passed checks
- `summary.failed`: count of failed checks
- `summary.skipped`: count of skipped checks

**Rule:**
- Any **MUST** failure in `compatibility.json` results in a verdict of `NOT CONFORMANT`.
- Any **SHOULD** failure is reported under "Conformance Warnings".
- **MAY** skips are noted as omitted optional features.
