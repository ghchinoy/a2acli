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
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// cardWith builds an agent card advertising the given transports, in order.
func cardWith(tps ...a2a.TransportProtocol) *a2a.AgentCard {
	c := &a2a.AgentCard{}
	for _, tp := range tps {
		c.SupportedInterfaces = append(c.SupportedInterfaces, &a2a.AgentInterface{ProtocolBinding: tp})
	}
	return c
}

func TestParseTransport(t *testing.T) {
	tests := []struct {
		in      string
		want    a2a.TransportProtocol
		wantErr bool
	}{
		{"grpc", a2a.TransportProtocolGRPC, false},
		{"GRPC", a2a.TransportProtocolGRPC, false},
		{"jsonrpc", a2a.TransportProtocolJSONRPC, false},
		{"rest", a2a.TransportProtocolHTTPJSON, false},
		{"httpjson", a2a.TransportProtocolHTTPJSON, false},
		{" grpc ", a2a.TransportProtocolGRPC, false},
		{"bogus", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseTransport(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTransport(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseTransport(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestSelectTransportSingleForces verifies that exactly one --transport forces
// that transport, unchanged from prior behavior — even when the card does not
// advertise it (back-compat guarantee of Roadmap A3).
func TestSelectTransportSingleForces(t *testing.T) {
	card := cardWith(a2a.TransportProtocolJSONRPC)
	got, forced, err := selectTransport([]string{"grpc"}, card)
	if err != nil {
		t.Fatal(err)
	}
	if got != a2a.TransportProtocolGRPC || !forced {
		t.Errorf("selectTransport(single grpc) = (%q, forced=%v), want (grpc, true)", got, forced)
	}
}

// TestSelectTransportNoneIsDynamic verifies that with no --transport the choice
// is dynamic (auto) from the card, forced=false.
func TestSelectTransportNoneIsDynamic(t *testing.T) {
	card := cardWith(a2a.TransportProtocolJSONRPC, a2a.TransportProtocolGRPC)
	got, forced, err := selectTransport(nil, card)
	if err != nil {
		t.Fatal(err)
	}
	// dynamic priority is gRPC > JSON-RPC > HTTP+JSON.
	if got != a2a.TransportProtocolGRPC || forced {
		t.Errorf("selectTransport(none) = (%q, forced=%v), want (grpc, false)", got, forced)
	}
}

// TestSelectTransportOrderedPreference verifies that a repeatable, ordered
// --transport list picks the first requested transport the card advertises
// (Roadmap A3), independent of the card's own ordering.
func TestSelectTransportOrderedPreference(t *testing.T) {
	// Card advertises jsonrpc and grpc (in that order).
	card := cardWith(a2a.TransportProtocolJSONRPC, a2a.TransportProtocolGRPC)

	// Request rest first (not advertised), then grpc (advertised) → grpc.
	got, forced, err := selectTransport([]string{"rest", "grpc"}, card)
	if err != nil {
		t.Fatal(err)
	}
	if got != a2a.TransportProtocolGRPC || !forced {
		t.Errorf("ordered [rest,grpc] = (%q, forced=%v), want (grpc, true)", got, forced)
	}

	// Request jsonrpc first, then grpc → jsonrpc wins by preference order even
	// though the card lists it first too.
	got, forced, err = selectTransport([]string{"jsonrpc", "grpc"}, card)
	if err != nil {
		t.Fatal(err)
	}
	if got != a2a.TransportProtocolJSONRPC || !forced {
		t.Errorf("ordered [jsonrpc,grpc] = (%q, forced=%v), want (jsonrpc, true)", got, forced)
	}
}

// TestSelectTransportOrderedFallback verifies that when none of the requested
// transports are advertised, selection falls back to the card's dynamic choice.
func TestSelectTransportOrderedFallback(t *testing.T) {
	card := cardWith(a2a.TransportProtocolHTTPJSON)
	got, forced, err := selectTransport([]string{"grpc", "jsonrpc"}, card)
	if err != nil {
		t.Fatal(err)
	}
	if got != a2a.TransportProtocolHTTPJSON || forced {
		t.Errorf("ordered fallback = (%q, forced=%v), want (httpjson, false)", got, forced)
	}
}

// TestSelectTransportInvalidValue verifies a bad transport value is an error.
func TestSelectTransportInvalidValue(t *testing.T) {
	if _, _, err := selectTransport([]string{"grpc", "bogus"}, cardWith(a2a.TransportProtocolGRPC)); err == nil {
		t.Error("expected error for invalid transport value in ordered list")
	}
	if _, _, err := selectTransport([]string{"bogus"}, nil); err == nil {
		t.Error("expected error for invalid single transport value")
	}
}

// TestPrimaryTransport verifies primaryTransport returns the first requested
// value or empty string.
func TestPrimaryTransport(t *testing.T) {
	transports = nil
	if got := primaryTransport(); got != "" {
		t.Errorf("primaryTransport() with none = %q, want empty", got)
	}
	transports = []string{"grpc", "rest"}
	if got := primaryTransport(); got != "grpc" {
		t.Errorf("primaryTransport() = %q, want grpc", got)
	}
	transports = nil
}
