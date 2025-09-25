// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"os"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_WithBytes(t *testing.T) {
	// --- Given ---
	want := []byte("{}\n{}\n")
	tst := &Tester{}

	// --- When ---
	WithBytes(want)(tst)

	// --- Then ---
	assert.Equal(t, "{}\n{}\n", tst.String())
}

func Test_WithString(t *testing.T) {
	// --- Given ---
	tst := &Tester{}

	// --- When ---
	WithString("{}\n{}\n")(tst)

	// --- Then ---
	assert.Equal(t, "{}\n{}\n", tst.String())
}

func Test_WithConfig(t *testing.T) {
	// --- Given ---
	cfg := DefaultConfig()
	tst := &Tester{}

	// --- When ---
	WithConfig(cfg)(tst)

	// --- Then ---
	assert.Same(t, cfg, tst.cfg)
}

func Test_New(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		tst := New(tspy)

		// --- Then ---
		assert.NotNil(t, tst.cfg)
		assert.NotNil(t, tst.buf)
		assert.Equal(t, 0, tst.cnt)
		assert.Nil(t, tst.matchers)
		assert.Equal(t, -1, tst.matchIdx)
		assert.Same(t, tspy, tst.t)
	})

	t.Run("WithBytes option", func(t *testing.T) {
		// --- Given ---
		lin0 := `{"level":"info", "str":"abc", "message":"msg0"}`
		lin1 := `{"level":"info", "str":"def", "message":"msg1"}`

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		tst := New(tspy, WithBytes([]byte(lin0+"\n"+lin1)))

		// --- Then ---
		assert.Equal(t, 2, tst.Len())
		assert.Equal(t, lin0, tst.Entries().Entry(0).String())
		assert.Equal(t, lin1, tst.Entries().Entry(1).String())
	})

	t.Run("WithString option", func(t *testing.T) {
		// --- Given ---
		lin0 := `{"level":"info", "str":"abc", "message":"msg0"}`
		lin1 := `{"level":"info", "str":"def", "message":"msg1"}`

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		tst := New(tspy, WithString(lin0+"\n"+lin1))

		// --- Then ---
		assert.Equal(t, 2, tst.Len())
		assert.Equal(t, lin0, tst.Entries().Entry(0).String())
		assert.Equal(t, lin1, tst.Entries().Entry(1).String())
	})
}

func Test_Load(t *testing.T) {
	t.Run("load log file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		tst := Load(tspy, "testdata/log.log")

		// --- Then ---
		assert.Equal(t, 2, tst.Len())
		want := must.Value(os.ReadFile("testdata/log.log"))
		assert.Equal(t, string(want), tst.String())
	})

	t.Run("error - file does not exist error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "open testdata/not_existing.log: no such file or directory"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		// --- When ---
		tst := Load(tspy, "testdata/not_existing.log")

		// --- Then ---
		assert.Nil(t, tst)
	})
}

func Test_Tester_Write(t *testing.T) {
	t.Run("write line", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)

		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have, err := tst.Write(lin0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 47, have)

		assert.Equal(t, string(lin0), string(tst.buf))
		assert.Equal(t, 1, tst.cnt)
		assert.Equal(t, -1, tst.matchIdx)
	})

	t.Run("with matchers", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"info", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		mcr0 := NewMatcher(tspy, nil, CheckMsg("msg0"))
		mcr2 := NewMatcher(tspy, nil, CheckMsg("msg2"))

		tst := New(tspy)
		tst.matchers = append(tst.matchers, mcr0, mcr2)

		// --- When --- add first line ---
		have, err := tst.Write(lin0)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 47, have)

		assert.Equal(t, string(lin0), string(tst.buf))
		assert.Equal(t, 1, tst.cnt)
		assert.Equal(t, 0, tst.matchIdx)
		assert.Len(t, 1, tst.matchers)
		assert.Same(t, mcr2, tst.matchers[0])

		// --- When --- add second line ---
		have, err = tst.Write(lin1)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 47, have)

		wantBuf := append(lin0, lin1...) // nolint: gocritic
		assert.Equal(t, string(wantBuf), string(tst.buf))
		assert.Equal(t, 2, tst.cnt)
		assert.Equal(t, 0, tst.matchIdx)
		assert.Len(t, 1, tst.matchers)
		assert.Same(t, mcr2, tst.matchers[0])

		// --- When --- add third line ---
		have, err = tst.Write(lin2)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 47, have)

		wantBuf = append(lin0, lin1...) // nolint: gocritic
		wantBuf = append(wantBuf, lin2...)
		assert.Equal(t, string(wantBuf), string(tst.buf))
		assert.Equal(t, 3, tst.cnt)
		assert.Equal(t, 2, tst.matchIdx)
		assert.Len(t, 0, tst.matchers)
	})

	t.Run("done matcher is not being run", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"info", "str":"def", "message":"msg1"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		mcr := NewMatcher(tspy, nil, CheckMsg("msg1"))
		mcr.discard()

		tst := New(tspy)
		tst.matchers = append(tst.matchers, mcr)

		// --- When ---
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))

		// --- Then ---
		assert.Equal(t, string(lin0)+string(lin1), string(tst.buf))
		assert.Equal(t, 2, tst.cnt)
		assert.Equal(t, -1, tst.matchIdx)
	})
}

