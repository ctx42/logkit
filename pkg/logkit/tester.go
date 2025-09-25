// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// WithBytes is an option for [New] setting buffer to use inside the [Tester].
func WithBytes(buf []byte) func(*Tester) {
	return func(tst *Tester) { tst.buf = buf }
}

// WithString is an option for [New] which sets the value of the buffer inside the
// [Tester].
func WithString(content string) func(*Tester) {
	return func(tst *Tester) { tst.buf = []byte(content) }
}

// WithConfig is an option for [New] which sets [Tester] configuration
func WithConfig(cfg *Config) func(*Tester) {
	return func(tst *Tester) { tst.cfg = cfg }
}

// Tester represents a test utility for structured JSON log messages.
//
// Example usage:
//
//	tst := logkit.New(t)
//	log := zerolog.New(tst)
//	log.Info().Str("key0", "val0").Send()
//
//	tst.Entries().Summary() // Print logged messages.
type Tester struct {
	cfg      *Config      // Tester configuration.
	buf      []byte       // Buffer for logger writes.
	cnt      int          // Number of all log messages (calls to Write).
	matchers []*Matcher   // Log line matchers.
	matchIdx int          // Last matched log entry index (-1 means none).
	mx       sync.RWMutex // Guards the structure fields.
	t        tester.T     // Test manager.
}

// New creates a new instance of [Tester].
func New(t tester.T, opts ...func(*Tester)) *Tester {
	t.Helper()
	tst := &Tester{
		cfg:      DefaultConfig(),
		matchIdx: -1,
		t:        t,
	}
	for _, opt := range opts {
		opt(tst)
	}
	if tst.buf == nil {
		tst.buf = make([]byte, 0, 512)
	}

	scn := bufio.NewScanner(bytes.NewReader(tst.buf))
	for scn.Scan() {
		tst.cnt++
	}
	if err := scn.Err(); err != nil {
		t.Error(err)
	}
	return tst
}

// Load loads the existing log from the path.
func Load(t tester.T, pth string) *Tester {
	t.Helper()
	buf, err := os.ReadFile(pth)
	if err != nil {
		t.Error(err)
		return nil
	}
	return New(t, WithBytes(buf))
}

// Write implements [io.Writer] interface. It expects p to be a single log
// entry which is appended to the buffer. Every time it's called the count, the
// cnt counter is increased.
//
// If there are any matchers, it checks if the message matches the first one.
// If a match is found, it sets the matchIdx index to the value of cnt and
// removes the matcher from the "matchers" slice. This logic allows matching
// log lines in a specific order.
//
// It returns the number of bytes written and a nil error.
func (tst *Tester) Write(p []byte) (n int, err error) {
	tst.mx.Lock()
	defer tst.mx.Unlock()

	tst.cnt++
	tst.buf = append(tst.buf, p...)

	if len(tst.matchers) == 0 {
		return len(p), nil
	}

	cpy := slices.Clone(p)
	m := tst.matchers[0]
	if ent := m.MatchLine(tst.cnt-1, cpy); !ent.IsZero() {
		tst.matchIdx = tst.cnt - 1
		tst.matchers = tst.matchers[1:]
	}

	return len(p), nil
}

// Len returns a number of log messages written to the [Tester].
func (tst *Tester) Len() int {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	return tst.cnt
}

// String implements [fmt.Stringer] interface and returns everything written
// to the [Tester] so far.
func (tst *Tester) String() string {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	return string(tst.buf)
}

// Bytes returns everything written to the [Tester] so far.
func (tst *Tester) Bytes() []byte {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	return bytes.Clone(tst.buf)
}

// Entries returns all logged entries in the order they were logged. It marks
// the test as failed if log entries cannot be unmarshaled.
func (tst *Tester) Entries() Entries {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	tst.t.Helper()
	return tst.entries()
}

// entries returns [Entries] object containing parsed log entries from Tester's
// buffer. It uses a [json.NewDecoder] to iterate through the buffer and decode
// each entry into a map[string]any. It then creates a new [Entry] object for
// each decoded line and populates it with the necessary fields. Finally, it
// returns an [Entries] object containing the decoded entries. It marks the
// test as failed if log entries cannot be unmarshaled.
func (tst *Tester) entries() Entries {
	tst.t.Helper()

	ets := make([]Entry, 0, tst.cnt)

	var off int64
	dec := json.NewDecoder(bytes.NewReader(tst.buf))
	idx := 0
	for dec.More() {
		m := make(map[string]any)
		if err := dec.Decode(&m); err != nil {
			tst.t.Error(err)
			return Entries{cfg: tst.cfg, t: tst.t}
		}

		tmp := tst.buf[off:dec.InputOffset()]
		off = dec.InputOffset()
		ets = append(ets, Entry{
			cfg: tst.cfg,
			raw: string(bytes.TrimSpace(tmp)),
			m:   m,
			idx: idx,
			t:   tst.t,
		})
		idx++
	}
	return Entries{cfg: tst.cfg, ets: ets, t: tst.t}
}

