// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package logkit

import (
	"errors"
	"strconv"
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
)

// Log entry field checking errors.
var (
	// ErrMissing represents an error for missing log entry field.
	ErrMissing = errors.New("missing log entry field")

	// ErrType represents an error for invalid log entry field type.
	ErrType = errors.New("invalid log entry field type")

	// ErrFormat represents an error for invalid log entry field format.
	ErrFormat = errors.New("invalid log entry field format")

	// ErrValue represents an error for invalid log entry field value.
	ErrValue = errors.New("invalid log entry field value")
)

// CheckBool returns a function that takes an [Entry] and checks if the
// specified field exists with a boolean value equal to the given value.
// Returns nil if the field exists, is a boolean, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a
// boolean, or does not match, respectively.
func CheckBool(field string, want bool) Checker {
	return func(ent Entry) error {
		have, err := HasBool(ent, field)
		if err != nil {
			return err
		}
		if err = check.Equal(want, have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckStr returns a function that takes an [Entry] and checks if the
// specified field exists with a string value equal to the given value.
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckStr(field, want string) Checker {
	return func(ent Entry) error {
		have, err := HasStr(ent, field)
		if err != nil {
			return err
		}
		if err = check.Equal(want, have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckStrErr returns a function that takes an [Entry] and checks if the
// specified field exists with a string value equal to the given error message.
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckStrErr(field string, want error) Checker {
	return func(ent Entry) error {
		have, err := HasStr(ent, field)
		if err != nil {
			return err
		}
		if err = check.Equal(want.Error(), have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckContain returns a function that takes an [Entry] and checks if the
// specified field exists with a string value containing the given value.
// Returns nil if the field exists, is a string, and contains the value.
// Returns [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not
// a string, or does not contain the value, respectively.
func CheckContain(field, want string) Checker {
	return func(ent Entry) error {
		have, err := HasStr(ent, field)
		if err != nil {
			return err
		}
		if err = check.Contain(want, have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Wrap(ErrValue)

		}
		return nil
	}
}

// CheckMsg returns a function that takes an [Entry] and checks if the
// [Config.MessageField] field exists with a string value equal to the given
// value. Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckMsg(want string) Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.MessageField, want)(ent)
	}
}

// CheckMsgContain returns a function that takes an [Entry] and checks if the
// [Config.MessageField] field exists with a string value containing the given
// value. Returns nil if the field exists, is a string, and contains the value.
// Returns [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not
// a string, or does not contain the value, respectively.
func CheckMsgContain(want string) Checker {
	return func(ent Entry) error {
		return CheckContain(ent.cfg.MessageField, want)(ent)
	}
}

// CheckErrContain returns a function that takes an [Entry] and checks if the
// [Config.ErrorField] field exists with a string value containing the given
// value. Returns nil if the field exists, is a string, and contains the value.
// Returns [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not
// a string, or does not contain the value, respectively.
func CheckErrContain(want string) Checker {
	return func(ent Entry) error {
		return CheckContain(ent.cfg.ErrorField, want)(ent)
	}
}

// CheckTime returns a function that takes an [Entry] and checks if the
// specified field exists with a time value, parsed using [Config.TimeFormat],
// equal to the given time. Returns nil if the field exists, is a valid time,
// and matches. Returns [ErrMissing], [ErrType], or [ErrValue] if the field is
// missing, not a valid time, or does not match, respectively.
func CheckTime(field string, want time.Time) Checker {
	return func(ent Entry) error {
		have, err := HasTime(ent, field)
		if err != nil {
			return err
		}
		if err = check.Time(want, have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckDuration returns a function that takes an [Entry] and checks if the
// specified field exists with an integer value, in [DurationFieldUnit], equal
// to the given duration. Returns nil if the field exists, is an integer, and
// matches. Returns [ErrMissing], [ErrType], or [ErrValue] if the field is
// missing, not an integer, or does not match, respectively.
func CheckDuration(field string, want time.Duration) Checker {
	return func(ent Entry) error {
		have, err := HasDur(ent, field)
		if err != nil {
			return err
		}
		if err = check.Duration(want, have); err != nil {
			return notice.From(err, "log entry").
				Prepend("field", "%s", field).
				Want("%d (%s)", want/ent.cfg.DurationUnit, want.String()).
				Have("%d (%s)", have/ent.cfg.DurationUnit, have.String()).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckLevel returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field exists with a string value equal to the given
// value. Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckLevel(want string) Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, want)(ent)
	}
}

// CheckDebug returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelDebugValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckDebug() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelDebugValue)(ent)
	}
}

// CheckInfo returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelInfoValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckInfo() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelInfoValue)(ent)
	}
}

// CheckWarn returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelWarnValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckWarn() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelWarnValue)(ent)
	}
}

// CheckError returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelErrorValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckError() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelErrorValue)(ent)
	}
}

// CheckFatal returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelFatalValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckFatal() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelFatalValue)(ent)
	}
}

// CheckPanic returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelPanicValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckPanic() Checker {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelPanicValue)(ent)
	}
}

// CheckTrace returns a function that takes an [Entry] and checks if the
// [Config.LevelField] field is a string equal to [Config.LevelTraceValue].
// Returns nil if the field exists, is a string, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a string,
// or does not match, respectively.
func CheckTrace() func(ent Entry) error {
	return func(ent Entry) error {
		return CheckStr(ent.cfg.LevelField, ent.cfg.LevelTraceValue)(ent)
	}
}

// CheckNumber returns a function that takes an [Entry] and checks if the
// specified field exists with a number value equal to the given value. Returns
// nil if the field exists, is a number, and matches. Returns [ErrMissing],
// [ErrType], or [ErrValue] if the field is missing, not a number, or does not
// match, respectively.
func CheckNumber(field string, want float64) Checker {
	return func(ent Entry) error {
		have, err := HasNum(ent, field)
		if err != nil {
			return err
		}
		if err = check.Equal(want, have); err != nil {
			wantStr := strconv.FormatFloat(want, 'f', -1, 64)
			haveStr := strconv.FormatFloat(have, 'f', -1, 64)
			return notice.New("error checking log entry").
				Prepend("field", "%s", field).
				Want("%s", wantStr).
				Have("%s", haveStr).
				Wrap(ErrValue)
		}
		return nil
	}
}

// CheckMap returns a function that takes an [Entry] and checks if the
// specified field exists with a map[string]any value deeply equal to the given
// value. Returns nil if the field exists, is a map, and matches. Returns
// [ErrMissing], [ErrType], or [ErrValue] if the field is missing, not a map,
// or does not match, respectively.
func CheckMap(field string, want map[string]any) Checker {
	return func(ent Entry) error {
		have, err := HasMap(ent, field)
		if err != nil {
			return err
		}
		if err = check.Equal(want, have); err != nil {
			return notice.From(err, "log entry").Wrap(ErrValue)
		}
		return nil
	}
}
