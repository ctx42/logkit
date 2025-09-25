// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_CheckBool(t *testing.T) {
	t.Run("equal true", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"bool_t": true, "number": 42.0}}

		// --- When ---
		err := CheckBool("bool_t", true)(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("equal false", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"bool_f": false, "number": 42.0}}

		// --- When ---
		err := CheckBool("bool_f", false)(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"bool_t": true, "number": 42.0}}

		// --- When ---
		err := CheckBool("bool_t", false)(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: bool_t\n" +
			"   want: false\n" +
			"   have: true"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"bool_t": true, "number": 42.0}}

		// --- When ---
		err := CheckBool("missing", true)(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckStr(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc", "number": 42.0}}

		// --- When ---
		err := CheckStr("str", "abc")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc", "number": 42.0}}

		// --- When ---
		err := CheckStr("str", "xyz")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: str\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckStr("missing", "abc")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckStrErr(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc", "number": 42.0}}

		// --- When ---
		err := CheckStrErr("str", errors.New("abc"))(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc", "number": 42.0}}

		// --- When ---
		err := CheckStrErr("str", errors.New("xyz"))(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: str\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckStrErr("missing", errors.New("abc"))(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckContain(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc def ghi", "number": 42.0}}

		// --- When ---
		err := CheckContain("str", "def")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"str": "abc def ghi", "number": 42.0}}

		// --- When ---
		err := CheckContain("str", "xyz")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected string to contain substring:\n" +
			"      field: str\n" +
			"     string: \"abc def ghi\"\n" +
			"  substring: \"xyz\""
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckContain("missing", "abc")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckMsg(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc", "number": 42.0},
		}

		// --- When ---
		err := CheckMsg("abc")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc", "number": 42.0},
		}

		// --- When ---
		err := CheckMsg("xyz")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: %s\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		wMsg = fmt.Sprintf(wMsg, "message")
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"number": 42.0},
		}

		// --- When ---
		err := CheckMsg("abc")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckMsgContain(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc def ghi"},
		}

		// --- When ---
		err := CheckMsgContain("def")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"message": "abc def ghi"},
		}

		// --- When ---
		err := CheckMsgContain("xyz")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected string to contain substring:\n" +
			"      field: %s\n" +
			"     string: \"abc def ghi\"\n" +
			"  substring: \"xyz\""
		wMsg = fmt.Sprintf(wMsg, "message")
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"number": 42.0},
		}

		// --- When ---
		err := CheckMsgContain("abc")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckErrContain(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc def ghi"},
		}

		// --- When ---
		err := CheckErrContain("def")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"error": "abc def ghi"},
		}

		// --- When ---
		err := CheckErrContain("xyz")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected string to contain substring:\n" +
			"      field: %s\n" +
			"     string: \"abc def ghi\"\n" +
			"  substring: \"xyz\""
		wMsg = fmt.Sprintf(wMsg, "error")
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"number": 42.0},
		}

		// --- When ---
		err := CheckErrContain("abc")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckTime(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS, "number": 42.0},
		}

		// --- When ---
		err := CheckTime("time", entTim)(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		entTim := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
		entTimS := entTim.Format(time.RFC3339)
		wantTim := time.Date(2222, 1, 2, 3, 4, 5, 0, time.UTC)
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"time": entTimS, "number": 42.0},
		}

		// --- When ---
		err := CheckTime("time", wantTim)(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected equal dates:\n" +
			"  field: time\n" +
			"   want: 2222-01-02T03:04:05Z\n" +
			"   have: 2000-01-02T03:04:05Z\n" +
			"   diff: 1946016h0m0s"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckTime("missing", time.Now())(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckDuration(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"dur": 1000.0, "number": 42.0},
		}

		// --- When ---
		err := CheckDuration("dur", time.Second)(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"dur": 1000.0, "number": 42.0},
		}

		// --- When ---
		err := CheckDuration("dur", time.Hour)(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected equal time durations:\n" +
			"  field: dur\n" +
			"   want: 3600000 (1h0m0s)\n" +
			"   have: 1000 (1s)"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckDuration("missing", time.Second)(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckNumber(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"float": 42.0, "str": "abc"}}

		// --- When ---
		err := CheckNumber("float", 42)(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"float": 42.0, "str": "abc"}}

		// --- When ---
		err := CheckNumber("float", 43)(ent)

		// --- Then ---
		wMsg := "error checking log entry:\n" +
			"  field: float\n" +
			"   want: 43\n" +
			"   have: 42"
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckNumber("missing", 42)(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_CheckLevel(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"level": "info"},
		}

		// --- When ---
		err := CheckLevel("info")(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   map[string]any{"level": "info"},
		}

		// --- When ---
		err := CheckLevel("error")(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  field: level\n" +
			"   want: \"error\"\n" +
			"   have: \"info\""
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{
			cfg: DefaultConfig(),
			m:   make(map[string]any),
		}

		// --- When ---
		err := CheckLevel("info")(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}

func Test_check_level_success_tabular(t *testing.T) {
	tt := []struct {
		testN string

		check Checker
		level string
	}{
		{"debug", CheckDebug(), DefaultConfig().LevelDebugValue},
		{"info", CheckInfo(), DefaultConfig().LevelInfoValue},
		{"warn", CheckWarn(), DefaultConfig().LevelWarnValue},
		{"error", CheckError(), DefaultConfig().LevelErrorValue},
		{"fatal", CheckFatal(), DefaultConfig().LevelFatalValue},
		{"panic", CheckPanic(), DefaultConfig().LevelPanicValue},
		{"trace", CheckTrace(), DefaultConfig().LevelTraceValue},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t, 0)
			tspy.Close()

			ent := Entry{
				cfg: DefaultConfig(),
				m:   map[string]any{"level": tc.level},
			}

			// --- When ---
			err := tc.check(ent)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_check_level_failure_tabular(t *testing.T) {
	tt := []struct {
		testN string

		check Checker
		want  string
		have  string
	}{
		{"debug", CheckDebug(), DefaultConfig().LevelDebugValue, "info"},
		{"info", CheckInfo(), "info", DefaultConfig().LevelTraceValue},
		{"warn", CheckWarn(), DefaultConfig().LevelWarnValue, "info"},
		{"error", CheckError(), DefaultConfig().LevelErrorValue, "info"},
		{"fatal", CheckFatal(), DefaultConfig().LevelFatalValue, "info"},
		{"panic", CheckPanic(), DefaultConfig().LevelPanicValue, "info"},
		{"trace", CheckTrace(), DefaultConfig().LevelTraceValue, "info"},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			tspy := tester.New(t, 0)
			tspy.Close()

			ent := Entry{
				cfg: DefaultConfig(),
				m:   map[string]any{"level": tc.have},
			}

			// --- When ---
			err := tc.check(ent)

			// --- Then ---
			wMsg := "" +
				"[log entry] expected values to be equal:\n" +
				"  field: level\n" +
				"   want: %q\n" +
				"   have: %q"
			wMsg = fmt.Sprintf(wMsg, tc.want, tc.have)
			assert.ErrorEqual(t, wMsg, err)
			assert.ErrorIs(t, ErrValue, err)
		})
	}
}

func Test_CheckMap(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"map": map[string]any{"str": "abc"}}}

		// --- When ---
		err := CheckMap("map", map[string]any{"str": "abc"})(ent)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - when a field is not equal", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: map[string]any{"map": map[string]any{"str": "abc"}}}

		// --- When ---
		err := CheckMap("map", map[string]any{"str": "xyz"})(ent)

		// --- Then ---
		wMsg := "" +
			"[log entry] expected values to be equal:\n" +
			"  trail: map[\"str\"]\n" +
			"   want: \"xyz\"\n" +
			"   have: \"abc\""
		assert.ErrorEqual(t, wMsg, err)
		assert.ErrorIs(t, ErrValue, err)
	})

	t.Run("error - when a field does not exist", func(t *testing.T) {
		// --- Given ---
		ent := Entry{m: make(map[string]any)}

		// --- When ---
		err := CheckMap("missing", map[string]any{"str": "abc"})(ent)

		// --- Then ---
		assert.ErrorIs(t, ErrMissing, err)
	})
}
