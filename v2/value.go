package golog

import (
	"fmt"
	"slices"
	"time"
)

// Value is a tagged union: it pairs a [Kind] with the underlying Go value in
// Value.any. Use the typed helpers (e.g. [StringValue], [Int64Value],
// [AnyValue]) or read the kind with [Value.Kind] and the payload with
// [Value.Any] / the type-specific accessors.
//
// The zero Value is [KindAny] with a nil any payload, which represents “no
// value” / nil.
//
// Do not compare Values with == or !=: equality of the struct does not match
// semantic equality of the logged content, and comparing interface values can
// panic if the
// dynamic types are not comparable. Use [Value.Equal] instead.
//
// The unnamed field exists for parity with [log/slog.Value] and documents that
// intent. It is a zero-sized [0]func(); note that function types are comparable
// in Go, so this
// does not by itself make [Value] incomparable at compile time.
type Value struct {
	_    [0]func() // zero-width marker; see Value doc (discourage ==, use Equal)
	kind Kind
	any  any
}

// Kind classifies the dynamic type stored in a [Value]. Inspect with
// [Value.Kind] before
// calling type-specific accessors ([Value.Int64], [Value.String], etc.).
type Kind int

const (
	KindAny Kind = iota
	KindBool
	KindDuration
	KindFloat64
	KindInt64
	KindString
	KindTime
	KindUint64
	KindGroup
)

var kindStrings = []string{
	"Any",
	"Bool",
	"Duration",
	"Float64",
	"Int64",
	"String",
	"Time",
	"Uint64",
	"Group",
}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(kindStrings) {
		return kindStrings[k]
	}

	return "<unknown Kind>"
}

// Kind returns v's Kind.
func (v Value) Kind() Kind {
	return v.kind
}

//////////////// Constructors

// StringValue returns a new [Value] for a string.
func StringValue(s string) Value {
	return Value{kind: KindString, any: s}
}

// IntValue returns a [Value] for an int.
func IntValue(v int) Value {
	return Int64Value(int64(v))
}

// Int64Value returns a [Value] for an int64.
func Int64Value(v int64) Value {
	return Value{kind: KindInt64, any: v}
}

// Uint64Value returns a [Value] for a uint64.
func Uint64Value(v uint64) Value {
	return Value{kind: KindUint64, any: v}
}

// Float64Value returns a [Value] for a floating-point number.
func Float64Value(v float64) Value {
	return Value{kind: KindFloat64, any: v}
}

// BoolValue returns a [Value] for a bool.
func BoolValue(v bool) Value {
	return Value{kind: KindBool, any: v}
}

// TimeValue returns a [Value] for a [time.Time].
// It discards the monotonic portion.
func TimeValue(v time.Time) Value {
	return Value{kind: KindTime, any: v.Round(0)}
}

// DurationValue returns a [Value] for a [time.Duration].
func DurationValue(v time.Duration) Value {
	return Value{kind: KindDuration, any: v}
}

// GroupValue returns a new [Value] for a list of [Attr]s.
// The caller must not subsequently mutate the argument slice.
func GroupValue(as ...Attr) Value {
	if n := countEmptyGroups(as); n > 0 {
		as2 := make([]Attr, 0, len(as)-n)
		for _, a := range as {
			if !a.Value.isEmptyGroup() {
				as2 = append(as2, a)
			}
		}

		as = as2
	}

	return Value{kind: KindGroup, any: as}
}

func countEmptyGroups(as []Attr) int {
	n := 0

	for _, a := range as {
		if a.Value.isEmptyGroup() {
			n++
		}
	}

	return n
}

// AnyValue returns a [Value] for the supplied value.
func AnyValue(v any) Value {
	switch v := v.(type) {
	case string:
		return StringValue(v)
	case int:
		return Int64Value(int64(v))
	case uint:
		return Uint64Value(uint64(v))
	case int64:
		return Int64Value(v)
	case uint64:
		return Uint64Value(v)
	case bool:
		return BoolValue(v)
	case time.Duration:
		return DurationValue(v)
	case time.Time:
		return TimeValue(v)
	case uint8:
		return Uint64Value(uint64(v))
	case uint16:
		return Uint64Value(uint64(v))
	case uint32:
		return Uint64Value(uint64(v))
	case uintptr:
		return Uint64Value(uint64(v))
	case int8:
		return Int64Value(int64(v))
	case int16:
		return Int64Value(int64(v))
	case int32:
		return Int64Value(int64(v))
	case float64:
		return Float64Value(v)
	case float32:
		return Float64Value(float64(v))
	case []Attr:
		return GroupValue(v...)
	case Value:
		return v
	default:
		return Value{kind: KindAny, any: v}
	}
}

