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
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// parseTransport maps a --transport value to its A2A TransportProtocol.
// "rest" and "httpjson" both map to HTTP+JSON.
func parseTransport(s string) (a2a.TransportProtocol, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "grpc":
		return a2a.TransportProtocolGRPC, nil
	case "jsonrpc":
		return a2a.TransportProtocolJSONRPC, nil
	case "rest", "httpjson":
		return a2a.TransportProtocolHTTPJSON, nil
	default:
		return "", fmt.Errorf("unsupported transport: %s", s)
	}
}

// cardAdvertised returns the transports advertised by the card's supported
// interfaces, deduplicated and in the card's declared order.
func cardAdvertised(card *a2a.AgentCard) []a2a.TransportProtocol {
	if card == nil {
		return nil
	}
	var out []a2a.TransportProtocol
	seen := make(map[a2a.TransportProtocol]bool)
	for _, iface := range card.SupportedInterfaces {
		if iface != nil && iface.ProtocolBinding != "" && !seen[iface.ProtocolBinding] {
			seen[iface.ProtocolBinding] = true
			out = append(out, iface.ProtocolBinding)
		}
	}
	return out
}

// dynamicTransport picks a transport from the card by the fixed priority
// gRPC > JSON-RPC > HTTP+JSON, defaulting to JSON-RPC when the card advertises
// none of them.
func dynamicTransport(card *a2a.AgentCard) a2a.TransportProtocol {
	available := make(map[a2a.TransportProtocol]bool)
	for _, tp := range cardAdvertised(card) {
		available[tp] = true
	}
	switch {
	case available[a2a.TransportProtocolGRPC]:
		return a2a.TransportProtocolGRPC
	case available[a2a.TransportProtocolJSONRPC]:
		return a2a.TransportProtocolJSONRPC
	case available[a2a.TransportProtocolHTTPJSON]:
		return a2a.TransportProtocolHTTPJSON
	default:
		return a2a.TransportProtocolJSONRPC
	}
}

// selectTransport resolves the transport to use from an ordered --transport
// preference list (Roadmap A3) and the agent card. It returns the chosen
// transport and whether it was forced (explicitly requested) rather than
// auto-selected.
//
// Semantics:
//   - No --transport: dynamic selection from the card (auto).
//   - Exactly one --transport: that transport is forced, exactly as before —
//     it is used even if the card does not advertise it (back-compat).
//   - More than one --transport: the values are an ordered preference list; the
//     first requested transport the card advertises is selected. When none of
//     the requested transports are advertised, selection falls back to the
//     card's advertised transports (dynamic priority).
func selectTransport(reqs []string, card *a2a.AgentCard) (a2a.TransportProtocol, bool, error) {
	switch len(reqs) {
	case 0:
		return dynamicTransport(card), false, nil
	case 1:
		tp, err := parseTransport(reqs[0])
		if err != nil {
			return "", false, err
		}
		return tp, true, nil
	default:
		available := make(map[a2a.TransportProtocol]bool)
		for _, tp := range cardAdvertised(card) {
			available[tp] = true
		}
		parsed := make([]a2a.TransportProtocol, 0, len(reqs))
		for _, r := range reqs {
			tp, err := parseTransport(r)
			if err != nil {
				return "", false, err
			}
			parsed = append(parsed, tp)
		}
		for _, tp := range parsed {
			if available[tp] {
				return tp, true, nil
			}
		}
		return dynamicTransport(card), false, nil
	}
}

// primaryTransport returns the first requested transport value, or "" when none
// was given. It exists for callers that only support a single forced transport
// (e.g. the mock `serve` command and config display).
func primaryTransport() string {
	if len(transports) > 0 {
		return transports[0]
	}
	return ""
}
