// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

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
	if err := check.SameType(Entry{}, have, stOpt); err != nil {
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
func JSON2Map(t tester.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
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
			panic(fmt.Sprintf(format, len(format), n))
		}
	}
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

		tspy := tester.New(t, 0)
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
		have := JSON2Map(tspy, []byte(`{"f_str": "abc"}`))

		// --- Then ---
		assert.Equal(t, map[string]any{"f_str": "abc"}, have)
	})

	t.Run("fail", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("invalid character")
		tspy.Close()

		// --- When ---
		have := JSON2Map(tspy, []byte("{!!!}"))

		// --- Then ---
		assert.Nil(t, have)
	})
}
