# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests with race detector
go test -race ./...

# Run tests for the main package only
go test -race ./pkg/logkit/...

# Run a single test
go test -race -run Test_Tester_WaitFor ./pkg/logkit/...

# Run examples module tests (separate go.mod)
cd internal/examples && go test -race ./...
```

There is no separate build or lint step beyond standard `go test`.

## Module Structure

The repo contains two Go modules:

| Path                | Module              | Purpose                                       |
|---------------------|---------------------|-----------------------------------------------|
| `.`                 | `github.com/ctx42/logkit` | Public library                          |
| `internal/examples` | `example`           | Integration examples (uses replace directive) |

All public code lives in `pkg/logkit/`. The `internal/examples` module replaces `github.com/ctx42/logkit` with `../..` and pulls in zerolog, zap, logrus, and slog to demonstrate real-world integration.

## Architecture

The library provides test helpers for asserting structured JSON log output. Its public surface is a single package (`pkg/logkit`).

### Core types

**`Tester`** — the primary entry point. Implements `io.Writer` so it can be passed directly to any logger (zerolog, slog, zap, logrus). Accumulates raw log lines in a byte buffer, counts writes, and manages a queue of `Matcher`s for asynchronous matching. Created with `New(t, ...opts)` or `Load(t, path)`.

**`Entry`** — a single decoded log line. Holds the raw JSON string, the `map[string]any` decoded from it, and the `*Config` that names the logger's fields. All `Assert*` methods mark the test as failed and continue execution (non-fatal).

**`Entries`** — an ordered slice of `Entry` values returned by `Tester.Entries()` or `Tester.Filter(...)`. Provides collection-level `Assert*` methods that pass when _any_ entry satisfies the condition.

**`Matcher`** — holds a list of `Checker` functions and runs them all against an `Entry` or raw line. Used internally by `Tester.WaitFor`, `Tester.Filter`, and `Tester.Match`, and can also be instantiated directly via `NewMatcher`.

**`Checker`** — `type Checker func(Entry) error`. The composable predicate type. The `Check*` family of functions (`CheckStr`, `CheckNumber`, `CheckLevel`, `CheckMsg`, etc.) all return `Checker`.

**`Config`** — maps logger-specific field names and level strings to logkit's internal expectations. Four presets are provided: `DefaultConfig()` (zerolog), `SlogConfig()`, `ZapConfig()`, `LogrusConfig()`.

**`Trait`** — a composable struct for embedding in test helpers. Wraps a `Tester` and fails the test in cleanup if logs were written but never examined (`ExamineLog()` not called).

### Data flow

```
Logger → Tester.Write() → buf ([]byte)
                        → Matcher queue (for WaitFor)

Tester.Entries()   → json.Decoder over buf → []Entry → Entries
Tester.Filter(...) → Entries + Matcher.MatchEntry per entry → filtered Entries
Tester.WaitFor(...)→ checks existing entries, then registers Matcher + channel,
                     blocks until match or timeout
Tester.Match(mcr)  → linear scan of Entries using Matcher.MatchEntry
```

### `has.go` vs `check.go`

`has.go` provides `Has*` functions (e.g., `HasStr`, `HasNum`, `HasBool`) that extract and type-assert fields from an `Entry`, returning typed values and errors wrapped in `ErrMissing` / `ErrType`. `check.go` builds `Checker` functions on top of `Has*` functions, adding value comparison and `ErrValue` wrapping.

### Test helpers (`all_test.go`)

`all_test.go` is compiled only for tests and contains shared helpers used across the package's test files: `MustEntries`, `JSON2Map`, `MustWriteLine`, `ReceiveValue`, and the `EntryCheck` custom type-checker registered with the `ctx42/testing` framework. Tests use `tester.New(t)` from `github.com/ctx42/testing` as a spy `T` to assert that methods fail the test when expected.

### Test ordering convention

Test functions in `*_test.go` files must appear in the same order as the corresponding functions in the production file they test.