// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"iter"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/a2aproject/a2a-go/v2/a2a"
	a2agrpc "github.com/a2aproject/a2a-go/v2/a2agrpc/v1"
	"github.com/a2aproject/a2a-go/v2/a2asrv"
)

var (
	servePort  int
	serveHost  string
	serveEcho  bool
	serveProxy string
	serveExec  string
)

func setupServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		GroupID: GroupServer,
		Short:   "Start an A2A-compliant mock server",
		Long: `Spin up a local A2A-compliant agent for testing purposes.
Currently, this CLI supports the --echo mock mode.`,
		Example: `  a2acli serve --echo --port 9001`,
		Run:     runServe,
	}

	cmd.Flags().IntVar(&servePort, "port", 9001, "Listen port")
	cmd.Flags().StringVar(&serveHost, "host", "127.0.0.1", "Bind address")
	cmd.Flags().BoolVar(&serveEcho, "echo", false, "Echo mode: return the user's message as a response")
	cmd.Flags().StringVar(&serveProxy, "proxy", "", "Proxy mode (not yet supported)")
	cmd.Flags().StringVar(&serveExec, "exec", "", "Exec mode (not yet supported)")

	return cmd
}

func runServe(_ *cobra.Command, _ []string) {
	if serveProxy != "" || serveExec != "" {
		fatalf("unsupported mode", nil, "Proxy and Exec modes are not yet supported in a2acli. Please use --echo.")
	}
	if !serveEcho {
		fatalf("missing mode", nil, "You must specify a mode to serve, e.g., --echo")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	addr := fmt.Sprintf("%s:%d", serveHost, servePort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fatalf("failed to listen", err, "")
	}

	card := &a2a.AgentCard{
		Name:         "a2acli-mock-agent",
		Description:  "A simple echo agent spun up via a2acli",
		Version:      "1.0.0",
		Capabilities: a2a.AgentCapabilities{Streaming: true},
	}

	// Determine transport
	selectedTransport := a2a.TransportProtocolHTTPJSON
	switch transport {
	case "jsonrpc":
		selectedTransport = a2a.TransportProtocolJSONRPC
	case "grpc":
		selectedTransport = a2a.TransportProtocolGRPC
	}

	if selectedTransport == a2a.TransportProtocolGRPC {
		card.SupportedInterfaces = []*a2a.AgentInterface{a2a.NewAgentInterface(addr, selectedTransport)}
	} else {
		card.SupportedInterfaces = []*a2a.AgentInterface{a2a.NewAgentInterface("http://"+addr, selectedTransport)}
	}

	handler := a2asrv.NewHandler(&echoExecutor{})

	if !disableTUI {
		fmt.Printf("Starting Mock Agent (%s) on %s\n", selectedTransport, addr)
	}

	if selectedTransport == a2a.TransportProtocolGRPC {
		s := grpc.NewServer()
		a2agrpc.NewHandler(handler).RegisterWith(s)

		cardMux := http.NewServeMux()
		cardMux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

		go func() {
			cardListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", serveHost, servePort+1))
			if err == nil {
				if !disableTUI {
					fmt.Printf("Agent card HTTP server running on %s%s\n", cardListener.Addr(), a2asrv.WellKnownAgentCardPath)
				}
				_ = http.Serve(cardListener, cardMux)
			}
		}()

		go func() {
			<-ctx.Done()
			s.GracefulStop()
		}()

		if err := s.Serve(listener); err != nil {
			fatalf("grpc server failed", err, "")
		}
	} else {
		mux := http.NewServeMux()
		mux.Handle(a2asrv.WellKnownAgentCardPath, a2asrv.NewStaticAgentCardHandler(card))

		if selectedTransport == a2a.TransportProtocolJSONRPC {
			mux.Handle("/", a2asrv.NewJSONRPCHandler(handler))
		} else {
			mux.Handle("/", a2asrv.NewRESTHandler(handler))
		}

		srv := &http.Server{Handler: mux}

		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(context.Background())
		}()

		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			fatalf("http server failed", err, "")
		}
	}
}

type echoExecutor struct{}

func (e *echoExecutor) Execute(_ context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		if execCtx.StoredTask == nil {
			if !yield(a2a.NewSubmittedTask(execCtx, execCtx.Message), nil) {
				return
			}
		}
		if !yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateWorking, nil), nil) {
			return
		}

		var text string
		for _, part := range execCtx.Message.Parts {
			if tp, ok := part.Content.(a2a.Text); ok {
				text += string(tp)
			}
		}

		evt := a2a.NewArtifactEvent(execCtx, a2a.NewTextPart(text))
		evt.LastChunk = true
		if !yield(evt, nil) {
			return
		}
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCompleted, nil), nil)
	}
}

func (e *echoExecutor) Cancel(_ context.Context, execCtx *a2asrv.ExecutorContext) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		yield(a2a.NewStatusUpdateEvent(execCtx, a2a.TaskStateCanceled, nil), nil)
	}
}