// Filter returns entries matching the provided [Matcher].
func (tst *Tester) Filter(checks ...Checker) Entries {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	tst.t.Helper()

	mcr := NewMatcher(tst.t, tst.cfg, checks...)
	ets := make([]Entry, 0)
	for _, ent := range tst.Entries().Get() {
		if mcr.MatchEntry(ent) {
			ets = append(ets, ent)
		}
	}
	return Entries{cfg: tst.cfg, ets: ets, t: tst.t}
}

// FirstEntry returns the first log entry or zero value Entry if no log entries
// written to the [Tester]. It marks the test as failed if log entries cannot
// be unmarshaled.
func (tst *Tester) FirstEntry() Entry {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	tst.t.Helper()

	ets := tst.Entries().Get()
	if len(ets) == 0 {
		return Entry{t: tst.t}
	}
	return ets[0]
}

// LastEntry returns the first log entry or zero value Entry if no log entries
// written to the [Tester]. It marks the test as failed if log entries cannot
// be unmarshaled.
func (tst *Tester) LastEntry() Entry {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	tst.t.Helper()

	ets := tst.Entries().Get()
	if len(ets) == 0 {
		return Entry{t: tst.t}
	}
	return ets[len(ets)-1]
}

// ResetLastMatch resets the value of the matchIdx field to -1. This field is
// used to keep track of the last successfully matched log line. By resetting
// it to -1, the matching process starts from the beginning of the log lines.
func (tst *Tester) ResetLastMatch() {
	tst.mx.Lock()
	defer tst.mx.Unlock()
	tst.matchIdx = -1
}

// WaitFor waits for a log entry that satisfies the specified conditions within
// the given timeout duration. If the entry is not logged within the given
// timeout, it will mark the test as failed and return zero value [Entry].
func (tst *Tester) WaitFor(timeout string, checks ...Checker) Entry {
	tst.mx.Lock()
	tst.t.Helper()

	to, err := time.ParseDuration(timeout)
	if err != nil {
		tst.t.Error(err)
		return ZeroEntry(tst.t, tst.cfg)
	}

	mcr := NewMatcher(tst.t, tst.cfg, checks...)

	// Check if we already have the entry.
	for i, ent := range tst.entries().Get() {
		if i <= tst.matchIdx {
			continue
		}
		if mcr.MatchEntry(ent) {
			tst.matchIdx = i
			tst.mx.Unlock()
			return ent
		}
	}

	found := mcr.Notify()
	tst.matchers = append(tst.matchers, mcr)
	timer := time.NewTimer(to)
	defer timer.Stop()
	tst.mx.Unlock()

	var ent Entry
	select {
	case ent = <-found:
		mcr.NotifyStop()
		if !timer.Stop() {
			<-timer.C
		}

	case <-timer.C:
		mcr.NotifyStop()
	}

	if !ent.IsZero() {
		return ent
	}

	mHeader := "timeout waiting for log entry reached"
	tst.t.Error(notice.New(mHeader).Append("timeout", "%s", timeout))
	tst.t.Error(tst.Entries().summary(1))
	return ZeroEntry(tst.t, tst.cfg)
}

// WaitForAny works like [Tester.WaitFor] but resets the last match before it
// returns. It can be used to match log entries in any order.
func (tst *Tester) WaitForAny(timeout string, checks ...Checker) Entry {
	tst.t.Helper()
	defer tst.ResetLastMatch()
	return tst.WaitFor(timeout, checks...)
}

// Match uses [Matcher] to find the first entry in far logged entries. When
// entry is not found, it will mark the test as failed and return zero value
// [Entry].
func (tst *Tester) Match(mch *Matcher) Entry {
	tst.mx.RLock()
	defer tst.mx.RUnlock()
	tst.t.Helper()

	for _, ent := range tst.entries().Get() {
		if mch.MatchEntry(ent) {
			return ent
		}
	}
	tst.t.Error(notice.New("log entry not found"))
	tst.t.Error(tst.Entries().summary(1))
	return Entry{t: tst.t}
}

// Reset resets the Tester.
func (tst *Tester) Reset() {
	tst.mx.Lock()
	defer tst.mx.Unlock()

	tst.cnt = 0
	tst.buf = tst.buf[:0]
	tst.matchers = tst.matchers[:0]
}
