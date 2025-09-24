// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_HasBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"bool_t": true, "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasBool(ent, "bool_t")

		// --- Then ---
		assert.NoError(t, err)
		assert.True(t, have)
	})

	t.Run("false", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"bool_f": false, "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasBool(ent, "bool_f")

		// --- Then ---
		assert.NoError(t, err)
		assert.False(t, have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"bool_t": true, "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasBool(ent, "number")

		// --- Then ---
		wMsg := "" +
			"[log entry] expected same types:\n" +
			"  field: number\n" +
			"   want: bool\n" +
			"   have: float64"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.False(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasBool(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"   type: bool\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.False(t, have)
	})
}

func Test_HasStr(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"str": "abc", "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasStr(ent, "str")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc", have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"str": "abc", "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasStr(ent, "number")

		// --- Then ---
		wMsg := "[log entry] expected same types:\n" +
			"  field: number\n" +
			"   want: string\n" +
			"   have: float64"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.Empty(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasStr(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"   type: string\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.Empty(t, have)
	})
}

func Test_HasTime(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)

		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS, "number": 42.0},
			t:   tspy,
		}

		// --- When ---
		have, err := HasTime(ent, "time")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, entTim, have)
	})

	t.Run("error - field has a wrong format", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": "2000-01-02", "number": 42.0},
			t:   tspy,
		}

		// --- When ---
		have, err := HasTime(ent, "time")

		// --- Then ---
		wMsg := "[log entry] expected log entry field to have formatted time:\n" +
			"  field: time\n" +
			"   want: 2006-01-02T15:04:05Z07:00\n" +
			"   have: 2000-01-02"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrFormat, err)
		assert.Empty(t, have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"time": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasTime(ent, "time")

		// --- Then ---
		wMsg := "[log entry] expected same types:\n" +
			"  field: time\n" +
			"   want: string\n" +
			"   have: float64"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.Empty(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasTime(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"   type: string\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.Empty(t, have)
	})
}

func Test_HasDur(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"dur": 1000.0, "number": 42.0},
			t:   tspy,
		}

		// --- When ---
		have, err := HasDur(ent, "dur")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, time.Second, have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"str": "abc", "number": 42.0},
			t: tspy,
		}

		// --- When ---
		have, err := HasDur(ent, "str")

		// --- Then ---
		wMsg := "[log entry] expected same types:\n" +
			"  field: str\n" +
			"   want: float64\n" +
			"   have: string"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.Empty(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasDur(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"   type: number\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.Empty(t, have)
	})
}

func Test_HasNum(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"float": 42.0, "str": "abc"},
			t: tspy,
		}

		// --- When ---
		have, err := HasNum(ent, "float")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, 42.0, have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"float": 42.0, "str": "abc"},
			t: tspy,
		}

		// --- When ---
		have, err := HasNum(ent, "str")

		// --- Then ---
		wMsg := "[log entry] expected same types:\n" +
			"  field: str\n" +
			"   want: float64\n" +
			"   have: string"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.Empty(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasNum(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"   type: number\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.Empty(t, have)
	})
}

func Test_HasMap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"map": map[string]any{"str": "abc"}},
			t: tspy,
		}

		// --- When ---
		have, err := HasMap(ent, "map")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{"str": "abc"}, have)
	})

	t.Run("error - field has a wrong type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: map[string]any{"str": "abc"},
			t: tspy,
		}

		// --- When ---
		have, err := HasMap(ent, "str")

		// --- Then ---
		wMsg := "[log entry] expected same types:\n" +
			"  field: str\n" +
			"   want: map[string]interface {}\n" +
			"   have: string"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrType, err)
		assert.Empty(t, have)
	})

	t.Run("error - field does not exist", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 0)
		tspy.Close()

		ent := Entry{
			m: make(map[string]any),
			t: tspy,
		}

		// --- When ---
		have, err := HasMap(ent, "missing")

		// --- Then ---
		wMsg := "[log entry] expected map to have a key:\n" +
			"  field: missing\n" +
			"    map: map[string]any{}"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrMissing, err)
		assert.Empty(t, have)
	})
}
