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
	"os"

	"github.com/spf13/cobra"
)

const defaultServiceURL = "http://127.0.0.1:9001"

// addGlobalFlags registers all persistent (global) flags on cmd, including the
// canonical spec/OFFICIAL flag aliases (Roadmap A1). It is called from main() for
// rootCmd and is exported at package level so tests can register the same flag
// set on a throwaway command and exercise alias resolution and precedence.
//
// Alias flags bind to the SAME variable as their legacy target and carry the
// SAME default, so: (a) any spelling sets the one value; (b) registration order
// cannot clobber the default; and (c) when a setting is given under two spellings
// on one command line, pflag applies them left-to-right and the last wins — never
// a silent double-apply.
//
// Short-flag collisions: the spec pairs --endpoint with -e, but -e is already
// taken here by --env, so --endpoint is registered long-only; --agent-card keeps
// its canonical -a (free).
func addGlobalFlags(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()

	pf.StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/a2acli/config.yaml)")
	pf.StringVarP(&envName, "env", "e", "", "environment name to load from config")
	pf.StringVarP(&serviceURL, "service-url", "u", defaultServiceURL, "Base URL of the A2A service")
	pf.StringVarP(&authToken, "token", "t", "", "Auth token (legacy alias for --bearer; --bearer wins if both are set)")
	// Canonical Tier-1 credential flags (SPEC §12.1 / §7.2). --bearer attaches
	// an Authorization: Bearer <token> header; --api-key attaches the API-key
	// header named by the agent card's declared APIKeySecurityScheme (default
	// X-Api-Key). Each has a canonical env equivalent (A2ACLI_BEARER /
	// A2ACLI_API_KEY); an explicit flag overrides the env. These are additive to
	// the existing --token/--auth mechanisms (back-compat). When both --bearer and
	// the legacy --token are supplied, --bearer takes precedence.
	pf.StringVar(&bearerToken, "bearer", "", "Bearer token credential (Authorization: Bearer <token>); env A2ACLI_BEARER. Takes precedence over --token when both are set")
	pf.StringVar(&apiKey, "api-key", "", "API key credential attached per the card's declared scheme (default header X-Api-Key); env A2ACLI_API_KEY")
	pf.StringSliceVar(&authHeaders, "auth", nil, "Authorization headers to send (e.g. 'Bearer ...')")
	pf.StringSliceVar(&svcParams, "svc-param", nil, "Service parameters to send (e.g. 'key=value')")
	pf.StringVarP(&targetTaskID, "task", "k", "", "Existing Task ID to continue (for active tasks)")
	pf.StringVar(&contextID, "context", "", "Context ID for multi-turn conversation thread")
	pf.StringVarP(&refTaskID, "ref", "r", "", "Task ID to reference for cross-task artifact chaining (does not continue conversation)")
	pf.BoolVar(&strictMode, "strict", false, "Fail fast on warnings (e.g. continuing terminal tasks)")
	pf.BoolVar(&noCache, "no-cache", false, "Bypass agent card disk cache and fetch fresh")
	pf.BoolVarP(&disableTUI, "no-tui", "n", false, "Disable the Terminal UI — alias for --output json (backwards compat)")
	pf.StringVarP(&outputMode, "output", "o", "", "Output mode: tui (default), text (plain, no animations), json (single doc, or JSONL with --stream), jsonl (alias of json for scripting)")
	pf.DurationVar(&requestTimeout, "timeout", 0, "Request timeout, e.g. 30s, 2m (0 = no timeout)")
	pf.BoolVarP(&verbose, "verbose", "v", false, "Print diagnostic info to stderr (also: A2ACLI_VERBOSE=true)")
	// --transport is repeatable and ordered (Roadmap A3): each occurrence appends
	// to an ordered preference list (highest preference first). A single
	// --transport behaves exactly as before — it forces that transport. With more
	// than one, the first requested transport the card advertises is selected,
	// falling back to the card's advertised transports when none match.
	pf.StringArrayVar(&transports, "transport", nil, "Transport preference (grpc, jsonrpc, rest); repeatable and ordered, highest preference first. A single value forces that transport.")
	pf.StringVarP(&protocol, "protocol", "p", "1.0.0", "A2A protocol version (1.0.0 or 0.3.0)")

	// Canonical (spec/OFFICIAL) flag spellings, registered as additive aliases
	// bound to the SAME underlying variables as OURS' existing flags (Roadmap A1).
	pf.StringVarP(&serviceURL, "agent-card", "a", defaultServiceURL, "Alias of --service-url/-u: base URL (or agent-card URL) of the A2A service; env A2ACLI_AGENT_CARD")
	pf.StringVar(&serviceURL, "endpoint", defaultServiceURL, "Alias of --service-url/-u: base URL of the A2A service; env A2ACLI_ENDPOINT (short -e is taken by --env, so long form only)")
	pf.StringVar(&contextID, "context-id", "", "Alias of --context: Context ID for a multi-turn conversation thread")
	pf.StringVar(&targetTaskID, "task-id", "", "Alias of --task/-k: existing Task ID to continue")
	pf.StringVar(&protocol, "a2a-version", "1.0.0", "Alias of --protocol/-p: A2A protocol version (1.0.0 or 0.3.0)")
}

// serviceURLSpellings lists every accepted flag name that sets the service URL:
// the legacy --service-url and the canonical aliases --endpoint / --agent-card
// (Roadmap A1). They share one variable, so any of them being set counts as an
// explicit service-URL flag.
var serviceURLSpellings = []string{"service-url", "endpoint", "agent-card"}

// serviceURLFlagChanged reports whether the service URL was set explicitly on the
// command line under any of its accepted spellings. It is used so that an env
// alias or a config profile does not override an explicit flag (precedence:
// flag > env > config-file > default).
func serviceURLFlagChanged() bool {
	for _, n := range serviceURLSpellings {
		if f := rootCmd.PersistentFlags().Lookup(n); f != nil && f.Changed {
			return true
		}
	}
	return false
}

// transportFlagChanged reports whether --transport was set on the command line.
func transportFlagChanged() bool {
	if f := rootCmd.PersistentFlags().Lookup("transport"); f != nil {
		return f.Changed
	}
	return false
}

// resolveAliasEnv applies canonical environment aliases for the service URL when
// no service-URL flag was passed (Roadmap A1). Precedence is flag > env >
// config-file > default: an explicit flag (any spelling) is never overridden, and
// an env alias overrides a config-profile value. It runs after initConfig so an
// env alias wins over a config-profile service_url. Credential env aliases
// (A2ACLI_BEARER / A2ACLI_API_KEY) are handled separately in resolveCredentials.
func resolveAliasEnv() {
	if serviceURLFlagChanged() {
		return
	}
	for _, k := range []string{"A2ACLI_AGENT_CARD", "A2ACLI_ENDPOINT", "A2ACLI_SERVICE_URL"} {
		if v := os.Getenv(k); v != "" {
			serviceURL = v
			return
		}
	}
}
