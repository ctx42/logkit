// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package examples

import (
	"log"
	"log/slog"
	"testing"

	"github.com/ctx42/logkit/pkg/logkit"
)

func Test_Hijack(t *testing.T) {
	// --- Given ---
	// Redirect the process-wide log and slog loggers to logkit. The loggers
	// are restored when the test ends. Not allowed in parallel tests.
	tst := logkit.Hijack(t)

	// --- When ---
	// Code under test uses the process-wide loggers.
	log.Printf("msg %d", 0)
	slog.Warn("msg 1", "A", 1, "B", "y")
	slog.Error("msg 2", "A", 2, "B", "z")

	// --- Then ---
	ets := tst.Entries()
	ets.AssertMsg("msg 0")   // Success.
	ets.AssertNumber("A", 2) // Success.
	ets.AssertStr("B", "z")  // Success.

	t.Log(tst.Entries().Summary())
}