func Test_Tester_Len(t *testing.T) {
	t.Run("without writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.Len()

		// --- Then ---
		assert.Equal(t, 0, have)
	})

	t.Run("with one writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write([]byte("test_0")))

		// --- When ---
		have := tst.Len()

		// --- Then ---
		assert.Equal(t, 1, have)
	})

	t.Run("with multiple writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write([]byte("test_0")))
		must.Value(tst.Write([]byte("test_1")))
		must.Value(tst.Write([]byte("test_2")))

		// --- When ---
		have := tst.Len()

		// --- Then ---
		assert.Equal(t, 3, have)
	})
}

func Test_Tester_String(t *testing.T) {
	t.Run("without writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.String()

		// --- Then ---
		assert.Equal(t, "", have)
	})

	t.Run("with writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write([]byte("test_0")))
		must.Value(tst.Write([]byte(" test_1")))

		// --- When ---
		have := tst.String()

		// --- Then ---
		assert.Equal(t, "test_0 test_1", have)
	})
}

func Test_Tester_Bytes(t *testing.T) {
	t.Run("without writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.Bytes()

		// --- Then ---
		assert.Len(t, 0, have)
	})

	t.Run("with writes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write([]byte("test_0")))
		must.Value(tst.Write([]byte(" test_1")))

		// --- When ---
		have := tst.Bytes()

		// --- Then ---
		assert.Equal(t, []byte("test_0 test_1"), have)
	})
}

func Test_Tester_Entries(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.Entries()

		// --- Then ---
		assert.NotNil(t, have)
		assert.Len(t, 0, have.Get())
	})

	t.Run("couple of entries", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"info", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.Close()

		cfg := DefaultConfig()
		tst := New(tspy, WithConfig(cfg))
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- When ---
		have := tst.Entries()

		// --- Then ---
		ent := have.Entry(0)
		assert.Same(t, cfg, ent.cfg)
		assert.Equal(t, string(lin0), ent.String())
		want := map[string]any{
			"level":   "info",
			"str":     "abc",
			"message": "msg0",
		}
		assert.Equal(t, want, ent.m)
		assert.Equal(t, 0, ent.idx)
		assert.Same(t, tspy, ent.t)

		ent = have.Entry(1)
		assert.Same(t, cfg, ent.cfg)
		assert.Equal(t, string(lin1), ent.String())
		want = map[string]any{
			"level":   "info",
			"str":     "def",
			"message": "msg1",
		}
		assert.Equal(t, want, ent.m)
		assert.Equal(t, 1, ent.idx)
		assert.Same(t, tspy, ent.t)

		ent = have.Entry(2)
		assert.Same(t, cfg, ent.cfg)
		assert.Equal(t, string(lin2), ent.String())
		want = map[string]any{
			"level":   "info",
			"str":     "ghi",
			"message": "msg2",
		}
		assert.Equal(t, want, ent.m)
		assert.Equal(t, 2, ent.idx)
		assert.Same(t, tspy, ent.t)

		assert.Len(t, 3, have.Get())
	})

	t.Run("error - decoding", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("invalid character")
		tspy.Close()

		cfg := DefaultConfig()
		tst := New(tspy, WithConfig(cfg))
		must.Value(tst.Write([]byte("{!!!}")))

		// --- When ---
		have := tst.Entries()

		// --- Then ---
		assert.Same(t, cfg, have.cfg)
		assert.Len(t, 0, have.Get())
		assert.Same(t, tspy, have.t)
	})
}

