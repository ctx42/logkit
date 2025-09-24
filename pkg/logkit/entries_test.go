// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"errors"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_Entries_Get(t *testing.T) {
	t.Run("with entries", func(t *testing.T) {
		// --- Given ---
		const lin0 = `{"level": "error", "number": 0.0,   "message": "msg0"}`
		const lin1 = `{"level": "info",  "bool_t": true,  "message": "msg1"}`
		const lin2 = `{"level": "info",  "bool_f": false, "message": "msg2"}`

		tspy := tester.New(t, 0)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.Get()

		// --- Then ---
		assert.Len(t, 3, have)
		assert.Equal(t, lin0, have[0].String())
		assert.Equal(t, lin1, have[1].String())
		assert.Equal(t, lin2, have[2].String())
	})

	t.Run("without entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.Get()

		// --- Then ---
		assert.Empty(t, have)
		assert.NotNil(t, have)
	})
}

func Test_Entries_MetaAll(t *testing.T) {
	// --- Given ---
	tst := New(t)
	MustWriteLine(tst, `{"level": "trace", "bools": [true, false, true]}`)
	MustWriteLine(tst, `{"level": "debug", "bool": true}`)
	MustWriteLine(tst, `{"level": "info",  "time": "2000-01-02T03:04:05Z"}`)
	MustWriteLine(tst, `{"level": "warn", "bytes": "\u0001\u0002\u0003"}`)
	MustWriteLine(tst, `{"level": "fatal", "str": "abc"}`)
	MustWriteLine(tst, `{"level": "panic", "int": 123, "int64": 456}`)

	// --- When ---
	have := tst.Entries().MetaAll()

	// --- Then ---
	want := []map[string]any{
		{"level": "trace", "bools": []any{true, false, true}},
		{"level": "debug", "bool": true},
		{"level": "info", "time": "2000-01-02T03:04:05Z"},
		{"level": "warn", "bytes": "\x01\x02\x03"},
		{"level": "fatal", "str": "abc"},
		{"level": "panic", "int": float64(123), "int64": float64(456)},
	}
	assert.Equal(t, want, have)
}

func Test_Entries_Entry(t *testing.T) {
	const lin0 = `{"level": "error", "number": 0.0,   "message": "msg0"}`
	const lin1 = `{"level": "info",  "bool_t": true,  "message": "msg1"}`
	const lin2 = `{"level": "info",  "bool_f": false, "message": "msg2"}`

	t.Run("found entry at index", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.Entry(1)

		// --- Then ---
		assert.Equal(t, 1, have.idx)
		assert.Equal(t, lin1, have.raw)
	})

	t.Run("error - not existing index", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected log entry to exist:\n" +
			"  index: 3"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.Entry(3)

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("error - without entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected log entry to exist:\n" +
			"  index: 0"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.Entry(0)

		// --- Then ---
		assert.Zero(t, have)
	})
}

