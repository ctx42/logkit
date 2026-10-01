// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

func init() { check.RegisterTypeChecker(Entry{}, EntryCheck) }

// EntryCheck is a custom checker, matching [check.Checker] signature,
// comparing two instances of [Entry].
func EntryCheck(want, have any, opts ...any) error {
	ops := check.DefaultOptions(opts...)
	stOpt := check.WithOptions(ops)
	if _, err := check.SameType(Entry{}, have, stOpt); err != nil {
		return err
	}
	w, h := want.(Entry), have.(Entry)

	fName := check.FieldName(ops, "Entry")
	ers := []error{
		check.Equal(w.cfg, h.cfg, fName("cfg")),
		check.Equal(w.m, h.m, fName("m")),
		check.Equal(w.raw, h.raw, fName("raw")),
		check.Equal(w.idx, h.idx, fName("idx")),
		check.Fields(5, w, fName("{field count}")),
	}
	return notice.Join(ers...)
}

// MustEntries builds a [Entries] from a list of raw log messages using the
// [DefaultConfig]. Panics if a log message is not valid JSON.
func MustEntries(t tester.T, raws ...string) Entries {
	t.Helper()
	cfg := DefaultConfig()
	ets := Entries{cfg: cfg, ets: make([]Entry, 0, 10), t: t}
	for i, raw := range raws {
		ent := Entry{cfg: cfg, raw: raw, idx: i, t: t}
		if err := json.Unmarshal([]byte(raw), &ent.m); err != nil {
			panic(err.Error())
		}
		ets.ets = append(ets.ets, ent)
	}
	return ets
}

// JSON2Map unmarshalls JSON to `map[string]any` and returns it on success. On
// failure, marks the test as failed and returns nil.
func JSON2Map(t tester.T, data string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Error(err)
		return nil
	}
	return m
}

// MustWriteLine writes the string and new line to the writer. Panics on error.
func MustWriteLine(w io.Writer, lines ...string) {
	for _, line := range lines {
		n, err := w.Write([]byte(line + "\n"))
		if err != nil {
			panic(err)
		}
		if n != len(line)+1 {
			format := "expected to write %d bytes, wrote %d"
			panic(fmt.Sprintf(format, len(line)+1, n))
		}
	}
}

// shortWriter is an [io.Writer] which reports writing a single byte.
type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 1, nil }

// ReceiveValue waits in a separate goroutine for up to the given timeout for a
// value to be received from the channel. Returns a channel on which the
// received value will be sent before it is closed. If the timeout is reached,
// the test is marked as failed, the error message is logged, and the returned
// channel is closed without a value.
func ReceiveValue[T any](t tester.T, timeout string, from <-chan T) <-chan T {
	t.Helper()

	to, err := time.ParseDuration(timeout)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	timer := time.NewTimer(to)

	received := make(chan T, 1)
	go func() {
		defer close(received)
		select {
		case v := <-from:
			timer.Stop()
			received <- v

		case <-timer.C:
			mHeader := "timeout receiving a value from the channel"
			t.Error(notice.New(mHeader).Append("timeout", "%s", timeout))
		}
	}()
	return received
}

// Contend calls fn n times while another goroutine repeatedly locks mx for
// writing. Returns true when all calls finish within the timeout. Otherwise,
// marks the test as failed and returns false.
func Contend(
	t tester.T,
	mx *sync.RWMutex,
	timeout string,
	n int,
	fn func(),
) bool {

	t.Helper()

	to, err := time.ParseDuration(timeout)
	if err != nil {
		t.Error(err)
		return false
	}

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-stop:
				return

			default:
				// Hold the write lock long enough to queue readers behind it.
				mx.Lock()
				runtime.Gosched()
				mx.Unlock()
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		for range n {
			fn()
		}
		close(done)
	}()

	select {
	case <-done:
		return true

	case <-time.After(to):
		mHeader := "timeout waiting for calls under lock contention"
		t.Error(notice.New(mHeader).Append("timeout", "%s", timeout))
		return false
	}
}

