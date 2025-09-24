// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"time"
)

// Config holds information about the log messages fields and their formats.
type Config struct {
	TimeField    string // Log message time field name.
	LevelField   string // Log message level field name.
	MessageField string // Log message field name.
	ErrorField   string // Log message error field name.

	LevelTraceValue string // The [Config.LevelField] trace level value.
	LevelDebugValue string // The [Config.LevelField] debug level value.
	LevelInfoValue  string // The [Config.LevelField] info level value.
	LevelWarnValue  string // The [Config.LevelField] warn level value.
	LevelErrorValue string // The [Config.LevelField] error level value.
	LevelFatalValue string // The [Config.LevelField] fatal level value.
	LevelPanicValue string // The [Config.LevelField] panic level value.

	TimeFormat   string        // The [Config.TimeField] time format.
	DurationUnit time.Duration // The [time.Duration] unit.
}

// DefaultConfig returns the default instance of [Config] which matches the
// `zerolog` defaults.
func DefaultConfig() *Config {
	return &Config{
		TimeField:    "time",
		LevelField:   "level",
		MessageField: "message",
		ErrorField:   "error",

		TimeFormat:   time.RFC3339,
		DurationUnit: time.Millisecond,

		LevelTraceValue: "trace",
		LevelDebugValue: "debug",
		LevelInfoValue:  "info",
		LevelWarnValue:  "warn",
		LevelErrorValue: "error",
		LevelFatalValue: "fatal",
		LevelPanicValue: "panic",
	}
}

// SlogConfig returns the default instance of [Config] which matches the
// `log/slog` defaults.
func SlogConfig() *Config {
	return &Config{
		TimeField:    "time",
		LevelField:   "level",
		MessageField: "msg",
		ErrorField:   "error",

		TimeFormat:   time.RFC3339,
		DurationUnit: time.Millisecond,

		LevelTraceValue: "TRACE", // Not supported by slog.
		LevelDebugValue: "DEBUG",
		LevelInfoValue:  "INFO",
		LevelWarnValue:  "WARN",
		LevelErrorValue: "ERROR",
		LevelFatalValue: "FATAL", // Not supported by slog.
		LevelPanicValue: "PANIC", // Not supported by slog.
	}
}

// LogrusConfig returns the default instance of [Config] which matches the
// `logrus` defaults.
func LogrusConfig() *Config {
	return &Config{
		TimeField:    "time",
		LevelField:   "level",
		MessageField: "msg",
		ErrorField:   "error",

		TimeFormat:   time.RFC3339,
		DurationUnit: time.Nanosecond,

		LevelTraceValue: "trace",
		LevelDebugValue: "debug",
		LevelInfoValue:  "info",
		LevelWarnValue:  "warning",
		LevelErrorValue: "error",
		LevelFatalValue: "fatal",
		LevelPanicValue: "panic",
	}
}

// ZapConfig returns the instance of [Config] configured for `zap`.
func ZapConfig() *Config {
	return &Config{
		TimeField:    "ts",
		LevelField:   "level",
		MessageField: "msg",
		ErrorField:   "", // Not used in zap.

		TimeFormat:   time.RFC3339,
		DurationUnit: time.Second,

		LevelTraceValue: "trace", // Not supported by zap.
		LevelDebugValue: "debug",
		LevelInfoValue:  "info",
		LevelWarnValue:  "warn",
		LevelErrorValue: "error",
		LevelFatalValue: "fatal",
		LevelPanicValue: "panic",
	}
}
