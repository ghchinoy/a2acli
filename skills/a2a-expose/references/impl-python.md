# Python Implementation Reference (`a2a-sdk` >=1.0)

> **Version Guard:** This reference applies to `a2a-sdk` version **1.0.0 or higher**. Before writing code, verify the installed SDK version via `python -c "import a2a; print(a2a.__version__)"`. Do NOT use pre-1.0 SDK APIs.

---

## 1. Minimal FastAPI / Starlette Server Layout

```python
import asyncio
from typing import AsyncGenerator
from fastapi import FastAPI
from uvicorn import run

from a2a.server.agent_execution import AgentExecutor, RequestContext
from a2a.server.apps import A2AStarletteApp
from a2a.server.request_handlers import DefaultRequestHandlerV2
from a2a.server.task_store import InMemoryTaskStore
from a2a.types import (
    AgentCard,
    AgentCapabilities,
    AgentInterface,
    AgentSkill,
    Event,
    Message,
    Task,
    TaskState,
    TaskStatusUpdateEvent,
    TaskArtifactUpdateEvent,
    TransportProtocol,
)

# 1. Implement AgentExecutor
class MyAgentExecutor(AgentExecutor):
    async def execute(self, ctx: RequestContext) -> AsyncGenerator[Event, None]:
        # Emit submitted task if new
        if not ctx.stored_task:
            yield Task(id=ctx.task_id, context_id=ctx.context_id, status={"state": TaskState.SUBMITTED})

        # Emit working status
        yield TaskStatusUpdateEvent(
            task_id=ctx.task_id,
            context_id=ctx.context_id,
            status={"state": TaskState.WORKING}
        )

        # Emit result artifact
        yield TaskArtifactUpdateEvent(
            task_id=ctx.task_id,
            context_id=ctx.context_id,
            artifact={"name": "Output Report", "parts": [{"text": "Task finished."}]},
            last_chunk=True
        )

        # Emit completed status
        yield TaskStatusUpdateEvent(
            task_id=ctx.task_id,
            context_id=ctx.context_id,
            status={"state": TaskState.COMPLETED}
        )

    async def cancel(self, ctx: RequestContext) -> AsyncGenerator[Event, None]:
        yield TaskStatusUpdateEvent(
            task_id=ctx.task_id,
            context_id=ctx.context_id,
            status={"state": TaskState.CANCELED}
        )

# 2. Build AgentCard
card = AgentCard(
    name="Python Custom Agent",
    version="1.0",
    description="A Python A2A service.",
    supported_interfaces=[
        AgentInterface(url="http://127.0.0.1:9001/invoke", protocol_binding=TransportProtocol.JSONRPC),
        AgentInterface(url="http://127.0.0.1:9001/", protocol_binding=TransportProtocol.HTTP_JSON)
    ],
    default_input_modes=["text"],
    default_output_modes=["text"],
    capabilities=AgentCapabilities(streaming=True),
    skills=[
        AgentSkill(
            id="python_task",
            name="Python Task",
            description="Executes python task.",
            tags=["python"],
            examples=["Run python task"]
        )
    ]
)

# 3. Create RequestHandler & Application
task_store = InMemoryTaskStore()
handler = DefaultRequestHandlerV2(
    agent_executor=MyAgentExecutor(),
    task_store=task_store,
    agent_card=card
)

app = FastAPI()
a2a_app = A2AStarletteApp(request_handler=handler, agent_card=card)
app.mount("/", a2a_app)

if __name__ == "__main__":
    run(app, host="127.0.0.1", port=9001)
```

---

## 2. Verification Checklist for Python Developers

- Ensure `a2a-sdk` is installed with `fastapi` or `http-server` extra (`pip install "a2a-sdk[fastapi]"`).
- Mount the A2A sub-app at the root path (`/`) so standard routes (`/.well-known/agent-card.json`, `/invoke`, etc.) resolve properly.
- Verify async generator (`AsyncGenerator[Event, None]`) yields terminal state before exiting.
