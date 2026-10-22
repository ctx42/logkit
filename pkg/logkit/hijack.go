// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"log"
	"log/slog"

	"github.com/ctx42/testing/pkg/tester"
)

// Hijack redirects the process-wide loggers, [log] and [log/slog], to the
// returned [Tester] for the duration of the test. Both write JSON lines in the
// [slog.JSONHandler] format, so the [Tester] uses [SlogConfig] unless options
// override it. A [log] call becomes a record at the INFO level with the
// message, without the prefix, in the "msg" field. Use [WithSlogLevel] to
// change the minimum level of the captured records.
//
// Hijack calls [testing.T.Setenv] with [HijackEnv] before changing global
// state, so it panics in parallel tests. The cleanup restores the default
// [slog.Logger], the [slog.SetLogLoggerLevel] level, and the [log] output,
// flags and prefix.
func Hijack(t tester.T, opts ...func(*Tester)) *Tester {
	t.Helper()
	t.Setenv(HijackEnv, "1")

	orgDef := slog.Default()
	orgLvl := slog.SetLogLoggerLevel(slog.LevelInfo)
	orgOut, orgFlags, orgPrefix := log.Writer(), log.Flags(), log.Prefix()

	opts = append([]func(*Tester){WithConfig(SlogConfig())}, opts...)
	tst := New(t, opts...)
	lvl := tst.slvl
	if lvl == nil {
		lvl = slog.LevelDebug
	}

	// The prefix would become part of the message, and the flags must be
	// zeroed before slog.SetDefault captures them.
	log.SetFlags(0)
	log.SetPrefix("")
	hop := &slog.HandlerOptions{Level: lvl}
	slog.SetDefault(slog.New(slog.NewJSONHandler(tst, hop)))

	t.Cleanup(func() {
		// slog.SetDefault leaves the log output as is when given the
		// built-in default logger, so restore the log state explicitly.
		slog.SetDefault(orgDef)
		slog.SetLogLoggerLevel(orgLvl)
		log.SetOutput(orgOut)
		log.SetFlags(orgFlags)
		log.SetPrefix(orgPrefix)
	})
	return tst
}
