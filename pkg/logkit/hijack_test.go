// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"bytes"
	"log"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_Hijack(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		have := Hijack(tspy)

		// --- Then ---
		assert.Equal(t, SlogConfig(), have.cfg)
		assert.Equal(t, 0, have.Len())
		assert.Equal(t, "1", os.Getenv(HijackEnv))
	})

	t.Run("options override config", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		cfg := DefaultConfig()

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		// --- When ---
		have := Hijack(tspy, WithConfig(cfg))

		// --- Then ---
		assert.Same(t, cfg, have.cfg)
	})

	t.Run("log", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		log.Printf("msg %d", 0)

		// --- Then ---
		ent := tst.LastEntry()
		assert.Equal(t, 1, tst.Len())
		assert.NoError(t, CheckMsg("msg 0")(ent))
		assert.NoError(t, CheckLevel("INFO")(ent))
	})

	t.Run("log prefix and flags are not in message", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		log.SetPrefix("pfx: ")
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		log.Print("msg")

		// --- Then ---
		assert.NoError(t, CheckMsg("msg")(tst.LastEntry()))
	})

	t.Run("log level ignores slog log logger level", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		slog.SetLogLoggerLevel(slog.LevelWarn)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		log.Print("msg")

		// --- Then ---
		assert.NoError(t, CheckLevel("INFO")(tst.LastEntry()))
	})

	t.Run("slog", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		slog.Warn("msg", "A", 1)

		// --- Then ---
		ent := tst.LastEntry()
		assert.Equal(t, 1, tst.Len())
		assert.NoError(t, CheckMsg("msg")(ent))
		assert.NoError(t, CheckLevel("WARN")(ent))
		assert.NoError(t, CheckNumber("A", 1)(ent))
	})

	t.Run("slog debug captured by default", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		slog.Debug("msg")

		// --- Then ---
		assert.NoError(t, CheckLevel("DEBUG")(tst.LastEntry()))
	})

	t.Run("WithSlogLevel option", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy, WithSlogLevel(slog.LevelWarn))

		// --- When ---
		log.Print("msg 0")
		slog.Info("msg 1")
		slog.Warn("msg 2")

		// --- Then ---
		assert.Equal(t, 1, tst.Len())
		assert.NoError(t, CheckMsg("msg 2")(tst.LastEntry()))
	})

	t.Run("concurrent log and read", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		tst := Hijack(tspy)

		// --- When ---
		var wg sync.WaitGroup
		for i := range 10 {
			wg.Go(func() { log.Printf("msg %d", i) })
			wg.Go(func() { _ = tst.String() })
		}
		wg.Wait()

		// --- Then ---
		assert.Equal(t, 10, tst.Len())
		assert.Len(t, 10, tst.Entries().Get())
	})

	t.Run("cleanup restores built-in default", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		slog.SetLogLoggerLevel(slog.LevelWarn)
		log.SetFlags(log.Lshortfile)
		log.SetPrefix("pfx: ")
		def, out := slog.Default(), log.Writer()

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		Hijack(tspy)

		// --- When ---
		tspy.Finish()

		// --- Then ---
		assert.Same(t, def, slog.Default())
		assert.Equal(t, slog.LevelWarn, slog.SetLogLoggerLevel(slog.LevelWarn))
		assert.Same(t, out, log.Writer())
		assert.Equal(t, log.Lshortfile, log.Flags())
		assert.Equal(t, "pfx: ", log.Prefix())
	})

	t.Run("cleanup restores custom default", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		buf := &bytes.Buffer{}
		slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
		def, out := slog.Default(), log.Writer()

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		Hijack(tspy)

		// --- When ---
		tspy.Finish()

		// --- Then ---
		assert.Same(t, def, slog.Default())
		assert.Same(t, out, log.Writer())

		log.Print("msg")
		assert.Contain(t, "level=INFO msg=msg", buf.String())
	})

	t.Run("cleanup restores prefix changed during test", func(t *testing.T) {
		// --- Given ---
		KeepLoggers(t)
		log.SetPrefix("pfx: ")

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(1)
		tspy.Close()

		Hijack(tspy)
		log.SetPrefix("changed: ")

		// --- When ---
		tspy.Finish()

		// --- Then ---
		assert.Equal(t, "pfx: ", log.Prefix())
	})

	t.Run("panics in parallel test", func(t *testing.T) {
		t.Parallel()

		// --- Given ---
		def, out := slog.Default(), log.Writer()

		tspy := tester.New(t)
		tspy.ExpectSetenv(HijackEnv, "1")
		tspy.ExpectCleanups(0)
		tspy.Close()

		// --- When ---
		have := func() { Hijack(tspy) }

		// --- Then ---
		assert.PanicContain(t, "t.Parallel", have)
		assert.Same(t, def, slog.Default())
		assert.Same(t, out, log.Writer())
	})
}
