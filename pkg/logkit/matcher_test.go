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
		tspy.Close()

		cfg := DefaultConfig()

		// --- When ---
		mcr := NewMatcher(tspy, cfg)

		// --- Then ---
		assert.Same(t, cfg, mcr.cfg)
		assert.Nil(t, mcr.checks)
		assert.Equal(t, 0, mcr.cnt)
		assert.Same(t, tspy, mcr.t)
	})

	t.Run("nil config means default", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		mcr := NewMatcher(tspy, nil)

		// --- Then ---
		assert.Equal(t, DefaultConfig(), mcr.cfg)
		assert.Nil(t, mcr.checks)
		assert.Equal(t, 0, mcr.cnt)
		assert.Same(t, tspy, mcr.t)
	})

	t.Run("with fields to match", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		fn := func(Entry) error { return nil }

		// --- When ---
		mcr := NewMatcher(tspy, nil, fn)

		// --- Then ---
		assert.Len(t, 1, mcr.checks)
		assert.Equal(t, 0, mcr.cnt)
		assert.Same(t, fn, mcr.checks[0])
	})
}

func Test_Matcher_Checks(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	chk0 := CheckMsg("msg0")
	chk1 := CheckStr("A", "a")
	mcr := NewMatcher(tspy, nil, chk0, chk1)

	// --- When ---
	have := mcr.Checks()

	// --- Then ---
	assert.Len(t, 2, have)
	assert.NotSame(t, mcr.checks, have)
	assert.Same(t, chk0, have[0])
	assert.Same(t, chk1, have[1])
}

func Test_Matcher_Matched(t *testing.T) {
	// --- Given ---
	mcr := &Matcher{cnt: 42}

	// --- When ---
	have := mcr.Matched()

	// --- Then ---
	assert.Equal(t, 42, have)
}

func Test_Matcher_Notify(t *testing.T) {
	t.Run("returned chan is closed at the test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		mcr := NewMatcher(tspy, nil)

		// --- When ---
		have := mcr.Notify()

		// --- Then ---
		tspy.Finish()
		assert.ChannelWillClose(t, "1s", have)
	})

	t.Run("returns the same chan on multiple calls", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		mcr := NewMatcher(tspy, nil)

		// --- When ---
		have0 := mcr.Notify()
		have1 := mcr.Notify()

		// --- Then ---
		assert.Same(t, have0, have1)
	})
}

func Test_Matcher_NotifyStop(t *testing.T) {
	t.Run("existing notification", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		mcr := NewMatcher(tspy, nil)
		notify := mcr.Notify()

		// --- When ---
		mcr.NotifyStop()

		// --- Then ---
		_, open := <-notify
		assert.False(t, open)
	})

	t.Run("call without a previous call to Notify", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		mcr := NewMatcher(tspy, nil)

		// --- When ---
		mcr.NotifyStop()
	})
}

func Test_Matcher_MatchEntry(t *testing.T) {
	t.Run("without checks", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil)

		// --- When ---
		have := mcr.MatchEntry(ent)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("match a check", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckStr("str", "abc"))

		// --- When ---
		have := mcr.MatchEntry(ent)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("match multiple times", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckStr("str", "abc"))

		// --- When ---
		mcr.MatchEntry(ent)
		mcr.MatchEntry(ent)

		// --- Then ---
		assert.Equal(t, 2, mcr.cnt)
	})

	t.Run("success - line must match all checks", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckLevel("info"), CheckMsg("msg0"))

		// --- When ---
		have := mcr.MatchEntry(ent)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("failure - line must match all checks", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckLevel("info"), CheckMsg("msg1"))

		// --- When ---
		have := mcr.MatchEntry(ent)

		// --- Then ---
		assert.False(t, have)
		assert.Equal(t, 0, mcr.cnt)
	})

	t.Run("sends entry on the notification channel", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckStr("str", "abc"))
		received := ReceiveValue(t, "1s", mcr.Notify())

		// --- When ---
		have := mcr.MatchEntry(ent)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, ent, <-received)
	})
}

func Test_Matcher_MatchLine(t *testing.T) {
	t.Run("without checks", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		mcr := NewMatcher(tspy, nil)

		// --- When ---
		have := mcr.MatchLine(1, []byte(lin))

		// --- Then ---
		want := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}
		assert.Equal(t, want, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("match a check", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		mcr := NewMatcher(tspy, nil, CheckStr("str", "abc"))

		// --- When ---
		have := mcr.MatchLine(1, []byte(lin))

		// --- Then ---
		want := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}
		assert.Equal(t, want, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("line must match all checks - success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		mcr := NewMatcher(tspy, nil, CheckLevel("info"), CheckMsg("msg0"))

		// --- When ---
		have := mcr.MatchLine(1, []byte(lin))

		// --- Then ---
		want := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}
		assert.Equal(t, want, have)
		assert.Equal(t, 1, mcr.cnt)
	})

	t.Run("line must match all checks - failure", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		mcr := NewMatcher(tspy, nil, CheckLevel("info"), CheckMsg("msg1"))

		// --- When ---
		have := mcr.MatchLine(1, []byte(lin))

		// --- Then ---
		assert.Zero(t, have)
		assert.Equal(t, 0, mcr.cnt)
	})

	t.Run("sends entry on the notification channel", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		lin := `{"level":"info", "str":"abc", "message":"msg0"}`
		ent := Entry{
			cfg: DefaultConfig(),
			raw: lin,
			m:   JSON2Map(t, lin),
			idx: 1,
			t:   tspy,
		}

		mcr := NewMatcher(tspy, nil, CheckStr("str", "abc"))
		received := ReceiveValue(t, "1s", mcr.Notify())

		// --- When ---
		have := mcr.MatchLine(1, []byte(lin))

		// --- Then ---
		assert.Equal(t, ent, have)
		assert.Equal(t, ent, <-received)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("matcher line 1: invalid character")
		tspy.Close()

		mcr := NewMatcher(tspy, nil, CheckMsg("msg0"))

		// --- When ---
		have := mcr.MatchLine(1, []byte("{!!!}"))

		// --- Then ---
		assert.Zero(t, have)
	})
}
