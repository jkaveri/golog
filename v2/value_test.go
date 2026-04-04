package golog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKind_String(t *testing.T) {
	require.Equal(t, "String", KindString.String())
	require.Equal(t, "<unknown Kind>", Kind(999).String())
}

func TestValue_zeroIsKindAnyNil(t *testing.T) {
	var v Value
	require.Equal(t, KindAny, v.Kind())
	require.Nil(t, v.Any())
}

func TestAnyValue_dispatch(t *testing.T) {
	type Args struct {
		in any
	}
	type Expects struct {
		want Kind
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "string-kind", args: Args{in: "s"}, expects: Expects{want: KindString}},
		{name: "int-kind", args: Args{in: 42}, expects: Expects{want: KindInt64}},
		{name: "uint-kind", args: Args{in: uint(3)}, expects: Expects{want: KindUint64}},
		{name: "int64-negative", args: Args{in: int64(-1)}, expects: Expects{want: KindInt64}},
		{name: "int8-kind", args: Args{in: int8(9)}, expects: Expects{want: KindInt64}},
		{name: "int16-kind", args: Args{in: int16(10)}, expects: Expects{want: KindInt64}},
		{name: "int32-kind", args: Args{in: int32(11)}, expects: Expects{want: KindInt64}},
		{name: "bool-kind", args: Args{in: true}, expects: Expects{want: KindBool}},
		{name: "duration-kind", args: Args{in: time.Duration(1)}, expects: Expects{want: KindDuration}},
		{name: "time-kind", args: Args{in: time.Unix(1, 0).UTC()}, expects: Expects{want: KindTime}},
		{name: "float64-kind", args: Args{in: float64(1.5)}, expects: Expects{want: KindFloat64}},
		{name: "float32-kind", args: Args{in: float32(2)}, expects: Expects{want: KindFloat64}},
		{name: "uint8-kind", args: Args{in: uint8(4)}, expects: Expects{want: KindUint64}},
		{name: "uint16-kind", args: Args{in: uint16(5)}, expects: Expects{want: KindUint64}},
		{name: "uint32-kind", args: Args{in: uint32(6)}, expects: Expects{want: KindUint64}},
		{name: "uintptr-kind", args: Args{in: uintptr(7)}, expects: Expects{want: KindUint64}},
		{name: "group-from-attrs", args: Args{in: []Attr{String("k", "v")}}, expects: Expects{want: KindGroup}},
		{name: "int64-value-wrapper", args: Args{in: Int64Value(7)}, expects: Expects{want: KindInt64}},
		{name: "struct-fallback-any", args: Args{in: struct{ x int }{1}}, expects: Expects{want: KindAny}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := AnyValue(tc.args.in)
			require.Equal(t, tc.expects.want, v.Kind(), "AnyValue(%T)", tc.args.in)
		})
	}
}

func TestIntValue(t *testing.T) {
	v := IntValue(42)
	require.Equal(t, KindInt64, v.Kind())
	require.Equal(t, int64(42), v.Int64())
}

func TestGroupValue_dropsEmptyChildGroups(t *testing.T) {
	empty := Group("inner")
	nonEmpty := Group("inner", String("k", "v"))
	v := GroupValue(empty, nonEmpty)
	require.Equal(t, KindGroup, v.Kind())
	attrs := v.Group()
	require.Len(t, attrs, 1)
}

func TestValue_String_stringKindNoAllocPath(t *testing.T) {
	v := StringValue("plain")
	require.Equal(t, "plain", v.String())
}

func TestValue_String_nonStringUsesTextFormat(t *testing.T) {
	v := Int64Value(42)
	require.Contains(t, v.String(), "42")
}

func TestValue_Any(t *testing.T) {
	require.Equal(t, "x", AnyValue("x").Any())
	require.Equal(t, int64(3), AnyValue(int64(3)).Any())
	g := GroupValue(String("a", "b"))
	require.Equal(t, KindGroup, AnyValue([]Attr{String("a", "b")}).Kind())
	require.Len(t, g.Group(), 1)
}