//////////////// Accessors

// Any returns v's value as an any.
func (v Value) Any() any {
	switch v.Kind() {
	case KindAny:
		return v.any
	case KindGroup:
		return v.any
	case KindInt64:
		return v.Int64()
	case KindUint64:
		return v.Uint64()
	case KindFloat64:
		return v.Float64()
	case KindString:
		return v.any
	case KindBool:
		return v.Bool()
	case KindDuration:
		return v.Duration()
	case KindTime:
		return v.Time()
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

// String returns Value's value as a string, formatted like [fmt.Sprint]. Unlike
// the methods Int64, Float64, and so on, which panic if v is of the
// wrong kind, String never panics.
//
// Formatting matches the text representation used for attributes in
// [TextWriter].
func (v Value) String() string {
	if v.Kind() == KindString {
		return v.any.(string)
	}

	return string(formatTextAttr(Attr{Value: v}))
}

func (v Value) str() string {
	return v.any.(string)
}

// Int64 returns v's value as an int64. It panics
// if v is not a signed integer.
func (v Value) Int64() int64 {
	if g, w := v.Kind(), KindInt64; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.any.(int64)
}

// Uint64 returns v's value as a uint64. It panics
// if v is not an unsigned integer.
func (v Value) Uint64() uint64 {
	if g, w := v.Kind(), KindUint64; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.any.(uint64)
}

// Bool returns v's value as a bool. It panics
// if v is not a bool.
func (v Value) Bool() bool {
	if g, w := v.Kind(), KindBool; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.bool()
}

func (v Value) bool() bool {
	return v.any.(bool)
}

// Duration returns v's value as a [time.Duration]. It panics
// if v is not a time.Duration.
func (v Value) Duration() time.Duration {
	if g, w := v.Kind(), KindDuration; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.duration()
}

func (v Value) duration() time.Duration {
	return v.any.(time.Duration)
}

// Float64 returns v's value as a float64. It panics
// if v is not a float64.
func (v Value) Float64() float64 {
	if g, w := v.Kind(), KindFloat64; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.float()
}

func (v Value) float() float64 {
	return v.any.(float64)
}

// Time returns v's value as a [time.Time]. It panics
// if v is not a time.Time.
func (v Value) Time() time.Time {
	if g, w := v.Kind(), KindTime; g != w {
		panic(fmt.Sprintf("Value kind is %s, not %s", g, w))
	}

	return v.time()
}

func (v Value) time() time.Time {
	return v.any.(time.Time)
}

// Group returns v's value as a []Attr.
// It panics if v's [Kind] is not [KindGroup].
func (v Value) Group() []Attr {
	if v.Kind() == KindGroup {
		return v.group()
	}

	panic("Group: bad kind")
}

func (v Value) group() []Attr {
	return v.any.([]Attr)
}

//////////////// Other

// Equal reports whether v and w represent the same Go value.
func (v Value) Equal(w Value) bool {
	k1 := v.Kind()

	k2 := w.Kind()
	if k1 != k2 {
		return false
	}

	switch k1 {
	case KindInt64:
		return v.Int64() == w.Int64()
	case KindUint64:
		return v.Uint64() == w.Uint64()
	case KindBool:
		return v.Bool() == w.Bool()
	case KindDuration:
		return v.Duration() == w.Duration()
	case KindString:
		return v.str() == w.str()
	case KindFloat64:
		return v.float() == w.float()
	case KindTime:
		return v.time().Equal(w.time())
	case KindAny:
		return v.any == w.any // may panic if non-comparable
	case KindGroup:
		return slices.EqualFunc(v.group(), w.group(), Attr.Equal)
	default:
		panic(fmt.Sprintf("bad kind: %s", k1))
	}
}

func (v Value) isEmptyGroup() bool {
	if v.Kind() != KindGroup {
		return false
	}

	return len(v.group()) == 0
}
