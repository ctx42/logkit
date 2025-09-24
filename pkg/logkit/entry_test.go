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

func Test_Entry_IsZero(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{raw: "", t: tspy}

		// --- When ---
		have := ent.IsZero()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not zero", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{raw: "abc", t: tspy}

		// --- When ---
		have := ent.IsZero()

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Index(t *testing.T) {
	// --- Given ---
	tst := New(t)
	MustWriteLine(tst, `{"level": "error", "A": 1}`)
	MustWriteLine(tst, `{"level": "error", "B": 2}`)

	// --- When ---
	have := tst.Entries()

	// --- Then ---
	assert.Equal(t, 0, have.Entry(0).Index())
	assert.Equal(t, 1, have.Entry(1).Index())
}

func Test_Entry_String(t *testing.T) {
	// --- Given ---
	tst := New(t)
	MustWriteLine(tst, `{"level": "error", "A": 1}`)
	ent := tst.LastEntry()

	// --- When ---
	have := ent.String()

	// --- Then ---
	assert.Equal(t, `{"level": "error", "A": 1}`, have)
}

func Test_Entry_Bytes(t *testing.T) {
	// --- Given ---
	tst := New(t)
	MustWriteLine(tst, `{"level": "error", "A": 1}`)
	ent := tst.LastEntry()

	// --- When ---
	have := ent.Bytes()

	// --- Then ---
	assert.Equal(t, []byte(`{"level": "error", "A": 1}`), have)
}

func Test_Entry_MetaAll(t *testing.T) {
	tst := New(t)
	msg := `{
        "level": "error", 
        "bools": [true, false, true],
        "bool":  true,
        "time":  "2000-01-02T03:04:05Z",
        "bytes": "\u0001\u0002\u0003",
        "str":   "abc",
        "int":   123,
        "int64": 456
    }`
	MustWriteLine(tst, msg)
	ent := tst.LastEntry()

	// --- When ---
	have := ent.MetaAll()

	// --- Then ---
	want := map[string]any{
		"level": "error",
		"bools": []any{true, false, true},
		"bool":  true,
		"time":  "2000-01-02T03:04:05Z",
		"bytes": "\x01\x02\x03",
		"str":   "abc",
		"int":   float64(123),
		"int64": float64(456),
	}
	assert.Equal(t, want, have)
}

func Test_Entry_AssertRaw(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			raw: `{"A": 1}`,
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertRaw(`{"A": 1}`)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		wMsg := "" +
			"[log entry] expected JSON strings to be equal:\n" +
			"  want: {\"A\":2}\n" +
			"  have: {\"A\":1}"
		tspy.ExpectLogEqual(wMsg)
		tspy.ExpectError()
		tspy.Close()

		ent := &Entry{
			raw: `{"A": 1}`,
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertRaw(`{"A": 2}`)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertExist(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertExist("str")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected log entry field to be present:\n  field: missing"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have := ent.AssertExist("missing")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertNotExist(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected log entry field not to be present:\n  field: str"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertNotExist("str")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have := ent.AssertNotExist("missing")

		// --- Then ---
		assert.True(t, have)
	})
}

func Test_Entry_AssertFieldCount(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc", "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertFieldCount(2)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"expected log entry to have N fields:\n" +
			"  want: 3\n" +
			"  have: 2"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc", "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertFieldCount(3)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertFieldType_tabular(t *testing.T) {
	tspy := tester.New(t)
	tspy.Close()

	ent := &Entry{
		idx: 0,
		m: map[string]any{
			"bool":        true,
			"string":      "val",
			"integer":     42,
			"float64":     42.42,
			"number":      42.42,
			"time":        time.Now(),
			"dur":         time.Second,
			"map":         map[string]any{"k": "v"},
			"unsupported": struct{}{},
		},
		t: tspy,
	}

	tt := []struct {
		testN string
		want  FieldType
	}{
		{"bool", TypBool},
		{"string", TypString},
		{"integer", TypNumber},
		{"float64", TypNumber},
		{"number", TypNumber},
		{"time", TypTime},
		{"dur", TypDur},
		{"map", TypMap},
		{"unsupported", TypUnsupported},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := ent.AssertFieldType(tc.testN, tc.want)

			// --- Then ---
			assert.True(t, have)
		})
	}
}