// AwaitMatchers waits up to one second until tst has n queued matchers.
// Returns true when it does. Otherwise, marks the test as failed and returns
// false.
func AwaitMatchers(t tester.T, tst *Tester, n int) bool {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		tst.mx.RLock()
		have := len(tst.matchers)
		tst.mx.RUnlock()
		if have == n {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	t.Errorf("expected %d queued matchers", n)
	return false
}

// /////////////////////////////////////////////////////////////////////////////

func Test_MustEntries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		const lin0 = `{"level": "error", "number": 0.0,   "message": "msg0"}`
		const lin1 = `{"level": "info",  "bool_t": true,  "message": "msg1"}`

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := MustEntries(tspy, lin0, lin1)

		// --- Then ---
		assert.Len(t, 2, have.Get())
		assert.Equal(t, lin0, have.Entry(0).String())
		assert.Equal(t, lin1, have.Entry(1).String())
	})

	t.Run("panics", func(t *testing.T) {
		// --- Given ---
		const lin0 = `{!!!}`

		tspy := tester.New(t, 1)
		tspy.Close()

		// --- When ---
		msg := assert.PanicMsg(t, func() { MustEntries(tspy, lin0) })

		// --- Then ---
		assert.NotNil(t, msg)
		assert.Contain(t, "invalid character", *msg)
	})
}

func Test_JSON2Map(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := JSON2Map(tspy, `{"f_str": "abc"}`)

		// --- Then ---
		assert.Equal(t, map[string]any{"f_str": "abc"}, have)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("invalid character")
		tspy.Close()

		// --- When ---
		have := JSON2Map(tspy, "{!!!}")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_MustWriteLine(t *testing.T) {
	t.Run("writes lines", func(t *testing.T) {
		// --- Given ---
		buf := &bytes.Buffer{}

		// --- When ---
		MustWriteLine(buf, "abc", "def")

		// --- Then ---
		assert.Equal(t, "abc\ndef\n", buf.String())
	})

	t.Run("error - short write", func(t *testing.T) {
		// --- Given ---
		w := shortWriter{}
		fn := func() { MustWriteLine(w, "abc") }

		// --- When ---
		msg := assert.PanicMsg(t, fn)

		// --- Then ---
		assert.NotNil(t, msg)
		assert.Equal(t, "expected to write 4 bytes, wrote 1", *msg)
	})
}

func Test_ReceiveValue(t *testing.T) {
	t.Run("value received", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		from := make(chan int)
		go func() { from <- 42 }()

		// --- When ---
		have := ReceiveValue(tspy, "1s", from)

		// --- Then ---
		assert.Equal(t, 42, <-have)
	})

	t.Run("error - timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"timeout receiving a value from the channel:\n" +
			"  timeout: 10ms"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		from := make(chan int)

		// --- When ---
		have := ReceiveValue(tspy, "10ms", from)

		// --- Then ---
		assert.ChannelWillClose(t, "1s", have)
	})
}

func Test_Contend(t *testing.T) {
	t.Run("all calls finish", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		var mx sync.RWMutex
		var cnt int
		fn := func() {
			mx.RLock()
			cnt++
			mx.RUnlock()
		}

		// --- When ---
		have := Contend(tspy, &mx, "1s", 10, fn)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, 10, cnt)
	})

	t.Run("error - timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"timeout waiting for calls under lock contention:\n" +
			"  timeout: 10ms"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		var mx sync.RWMutex
		fn := func() { time.Sleep(100 * time.Millisecond) }

		// --- When ---
		have := Contend(tspy, &mx, "10ms", 1, fn)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - invalid timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("time: invalid duration \"abc\"")
		tspy.Close()

		var mx sync.RWMutex

		// --- When ---
		have := Contend(tspy, &mx, "abc", 1, func() {})

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_AwaitMatchers(t *testing.T) {
	t.Run("queued", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		tst := New(tspy)
		tst.matchers = append(tst.matchers, NewMatcher(tspy, nil))

		// --- When ---
		have := AwaitMatchers(tspy, tst, 1)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - not queued", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected 1 queued matchers")
		tspy.Close()

		tst := New(tspy)

		// --- When ---
		have := AwaitMatchers(tspy, tst, 1)

		// --- Then ---
		assert.False(t, have)
	})
}