func Test_AssertRaw(t *testing.T) {
	t.Run("entries match", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertRaw(lin0, lin1, lin2)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("entries do not match", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		wMsg := "" +
			"[log entry] expected JSON strings to be equal:\n" +
			"  index: 2\n" +
			"   want: {\"level\":\"info\",\"str\":\"msg3\"}\n" +
			"   have: {\"level\":\"info\",\"str\":\"msg2\"}"
		tspy.ExpectLogEqual(wMsg)
		tspy.ExpectError()
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`
		const lin3 = `{"level": "info", "str": "msg3"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertRaw(lin0, lin1, lin3)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("have has more lines than want", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		wMsg := "" +
			"[log entry] expected N log entries:\n" +
			"       want: 1\n" +
			"       have: 2\n" +
			"  have logs:\n" +
			"             {\"level\": \"info\", \"str\": \"msg0\"}\n" +
			"             {\"level\": \"info\", \"str\": \"msg1\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.ExpectError()
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertRaw(lin0)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("want has more lines than have", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		wMsg := "" +
			"[log entry] expected log entry to exist:\n" +
			"  index: 1"
		tspy.ExpectLogEqual(wMsg)
		tspy.ExpectError()
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`

		ets := MustEntries(tspy, lin0)

		// --- When ---
		have := ets.AssertRaw(lin0, lin1)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertLen(t *testing.T) {
	const lin0 = `{"level": "error", "number": 0.0,   "message": "msg0"}`
	const lin1 = `{"level": "info",  "bool_t": true,  "message": "msg1"}`
	const lin2 = `{"level": "info",  "bool_f": false, "message": "msg2"}`

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertLen(3)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("zero length", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.AssertLen(0)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - wrong number of entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected N log entries:\n" +
			"  want: 10\n" +
			"  have: 3"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertLen(10)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertMsg(t *testing.T) {
	lin0 := `{"level": "info",  "message": "msg0"}`
	lin1 := `{"level": "debug", "message": "msg1"}`
	lin2 := `{"level": "debug", "message": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertMsg("msg1")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertMsg("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoMsg(t *testing.T) {
	lin0 := `{"level": "info",  "message": "msg0"}`
	lin1 := `{"level": "debug", "message": "msg1"}`
	lin2 := `{"level": "debug", "message": "msg2"}`

	t.Run("field name with value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoMsg("xyz")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoMsg("msg1")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertMsgContain(t *testing.T) {
	lin0 := `{"level": "info",  "message": "msg0 abc"}`
	lin1 := `{"level": "debug", "message": "msg1 abc"}`
	lin2 := `{"level": "debug", "message": "msg2 abc"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertMsgContain("msg1")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertMsgContain("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoMsgContain(t *testing.T) {
	lin0 := `{"level": "info",  "message": "msg0 abc"}`
	lin1 := `{"level": "debug", "message": "msg1 abc"}`
	lin2 := `{"level": "debug", "message": "msg2 abc"}`

	t.Run("field name with value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoMsgContain("xyz")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoMsgContain("msg1")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertError(t *testing.T) {
	lin0 := `{"level": "info",  "error": "msg0"}`
	lin1 := `{"level": "debug", "error": "msg1"}`
	lin2 := `{"level": "debug", "error": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertError("msg1")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertError("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertErrorContain(t *testing.T) {
	lin0 := `{"level": "info",  "error": "msg0 abc"}`
	lin1 := `{"level": "debug", "error": "msg1 abc"}`
	lin2 := `{"level": "debug", "error": "msg2 abc"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertErrorContain("msg1")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertErrorContain("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoError(t *testing.T) {
	lin0 := `{"level": "info",  "error": "msg0"}`
	lin1 := `{"level": "debug", "error": "msg1"}`
	lin2 := `{"level": "debug", "error": "msg2"}`

	t.Run("field name with value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoError("xyz")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoError("msg1")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertErr(t *testing.T) {
	lin0 := `{"level": "info",  "error": "msg0"}`
	lin1 := `{"level": "debug", "error": "msg1"}`
	lin2 := `{"level": "debug", "error": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertErr(errors.New("msg1"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertErr(errors.New("xyz"))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoErr(t *testing.T) {
	lin0 := `{"level": "info",  "error": "msg0"}`
	lin1 := `{"level": "debug", "error": "msg1"}`
	lin2 := `{"level": "debug", "error": "msg2"}`

	t.Run("field name with value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoErr(errors.New("xyz"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoErr(errors.New("msg1"))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertContain(t *testing.T) {
	const lin0 = `{"level": "debug", "str": "abc def ghi", "message": "msg0"}`
	const lin1 = `{"level": "debug", "str": "jkl mno pqr", "message": "msg1"}`
	const lin2 = `{"level": "debug", "str": "stu vwx yz",  "message": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertContain("str", "abc")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertContain("str", "xxx")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertStr(t *testing.T) {
	const lin0 = `{"level": "info",  "bool_f": false, "message": "msg0"}`
	const lin1 = `{"level": "debug", "number": 3.0,   "message": "msg1"}`
	const lin2 = `{"level": "debug", "str":    "abc", "message": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertStr("str", "abc")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertStr("str", "xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoStr(t *testing.T) {
	const lin0 = `{"level": "info",  "bool_f": false, "message": "msg0"}`
	const lin1 = `{"level": "debug", "number": 3.0,   "message": "msg1"}`
	const lin2 = `{"level": "debug", "str":    "abc", "message": "msg2"}`

	t.Run("field name exists with different value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoStr("str", "xyz")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("field name does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoStr("missing", "xyz")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoStr("str", "abc")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNumber(t *testing.T) {
	const lin0 = `{"level": "info",  "bool_f": false, "message": "msg0"}`
	const lin1 = `{"level": "debug", "number": 3.0,   "message": "msg1"}`
	const lin2 = `{"level": "debug", "number": 4.0,   "message": "msg2"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNumber("number", 4)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNumber("number", 5)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoNumber(t *testing.T) {
	const lin0 = `{"level": "info",  "bool_f": false, "message": "msg0"}`
	const lin1 = `{"level": "debug", "number": 3.0,   "message": "msg1"}`
	const lin2 = `{"level": "debug", "number": 4.0,   "message": "msg2"}`

	t.Run("field name exists with different value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoNumber("number", 5)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.AssertNoNumber("number", 4)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertBool(t *testing.T) {
	const lin0 = `{"level": "error", "number": 0.0,   "message": "msg0"}`
	const lin1 = `{"level": "info",  "bool_t": true,  "message": "msg1"}`
	const lin2 = `{"level": "info",  "bool_f": false, "message": "msg2"}`
	const lin3 = `{"level": "debug", "number": 3.0,   "message": "msg3"}`

	t.Run("field with the value of true found", func(t *testing.T) {
		// --- Given ---

		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2, lin3)

		// --- When ---
		have := ets.AssertBool("bool_t", true)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("field with the value of false found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2, lin3)

		// --- When ---
		have := ets.AssertBool("bool_f", false)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name is not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2, lin3)

		// --- When ---
		have := ets.AssertBool("missing", true)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertTime(t *testing.T) {
	const lin0 = `{"level": "info",  "tim": "2000-01-02T03:04:05Z", "message": "msg0"}`
	const lin1 = `{"level": "debug", "str": "abc",                  "message": "msg1"}`

	t.Run("entry with the field value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertTime("tim", time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertTime("tim", time.Now())

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoTime(t *testing.T) {
	const lin0 = `{"level": "info",  "tim": "2000-01-02T03:04:05Z", "message": "msg0"}`
	const lin1 = `{"level": "debug", "str": "abc",                  "message": "msg1"}`

	t.Run("field name exists with different value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertNoTime("tim", time.Now())

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertNoTime("tim", time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertDuration(t *testing.T) {
	const lin0 = `{"level": "info",  "dur": 1000,  "message": "msg0"}`
	const lin1 = `{"level": "debug", "str": "abc", "message": "msg1"}`

	t.Run("field and value found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertDuration("dur", time.Second)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name and value not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertDuration("dur", time.Hour)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entries_AssertNoDuration(t *testing.T) {
	const lin0 = `{"level": "info",  "dur": 1000,  "message": "msg0"}`
	const lin1 = `{"level": "debug", "str": "abc", "message": "msg1"}`

	t.Run("field name exists with different value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertNoDuration("dur", time.Hour)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - field name exists with the value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1)

		// --- When ---
		have := ets.AssertNoDuration("dur", time.Second)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_exp(t *testing.T) {
	t.Run("entry found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// fn is a function always returning nil.
		fn := func(ent Entry) error { return nil }

		// --- When ---
		have := ets.exp(fn)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - empty log - no entries found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		// fn is a function always returning nil.
		fn := func(ent Entry) error { return nil }

		// --- When ---
		have := Entries{t: tspy}.exp(fn)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - no entries found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] no matching log entry found")
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// fn is a function always returning an error.
		fn := func(ent Entry) error { return errors.New("test message") }

		// --- When ---
		have := ets.exp(fn)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_notExp(t *testing.T) {
	t.Run("no entries found - fn returns no error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// fn is a function always returning nil.
		fn := func(ent Entry) error { return nil }

		// --- When ---
		have := Entries{t: tspy}.notExp(fn)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("no entries found - fn returns error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// fn is a function always returning an error.
		fn := func(ent Entry) error { return errors.New("test message") }

		// --- When ---
		have := Entries{t: tspy}.notExp(fn)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("error - found entry", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("[log entry] matching log entry found")
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.notExp(CheckStr("str", "msg1"))

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("not found entry", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.notExp(CheckStr("str", "xyz"))

		// --- Then ---
		assert.True(t, have)
	})
}

func Test_Summary(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.Summary()

		// --- Then ---
		assert.Equal(t, "no entries logged so far", have)
	})

	t.Run("some entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.Summary()

		// --- Then ---
		want := "" +
			"entries logged so far:\n" +
			"  " + lin0 + "\n" +
			"  " + lin1 + "\n" +
			"  " + lin2 + "\n"
		assert.Equal(t, want, have)
	})
}

func Test_summary(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.summary(3)

		// --- Then ---
		assert.Equal(t, "no entries logged so far", have)
	})

	t.Run("some entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.summary(2)

		// --- Then ---
		want := "" +
			"entries logged so far:\n" +
			"    " + lin0 + "\n" +
			"    " + lin1 + "\n" +
			"    " + lin2 + "\n"
		assert.Equal(t, want, have)
	})
}

func Test_print(t *testing.T) {
	t.Run("no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		have := ets.print()

		// --- Then ---
		assert.Equal(t, "", have)
	})

	t.Run("some entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		have := ets.print()

		// --- Then ---
		want := "" +
			lin0 + "\n" +
			lin1 + "\n" +
			lin2 + "\n"
		assert.Equal(t, want, have)
	})
}

func Test_Print(t *testing.T) {
	t.Run("error - no entries", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectLogEqual("no entries logged so far")
		tspy.Close()

		ets := MustEntries(tspy)

		// --- When ---
		ets.Print()
	})

	t.Run("some entries", func(t *testing.T) {
		// --- Given ---
		const lin0 = `{"level": "info", "str": "msg0"}`
		const lin1 = `{"level": "info", "str": "msg1"}`
		const lin2 = `{"level": "info", "str": "msg2"}`

		tspy := tester.New(t)
		wMsg := "" +
			"entries logged so far:\n" +
			"  " + lin0 + "\n" +
			"  " + lin1 + "\n" +
			"  " + lin2 + "\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ets := MustEntries(tspy, lin0, lin1, lin2)

		// --- When ---
		ets.Print()
	})
}
