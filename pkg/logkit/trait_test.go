// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_NewTrait(t *testing.T) {
	t.Run("no logs written", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		tr := NewTrait(tspy)

		// --- Then ---
		assert.NotNil(t, tr)
	})

	t.Run("error - logs written are not examined", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		wMsg := "expected logs to be examined:\n" +
			"  message cnt: 2\n" +
			"          log:\n" +
			"                {\"level\":\"debug\",\"message\":\"msg0\"}\n" +
			"                {\"level\":\"info\",\"message\":\"msg1\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		tr := NewTrait(tspy)

		// --- When ---
		MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
		MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)
	})
}

func Test_Trait_ExamineLog(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	tr := NewTrait(tspy)
	MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
	MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)

	// --- When ---
	have := tr.ExamineLog()

	// --- Then ---
	ets := have.Entries()
	assert.Equal(t, `{"level":"debug","message":"msg0"}`, ets.Entry(0).String())
	assert.Equal(t, `{"level":"info","message":"msg1"}`, ets.Entry(1).String())
	assert.Len(t, 2, ets.Get())
}

func Test_Trait_IgnoreLogs(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	tr := NewTrait(tspy)
	MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
	MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)
	MustWriteLine(tr.tlog, `{"level":"error","message":"msg2"}`)

	// --- When ---
	have := tr.IgnoreLogs()

	// --- Then ---
	assert.Same(t, tr, have)
}

func Test_Trait_IgnoreNonErrorLogs(t *testing.T) {
	t.Run("error - with error", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.ExpectError()
		wMsg := "" +
			"expected logs to be examined:\n" +
			"  message cnt: 3\n" +
			"          log:\n" +
			"                {\"level\":\"debug\",\"message\":\"msg0\"}\n" +
			"                {\"level\":\"info\",\"message\":\"msg1\"}\n" +
			"                {\"level\":\"error\",\"message\":\"msg2\"}\n"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		tr := NewTrait(tspy)
		MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
		MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)
		MustWriteLine(tr.tlog, `{"level":"error","message":"msg2"}`)

		// --- When ---
		have := tr.IgnoreNonErrorLogs()

		// --- Then ---
		assert.Same(t, tr, have)
	})

	t.Run("without errors", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		tr := NewTrait(tspy)
		MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
		MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)

		// --- When ---
		have := tr.IgnoreNonErrorLogs()

		// --- Then ---
		assert.Same(t, tr, have)
	})
}

func Test_Trait_ResetLog(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.ExpectCleanups(1)
	tspy.Close()

	tr := NewTrait(tspy)
	MustWriteLine(tr.tlog, `{"level":"debug","message":"msg0"}`)
	MustWriteLine(tr.tlog, `{"level":"info","message":"msg1"}`)

	// --- When ---
	have := tr.ResetLog()

	// --- Then ---
	assert.Same(t, tr, have)
	assert.Len(t, 0, tr.ExamineLog().Entries().Get())
}
