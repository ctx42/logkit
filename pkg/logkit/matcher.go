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

// subscription represents the notification channel returned by
// [Matcher.Notify].
type subscription struct {
	ch   chan Entry     // Receives matched entries.
	done chan struct{}  // Closed when the subscription ends.
	wg   sync.WaitGroup // Counts sends in progress.
}

// Matcher represents log line matcher.
type Matcher struct {
	// Log messages fields and their formats.
	cfg *Config

	// Checks to run against the lines.
	checks []Checker

	// Number of times the matcher matched a line or entry.
	cnt int

	// Guards the structure fields.
	mx sync.Mutex

	// When not nil, matched log lines and entries are sent to its channel.
	sub *subscription

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
	if mcr.sub == nil {
		mcr.sub = &subscription{
			ch:   make(chan Entry),
			done: make(chan struct{}),
		}
		mcr.t.Cleanup(mcr.NotifyStop)
	}
	return mcr.sub.ch
}

// NotifyStop closes the notification channel returned by [Matcher.Notify].
// A match waiting for a receiver on that channel gives up and returns.
func (mcr *Matcher) NotifyStop() {
	mcr.mx.Lock()
	sub := mcr.sub
	mcr.sub = nil
	if sub != nil {
		close(sub.done)
	}
	mcr.mx.Unlock()

	if sub != nil {
		sub.wg.Wait()
		close(sub.ch)
	}
}

// MatchEntry runs all checks on the provided [Entry]. Returns true if all
// checks pass; otherwise, returns false.
//
// When [Matcher.Notify] was called, a matching entry is sent to the returned
// channel, and this call blocks until it is received or [Matcher.NotifyStop]
// is called.
func (mcr *Matcher) MatchEntry(ent Entry) bool {
	return mcr.match(ent)
}

// MatchLine decodes a log line into a map[string]any, creates an [Entry], and
// runs all checks on it. Returns the entry if all checks pass; otherwise,
// returns a zero-value entry.
func (mcr *Matcher) MatchLine(idx int, line []byte) Entry {
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
	if !mcr.match(ent) {
		return ZeroEntry(mcr.t, mcr.cfg)
	}
	return ent
}

// match runs all checks on ent. When they all pass, it counts the match and
// sends ent to the notification channel, if there is one, without holding
// the lock, so [Matcher.NotifyStop] can cancel the send. Returns true if all
// checks pass.
func (mcr *Matcher) match(ent Entry) bool {
	mcr.mx.Lock()
	for _, chk := range mcr.checks {
		if err := chk(ent); err != nil {
			mcr.mx.Unlock()
			return false
		}
	}
	mcr.cnt++
	sub := mcr.sub
	if sub != nil {
		sub.wg.Add(1)
	}
	mcr.mx.Unlock()

	if sub != nil {
		defer sub.wg.Done()
		select {
		case sub.ch <- ent:
		case <-sub.done:
		}
	}
	return true
}