func Test_Tester_Filter(t *testing.T) {
	t.Run("some found", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- When ---
		ets := tst.Filter("info")

		// --- Then ---
		assert.Same(t, tspy, ets.t)
		assert.Len(t, 2, ets.ets)

		ent := ets.ets[0]
		assert.Equal(t, string(lin0), ent.String())
		assert.Equal(t, 0, ent.idx)

		ent = ets.ets[1]
		assert.Equal(t, string(lin2), ent.String())
		assert.Equal(t, 2, ent.idx)
	})

	t.Run("none found", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- When ---
		ets := tst.Filter("error")

		// --- Then ---
		assert.Same(t, tspy, ets.t)
		assert.Len(t, 0, ets.ets)
	})
}

func Test_Tester_FirstEntry(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- When ---
		have := tst.FirstEntry()

		// --- Then ---
		assert.Equal(t, string(lin0), have.String())
		assert.Equal(t, 0, have.idx)
	})

	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.FirstEntry()

		// --- Then ---
		assert.Zero(t, have)
		assert.Same(t, tspy, have.t)
	})
}

func Test_Tester_LastEntry(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"def", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"ghi", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- When ---
		have := tst.LastEntry()

		// --- Then ---
		assert.Equal(t, string(lin2), have.String())
		assert.Equal(t, 2, have.idx)
	})

	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.LastEntry()

		// --- Then ---
		assert.Zero(t, have)
		assert.Same(t, tspy, have.t)
	})
}

func Test_Tester_ResetLastMatch(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	tst := New(tspy)
	tst.matchIdx = 3

	// --- When ---
	tst.ResetLastMatch()

	// --- Then ---
	assert.Equal(t, -1, tst.matchIdx)
}

func Test_Tester_WaitFor(t *testing.T) {
	t.Run("success level error", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := New(tspy)

		started, exited := make(chan struct{}), make(chan struct{})
		var ent Entry
		go func() {
			close(started)
			chk0 := CheckLevel("debug")
			chk1 := CheckStr("str", "abc")
			ent = tst.WaitFor("500ms", chk0, chk1)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Equal(t, 3, tst.cnt)
		assert.Equal(t, 1, tst.matchIdx)

		assert.Equal(t, string(lin1), ent.String())
		assert.Same(t, tspy, ent.t)
		assert.Equal(t, 1, ent.Index())
	})

	t.Run("match first existing", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))

		started, exited := make(chan struct{}), make(chan struct{})
		var ent Entry
		go func() {
			close(started)
			chk0 := CheckMsg("msg0")
			ent = tst.WaitFor("500ms", chk0)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Equal(t, 3, tst.cnt)
		assert.Equal(t, 0, tst.matchIdx)

		assert.Equal(t, string(lin0), ent.String())
		assert.Same(t, tspy, ent.t)
		assert.Equal(t, 0, ent.Index())
	})

	t.Run("error - wait timeout", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		wMsg := "timeout waiting for log entry reached:\n" +
			"  timeout: 500ms\n" +
			"entries logged so far:\n" +
			"   {\"level\":\"info\", \"str\":\"abc\", \"message\":\"msg0\"}\n" +
			"   {\"level\":\"debug\", \"str\":\"abc\", \"message\":\"msg1\"}\n" +
			"   {\"level\":\"info\", \"str\":\"abc\", \"message\":\"msg2\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		tst := New(tspy)

		started, exited := make(chan struct{}), make(chan struct{})
		var ent Entry
		go func() {
			close(started)
			chk0 := CheckLevel("debug")
			chk1 := CheckStr("str", "xyz")
			ent = tst.WaitFor("500ms", chk0, chk1)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Zero(t, ent)
		assert.Same(t, tspy, ent.t)
	})

	t.Run("already existing", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		started, exited := make(chan struct{}), make(chan struct{})
		var ent Entry
		go func() {
			close(started)
			chk0 := CheckLevel("debug")
			chk1 := CheckStr("str", "abc")
			ent = tst.WaitFor("500ms", chk0, chk1)
			close(exited)
		}()
		<-started

		// --- Then ---
		<-exited
		assert.Equal(t, string(lin1), ent.String())
		assert.Same(t, tspy, ent.t)
		assert.Equal(t, 1, ent.Index())
	})

	t.Run("error - order matters", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"def", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.ExpectError()
		wMsg := "timeout waiting for log entry reached:\n" +
			"  timeout: 50ms\n" +
			"entries logged so far:\n" +
			"   {\"level\":\"info\", \"str\":\"abc\", \"message\":\"msg0\"}\n" +
			"   {\"level\":\"debug\", \"str\":\"abc\", \"message\":\"msg1\"}\n" +
			"   {\"level\":\"info\", \"str\":\"def\", \"message\":\"msg2\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		tst := New(tspy)

		started, exited := make(chan struct{}), make(chan struct{})
		var ent0, ent1 Entry
		go func() {
			close(started)
			// Start waiting for the last log entry.
			chk00 := CheckLevel("info")
			chk01 := CheckStr("str", "def")
			ent0 = tst.WaitFor("50ms", chk00, chk01)

			// Start waiting for first log entry which will fail since we
			// always wait for the log entry after the last matched entry.
			chk10 := CheckLevel("info")
			chk11 := CheckStr("str", "abc")
			ent1 = tst.WaitFor("50ms", chk10, chk11)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Equal(t, string(lin2), ent0.String())
		assert.Same(t, tspy, ent0.t)
		assert.Zero(t, ent1)
		assert.Same(t, tspy, ent1.t)
	})

	t.Run("error - invalid time duration", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("time: invalid duration \"abc\"")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := tst.WaitFor("abc")

		// --- Then ---
		assert.Zero(t, have)
		assert.Same(t, tspy, have.t)
	})
}

