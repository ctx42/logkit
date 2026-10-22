// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

// Package logkit provides facilities to test structured JSON log messages.
package logkit

// HijackEnv is the environment variable [Hijack] sets to "1" for the duration
// of the test. Code under test can see it.
const HijackEnv = "LOGKIT_HIJACK"
