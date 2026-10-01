[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/logkit)](https://goreportcard.com/report/github.com/ctx42/logkit)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/logkit)
[![Tests](https://github.com/ctx42/logkit/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/ctx42/logkit/actions/workflows/go.yml)

# Logkit: Structured JSON Log Testing for Go

`logkit` is a Go library for testing structured JSON logs. It captures, filters,
waits for, and asserts log entries, and works with any logging library that
writes JSON to a custom `io.Writer`.

<!-- TOC -->
* [Logkit: Structured JSON Log Testing for Go](#logkit-structured-json-log-testing-for-go)
  * [Why Logkit?](#why-logkit)
  * [Prerequisites](#prerequisites)
  * [Installation](#installation)
  * [Usage](#usage)
    * [With Zerolog](#with-zerolog)
    * [With Slog](#with-slog)
    * [With Zap](#with-zap)
    * [With Logrus](#with-logrus)
    * [Custom Matchers for Complex Tests](#custom-matchers-for-complex-tests)
    * [Filtering Entries](#filtering-entries)
    * [Waiting for Asynchronous Logs](#waiting-for-asynchronous-logs)
    * [Loading Logs from Files](#loading-logs-from-files)
    * [Failing Tests With Unexamined Logs](#failing-tests-with-unexamined-logs)
  * [Configuration](#configuration)
  * [Assertions](#assertions)
  * [License](#license)
<!-- TOC -->

## Why Logkit?

Logging is critical for debugging and monitoring applications, but testing logs
is often overlooked or cumbersome. The `logkit` module addresses this by
providing:

- **Simple Assertions**: Validate log levels, messages, and fields with typed
  assert methods.
- **Flexible Matching**: Define custom matchers to test complex log structures.
- **Asynchronous Support**: Wait for logs in concurrent applications with
  configurable timeouts.
- **File Integration**: Load and test logs directly from files.
- **Type Safety**: Ensure fields match expected types (e.g., string, number,
  time) with precise error reporting.
- **Examined Logs**: Fail a test whose logs were written but never examined.

## Prerequisites

- Go 1.26 or newer.

## Installation

```shell
go get github.com/ctx42/logkit
```

Import the package:

```go
import "github.com/ctx42/logkit/pkg/logkit"
```

## Usage

### With Zerolog

The [zerolog](https://github.com/rs/zerolog) log message format is supported
out of the box without any additional configuration.

<!-- gmmce:internal/examples/Example_zerolog -->
```go
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
```

### With Slog

The [log/slog](https://pkg.go.dev/log/slog) log message format is supported
by providing `logkit.SlogConfig()` configuration to the `logkit.Tester`.

<!-- gmmce:internal/examples/Example_slog -->
```go
t := &testing.T{} // Test manager.

opt := logkit.WithConfig(logkit.SlogConfig()) // Configure logkit.
tst := logkit.New(t, opt)                     // Initialize logkit.

// Set the Tester as the slog destination.
log := slog.New(slog.NewJSONHandler(tst, nil))

// Use the log instance in your application.
log.Info("msg 0", "A", 0, "B", "x")
log.Warn("msg 1", "A", 1, "B", "y")
log.Error("msg 2", "A", 2, "B", "z")

ets := tst.Entries()
fmt.Printf("has A=2: %v\n", ets.AssertNumber("A", 2))
fmt.Printf("has B=z: %v\n", ets.AssertStr("B", "z"))
fmt.Printf("has msg 1: %v\n", ets.AssertMsg("msg 1"))

// Output:
// has A=2: true
// has B=z: true
// has msg 1: true
```

### With Zap

The [zap](https://github.com/uber-go/zap) log message format is supported
by providing `logkit.ZapConfig()` configuration to the `logkit.Tester`.

<!-- gmmce:internal/examples/Example_zap -->
```go
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
```

### With Logrus

The [logrus](https://github.com/sirupsen/logrus) log message format is
supported by providing `logkit.LogrusConfig()` configuration to the
`logkit.Tester`.

<!-- gmmce:internal/examples/Example_logrus -->
```go
t := &testing.T{} // Test manager.

opt := logkit.WithConfig(logkit.LogrusConfig()) // Configure logkit.
tst := logkit.New(t, opt)                       // Initialize logkit.

// Set the Tester as the logrus destination.
log := logrus.New()
log.SetOutput(tst)
log.SetFormatter(&logrus.JSONFormatter{})

// Use the log instance in your application.
log.WithField("A", 0).WithField("B", "x").Info("msg 0")
log.WithField("A", 1).WithField("B", "y").Warn("msg 1")
log.WithField("A", 2).WithField("B", "z").Error("msg 2")

ets := tst.Entries()
fmt.Printf("has A=2: %v\n", ets.AssertNumber("A", 2))
fmt.Printf("has B=z: %v\n", ets.AssertStr("B", "z"))
fmt.Printf("has msg 1: %v\n", ets.AssertMsg("msg 1"))

// Output:
// has A=2: true
// has B=z: true
// has msg 1: true
```

### Custom Matchers for Complex Tests

Use `Matcher` to find a specific log entry with a set of checks on multiple
fields. Ideal for complex log entries.

<!-- gmmce:pkg/logkit/ExampleTester_Match -->
```go
t := &testing.T{}

mcr := logkit.NewMatcher(
	t,
	logkit.DefaultConfig(),
	logkit.CheckNumber("A", 1),
	logkit.CheckNumber("B", 2),
)

tst := logkit.New(t)

// Example logs.
lines := []string{
	`{"level": "info", "A": 1, "B": 44, "message": "msg 0"}`,
	`{"level": "info", "A": 1, "B": 42, "message": "msg 1"}`,
	`{"level": "info", "A": 1, "B": 2, "message": "msg 2"}`,
}
for _, line := range lines {
	_, _ = io.WriteString(tst, line)
}

// Find the first matching entry.
ent := tst.Match(mcr)

fmt.Printf("found: %v\n", ent.String())
// Output:
// found: {"level": "info", "A": 1, "B": 2, "message": "msg 2"}
```

### Filtering Entries

`Tester.Filter` takes the same `Check…` functions as a `Matcher` and returns
every matching entry as `Entries`, so the collection assertions run on the
filtered subset only, for example
`tst.Filter(logkit.CheckError()).AssertLen(1)`.

### Waiting for Asynchronous Logs

Test logs from goroutines or async processes with `WaitFor`, which supports
configurable timeouts.

<!-- gmmce:pkg/logkit/ExampleTester_WaitFor -->
```go
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
```

Successive `WaitFor` calls match entries in the order they are logged; each
waits for an entry after the previous match. Use `WaitForAny` when the order
does not matter.

### Loading Logs from Files

Load and test logs from a file, useful for debugging or validating production
logs. When the file cannot be read, `Load` fails the test immediately.

<!-- gmmce:pkg/logkit/ExampleLoad -->
```go
t := &testing.T{}

tst := logkit.Load(t, "testdata/log.log")

fmt.Println(tst.String())
// Output:
// {"level":"info", "str":"abc", "message":"msg0"}
// {"level":"info", "str":"def", "message":"msg1"}
```

### Failing Tests With Unexamined Logs

Embed `Trait` in a test helper to make sure logs are not silently ignored.
Create it with `logkit.NewTrait(t)` and pass `LogWriter()` to the logger. If
anything was logged and the test never called `ExamineLog()`, the test fails at
cleanup. `IgnoreLogs()` opts out, and `IgnoreNonErrorLogs()` fails only when
entries with the error or panic level were logged.

## Configuration

`logkit.New` expects the zerolog field names by default. For other loggers pass
one of the presets with `logkit.WithConfig`:

| Logger  | Preset                   | Message   | Time   | Duration unit |
|---------|--------------------------|-----------|--------|---------------|
| zerolog | `logkit.DefaultConfig()` | `message` | `time` | millisecond   |
| slog    | `logkit.SlogConfig()`    | `msg`     | `time` | millisecond   |
| zap     | `logkit.ZapConfig()`     | `msg`     | `ts`   | second        |
| logrus  | `logkit.LogrusConfig()`  | `msg`     | `time` | nanosecond    |

For a custom format, start from a preset and change the fields of the returned
`Config`, such as `MessageField`, `LevelField`, `TimeFormat`, or the
`Level…Value` strings.

## Assertions

The `logkit` library provides two primary types for working with log entries:

- `Entries`: A collection of log entries, offering methods to assert fields
  across all entries.
- `Entry`: A single log entry, providing methods to assert individual fields.

Both types offer a suite of assertion methods for testing log entries. Below is
a summary of the available assertions:

```go
Entries.AssertRaw(want ...string) bool
Entries.AssertLen(want int) bool
Entries.AssertMsg(want string) bool
Entries.AssertNoMsg(want string) bool
Entries.AssertMsgContain(want string) bool
Entries.AssertNoMsgContain(want string) bool
Entries.AssertError(want string) bool
Entries.AssertErrorContain(want string) bool
Entries.AssertNoError(want string) bool
Entries.AssertErr(want error) bool
Entries.AssertNoErr(want error) bool
Entries.AssertContain(field, want string) bool
Entries.AssertStr(field, want string) bool
Entries.AssertNoStr(field, want string) bool
Entries.AssertNumber(field string, want float64) bool
Entries.AssertNoNumber(field string, want float64) bool
Entries.AssertBool(field string, want bool) bool
Entries.AssertTime(field string, want time.Time) bool
Entries.AssertNoTime(field string, want time.Time) bool
Entries.AssertDuration(field string, want time.Duration) bool
Entries.AssertNoDuration(field string, want time.Duration) bool

Entry.AssertRaw(want string) bool
Entry.AssertExist(field string) bool
Entry.AssertNotExist(field string) bool
Entry.AssertFieldCount(want int) bool
Entry.AssertFieldType(field string, want FieldType) bool
Entry.AssertLevel(want string) bool
Entry.AssertMsg(want string) bool
Entry.AssertMsgErr(want error) bool
Entry.AssertError(want string) bool
Entry.AssertErr(want error) bool
Entry.AssertStr(field, want string) bool
Entry.AssertContain(field, want string) bool
Entry.AssertNumber(field string, want float64) bool
Entry.AssertBool(field string, want bool) bool
Entry.AssertTime(key string, want time.Time) bool
Entry.AssertWithin(field string, want time.Time, diff string) bool
Entry.AssertLoggedWithin(want time.Time, diff string) bool
Entry.AssertDuration(field string, want time.Duration) bool
Entry.AssertMap(field string, want map[string]any) bool
```

## License

Released under the MIT License; see [LICENSE.md](LICENSE.md).
