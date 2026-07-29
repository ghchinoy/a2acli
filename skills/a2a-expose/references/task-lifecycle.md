# Task Lifecycle, States & Event Delivery

This guide details the A2A Task state machine, event streams, and multi-turn execution patterns.

---

## 1. Task State Machine

A Task transitions through active, interrupted, and terminal states:

```
                  [Create Task]
                        │
                        ▼
               TASK_STATE_SUBMITTED
                        │
                        ▼
                TASK_STATE_WORKING
               ╱        │        ╲
              ╱         │         ╲
             ▼          ▼          ▼
TASK_STATE_INPUT_REQUIRED  TASK_STATE_AUTH_REQUIRED  (Processing)
             │          │          │
             └──────────┴──────────┼────────────────────────┐
                        │          │                        │
                        ▼          ▼                        ▼
                TASK_STATE_COMPLETED / TASK_STATE_FAILED / TASK_STATE_CANCELED / TASK_STATE_REJECTED
                                    (TERMINAL STATES)
```

### State Classification Table

| State Enum Value | Classification | Meaning / Behavior |
|---|---|---|
| `TASK_STATE_SUBMITTED` | Active | Task created and queued; work not yet started. |
| `TASK_STATE_WORKING` | Active | Agent is actively processing the request. |
| `TASK_STATE_INPUT_REQUIRED` | Interrupted | Agent needs user clarification/data to proceed. Accepts follow-up messages. |
| `TASK_STATE_AUTH_REQUIRED` | Interrupted | Agent requires mid-task credentials (in-task auth §7.6). Accepts follow-up messages. |
| `TASK_STATE_COMPLETED` | **Terminal** | Task finished successfully. Stream closes. Cannot accept further messages. |
| `TASK_STATE_FAILED` | **Terminal** | Task failed with error. Stream closes. Cannot accept further messages. |
| `TASK_STATE_CANCELED` | **Terminal** | Task was canceled by client or server. Stream closes. |
| `TASK_STATE_REJECTED` | **Terminal** | Server declined to execute the task. Stream closes. |

---

## 2. Emitting Events from an Agent Executor

When executing a task, an agent yields a sequence of `a2a.Event` objects:

1. **Initial Task Event:** `NewSubmittedTask(infoProvider, initialMessage)` (if task is new).
2. **Status Update (Working):** `NewStatusUpdateEvent(infoProvider, TaskStateWorking, nil)`.
3. **Artifact Update(s):** `NewArtifactEvent(infoProvider, parts...)` or `NewArtifactUpdateEvent(infoProvider, artifactID, parts...)`.
4. **Final Status Update (Completed/Failed):** `NewStatusUpdateEvent(infoProvider, TaskStateCompleted, nil)`.

---

## 3. Artifact Delivery Patterns

Per spec §3.7, outputs belong in **`Artifacts`**, not status messages:

```go
// Creating a text artifact
artEvent := a2a.NewArtifactEvent(execCtx, a2a.NewTextPart("Analysis report text..."))
artEvent.Artifact.Name = "Analysis Report"
artEvent.LastChunk = true
yield(artEvent, nil)

// Creating a file/data artifact
dataPart := a2a.NewDataPart(map[string]any{"count": 42, "status": "ok"})
artEvent := a2a.NewArtifactEvent(execCtx, dataPart)
artEvent.Artifact.Name = "Results JSON"
yield(artEvent, nil)
```

---

## 4. Multi-Turn Conversation Patterns

- **Continuing a Task:** Client sends message with `--task <taskId>` (or `taskId` set in `SendMessageRequest`). The server loads `StoredTask`, inspects prior messages and artifacts, and continues processing.
- **Referencing Completed Tasks:** Client sends message with `--ref <taskId>`. Server receives referenced tasks in `execCtx.RelatedTasks` to use as context without reopening the completed task.
- **Context Grouping:** Tasks sharing a `contextId` belong to the same logical conversation session.
