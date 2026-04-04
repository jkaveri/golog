// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package golog

import "time"

// Attr is one key-value pair on a log line. Construct with helpers such as
// [String], [Int],
// [Group], and [Any]; inspect via [Attr.Value] and [Value.Kind].
type Attr struct {
	Key   string
	Value Value
}

// Equal reports whether a and b have the same key and [Value.Equal] values.
func (a Attr) Equal(b Attr) bool {
	return a.Key == b.Key && a.Value.Equal(b.Value)
}

// String returns a compact debug form "key=value" using [Value.String] for the
// value.
func (a Attr) String() string {
	return a.Key + "=" + a.Value.String()
}

// String returns an Attr for a string value.
func String(key, value string) Attr {
	return Attr{Key: key, Value: StringValue(value)}
}

// Int64 returns an Attr for an int64.
func Int64(key string, value int64) Attr {
	return Attr{Key: key, Value: Int64Value(value)}
}

// Int converts an int to an int64 and returns
// an Attr with that value.
func Int(key string, value int) Attr {
	return Int64(key, int64(value))
}

// Uint64 returns an Attr for a uint64.
func Uint64(key string, v uint64) Attr {
	return Attr{Key: key, Value: Uint64Value(v)}
}

// Float64 returns an Attr for a floating-point number.
func Float64(key string, v float64) Attr {
	return Attr{Key: key, Value: Float64Value(v)}
}

// Bool returns an Attr for a bool.
func Bool(key string, v bool) Attr {
	return Attr{Key: key, Value: BoolValue(v)}
}

// Time returns an Attr for a [time.Time].
// It discards the monotonic portion.
func Time(key string, v time.Time) Attr {
	return Attr{Key: key, Value: TimeValue(v)}
}

// Duration returns an Attr for a [time.Duration].
func Duration(key string, v time.Duration) Attr {
	return Attr{Key: key, Value: DurationValue(v)}
}

// Group returns an Attr for a Group [Value].
// The first argument is the key; the remaining arguments are child [Attr]s.
//
// Use Group to collect several key-value pairs under a single
// key on a log line, or as the result of LogValue
// in order to log a single value as multiple Attrs.
func Group(key string, args ...Attr) Attr {
	return Attr{Key: key, Value: GroupValue(args...)}
}

// Any returns an Attr for the supplied value.
// See [AnyValue] for how values are treated.
func Any(key string, value any) Attr {
	return Attr{Key: key, Value: AnyValue(value)}
}