func Test_Entry_AssertFieldType(t *testing.T) {
	t.Run("error - not existing field", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected log entry field to be present:\n  field: missing"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have := ent.AssertFieldType("missing", TypString)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"expected log entry field type:\n" +
			"  want: number\n" +
			"  have: string"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertFieldType("str", TypNumber)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("error - unsupported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected log entry field type:\n" +
			"  want: number\n" +
			"  have: struct {}"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"struct": struct{}{}},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertFieldType("struct", TypNumber)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Level_tabular(t *testing.T) {
	tt := []struct {
		testN string

		m       map[string]any
		wantVal string
		wantErr error
	}{
		{
			"ok",
			map[string]any{"level": "info"},
			"info",
			nil,
		},
		{
			"missing",
			map[string]any{},
			"",
			ErrMissing,
		},
		{
			"wrong type",
			map[string]any{"level": 42.0},
			"",
			ErrType,
		},
		{
			"empty",
			map[string]any{"level": ""},
			"",
			ErrValue,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				cfg: DefaultConfig(),
				m:   tc.m,
				t:   tspy,
			}

			// --- When ---
			have, err := ent.Level()

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertLevel(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"level": "info"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertLevel("info")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: %s\n" +
			"   want: \"error\"\n" +
			"   have: \"info\""
		tspy.ExpectLogEqual(wMsg, "level")
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"level": "info"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertLevel("error")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertMsg(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertMsg("abc")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: %s\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg, "message")
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertMsg("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertMsgErr(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertMsgErr(errors.New("abc"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: message\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertMsgErr(errors.New("xyz"))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertError(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertError("abc")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: %s\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg, "error")
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertError("xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertErr(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertErr(errors.New("abc"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: %s\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg, "error")
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc"},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertErr(errors.New("xyz"))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Str_tabular(t *testing.T) {
	tt := []struct {
		field   string
		wantVal string
		wantErr error
	}{
		{"str", "abc", nil},
		{"number", "", ErrType},
		{"missing", "", ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				m: map[string]any{"str": "abc", "number": 42.0},
				t: tspy,
			}

			// --- When ---
			have, err := ent.Str(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertStr(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertStr("str", "abc")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: str\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertStr("str", "xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertContain(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc def ghi"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertContain("str", "def")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("does not contain", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected string to contain substring:\n" +
			"      field: str\n" +
			"     string: \"abc def ghi\"\n" +
			"  substring: \"xyz\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"str": "abc def ghi"},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertContain("str", "xyz")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Number_tabular(t *testing.T) {
	tt := []struct {
		field   string
		wantVal float64
		wantErr error
	}{
		{"number", 42.0, nil},
		{"str", 0.0, ErrType},
		{"missing", 0.0, ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				m: map[string]any{"number": 42.0, "str": "abc"},
				t: tspy,
			}

			// --- When ---
			have, err := ent.Number(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertNumber(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"number": 42.0},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertNumber("number", 42.0)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "error checking log entry:\n" +
			"  field: number\n" +
			"   want: 44.1\n" +
			"   have: 42"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"number": 42.0},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertNumber("number", 44.1)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Bool_tabular(t *testing.T) {
	tt := []struct {
		field   string
		wantVal bool
		wantErr error
	}{
		{"bool_t", true, nil},
		{"bool_f", false, nil},
		{"number", false, ErrType},
		{"missing", false, ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				m: map[string]any{
					"bool_t": true,
					"bool_f": false,
					"number": 42.0,
				},
				t: tspy,
			}

			// --- When ---
			val, err := ent.Bool(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func Test_Entry_AssertBool(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"bool_t": true, "bool_f": false},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertBool("bool_t", true)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: bool_t\n" +
			"   want: false\n" +
			"   have: true"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"bool_t": true, "bool_f": false},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertBool("bool_t", false)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Time_tabular(t *testing.T) {
	entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	entTimS := entTim.Format(time.RFC3339)

	tt := []struct {
		field   string
		wantVal time.Time
		wantErr error
	}{
		{"time", entTim, nil},
		{"format", time.Time{}, ErrFormat},
		{"number", time.Time{}, ErrType},
		{"missing", time.Time{}, ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				cfg: DefaultConfig(),
				m: map[string]any{
					"time":   entTimS,
					"format": "2000-01-01",
					"number": 42.0,
				},
				t: tspy,
			}

			// --- When ---
			have, err := ent.Time(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertTime(t *testing.T) {
	entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	entTimS := entTim.Format(time.RFC3339)

	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertTime("time", entTim)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected equal dates:\n" +
			"  field: time\n" +
			"   want: 2000-01-02T03:04:06Z\n" +
			"   have: 2000-01-02T03:04:05Z\n" +
			"   diff: 1s"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertTime("time", entTim.Add(time.Second))

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertWithin(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertWithin("time", entTim, "1s")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("within", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertWithin("time", entTim.Add(time.Second), "1s")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not within", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected dates to be within:\n" +
			"         field: time\n" +
			"          want: 2000-01-02T04:04:05Z\n" +
			"          have: 2000-01-02T03:04:05Z\n" +
			"  max diff +/-: 59m59s\n" +
			"     have diff: 1h0m0s"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertWithin("time", entTim.Add(time.Hour), "59m59s")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("missing", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected map to have a key:\n" +
			"  field: time\n" +
			"   type: string\n" +
			"    map: map[string]any{}"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertWithin("time", time.Now(), "1h")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("invalid diff", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] [within] failed to parse duration:\n" +
			"  field: time\n" +
			"  value: abc"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertWithin("time", entTim, "abc")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_AssertLoggedWithin(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertLoggedWithin(entTim, "1s")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("within", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertLoggedWithin(entTim.Add(time.Second), "1s")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not within", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected dates to be within:\n" +
			"         field: time\n" +
			"          want: 2000-01-02T04:04:05Z\n" +
			"          have: 2000-01-02T03:04:05Z\n" +
			"  max diff +/-: 59m59s\n" +
			"     have diff: 1h0m0s"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertLoggedWithin(entTim.Add(time.Hour), "59m59s")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Duration_tabular(t *testing.T) {
	tt := []struct {
		field   string
		wantVal time.Duration
		wantErr error
	}{
		{"dur", time.Second, nil},
		{"str", 0.0, ErrType},
		{"missing", 0.0, ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				cfg: DefaultConfig(),
				m:   map[string]any{"dur": 1000.0, "str": "abc"},
				t:   tspy,
			}

			// --- When ---
			have, err := ent.Duration(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertDuration(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"dur": 1000.0},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertDuration("dur", time.Second)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected equal time durations:\n" +
			"  field: dur\n" +
			"   want: 1000 (1s)\n" +
			"   have: 1001 (1.001s)"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"dur": 1001.0},
			t:   tspy,
		}

		// --- When ---
		have := ent.AssertDuration("dur", time.Second)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Entry_Map_tabular(t *testing.T) {
	tt := []struct {
		field   string
		wantVal map[string]any
		wantErr error
	}{
		{"map", map[string]any{"str": "abc"}, nil},
		{"number", nil, ErrType},
		{"missing", nil, ErrMissing},
	}

	for _, tc := range tt {
		t.Run(tc.field, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t)
			tspy.Close()

			ent := &Entry{
				m: map[string]any{
					"map":    map[string]any{"str": "abc"},
					"number": 42.0,
				},
				t: tspy,
			}

			// --- When ---
			have, err := ent.Map(tc.field)

			// --- Then ---
			assert.ErrorIs(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, have)
		})
	}
}

func Test_Entry_AssertMap(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"map": map[string]any{"str": "abc"}},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertMap("map", map[string]any{"str": "abc"})

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  trail: map[\"str\"]\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		ent := &Entry{
			m: map[string]any{"map": map[string]any{"str": "abc"}},
			t: tspy,
		}

		// --- When ---
		have := ent.AssertMap("map", map[string]any{"str": "xyz"})

		// --- Then ---
		assert.False(t, have)
	})
}
