// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
)

// HasBool checks if the specified boolean field exists in the Entry's map of
// fields. If the field is missing, it returns false, and the error has
// [ErrMissing] in its chain. If the field exists but its value is not of
// type bool, it returns false and error having [ErrType] in its chain.
// Otherwise, it returns the boolean value of the field and a nil error.
func HasBool(ent Entry, field string) (bool, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return false, notice.From(err, "log entry").
			Prepend("type", "%T", true).
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType(true, val); err != nil {
		return false, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	return val.(bool), nil // nolint: forcetypeassert
}

// HasStr checks if the specified string field exists in the Entry's map of
// fields. If the field is missing, it returns an empty string and error having
// [ErrMissing] in its chain. If the field exists but its value is not of
// type string, it returns an empty string and error having [ErrType] in its
// chain. Otherwise, it returns the string value of the field and a nil error.
func HasStr(ent Entry, field string) (string, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return "", notice.From(err, "log entry").
			Prepend("type", "%T", "").
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType("", val); err != nil {
		return "", notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	return val.(string), nil // nolint: forcetypeassert
}

// HasTime checks if the specified string field exists in the Entry's map of
// fields. If the field is missing, it returns zero value time and error having
// [ErrMissing] in its chain. If the field exists but its value is not of
// type string, it returns zero value time and error having [ErrType] in its
// chain. If the field exists but its value is not time formatted according to
// [Config.TimeFormat], it returns zero value time and error having
// [ErrFormat] in its chain. Otherwise, it returns the string value of the
// field and a nil error.
func HasTime(ent Entry, field string) (time.Time, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return time.Time{}, notice.From(err, "log entry").
			Prepend("type", "%T", "").
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType("", val); err != nil {
		return time.Time{}, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	haveStr := val.(string) // nolint: forcetypeassert
	have, err := time.Parse(ent.cfg.TimeFormat, haveStr)
	if err != nil {
		format := "[log entry] expected log entry field to have formatted time"
		return time.Time{}, notice.New(format).
			Append("field", "%s", field).
			Want("%s", ent.cfg.TimeFormat).
			Have("%s", haveStr).
			Wrap(ErrFormat)
	}
	return have, nil
}

// HasDur checks if the specified duration field exists in the Entry's map of
// fields. If the field is missing, it returns 0, and the error has
// [ErrMissing] in its chain. If the field exists but its value is not of
// type float64, it returns 0 and error having [ErrType] in its chain.
// Otherwise, it returns the duration value of the field and a nil error.
func HasDur(ent Entry, field string) (time.Duration, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return 0, notice.From(err, "log entry").
			Prepend("type", "number").
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType(1.1, val); err != nil {
		return 0, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	haveVal := val.(float64) // nolint: forcetypeassert
	have := time.Duration(int(haveVal)) * ent.cfg.DurationUnit
	return have, nil
}

// HasNum checks if the specified number field exists in the Entry's map of
// fields. If the field is missing, it returns 0, and the error has
// [ErrMissing] in its chain. If the field exists but its value is not a
// float64, it returns 0 and error having [ErrType] in its chain.
// Otherwise, it returns the float64 value of the field and a nil error.
func HasNum(ent Entry, field string) (float64, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return 0, notice.From(err, "log entry").
			Prepend("type", "number").
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType(1.1, val); err != nil {
		return 0, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	return val.(float64), nil // nolint: forcetypeassert
}

// HasMap checks if the specified map field exists in the Entry's map of
// fields. If the field is missing, it returns nil, and the error has
// [ErrMissing] in its chain. If the field exists but its value is not of
// type map[string]any, it returns nil and error having [ErrType] in its chain.
// Otherwise, it returns the map value of the field and a nil error.
func HasMap(ent Entry, field string) (map[string]any, error) {
	val, err := check.HasKey(field, ent.m)
	if err != nil {
		return nil, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Remove("key").
			Wrap(ErrMissing)
	}
	if err = check.SameType(map[string]any{}, val); err != nil {
		return nil, notice.From(err, "log entry").
			Prepend("field", "%s", field).
			Wrap(ErrType)
	}
	return val.(map[string]any), nil // nolint: forcetypeassert
}
