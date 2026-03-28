package golog

import (
	"fmt"
	"strconv"
	"strings"
)

// A Level is the importance or severity of a log event.
// The higher the level, the more important or severe the event.
//
// For practical guidance on when to use each level, see the package
// documentation
// (“Logging levels and philosophy”).
type Level uint8

// Names for common levels.
//
// Level numbers are inherently arbitrary,
// but we picked them to satisfy three constraints.
// Any system can map them to another numbering scheme if it wishes.
//
// First, we wanted the default level to be Info, Since Levels are ints, Info is
// the default value for int, zero.
//
// Second, we wanted to make it easy to use levels to specify logger verbosity.
// Since a larger level means a more severe event, a logger that accepts events
// with smaller (or more negative) level means a more verbose logger. Logger
// verbosity is thus the negation of event severity, and the default verbosity
// of 0 accepts all events at least as severe as INFO.
//
// Third, there is one integer step between [LevelInfo] and [LevelError] (level
// 1) for custom schemes; it stringifies as INFO+1. Levels at or above
// [LevelError] use ERROR+N.
//
// In typical use, prefer [LevelInfo] for normal operation and [LevelDebug] for
// troubleshooting; see package documentation for philosophy. [LevelError]
// records
// elevated severity at this layer.
const (
	// LevelDebug is for information useful when diagnosing problems. Turn it on
	// in development or when investigating issues; it is usually off in
	// production.
	LevelDebug Level = iota

	// LevelInfo is the default: messages that describe what the system is doing
	// under
	// normal operation (requests, state changes, milestones).
	LevelInfo

	// LevelError is for failures and exceptional conditions you are recording
	// at this layer. Prefer returning errors from functions rather than only
	// logging them; see package docs.
	LevelError
)

// String returns a name for the level.
// If the level has a name, then that name
// in uppercase is returned.
// If the level is between named values, then
// an integer is appended to the uppercased name.
// Examples:
//
//	(LevelInfo+1).String() => "INFO+1"
//	LevelError.String() => "ERROR"
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelDebug:
		return "DEBUG"
	default:
		panic(fmt.Sprintf("unknown level: %d", l))
	}
}

// MarshalJSON implements [encoding/json.Marshaler]
// by quoting the output of [Level.String].
func (l Level) MarshalJSON() ([]byte, error) {
	// AppendQuote is sufficient for JSON-encoding all Level strings.
	// They don't contain any runes that would produce invalid JSON
	// when escaped.
	return strconv.AppendQuote(nil, l.String()), nil
}

// UnmarshalJSON implements [encoding/json.Unmarshaler]
// It accepts any string produced by [Level.MarshalJSON],
// ignoring case.
// It also accepts numeric offsets that would result in a different string on
// output. For example, "ERROR-2" unmarshals to [LevelInfo].
func (l *Level) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	return l.parse(s)
}

// AppendText implements [encoding.TextAppender]
// by calling [Level.String].
func (l Level) AppendText(b []byte) ([]byte, error) {
	return append(b, l.String()...), nil
}

// MarshalText implements [encoding.TextMarshaler]
// by calling [Level.AppendText].
func (l Level) MarshalText() ([]byte, error) {
	return l.AppendText(nil)
}

// UnmarshalText implements [encoding.TextUnmarshaler].
// It accepts any string produced by [Level.MarshalText],
// ignoring case.
// It also accepts numeric offsets that would result in a different string on
// output. For example, "ERROR-2" unmarshals to [LevelInfo].
func (l *Level) UnmarshalText(data []byte) error {
	return l.parse(string(data))
}

func (l *Level) parse(s string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("golog: level string %q: %w", s, err)
		}
	}()

	name := strings.ToUpper(s)
	switch name {
	case "DEBUG":
		*l = LevelDebug
	case "INFO":
		*l = LevelInfo
	case "ERROR":
		*l = LevelError
	default:
		return fmt.Errorf("unknown level: %s", s)
	}

	return nil
}
