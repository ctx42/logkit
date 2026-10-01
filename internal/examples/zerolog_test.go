// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package examples

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"

	"github.com/ctx42/logkit/pkg/logkit"
)

func Test_Zerolog(t *testing.T) {
	// --- Given ---
	tst := logkit.New(t) // Initialize logkit.

	// Configure zerolog with Tester as the writer.
	log := zerolog.New(tst)

	// --- When ---
	// Use the log instance in your application.
	log.Info().Int("A", 0).Str("B", "x").Msg("msg 0")
	log.Warn().Int("A", 1).Str("B", "y").Msg("msg 1")
	log.Error().Int("A", 2).Str("B", "z").Msg("msg 2")

	// --- Then ---
	ets := tst.Entries()
	ets.AssertNumber("A", 2) // Success.
	ets.AssertStr("B", "z")  // Success.

	t.Log(tst.Entries().Summary())
}

func Example_zerolog() {
	t := &testing.T{} // Test manager.

	tst := logkit.New(t) // The zerolog format needs no configuration.

	// Set the Tester as the zerolog destination.
	log := zerolog.New(tst)

	// Use the log instance in your application.
	log.Info().Int("A", 0).Str("B", "x").Msg("msg 0")
	log.Warn().Int("A", 1).Str("B", "y").Msg("msg 1")
	log.Error().Int("A", 2).Str("B", "z").Msg("msg 2")

	ets := tst.Entries()
	fmt.Printf("has A=2: %v\n", ets.AssertNumber("A", 2))
	fmt.Printf("has B=z: %v\n", ets.AssertStr("B", "z"))
	fmt.Printf("has msg 1: %v\n", ets.AssertMsg("msg 1"))

	// Output:
	// has A=2: true
	// has B=z: true
	// has msg 1: true
}
