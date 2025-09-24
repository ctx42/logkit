// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"fmt"
	"maps"
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// FieldType represents a log entry field type.
type FieldType string

// Entry field supported types.
const (
	TypString      FieldType = "string"
	TypNumber      FieldType = "number"
	TypBool        FieldType = "bool"
	TypTime        FieldType = "time"
	TypDur         FieldType = "duration"
	TypMap         FieldType = "map"
	TypUnsupported FieldType = "unsupported"
)

// Entry represents a single log entry (line).
type Entry struct {
	cfg *Config        // Log message format configuration.
	raw string         // Log entry as it was written to the writer.
	m   map[string]any // JSON decoded log entry.
	idx int            // Log the message index in the [Entries] collection.
	t   tester.T       // Test manager.
}

// IsZero reports whether the raw string is empty. Returns true if the string
// is empty, and false otherwise.
func (ent Entry) IsZero() bool {
	return ent.raw == ""
}

// Index of the log entry on the list of log all logged messages.
func (ent Entry) Index() int {
	return ent.idx
}

// String returns the log entry as it was written to the writer.
func (ent Entry) String() string {
	return ent.raw
}

// Bytes return the log entry as it was written to the writer.
func (ent Entry) Bytes() []byte {
	return []byte(ent.raw)
}

// MetaAll returns JSON decoded log entry as a map.
func (ent Entry) MetaAll() map[string]any {
	return maps.Clone(ent.m)
}

// AssertRaw asserts if the raw log entry matches the provided string. If the
// log entry is not equal, the test is marked as failed, an error message is
// logged, and the method returns false.
func (ent Entry) AssertRaw(want string) bool {
	ent.t.Helper()
	if err := check.JSON(want, ent.raw); err != nil {
		err = notice.From(err, "log entry")
		ent.t.Error(err)
		return false

	}
	return true
}

// AssertExist asserts log entry has the given field name. If it doesn't, the
// test is marked as failed, an error message is logged, and the method returns
// false.
func (ent Entry) AssertExist(field string) bool {
	ent.t.Helper()
	if _, ok := ent.m[field]; ok {
		return true
	}
	const format = "expected log entry field to be present:\n  field: %s"
	ent.t.Errorf(format, field)
	return false
}

// AssertNotExist asserts that the log entry does not contain a field with the
// specified name. It returns true if the field is absent, otherwise it marks
// the test as failed, logs an error message, and returns false.
func (ent Entry) AssertNotExist(field string) bool {
	ent.t.Helper()
	if _, ok := ent.m[field]; !ok {
		return true
	}
	const format = "expected log entry field not to be present:\n  field: %s"
	ent.t.Errorf(format, field)
	return false
}

// AssertFieldCount asserts if the log entry has exactly the specified number
// of fields. It returns true if the field count matches, otherwise it marks
// the test as failed, logs an error message, and returns false.
func (ent Entry) AssertFieldCount(want int) bool {
	ent.t.Helper()
	have := len(ent.m)
	if have == want {
		return true
	}
	const format = "expected log entry to have N fields:\n" +
		"  want: %d\n" +
		"  have: %d"
	ent.t.Errorf(format, want, have)
	return false
}

// AssertFieldType asserts if the log entry contains a field with the specified
// name and type. It returns true if the field exists and matches the expected
// type, otherwise it marks the test as failed and returns false.
func (ent Entry) AssertFieldType(field string, want FieldType) bool {
	ent.t.Helper()

	if !ent.AssertExist(field) {
		return false
	}
	val := ent.m[field]
	var have FieldType

	switch val.(type) {
	case bool:
		have = TypBool
	case string:
		have = TypString
	case int:
		have = TypNumber
	case float64:
		have = TypNumber
	case time.Time:
		have = TypTime
	case time.Duration:
		have = TypDur
	case map[string]any:
		have = TypMap
	default:
		have = TypUnsupported
	}

	if want == have {
		return true
	}

	const format = "expected log entry field type:\n" +
		"  want: %s\n" +
		"  have: %T"
	ent.t.Errorf(format, want, val)
	return false
}

