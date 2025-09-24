// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/ctx42/testing/pkg/tester"
)

// Matcher represents log line matcher.
type Matcher struct {
	cfg    *Config             // Log configuration.
	checks []func(Entry) error // Checks to run against the lines.
	done   chan struct{}       // Closed by match od discard methods.
	ent    Entry               // The first found log entry.
	mx     sync.Mutex          // Guards Matcher checks.
	t      tester.T            // Test manager.
}

// NewMatcher returns a new instance of [Matcher] for all the provided checks.
func NewMatcher(t tester.T, cfg *Config, checks ...func(Entry) error) *Matcher {
	t.Helper()
	mcr := &Matcher{
		cfg:    cfg,
		checks: checks,
		done:   make(chan struct{}),
		mx:     sync.Mutex{},
		t:      t,
	}
	t.Cleanup(func() {
		mcr.mx.Lock()
		if mcr.done != nil {
			close(mcr.done)
			mcr.done = nil
		}
		mcr.mx.Unlock()
	})
	return mcr
}

// Checks returns a copy of the checks.
func (mcr *Matcher) Checks() []func(Entry) error {
	return slices.Clone(mcr.checks)
}

// Done returns the channel which is closed once the first [Entry] matching all
// checks is found.
func (mcr *Matcher) Done() <-chan struct{} {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()
	return mcr.done
}

// IsDone returns true if at least one [Entry] matching all checks was found.
func (mcr *Matcher) IsDone() bool {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()
	return mcr.done == nil
}

// Entry returns a copy of the first [Entry] matching all the checks or zero
// value if it was not yet found.
func (mcr *Matcher) Entry() Entry {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()

	ent := mcr.ent
	ent.m = maps.Clone(ent.m)
	return ent
}

// match processes JSON encoded log lines. It decodes the line as
// map[string]any, creates an [Entry] object from it and runs all the provided
// checks on it. It closes the done channel, saves the [Entry] and returns true
// only if the line matches all the checks. Otherwise, it returns false without
// closing the done channel.
func (mcr *Matcher) match(idx int, line []byte) bool {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()

	line = bytes.TrimSpace(line)
	dst := make(map[string]any)
	if err := json.Unmarshal(line, &dst); err != nil {
		mcr.t.Error(fmt.Errorf("matcher line %d: %w", idx, err))
		return false
	}

	var first Entry
	for _, chk := range mcr.checks {
		ent := Entry{
			cfg: mcr.cfg,
			raw: string(line),
			m:   maps.Clone(dst),
			idx: idx,
			t:   mcr.t,
		}
		if err := chk(ent); err != nil {
			return false
		}
		// The entry is set to the first one that passed the check.
		if first.IsZero() {
			// Entry.m may have been modified by the check function.
			first = Entry{
				cfg: mcr.cfg,
				raw: string(line),
				m:   dst,
				idx: idx,
				t:   mcr.t,
			}
		}
	}

	if mcr.done != nil {
		mcr.ent = first
		close(mcr.done)
		mcr.done = nil
	}
	return true
}

// discard discards the [Matcher] instance by closing the done channel and
// assigning nil to it.
func (mcr *Matcher) discard() {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()

	if mcr.done != nil {
		close(mcr.done)
	}
	mcr.done = nil
	mcr.ent = Entry{}
}