func TestValue_Any_accessorByKind(t *testing.T) {
	ts := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	d := 4 * time.Millisecond

	type Args struct {
		v Value
	}
	type Expects struct {
		wantKind Kind
		want     any
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name:    "uint64",
			args:    Args{v: Uint64Value(9)},
			expects: Expects{wantKind: KindUint64, want: uint64(9)},
		},
		{
			name:    "float64",
			args:    Args{v: Float64Value(2.5)},
			expects: Expects{wantKind: KindFloat64, want: 2.5},
		},
		{
			name:    "bool",
			args:    Args{v: BoolValue(true)},
			expects: Expects{wantKind: KindBool, want: true},
		},
		{
			name:    "duration",
			args:    Args{v: DurationValue(d)},
			expects: Expects{wantKind: KindDuration, want: d},
		},
		{
			name:    "time",
			args:    Args{v: TimeValue(ts)},
			expects: Expects{wantKind: KindTime, want: ts},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expects.wantKind, tc.args.v.Kind())
			got := tc.args.v.Any()
			switch tc.expects.wantKind {
			case KindTime:
				require.True(t, got.(time.Time).Equal(tc.expects.want.(time.Time)))
			default:
				require.Equal(t, tc.expects.want, got)
			}
		})
	}
}

func TestValue_Equal(t *testing.T) {
	require.True(t, Int64Value(1).Equal(Int64Value(1)))
	require.False(t, Int64Value(1).Equal(Int64Value(2)))
	require.False(t, Int64Value(1).Equal(StringValue("1")))
	ts := time.Unix(100, 0).UTC()
	require.True(t, TimeValue(ts).Equal(TimeValue(ts)))
	require.True(t, GroupValue(String("k", "v")).Equal(GroupValue(String("k", "v"))))
}

func TestValue_Equal_moreKinds(t *testing.T) {
	type Args struct {
		a Value
		b Value
	}
	type Expects struct {
		wantEqual bool
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{
			name: "uint64-same",
			args: Args{a: Uint64Value(3), b: Uint64Value(3)},
			expects: Expects{wantEqual: true},
		},
		{
			name: "uint64-diff",
			args: Args{a: Uint64Value(3), b: Uint64Value(4)},
			expects: Expects{wantEqual: false},
		},
		{
			name: "float64-same",
			args: Args{a: Float64Value(1.5), b: Float64Value(1.5)},
			expects: Expects{wantEqual: true},
		},
		{
			name: "float64-diff",
			args: Args{a: Float64Value(1.5), b: Float64Value(2.0)},
			expects: Expects{wantEqual: false},
		},
		{
			name: "bool-same",
			args: Args{a: BoolValue(true), b: BoolValue(true)},
			expects: Expects{wantEqual: true},
		},
		{
			name: "bool-diff",
			args: Args{a: BoolValue(true), b: BoolValue(false)},
			expects: Expects{wantEqual: false},
		},
		{
			name: "duration-same",
			args: Args{a: DurationValue(time.Second), b: DurationValue(time.Second)},
			expects: Expects{wantEqual: true},
		},
		{
			name: "duration-diff",
			args: Args{a: DurationValue(time.Second), b: DurationValue(2 * time.Second)},
			expects: Expects{wantEqual: false},
		},
		{
			name: "string-same",
			args: Args{a: StringValue("a"), b: StringValue("a")},
			expects: Expects{wantEqual: true},
		},
		{
			name: "string-diff",
			args: Args{a: StringValue("a"), b: StringValue("b")},
			expects: Expects{wantEqual: false},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expects.wantEqual, tc.args.a.Equal(tc.args.b))
		})
	}
}

func TestValue_Equal_kindAnyComparable(t *testing.T) {
	a := AnyValue(1)
	b := AnyValue(1)
	require.True(t, a.Equal(b))
}

func TestValue_GroupPanicsWrongKind(t *testing.T) {
	defer func() {
		require.NotNil(t, recover(), "expected panic")
	}()
	_ = StringValue("x").Group()
}

func TestValue_accessorPanicsWrongKind(t *testing.T) {
	type Args struct {
		call func()
	}

	testCases := []struct {
		name string
		args Args
	}{
		{name: "int64-on-string", args: Args{call: func() { StringValue("").Int64() }}},
		{name: "uint64-on-int64", args: Args{call: func() { Int64Value(1).Uint64() }}},
		{name: "bool-on-int64", args: Args{call: func() { Int64Value(1).Bool() }}},
		{name: "duration-on-int64", args: Args{call: func() { Int64Value(1).Duration() }}},
		{name: "float64-on-int64", args: Args{call: func() { Int64Value(1).Float64() }}},
		{name: "time-on-int64", args: Args{call: func() { Int64Value(1).Time() }}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				require.NotNil(t, recover(), "expected panic")
			}()
			tc.args.call()
		})
	}
}

func TestTimeValue_equalSelf(t *testing.T) {
	ts := time.Date(2026, 3, 28, 12, 0, 0, 123456789, time.UTC)
	v := TimeValue(ts)
	require.Equal(t, KindTime, v.Kind())
	require.True(t, v.Equal(TimeValue(ts)))
}
