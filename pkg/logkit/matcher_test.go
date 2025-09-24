// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_NewMatcher(t *testing.T) {
	t.Run("no checks", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()

		// --- When ---
		mcr := NewMatcher(tspy, cfg)

		// --- Then ---
		assert.Nil(t, mcr.checks)
		assert.NotNil(t, mcr.done)
		assert.Zero(t, mcr.ent)
		assert.Same(t, tspy, mcr.t)
	})

	t.Run("with fields to match", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		fn := func(Entry) error { return nil }

		// --- When ---
		mcr := NewMatcher(tspy, cfg, fn)

		// --- Then ---
		assert.Len(t, 1, mcr.checks)
		assert.Same(t, fn, mcr.checks[0])
	})

	t.Run("closes the done channel at the test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()

		// --- When ---
		mcr := NewMatcher(tspy, cfg)

		// --- Then ---
		assert.NotNil(t, mcr.done)
		tspy.Finish()
		assert.Nil(t, mcr.done)
	})

	t.Run("no panic when the done channel is nil", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg)

		// --- When ---
		mcr.done = nil

		// --- Then ---
		assert.NoPanic(t, func() { tspy.Finish() })
	})
}

func Test_Matcher_Checks(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	cfg := DefaultConfig()
	checks := []func(Entry) error{CheckMsg("msg0"), CheckStr("A", "a")}
	mcr := NewMatcher(tspy, cfg, checks...)

	// --- When ---
	have := mcr.Checks()

	// --- Then ---
	assert.Len(t, 2, have)
	assert.NotSame(t, checks, have)
	assert.Same(t, checks[0], have[0])
	assert.Same(t, checks[1], have[1])
}

func Test_Matcher_Done(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	cfg := DefaultConfig()
	mcr := NewMatcher(tspy, cfg)

	// --- When ---
	done := mcr.Done()

	// --- Then ---
	assert.NotNil(t, done)
	go tspy.Finish()
	assert.ChannelWillClose(t, "1s", done)
}

func Test_Matcher_IsDone(t *testing.T) {
	t.Run("no lines matched yet", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckMsg("msg0"))

		// --- When ---
		have := mcr.IsDone()

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("at least one line matched", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckMsg("msg0"))
		assert.True(t, mcr.match(0, lin0))

		// --- When ---
		have := mcr.IsDone()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true when discarded", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg)
		mcr.discard()

		// --- When ---
		have := mcr.IsDone()

		// --- Then ---
		tspy.Finish()
		assert.True(t, have)
	})
}

func Test_Matcher_Entry(t *testing.T) {
	t.Run("no lines matched yet", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg)

		// --- When ---
		have := mcr.Entry()

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("line matched", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		mcr := &Matcher{
			ent: Entry{
				raw: "abc",
				m:   map[string]any{"str": "def"},
				idx: 1,
				t:   tspy,
			},
		}

		// --- When ---
		have := mcr.Entry()

		// --- Then ---
		assert.Equal(t, "abc", have.String())
		assert.Equal(t, map[string]any{"str": "def"}, have.m)
		assert.Equal(t, 1, have.idx)
		assert.Same(t, tspy, have.t)
	})

	t.Run("the returned entry map is a clone", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		mcr := &Matcher{
			ent: Entry{
				raw: "abc",
				m:   map[string]any{"str": "def"},
			},
		}

		// --- When ---
		have := mcr.Entry()
		have.m["changed"] = true

		// --- Then ---
		assert.HasKey(t, "changed", have.m)
		assert.HasNoKey(t, "changed", mcr.ent.m)
	})
}

func Test_Matcher_match(t *testing.T) {
	t.Run("without checks", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg)

		// --- When ---
		have := mcr.match(1, lin0)

		// --- Then ---
		assert.True(t, have)
		assert.True(t, mcr.IsDone())
		assert.Zero(t, mcr.Entry())
		assert.ChannelWillClose(t, "1s", mcr.Done())
	})

	t.Run("match a check", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin0m := JSON2Map(t, lin0)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckStr("str", "abc"))

		// --- When ---
		have := mcr.match(1, lin0)

		// --- Then ---
		assert.True(t, have)
		assert.True(t, mcr.IsDone())

		want := Entry{
			cfg: DefaultConfig(),
			raw: string(lin0),
			m:   lin0m,
			idx: 1,
			t:   tspy,
		}
		assert.Equal(t, want, mcr.Entry())
	})

	t.Run("line must match all checks - success", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin0m := JSON2Map(t, lin0)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckLevel("info"), CheckMsg("msg0"))

		// --- When ---
		have := mcr.match(1, lin0)

		// --- Then ---
		assert.True(t, have)
		assert.True(t, mcr.IsDone())
		want := Entry{
			cfg: DefaultConfig(),
			raw: string(lin0),
			m:   lin0m,
			idx: 1,
			t:   tspy,
		}
		assert.Equal(t, want, mcr.Entry())
	})

	t.Run("line must match all checks - failure", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckLevel("info"), CheckMsg("msg1"))

		// --- When ---
		have := mcr.match(1, lin0)

		// --- Then ---
		assert.False(t, have)
		assert.False(t, mcr.IsDone())
		assert.Zero(t, mcr.Entry())
	})

	t.Run("closes the done channel on the first check", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg)

		// --- When ---
		have := mcr.match(1, lin0)

		// --- Then ---
		assert.True(t, have)
		assert.ChannelWillClose(t, "1s", mcr.Done())
	})

	t.Run("the first check is saved", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin0m := JSON2Map(t, lin0)
		lin1 := []byte(`{"level":"info", "str":"def", "message":"msg1"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckLevel("info"))

		// --- When ---
		have0 := mcr.match(0, lin0)
		have1 := mcr.match(1, lin1)

		// --- Then ---
		assert.True(t, have0)
		assert.True(t, have1)

		want := Entry{
			cfg: DefaultConfig(),
			raw: string(lin0),
			m:   lin0m,
			idx: 0,
			t:   tspy,
		}
		assert.Equal(t, want, mcr.Entry())
	})

	t.Run("check can modify the entry map", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin0m := JSON2Map(t, lin0)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(ent Entry) error { ent.m["edit"] = true; return nil }
		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, fn)

		// --- When ---
		have := mcr.match(0, lin0)

		// --- Then ---
		assert.True(t, have)
		assert.HasNoKey(t, "edit", lin0m)
		assert.HasNoKey(t, "edit", mcr.Entry().m)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		tspy.ExpectLogContain("matcher line 1: invalid character")
		tspy.Close()

		cfg := DefaultConfig()
		mcr := NewMatcher(tspy, cfg, CheckMsg("msg0"))

		// --- When ---
		have := mcr.match(1, []byte("{!!!}"))

		// --- Then ---
		assert.False(t, have)
		assert.False(t, mcr.IsDone())
	})
}

func Test_Matcher_discard(t *testing.T) {
	t.Run("sets done to nil and entry to zero value", func(t *testing.T) {
		// --- Given ---
		mcr := Matcher{
			done: make(chan struct{}),
			ent:  Entry{raw: "data"},
		}

		// --- When ---
		mcr.discard()

		// --- Then ---
		assert.Nil(t, mcr.done)
		assert.Zero(t, mcr.ent)
	})

	t.Run("double call does not panic", func(t *testing.T) {
		// --- Given ---
		mcr := Matcher{done: nil}

		// --- When ---
		mcr.discard()

		// --- Then ---
		assert.NoPanic(t, func() { mcr.discard() })
	})
}
