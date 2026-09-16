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
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// taskGetter is the minimal client surface waitForTask needs. *a2aclient.Client
// satisfies it; the interface exists so the polling loop can be unit-tested with
// a fake.
type taskGetter interface {
	GetTask(ctx context.Context, req *a2a.GetTaskRequest) (*a2a.Task, error)
}

// pollMaxSuccessiveFailures caps how many consecutive GetTask errors are
// tolerated before waitForTask gives up. Transient blips are ignored; a
// persistent failure aborts.
const pollMaxSuccessiveFailures = 3

// isWaitComplete reports whether a task state ends a --wait poll: a terminal
// state, or an interrupted state (input-required / auth-required) that requires
// the caller to act (SPEC §9.3).
func isWaitComplete(state a2a.TaskState) bool {
	return state.Terminal() ||
		state == a2a.TaskStateInputRequired ||
		state == a2a.TaskStateAuthRequired
}

// waitForTask polls GetTask until the task reaches a terminal or interrupted
// state, the context is cancelled/expires, or GetTask fails repeatedly (Roadmap
// A4; SPEC §9.3/§10.3). The first poll happens immediately; subsequent polls are
// spaced by interval. A context deadline surfaces as ctx.Err() so the caller can
// map it to A2ACLI_ERR_TIMEOUT.
func waitForTask(ctx context.Context, client taskGetter, req *a2a.GetTaskRequest, interval time.Duration) (*a2a.Task, error) {
	if interval <= 0 {
		interval = time.Second
	}
	successiveFailures := 0
	for {
		task, err := client.GetTask(ctx, req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			successiveFailures++
			if successiveFailures >= pollMaxSuccessiveFailures {
				return nil, fmt.Errorf("task %q: successive polling failures exceeded: %w", req.ID, err)
			}
		} else {
			successiveFailures = 0
			if isWaitComplete(task.Status.State) {
				return task, nil
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}
