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
	"errors"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// fakeTaskGetter returns a scripted sequence of (task,err) pairs on successive
// GetTask calls; the last entry repeats once exhausted.
type fakeTaskGetter struct {
	calls   int
	results []struct {
		task *a2a.Task
		err  error
	}
}

func (f *fakeTaskGetter) GetTask(_ context.Context, _ *a2a.GetTaskRequest) (*a2a.Task, error) {
	i := f.calls
	if i >= len(f.results) {
		i = len(f.results) - 1
	}
	f.calls++
	return f.results[i].task, f.results[i].err
}

func taskInState(s a2a.TaskState) *a2a.Task {
	return &a2a.Task{ID: "t1", Status: a2a.TaskStatus{State: s}}
}

// TestIsWaitComplete verifies which states end a --wait poll (SPEC §9.3):
// terminal states plus the interrupted input-required / auth-required states.
func TestIsWaitComplete(t *testing.T) {
	done := []a2a.TaskState{
		a2a.TaskStateCompleted,
		a2a.TaskStateFailed,
		a2a.TaskStateCanceled,
		a2a.TaskStateRejected,
		a2a.TaskStateInputRequired,
		a2a.TaskStateAuthRequired,
	}
	for _, s := range done {
		if !isWaitComplete(s) {
			t.Errorf("isWaitComplete(%v) = false, want true", s)
		}
	}
	notDone := []a2a.TaskState{
		a2a.TaskStateSubmitted,
		a2a.TaskStateWorking,
	}
	for _, s := range notDone {
		if isWaitComplete(s) {
			t.Errorf("isWaitComplete(%v) = true, want false", s)
		}
	}
}

// TestWaitForTaskPollsToTerminal verifies waitForTask loops over non-terminal
// states and returns once a terminal state is observed.
func TestWaitForTaskPollsToTerminal(t *testing.T) {
	f := &fakeTaskGetter{results: []struct {
		task *a2a.Task
		err  error
	}{
		{taskInState(a2a.TaskStateSubmitted), nil},
		{taskInState(a2a.TaskStateWorking), nil},
		{taskInState(a2a.TaskStateCompleted), nil},
	}}

	got, err := waitForTask(context.Background(), f, &a2a.GetTaskRequest{ID: "t1"}, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status.State != a2a.TaskStateCompleted {
		t.Errorf("final state = %v, want completed", got.Status.State)
	}
	if f.calls != 3 {
		t.Errorf("GetTask called %d times, want 3", f.calls)
	}
}

// TestWaitForTaskContextTimeout verifies a context deadline surfaces as ctx.Err()
// (mapped by the caller to A2ACLI_ERR_TIMEOUT) rather than looping forever.
func TestWaitForTaskContextTimeout(t *testing.T) {
	f := &fakeTaskGetter{results: []struct {
		task *a2a.Task
		err  error
	}{
		{taskInState(a2a.TaskStateWorking), nil}, // never terminal
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := waitForTask(ctx, f, &a2a.GetTaskRequest{ID: "t1"}, 5*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

// TestWaitForTaskToleratesTransientErrors verifies a couple of transient GetTask
// errors are tolerated (reset on success) before a terminal state is reached.
func TestWaitForTaskToleratesTransientErrors(t *testing.T) {
	f := &fakeTaskGetter{results: []struct {
		task *a2a.Task
		err  error
	}{
		{nil, errors.New("blip")},
		{nil, errors.New("blip")},
		{taskInState(a2a.TaskStateWorking), nil}, // resets failure counter
		{nil, errors.New("blip")},
		{taskInState(a2a.TaskStateCompleted), nil},
	}}

	got, err := waitForTask(context.Background(), f, &a2a.GetTaskRequest{ID: "t1"}, time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Status.State != a2a.TaskStateCompleted {
		t.Errorf("final state = %v, want completed", got.Status.State)
	}
}

// TestWaitForTaskGivesUpAfterSuccessiveFailures verifies persistent GetTask
// failures abort after the cap rather than looping forever.
func TestWaitForTaskGivesUpAfterSuccessiveFailures(t *testing.T) {
	f := &fakeTaskGetter{results: []struct {
		task *a2a.Task
		err  error
	}{
		{nil, errors.New("down")},
	}}

	_, err := waitForTask(context.Background(), f, &a2a.GetTaskRequest{ID: "t1"}, time.Millisecond)
	if err == nil {
		t.Fatal("expected error after successive polling failures")
	}
	if f.calls != pollMaxSuccessiveFailures {
		t.Errorf("GetTask called %d times, want %d", f.calls, pollMaxSuccessiveFailures)
	}
}
