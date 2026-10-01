// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package examples

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ctx42/logkit/pkg/logkit"
)

func Test_zap(t *testing.T) {
	// --- Given ---
	opt := logkit.WithConfig(logkit.ZapConfig()) // Configure logkit.
	tst := logkit.New(t, opt)                    // Initialize logkit.

	// Configure Zap.
	writer := zapcore.AddSync(tst) // Set the Tester as the destination.
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)
	log := zap.New(zapcore.NewCore(enc, writer, zapcore.InfoLevel))

	// --- When ---
	// Use the log instance in your application.
	log.Info("msg 0", zap.Int("A", 0), zap.String("B", "x"))
	log.Warn("msg 1", zap.Int("A", 1), zap.String("B", "y"))
	log.Error("msg 2", zap.Int("A", 2), zap.String("B", "z"))

	// --- Then ---
	ets := tst.Entries()
	ets.AssertNumber("A", 2) // Success.
	ets.AssertStr("B", "z")  // Success.

	t.Log(tst.Entries().Summary())
}

func Example_zap() {
	t := &testing.T{} // Test manager.

	opt := logkit.WithConfig(logkit.ZapConfig()) // Configure logkit.
	tst := logkit.New(t, opt)                    // Initialize logkit.

	// Set the Tester as the zap destination.
	writer := zapcore.AddSync(tst)
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	enc := zapcore.NewJSONEncoder(encCfg)
	log := zap.New(zapcore.NewCore(enc, writer, zapcore.InfoLevel))

	// Use the log instance in your application.
	log.Info("msg 0", zap.Int("A", 0), zap.String("B", "x"))
	log.Warn("msg 1", zap.Int("A", 1), zap.String("B", "y"))
	log.Error("msg 2", zap.Int("A", 2), zap.String("B", "z"))

	ets := tst.Entries()
	fmt.Printf("has A=2: %v\n", ets.AssertNumber("A", 2))
	fmt.Printf("has B=z: %v\n", ets.AssertStr("B", "z"))
	fmt.Printf("has msg 1: %v\n", ets.AssertMsg("msg 1"))

	// Output:
	// has A=2: true
	// has B=z: true
	// has msg 1: true
}
