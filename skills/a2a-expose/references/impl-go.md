# Go Implementation Reference (`github.com/a2aproject/a2a-go/v2` v2.3.1)

This reference provides exact, verified code snippets for building an A2A server in Go using `a2a-go/v2` v2.3.1.

---

## 1. Minimal Server Layout

```go
package main

import (
	"context"
	"fmt"
	"iter"
	"net"
	"net/http"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
	"github.com/a2aproject/a2a-go/v2/a2asrv/taskstore"
)

// 1. Implement a2asrv.AgentExecutor
type myExecutor struct{}

var _ a2asrv.AgentExecutor = (*myExecutor)(nil)

func (e *myExecutor) Execute(ctx context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		// Emit initial task if new
		if execCtx.StoredTask == nil {
			if !yield(a2a.NewSubmittedTask(execCtx, execCtx.Message), nil) {
				return
			}
		}

		// Emit working status
		if !yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateWorking, nil), nil) {
			return
		}

		// Produce output artifact
		text := "Execution completed successfully."
		artEvent := a2a.NewArtifactEvent(execCtx, a2a.NewTextPart(text))
		artEvent.Artifact.Name = "Result"
		artEvent.LastChunk = true
		if !yield(artEvent, nil) {
			return
		}

		// Emit completed status
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, nil), nil)
	}
}

func (e *myExecutor) Cancel(ctx context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCanceled, nil), nil)
	}
}

func main() {
	// 2. Build AgentCard
	card := &a2a.AgentCard{
		Name:        "My Custom Agent",
		Version:     "1.0",
		Description: "A custom Go A2A service.",
		SupportedInterfaces: []*a2a.AgentInterface{
			a2a.NewAgentInterface("http://127.0.0.1:9001/invoke", a2a.TransportProtocolJSONRPC),
			a2a.NewAgentInterface("http://127.0.0.1:9001/", a2a.TransportProtocolHTTPJSON),
		},
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		Capabilities:       a2a.AgentCapabilities{Streaming: true},
		Skills: []a2a.AgentSkill{
			{
				ID:          "run_task",
				Name:        "Run Task",
				Description: "Executes the custom task.",
				Tags:        []string{"custom", "task"},
				Examples:    []string{"Run my task"},
			},
		},
	}

	// 3. Create RequestHandler with options
	store := taskstore.NewInMemory(&taskstore.InMemoryStoreConfig{
		Authenticator: a2asrv.NewTaskStoreAuthenticator(),
	})

	handler := a2asrv.NewHandler(&myExecutor{},
		a2asrv.WithTaskStore(store),
		a2asrv.WithCapabilityChecks(&card.Capabilities),
	)

	// 4. Wire HTTP Mux
	mux := http.NewServeMux()
	mux.Handle("/invoke", a2asrv.NewJSONRPCHandler(handler))
	mux.Handle("/", a2asrv.NewRESTHandler(handler))
	mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

	// 5. Listen and serve
	fmt.Println("Server running on :9001...")
	_ = http.ListenAndServe(":9001", mux)
}
```

---

## 2. Gotchas & Verification Checklist for Go Developers

- **Initial Task Signature:** Note that `a2a.NewSubmittedTask(execCtx, execCtx.Message)` takes an initial `*a2a.Message` as the 2nd argument (do NOT pass `TaskStateSubmitted` directly).
- **gRPC Interface URL:** When declaring a gRPC interface in `supportedInterfaces`, omit `http://` (use e.g. `"127.0.0.1:9015"`).
- **Mount Path:** Always register the REST handler at `/` so internal routes (`/message:send`, `/tasks/{id}`, etc.) match.
- **Card Endpoint:** Use `a2asrv.WellKnownAgentCardPath` (`/.well-known/agent-card.json`).
