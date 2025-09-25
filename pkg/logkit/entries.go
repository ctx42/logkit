// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"strings"
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// Entries represents collection of log entries.
type Entries struct {
	cfg *Config  // Log configuration.
	ets []Entry  // Log entries.
	t   tester.T // Test manager.
}

// Get returns the slice of entries.
func (ets Entries) Get() []Entry {
	return ets.ets
}

// MetaAll returns entries as array of JSON decoded log entries.
func (ets Entries) MetaAll() []map[string]any {
	var etsMaps []map[string]any
	for _, ent := range ets.ets {
		etsMaps = append(etsMaps, ent.MetaAll())
	}
	return etsMaps
}

// Entry returns the nth log entry. If the index is out of range, the test is
// marked as failed, the method returns false but continues execution.
func (ets Entries) Entry(n int) Entry {
	ets.t.Helper()
	if n < len(ets.ets) {
		return ets.ets[n]
	}
	msg := notice.New("[log entry] expected log entry to exist").
		Append("index", "%d", n)
	ets.t.Error(msg)
	return Entry{}
}

// AssertRaw asserts that the raw log entries match the provided string.
// Returns true if they match. If not, it marks the test as failed, logs an
// error message, and returns false.
func (ets Entries) AssertRaw(want ...string) bool {
	ets.t.Helper()

	for i, wEnt := range want {
		hEnt := ets.Entry(i)
		if hEnt.IsZero() {
			return false
		}
		if e := check.JSON(wEnt, hEnt.raw); e != nil {
			e = notice.From(e, "log entry").Prepend("index", "%d", i)
			ets.t.Error(e)
		}
	}

	if ets.t.Failed() {
		return false
	}

	hCnt := len(ets.ets)
	wCnt := len(want)
	if hCnt == wCnt {
		return true
	}
	msg := notice.New("[log entry] expected N log entries").
		Want("%d", wCnt).
		Have("%d", hCnt).
		Append("have logs", "%s", ets.print())
	ets.t.Error(msg)
	return false
}

// AssertLen asserts that the number of log entries equals the provided length.
// Returns true if the count matches. If not, it marks the test as failed, logs
// an error message, and returns false.
func (ets Entries) AssertLen(want int) bool {
	ets.t.Helper()
	have := len(ets.ets)
	if have == want {
		return true
	}
	msg := notice.New("[log entry] expected N log entries").
		Want("%d", want).
		Have("%d", have)
	ets.t.Error(msg)
	return false
}

// AssertMsg asserts that at least one log entry in the collection has the
// field [Config.MessageField] with the specified value and type. Returns true
// if found and matches. If no entry has the field with the value and type, it
// marks the test as failed, logs an error message, and returns false.
func (ets Entries) AssertMsg(want string) bool {
	ets.t.Helper()
	return ets.exp(CheckStr(ets.cfg.MessageField, want))
}

// AssertNoMsg asserts that no log entry in the collection has the field
// [Config.MessageField] with the specified value and type. Returns true if
// none matches. If any entry has the field with the value and type, it marks
// the test as failed, logs an error message, and returns false.
func (ets Entries) AssertNoMsg(want string) bool {
	ets.t.Helper()
	return ets.notExp(CheckStr(ets.cfg.MessageField, want))
}

// AssertMsgContain asserts that at least one log entry in the collection has
// the field [Config.MessageField] containing the specified value and type.
// Returns true if found and matches. If no entry has the field with the value
// and type, it marks the test as failed, logs an error message, and returns
// false.
func (ets Entries) AssertMsgContain(want string) bool {
	ets.t.Helper()
	return ets.exp(CheckContain(ets.cfg.MessageField, want))
}

// AssertNoMsgContain asserts that no log entry in the collection has the
// field [Config.MessageField] containing the specified value and of the
// correct type. Returns true if none contains it. If any entry has the field
// containing the value and type, it marks the test as failed, logs an error
// message, and returns false.
func (ets Entries) AssertNoMsgContain(want string) bool {
	ets.t.Helper()
	return ets.notExp(CheckContain(ets.cfg.MessageField, want))
}

