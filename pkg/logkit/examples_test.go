// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ctx42/logkit/pkg/logkit"
)

func ExampleNew() {
	t := &testing.T{} // Test manager.

	tst := logkit.New(t) // Somewhere in your tests initialize log tester.

	// Use it as the destination [io.Writer] in your logging library.

	// Here we will write directly to the writer.
	_, _ = tst.Write([]byte(`{"level": "debug", "message": "msg 0"}`))
	_, _ = tst.Write([]byte(`{"level": "error", "message": "msg 1"}`))

	ets := tst.Entries() // Extract logged entries.

	// Assert on all entries.
	fmt.Printf("has two messages: %v\n", ets.AssertLen(2))
	fmt.Printf("has msg 0: %v\n", ets.AssertMsg("msg 0"))
	fmt.Printf("has msg 1: %v\n", ets.AssertMsg("msg 1"))

	// Assert on individual entries.
	ent := ets.Entry(0)
	fmt.Printf("debug: %v\n", ent.AssertLevel("debug"))

	ent = ets.Entry(1)
	fmt.Printf("error: %v\n", ent.AssertLevel("error"))

	// Output:
	// has two messages: true
	// has msg 0: true
	// has msg 1: true
	// debug: true
	// error: true
}

func ExampleTester_Match() {
	t := &testing.T{}

	mcr := logkit.NewMatcher(
		t,
		logkit.DefaultConfig(),
		logkit.CheckNumber("A", 1),
		logkit.CheckNumber("B", 2),
	)

	tst := logkit.New(t)

	// Example logs.
	_, _ = tst.Write([]byte(`{"level": "info", "A": 1, "B": 44, "message": "msg 0"}`))
	_, _ = tst.Write([]byte(`{"level": "info", "A": 1, "B": 42, "message": "msg 1"}`))
	_, _ = tst.Write([]byte(`{"level": "info", "A": 1, "B": 2, "message": "msg 2"}`))

	// Find the first matching entry.
	ent := tst.Match(mcr)

	fmt.Printf("found: %v\n", ent.String())
	// Output:
	// found: {"level": "info", "A": 1, "B": 2, "message": "msg 2"}
}

func ExampleTester_WaitFor() {
	t := &testing.T{}

	tst := logkit.New(t)

	go func() {
		_, _ = tst.Write([]byte(`{"level": "debug", "A": 0}`))
		time.Sleep(500 * time.Millisecond)
		_, _ = tst.Write([]byte(`{"level": "error", "A": 1}`))
	}()

	ent := tst.WaitFor("1s", logkit.CheckNumber("A", 1))

	fmt.Printf("found: %v\n", ent.String())
	// Output:
	// found: {"level": "error", "A": 1}
}

func ExampleLoad() {
	t := &testing.T{}

	tst := logkit.Load(t, "testdata/log.log")

	fmt.Println(tst.String())
	// Output:
	// {"level":"info", "str":"abc", "message":"msg0"}
	// {"level":"info", "str":"def", "message":"msg1"}
}
