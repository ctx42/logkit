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

// Checker represents a function which checks a log entry for a condition.
type Checker func(Entry) error

// Matcher represents log line matcher.
type Matcher struct {
	// Log messages fields and their formats.
	cfg *Config

	// Checks to run against the lines.
	checks []Checker

	// Number of times the marcher matched a line or entry.
	cnt int

	// Guards the structure fields.
	mx sync.Mutex

	// When not nil, it will be closed when a log line or [Entry] is matched.
	notify chan Entry

	// Test manager.
	t tester.T
}

// NewMatcher creates a new [Matcher] instance for the given checks.
// If no checks are provided, it matches all log lines.
// If config is nil, [DefaultConfig] is used.
func NewMatcher(t tester.T, cfg *Config, checks ...Checker) *Matcher {
	t.Helper()
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Matcher{cfg: cfg, checks: checks, t: t}
}

// Checks returns a copy of the checks.
func (mcr *Matcher) Checks() []Checker {
	return slices.Clone(mcr.checks)
}

// Matched returns the number of times the matcher matched a line or entry.
func (mcr *Matcher) Matched() int {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()
	return mcr.cnt
}

// Notify returns a channel for notifications when a log line or [Entry]
// matches. The channel closes automatically when the test ends.
func (mcr *Matcher) Notify() <-chan Entry {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()
	if mcr.notify == nil {
		mcr.notify = make(chan Entry)
		mcr.t.Cleanup(func() {
			mcr.mx.Lock()
			if mcr.notify != nil {
				close(mcr.notify)
				mcr.notify = nil
			}
			mcr.mx.Unlock()
		})
	}
	return mcr.notify
}

// NotifyStop closes the notification channel returned by [Matcher.Notify].
func (mcr *Matcher) NotifyStop() {
	mcr.mx.Lock()
	if mcr.notify != nil {
		close(mcr.notify)
		mcr.notify = nil
	}
	mcr.mx.Unlock()
}

// MatchEntry runs all checks on the provided [Entry]. Returns true if all
// checks pass; otherwise, returns false.
//
// When [Matcher.Notify] is called, it sends the entry to the channel returned
// if nothing listens on that channel, this call will block.
func (mcr *Matcher) MatchEntry(ent Entry) bool {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()

	for _, chk := range mcr.checks {
		if err := chk(ent); err != nil {
			return false
		}
	}
	if mcr.notify != nil {
		mcr.notify <- ent
	}
	mcr.cnt++
	return true
}

// MatchLine decodes a log line into a map[string]any, creates an [Entry], and
// runs all checks on it. Returns the entry if all checks pass; otherwise,
// returns a zero-value entry.
func (mcr *Matcher) MatchLine(idx int, line []byte) Entry {
	mcr.mx.Lock()
	defer mcr.mx.Unlock()

	line = bytes.TrimSpace(line)
	dst := make(map[string]any)
	if err := json.Unmarshal(line, &dst); err != nil {
		mcr.t.Error(fmt.Errorf("matcher line %d: %w", idx, err))
		return ZeroEntry(mcr.t, mcr.cfg)
	}

	ent := Entry{
		cfg: mcr.cfg,
		raw: string(line),
		m:   maps.Clone(dst),
		idx: idx,
		t:   mcr.t,
	}
	for _, chk := range mcr.checks {
		if err := chk(ent); err != nil {
			return ZeroEntry(mcr.t, mcr.cfg)
		}
	}
	if mcr.notify != nil {
		mcr.notify <- ent
	}
	mcr.cnt++
	return ent
}