// AssertError asserts that at least one log entry in the collection has the
// field [Config.ErrorField] with the specified value and type. Returns true if
// found and matches. If no entry has the field with the value and type, it
// marks the test as failed, logs an error message, and returns false.
func (ets Entries) AssertError(want string) bool {
	ets.t.Helper()
	return ets.exp(CheckStr(ets.cfg.ErrorField, want))
}

// AssertErrorContain asserts that at least one log entry in the collection has
// the field [Config.ErrorField] containing the specified value and type.
// Returns true if found and matches. If no entry has the field with the value
// and type, it marks the test as failed, logs an error message, and returns
// false.
func (ets Entries) AssertErrorContain(want string) bool {
	ets.t.Helper()
	return ets.exp(CheckContain(ets.cfg.ErrorField, want))
}

// AssertNoError tests if the field with [Config.ErrorField] name and
// provided value is not present in any of the log entries and has the
// requested value. If none of the log entries in the collection has the field
// with the provided value, the test is marked as failed, an error message is
// logged, and the method returns false.
func (ets Entries) AssertNoError(want string) bool {
	ets.t.Helper()
	return ets.notExp(CheckStr(ets.cfg.ErrorField, want))
}

// AssertErr asserts that at least one log entry in the collection has the
// field [Config.ErrorField] with a value equal to the specified error message
// and type. Returns true if found and matches. If no entry has the field with
// the value and type, it marks the test as failed, logs an error message, and
// returns false.
func (ets Entries) AssertErr(want error) bool {
	ets.t.Helper()
	return ets.AssertError(want.Error())
}

// AssertNoErr asserts that no log entry in the collection has the field
// [Config.ErrorField] with a value equal to the specified error message and
// type. Returns true if none matches. If any entry has the field with the
// value and type, it marks the test as failed, logs an error message, and
// returns false.
func (ets Entries) AssertNoErr(want error) bool {
	ets.t.Helper()
	return ets.AssertNoError(want.Error())
}

// AssertContain asserts that at least one log entry in the collection has the
// specified field containing the given string value and type. Returns true if
// found and matches. If no entry has the field with the value and type, it
// marks the test as failed, logs an error message, and returns false.
func (ets Entries) AssertContain(field, want string) bool {
	ets.t.Helper()
	return ets.exp(func(e Entry) error {
		return CheckContain(field, want)(e)
	})
}

// AssertStr asserts that at least one log entry in the collection has the
// specified field with the given string value and type. Returns true if found
// and matches. If no entry has the field with the value and type, it marks the
// test as failed, logs an error message, and returns false.
func (ets Entries) AssertStr(field, want string) bool {
	ets.t.Helper()
	return ets.exp(CheckStr(field, want))
}

// AssertNoStr asserts that no log entry in the collection has the specified
// field with the given string value and type. Returns true if none match. If
// any entry has the field with the value and type, it marks the test as
// failed, logs an error message, and returns false.
func (ets Entries) AssertNoStr(field, want string) bool {
	ets.t.Helper()
	return ets.notExp(CheckStr(field, want))
}

// AssertNumber asserts that at least one log entry in the collection has the
// specified field with the given number value and type. Returns true if found
// and matches. If no entry has the field with the value and type, it marks the
// test as failed, logs an error message, and returns false.
func (ets Entries) AssertNumber(field string, want float64) bool {
	ets.t.Helper()
	return ets.exp(func(e Entry) error { return CheckNumber(field, want)(e) })
}

// AssertNoNumber asserts that no log entry in the collection has the specified
// field with the given number value and type. Returns true if none match. If
// any entry has the field with the value and type, it marks the test as
// failed, logs an error message, and returns false.
func (ets Entries) AssertNoNumber(field string, want float64) bool {
	ets.t.Helper()
	return ets.notExp(func(e Entry) error { return CheckNumber(field, want)(e) })
}

// AssertBool asserts that at least one log entry in the collection has the
// specified field with the given boolean value and type. Returns true if found
// and matches. If no entry has the field with the value and type, it marks the
// test as failed, logs an error message, and returns false.
func (ets Entries) AssertBool(field string, want bool) bool {
	ets.t.Helper()
	return ets.exp(func(e Entry) error { return CheckBool(field, want)(e) })
}