func Test_Tester_WaitForAny(t *testing.T) {
	t.Run("matches", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))

		started, exited := make(chan struct{}), make(chan struct{})
		var ent Entry
		go func() {
			close(started)
			chk0 := CheckMsg("msg2")
			ent = tst.WaitForAny("500ms", chk0)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Equal(t, 3, tst.cnt)
		assert.Equal(t, -1, tst.matchIdx)

		assert.Equal(t, string(lin2), ent.String())
		assert.Equal(t, 2, ent.Index())
	})

	t.Run("order does not matter", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"def", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(2)
		tspy.Close()

		tst := New(tspy)

		started, exited := make(chan struct{}), make(chan struct{})
		var ent0, ent1 Entry
		go func() {
			close(started)
			// Start waiting for the last log entry.
			chk00 := CheckLevel("info")
			chk01 := CheckStr("str", "def")
			ent0 = tst.WaitForAny("50ms", chk00, chk01)

			// Start waiting for first log entry.
			chk10 := CheckLevel("info")
			chk11 := CheckStr("str", "abc")
			ent1 = tst.WaitForAny("50ms", chk10, chk11)
			close(exited)
		}()
		<-started

		// --- When ---
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		// --- Then ---
		<-exited
		assert.Equal(t, string(lin2), ent0.String())
		assert.Equal(t, string(lin0), ent1.String())
		assert.Equal(t, -1, tst.matchIdx)
	})
}

func Test_Tester_Match(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		mcr := NewMatcher(tspy, nil, CheckMsg("msg1"))

		// --- When ---
		have := tst.Match(mcr)

		// --- Then ---
		assert.Equal(t, string(lin1), have.String())
		assert.Same(t, tspy, have.t)
	})

	t.Run("error -  not found", func(t *testing.T) {
		// --- Given ---
		lin0 := []byte(`{"level":"info", "str":"abc", "message":"msg0"}`)
		lin1 := []byte(`{"level":"debug", "str":"abc", "message":"msg1"}`)
		lin2 := []byte(`{"level":"info", "str":"abc", "message":"msg2"}`)

		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		wMsg := "log entry not found\n" +
			"entries logged so far:\n" +
			"   {\"level\":\"info\", \"str\":\"abc\", \"message\":\"msg0\"}\n" +
			"   {\"level\":\"debug\", \"str\":\"abc\", \"message\":\"msg1\"}\n" +
			"   {\"level\":\"info\", \"str\":\"abc\", \"message\":\"msg2\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		tst := New(tspy)
		must.Value(tst.Write(lin0))
		must.Value(tst.Write(lin1))
		must.Value(tst.Write(lin2))

		mcr := NewMatcher(tspy, nil, CheckMsg("msg3"))

		// --- When ---
		have := tst.Match(mcr)

		// --- Then ---
		assert.Zero(t, have)
		assert.Same(t, tspy, have.t)
	})
}

func Test_Tester_Reset(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	mcr := NewMatcher(t, nil, CheckLevel("info"))

	tst := New(tspy)
	tst.matchers = append(tst.matchers, mcr)
	MustWriteLine(tst, `{"level": "info", "A": 1}`)
	MustWriteLine(tst, `{"level": "error", "B": 1}`)

	// --- When ---
	tst.Reset()

	// --- Then ---
	assert.Equal(t, 0, tst.Len())
	assert.Equal(t, "", tst.String())
	assert.Len(t, 0, tst.matchers)
}
