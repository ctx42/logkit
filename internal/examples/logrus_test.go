// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package examples

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/ctx42/logkit/pkg/logkit"
)

func Test_Logrus(t *testing.T) {
	// --- Given ---
	opt := logkit.WithConfig(logkit.LogrusConfig()) // Configure logkit.
	tst := logkit.New(t, opt)                       // Initialize logkit.

	// Configure Logrus.
	log := logrus.New()
	log.SetOutput(tst) // Set the Tester as the destination.
	log.SetFormatter(&logrus.JSONFormatter{})

	// --- When ---
	// Use the log instance in your application.
	log.WithField("A", 0).WithField("B", "x").Info("msg 0")
	log.WithField("A", 1).WithField("B", "y").Warn("msg 1")
	log.WithField("A", 2).WithField("B", "z").Error("msg 2")

	// --- Then ---
	ets := tst.Entries()
	ets.AssertNumber("A", 2) // Success.
	ets.AssertStr("B", "z")  // Success.

	t.Log(tst.Entries().Summary())
}