// Level retrieves the log level from the field named [Config.LevelField].
// Returns the level as a string and nil error if the field is valid.
// If missing, returns an empty string and [ErrMissing]. For invalid type or
// value, returns empty string and [ErrType] or [ErrValue], respectively.
func (ent Entry) Level() (string, error) {
	ent.t.Helper()
	val, err := HasStr(ent, ent.cfg.LevelField)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", fmt.Errorf("%w: %w", ErrValue, err)
	}
	return val, nil
}

// AssertLevel asserts that the log entry's [Config.LevelField] matches the
// requested level. Returns true if the field exists and matches. If the field
// is missing or the value doesn't match, it marks the test as failed, logs an
// error message, and returns false.
func (ent Entry) AssertLevel(want string) bool {
	ent.t.Helper()
	if err := CheckLevel(want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// AssertMsg asserts that the log entry's [Config.MessageField] matches the
// expected value. Returns true if the field exists and matches. If the field
// is missing or the value doesn't match, it marks the test as failed, logs an
// error message, and returns false.
func (ent Entry) AssertMsg(want string) bool {
	ent.t.Helper()
	return ent.AssertStr(ent.cfg.MessageField, want)
}

// AssertMsgErr asserts that the log entry's [Config.MessageField] matches the
// expected error message. Returns true if the field exists and matches. If the
// field is missing or the value doesn't match, it marks the test as failed,
// logs an error message, and returns false.
func (ent Entry) AssertMsgErr(want error) bool {
	ent.t.Helper()
	return ent.AssertMsg(want.Error())
}

// AssertError asserts that the log entry's [Config.ErrorField] matches the
// expected value. Returns true if the field exists and matches. If the field
// is missing or the value doesn't match, it marks the test as failed, logs an
// error message, and returns false.
func (ent Entry) AssertError(want string) bool {
	ent.t.Helper()
	return ent.AssertStr(ent.cfg.ErrorField, want)
}

// AssertErr asserts that the log entry's [Config.ErrorField] matches the
// expected error message. Returns true if the field exists and matches. If the
// field is missing or the value doesn't match, it marks the test as failed,
// logs an error message, and returns false.
func (ent Entry) AssertErr(want error) bool {
	ent.t.Helper()
	return ent.AssertError(want.Error())
}

// Str retrieves the string value of a field in the log entry. Returns the
// string and nil error if the field exists and is a string. If the field is
// missing or not a string, it returns an empty string and [ErrMissing] or
// [ErrType], respectively.
func (ent Entry) Str(field string) (string, error) {
	ent.t.Helper()
	return HasStr(ent, field)
}

// AssertStr asserts that the log entry's string field matches the expected
// value. Returns true if the field exists and matches. If the field is missing
// or the value doesn't match, it marks the test as failed, logs an error
// message, and returns false.
func (ent Entry) AssertStr(field, want string) bool {
	ent.t.Helper()
	if err := CheckStr(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// AssertContain asserts that the log entry's string field contains the
// expected message. Returns true if the field exists and contains the message.
// If the field is missing or the value doesn't contain it, marks the test as
// failed, logs an error message, and returns false.
func (ent Entry) AssertContain(field, want string) bool {
	ent.t.Helper()
	if err := CheckContain(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// Number retrieves the float64 value of a field in the log entry. Returns the
// value and nil error if the field exists and is a float64. If the field is
// missing or not a float64, returns 0 and [ErrMissing] or [ErrType],
// respectively.
func (ent Entry) Number(field string) (float64, error) {
	ent.t.Helper()
	return HasNum(ent, field)
}

// AssertNumber asserts that the log entry's number field matches the expected
// value. Returns true if the field exists and matches. If the field is missing
// or the value doesn't match, it marks the test as failed, logs an error
// message, and returns false.
func (ent Entry) AssertNumber(field string, want float64) bool {
	ent.t.Helper()
	if err := CheckNumber(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// Bool retrieves the boolean value of a field in the log entry. Returns the
// value and nil error if the field exists and is a boolean. If the field is
// missing or not a boolean, it returns false and [ErrMissing] or [ErrType],
// respectively.
func (ent Entry) Bool(field string) (bool, error) {
	ent.t.Helper()
	return HasBool(ent, field)
}

// AssertBool asserts that the log entry's boolean field matches the expected
// value. Returns true if the field exists and matches. If the field is missing
// or the value doesn't match, it marks the test as failed, logs an error
// message, and returns false.
func (ent Entry) AssertBool(field string, want bool) bool {
	ent.t.Helper()
	if err := CheckBool(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// Time retrieves the time value of a field in the log entry, parsed using
// [Config.TimeFormat]. Returns the time and nil error if the field exists and
// is valid. If the field is missing, has an invalid type, or invalid format,
// returns a zero time value and [ErrMissing], [ErrType], or [ErrFormat],
// respectively.
func (ent Entry) Time(field string) (time.Time, error) {
	ent.t.Helper()
	return HasTime(ent, field)
}

// AssertTime asserts that the log entry's time field matches the expected
// value. Returns true if the field exists and matches. If the field is missing
// or the value doesn't match, it marks the test as failed, logs an error
// message, and returns false.
func (ent Entry) AssertTime(key string, want time.Time) bool {
	ent.t.Helper()
	if err := CheckTime(key, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// AssertWithin asserts that the log entry's time field is within the given
// duration from the expected value. Returns true if the field exists and is
// within the duration. If the field is missing or not within the duration, it
// marks the test as failed, logs an error message, and returns false.
func (ent Entry) AssertWithin(field string, want time.Time, diff string) bool {
	ent.t.Helper()
	have, err := HasTime(ent, field)
	if err != nil {
		ent.t.Error(err)
		return false
	}
	if err = check.Within(want, diff, have); err != nil {
		err = notice.From(err, "log entry").
			Prepend("field", "%s", field)
		ent.t.Error(err)
		return false
	}
	return true
}

// AssertLoggedWithin asserts that the log entry's timestamp field is within
// the specified duration from the given time, using [Entry.AssertWithin].
// Returns true if the field exists and is within the duration. If the field is
// missing or not within the duration, it marks the test as failed, logs an
// error message, and returns false.
func (ent Entry) AssertLoggedWithin(want time.Time, diff string) bool {
	ent.t.Helper()
	return ent.AssertWithin(ent.cfg.TimeField, want, diff)
}

// Duration retrieves the [time.Duration] value of a field in the log entry.
// Returns the duration and nil error if the field exists and is an integer.
// If the field is missing or not an integer, returns 0 and [ErrMissing] or
// [ErrType], respectively.
func (ent Entry) Duration(field string) (time.Duration, error) {
	ent.t.Helper()
	return HasDur(ent, field)
}

// AssertDuration asserts that the log entry's field has the expected
// [time.Duration] value. Returns true if the field exists and matches. If the
// field is missing or not a [time.Duration], it marks the test as failed, logs
// an error message, and returns false.
func (ent Entry) AssertDuration(field string, want time.Duration) bool {
	ent.t.Helper()
	if err := CheckDuration(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}

// Map retrieves the log entry key as a map[string]any. Returns the map and
// nil error if the field exists and is valid. If the field is missing or not a
// map, returns nil and [ErrMissing] or [ErrType], respectively.
func (ent Entry) Map(field string) (map[string]any, error) {
	ent.t.Helper()
	return HasMap(ent, field)
}

// AssertMap asserts that the log entry's map field matches the provided "want"
// map. Returns true if the field exists and matches. If the field is missing
// or the value doesn't match, it marks the test as failed, logs an error
// message, and returns false.
func (ent Entry) AssertMap(field string, want map[string]any) bool {
	ent.t.Helper()
	if err := CheckMap(field, want)(ent); err != nil {
		ent.t.Error(err)
		return false
	}
	return true
}
