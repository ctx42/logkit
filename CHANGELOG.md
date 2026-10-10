## v0.6.1 (Sat, 10 Oct 2026 19:10:02 UTC)
- build(deps): update github.com/ctx42/testing to v0.57.0.

## v0.6.0 (Thu, 01 Oct 2026 13:41:16 UTC)
- build: bump ctx42/testing to v0.56.0.
- docs: correct AGENTS.md heading.
- fix(logkit): fix deadlocks and stale state in Tester waits.
- fix(logkit): correct entry checks and assertion results.
- fix(logkit)!: fail the test fatally when Load cannot read the log.
- test(logkit): fix test helpers and flaky WaitFor tests.
- style(logkit): apply code style rules.
- docs: rework README around runnable examples.

## v0.5.1 (Mon, 04 May 2026 15:25:14 UTC)
- fix: use ZeroEntry for error-path Entry returns.

## v0.5.0 (Fri, 01 May 2026 20:06:22 UTC)
- test(logkit): fix flaky WaitForAny tests.
- chore: Update to Go 1.26 and update dependencies.

## v0.4.1 (Thu, 30 Apr 2026 16:28:53 UTC)
- chore: Update dependencies.

## v0.4.0 (Fri, 26 Sep 2025 10:01:28 UTC)
- feat!: Change how `logkit.Matcher` notifies about matches.

## v0.3.0 (Thu, 25 Sep 2025 12:17:56 UTC)
- feat: Calling `logkit.NewMatcher` with nil config means use `DefaultConfig`.

## v0.2.0 (Thu, 25 Sep 2025 07:58:38 UTC)
- feat: Rename `logkit/Trait.Writer` to `logkit/Trait.LogWriter`.

## v0.1.0 (Wed, 24 Sep 2025 07:54:18 UTC)
- feat: Initial implementation of Tester, Matcher, Entries, Entry, checks and the documentation.
- test: Add GitHub workflows.

