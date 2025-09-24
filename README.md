[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/logkit)](https://goreportcard.com/report/github.com/ctx42/logkit)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/logkit)
![Tests](https://github.com/ctx42/logkit/actions/workflows/go.yml/badge.svg?branch=master)

# Logkit: Structured JSON Log Testing for Go

`logkit` is a lightweight, powerful Go library for testing structured JSON logs.
It simplifies capturing, filtering, waiting for, and asserting log entries,
ensuring your logging logic is robust and reliable. Designed for flexibility, it
integrates seamlessly with any logging library that supports JSON output and
custom `io.Writer` destinations.

<!-- TOC -->
* [Logkit: Structured JSON Log Testing for Go](#logkit-structured-json-log-testing-for-go)
  * [Why Logkit?](#why-logkit)
  * [Installation](#installation)
  * [Usage](#usage)
    * [With Zerolog](#with-zerolog)
    * [Slog](#slog)
    * [With Zap](#with-zap)
    * [With Logrus](#with-logrus)
  * [Assertions](#assertions)
    * [Custom Matchers for Complex Tests](#custom-matchers-for-complex-tests)
    * [Waiting for Asynchronous Logs](#waiting-for-asynchronous-logs)
    * [Loading Logs from Files](#loading-logs-from-files)
<!-- TOC -->

## Why Logkit?

Logging is critical for debugging and monitoring applications, but testing logs
is often overlooked or cumbersome. The `logkit` module addresses this by 
providing:

- **Simple Assertions**: Validate log levels, messages, and fields with clear, type-safe assert methods.
- **Flexible Matching**: Define custom matchers to test complex log structures.
- **Asynchronous Support**: Wait for logs in concurrent applications with configurable timeouts.
- **File Integration**: Load and test logs directly from files.
- **Type Safety**: Ensure fields match expected types (e.g., string, number, time) with precise error reporting.

## Installation

```shell
go get github.com/ctx42/logkit
```

## Usage

### With Zerolog

The [zerolog](https://github.com/rs/zerolog) log message format is supported 
out of the box without any additional configuration.

```go
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
```

### Slog

The [log/slog](https://pkg.go.dev/log/slog) log message format is supported 
by providing `logkit.SlogConfig()` configuration to the `logkit.Tester`.

```go
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
```

### With Zap

The [zap](https://github.com/uber-go/zap) log message format is supported
by providing `logkit.ZapConfig()` configuration to the `logkit.Tester`.

```go
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
```

### With Logrus

The [logrus](https://github.com/sirupsen/logrus) log message format is 
supported by providing `logkit.LogrusConfig()` configuration to the 
`logkit.Tester`.

```go
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
```

## Assertions

The `logkit` library provides two primary types for working with log entries:

- `Entries`: A collection of log entries, offering methods to assert fields across all entries.
- `Entry`: A single log entry, providing methods to assert individual fields.

Both types offer a suite of assertion methods for testing log entries. Below is 
a summary of the available assertions:

```
Entrues.AssertRaw(want ...string) bool
Entrues.AssertLen(want int) bool
Entrues.AssertMsg(want string) bool
Entrues.AssertNoMsg(want string) bool
Entrues.AssertMsgContain(want string) bool
Entrues.AssertNoMsgContain(want string) bool
Entrues.AssertError(want string) bool
Entrues.AssertErrorContain(want string) bool
Entrues.AssertNoError(want string) bool
Entrues.AssertErr(want error) bool
Entrues.AssertNoErr(want error) bool
Entrues.AssertContain(field, want string) bool
Entrues.AssertStr(field, want string) bool
Entrues.AssertNoStr(field, want string) bool
Entrues.AssertNumber(field string, want float64) bool
Entrues.AssertNoNumber(field string, want float64) bool
Entrues.AssertBool(field string, want bool) bool
Entrues.AssertTime(field string, want time.Time) bool
Entrues.AssertNoTime(field string, want time.Time) bool
Entrues.AssertDuration(field string, want time.Duration) bool
Entrues.AssertNoDuration(field string, want time.Duration) bool


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

### Custom Matchers for Complex Tests

Use `Matcher` to find a specific log entry with a set of assertions on multiple 
fields. Ideal for complex log entries.

```go
func Test_Match(t *testing.T) {
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
```

### Waiting for Asynchronous Logs

Test logs from goroutines or async processes with `WaitFor`, which supports 
configurable timeouts.

```go
func Test_WaitFor(t *testing.T) {
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
```

### Loading Logs from Files

Load and test logs from a file, useful for debugging or validating production logs.

```go
func Test_Load(t *testing.T) {
    tst := logkit.Load(t, "testdata/log.log")
    
    fmt.Println(tst.String())
    // Output:
    // {"level":"info", "str":"abc", "message":"msg0"}
    // {"level":"info", "str":"def", "message":"msg1"}
}
```
