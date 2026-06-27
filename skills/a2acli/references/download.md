# download — Download Artifacts from a Task

Downloads artifacts from a completed task to the local filesystem. Equivalent to `get --out-dir` but focused purely on artifact retrieval.

## Flags

| Flag | Short | Description |
|---|---|---|
| `--out-dir` | `-o` | Directory to save artifacts to |
| `--file` | `-f` | Save artifact to a specific filename (index appended for multiples) |

## Usage

```bash
# Download all artifacts to a directory
a2acli download <task_id> --out-dir ./downloads --service-url http://localhost:9001 -n

# Download to a specific filename
a2acli download <task_id> --file result.pdf --service-url http://localhost:9001 -n
```

If multiple artifacts are returned and `--file` is used, the CLI appends an index for subsequent files (e.g., `result.pdf`, `result_1.pdf`, `result_2.pdf`).