// AssertTime asserts that at least one log entry in the collection has the
// specified field with the given time value and type. Returns true if found
// and matches. If no entry has the field with the value and type, it marks the
// test as failed, logs an error message, and returns false.
func (ets Entries) AssertTime(field string, want time.Time) bool {
	ets.t.Helper()
	return ets.exp(func(e Entry) error { return CheckTime(field, want)(e) })
}

// AssertNoTime asserts that no log entry in the collection has the specified
// field with the given time value and type. Returns true if none match. If any
// entry has the field with the value and type, it marks the test as failed,
// logs an error message, and returns false.
func (ets Entries) AssertNoTime(field string, want time.Time) bool {
	ets.t.Helper()
	return ets.notExp(func(e Entry) error { return CheckTime(field, want)(e) })
}

// AssertDuration asserts that at least one log entry in the collection has the
// specified field with the given [time.Duration] value and type. Returns true
// if found and matches. If no entry has the field with the value and type, it
// marks the test as failed, logs an error message, and returns false.
func (ets Entries) AssertDuration(field string, want time.Duration) bool {
	ets.t.Helper()
	return ets.exp(func(e Entry) error { return CheckDuration(field, want)(e) })
}

// AssertNoDuration asserts that no log entry in the collection has the
// specified field with the given [time.Duration] value and type. Returns true
// if none match. If any entry has the field with the value and type, it marks
// the test as failed, logs an error message, and returns false.
func (ets Entries) AssertNoDuration(field string, want time.Duration) bool {
	ets.t.Helper()
	return ets.notExp(func(e Entry) error { return CheckDuration(field, want)(e) })
}

// exp expects the passed function fn to return nil at least once.
//
// It iterates through the log entries and applies the supplied function fn to
// each entry, breaking the loop and exiting with true the first time the
// function returns nil error. If none of the entries passed to the function
// cause it to return nil, the test is marked as failed, an error message is
// logged, and the method returns false.
func (ets Entries) exp(fn Checker) bool {
	ets.t.Helper()
	for idx := range ets.ets {
		if fn(ets.ets[idx]) == nil {
			return true
		}
	}
	ets.t.Error(notice.New("[log entry] no matching log entry found"))
	return false
}

// notExp expects the passed function fn never to return nil error.
//
// It iterates through the log entries and applies the supplied function fn
// to each entry, breaking the loop and exiting with false the first time the
// function returns nil error. Before exiting with false, it marks the test
// failed and logs the error message.
func (ets Entries) notExp(fn Checker) bool {
	ets.t.Helper()
	for idx := range ets.ets {
		if fn(ets.ets[idx]) == nil {
			ets.t.Error(notice.New("[log entry] matching log entry found"))
			return false
		}
	}
	return true
}

// Summary returns all log entries as a formatted string.
func (ets Entries) Summary() string {
	ets.t.Helper()
	return notice.Indent(0, ' ', ets.summary(0))
}

// summary returns a formatted string with all the entries logged so far.
// It takes an integer parameter `indent` that specifies the number of tabs
// to prepend to each entry in the output string. If there are no entries
// logged, it returns the string "no entries logged so far".
func (ets Entries) summary(indent int) string {
	ets.t.Helper()
	if len(ets.ets) == 0 {
		return notice.Indent(indent, ' ', "no entries logged so far")
	}

	sb := strings.Builder{}
	sb.WriteString("entries logged so far:\n")
	sb.WriteString(notice.Indent(indent+2, ' ', ets.print()))
	return sb.String()
}

// print returns a string with all the entries logged so far.
func (ets Entries) print() string {
	ets.t.Helper()
	sb := strings.Builder{}
	for _, e := range ets.ets {
		sb.WriteString(e.raw + "\n")
	}
	return sb.String()
}

// Print prints all log entries to test log.
func (ets Entries) Print() {
	ets.t.Helper()
	ets.t.Log(ets.Summary())
}
