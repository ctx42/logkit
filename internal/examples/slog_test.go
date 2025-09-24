// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package examples

import (
	"log/slog"
	"testing"

	"github.com/ctx42/logkit/pkg/logkit"
)

func Test_Slog(t *testing.T) {
	// --- Given ---
	opt := logkit.WithConfig(logkit.SlogConfig()) // Configure logkit.
	tst := logkit.New(t, opt)                     // Initialize logkit.

	// Configure slog.
	log := slog.New(slog.NewJSONHandler(tst, nil))

	// --- When ---
	// Use the log instance in your application.
	log.Info("msg 0", "A", 0, "B", "x")
	log.Warn("msg 1", "A", 1, "B", "y")
	log.Error("msg 2", "A", 2, "B", "z")

	// --- Then ---
	ets := tst.Entries()
	ets.AssertNumber("A", 2) // Success.
	ets.AssertStr("B", "z")  // Success.

	t.Log(tst.Entries().Summary())
}
