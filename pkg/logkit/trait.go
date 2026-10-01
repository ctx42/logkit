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

	trt := &Trait{
		tlog:     New(t),
		accessed: false,
	}

	// If there are log messages in the test log and log was not
	// accessed from the test, the cleanup function will fail the test.
	// This forces the test to examine logs.
	t.Cleanup(func() {
		t.Helper()
		n := trt.tlog.Len()
		if trt.accessed || n == 0 {
			return
		}

		// Mark the test as failed only if messages with error or panic
		// log level were logged.
		if trt.ignoreNonErrors {
			var hasErrors bool
			for _, ent := range trt.tlog.Entries().Get() {
				val, _ := HasStr(ent, trt.tlog.cfg.LevelField)
				if val == trt.tlog.cfg.LevelErrorValue ||
					val == trt.tlog.cfg.LevelPanicValue {
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
			Append("log", "\n%s", notice.Indent(1, ' ', trt.tlog.String()))
		t.Error(msg)
	})
	return trt
}

// LogWriter returns the writer a logger should use as a destination.
func (trt *Trait) LogWriter() io.Writer { return trt.tlog }

// ExamineLog returns the test log for examination. If it doesn't get called
// and there are messages logged, it will mark the test as failed.
func (trt *Trait) ExamineLog() *Tester {
	trt.accessed = true
	return trt.tlog
}

// IgnoreLogs doesn't mark the test as failed when the logs weren't examined.
func (trt *Trait) IgnoreLogs() *Trait {
	trt.accessed = true
	return trt
}

// IgnoreNonErrorLogs doesn't mark the test as failed when the logs weren't
// examined, and there are no log messages with error or panic log levels.
func (trt *Trait) IgnoreNonErrorLogs() *Trait {
	trt.ignoreNonErrors = true
	return trt
}

// ResetLog deletes all logged messages and resets the accessed flag.
func (trt *Trait) ResetLog() *Trait {
	trt.accessed = false
	trt.tlog.Reset()
	return trt
}
