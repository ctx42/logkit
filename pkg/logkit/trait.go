// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"io"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// Trait provides functionality for testers to fail a test if a log entry is written
// but not inspected.
type Trait struct {
	// Reports true if logs were inspected by calling the ExamineLog method.
	accessed bool

	// Treat logs as inspected unless there are messages with an error level.
	ignoreNonErrors bool

	// Log tester.
	tlog *Tester
}

// NewTrait returns new instance of [Trait].
func NewTrait(t tester.T) *Trait {
	t.Helper()

	tr := &Trait{
		tlog:     New(t),
		accessed: false,
	}

	// If there are log messages in the test log and log was not
	// accessed from the test, the cleanup function will fail the test.
	// This forces the test to examine logs.
	t.Cleanup(func() {
		t.Helper()
		n := tr.tlog.Len()
		if tr.accessed || n == 0 {
			return
		}

		// Mark the test as failed only if messages with error or panic
		// log level were logged.
		if tr.ignoreNonErrors {
			var hasErrors bool
			for _, ent := range tr.tlog.Entries().Get() {
				val, _ := HasStr(ent, tr.tlog.cfg.LevelField)
				if val == tr.tlog.cfg.LevelErrorValue ||
					val == tr.tlog.cfg.LevelPanicValue {
					hasErrors = true
					break
				}
			}
			if !hasErrors {
				return
			}
		}

		msg := notice.New("expected logs to be examined").
			Append("message cnt", "%d", n).
			Append("log", "\n%s", notice.Indent(1, ' ', tr.tlog.String()))
		t.Error(msg)
	})
	return tr
}

// Writer returns the writer logger should use as a destination.
func (tr *Trait) Writer() io.Writer { return tr.tlog }

// ExamineLog returns the test log for examination. If it doesn't get called
// and there are messages logged, it will mark the test as failed.
func (tr *Trait) ExamineLog() *Tester {
	tr.accessed = true
	return tr.tlog
}

// IgnoreLogs doesn't mark the test as failed when the logs weren't examined.
func (tr *Trait) IgnoreLogs() *Trait {
	tr.accessed = true
	return tr
}

// IgnoreNonErrorLogs doesn't mark the test as failed when the logs weren't
// examined, and there are no log messages with error or panic log levels.
func (tr *Trait) IgnoreNonErrorLogs() *Trait {
	tr.ignoreNonErrors = true
	return tr
}

// ResetLog deletes all logged messages and resets the accessed flag.
func (tr *Trait) ResetLog() *Trait {
	tr.accessed = false
	tr.tlog.Reset()
	return tr
}
