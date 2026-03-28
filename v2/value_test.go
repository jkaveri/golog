package golog

import (
	"strings"
	"testing"
	"time"
)

func TestKind_String(t *testing.T) {
	if got := KindString.String(); got != "String" {
		t.Fatalf("KindString: got %q", got)
	}
	if got := Kind(999).String(); got != "<unknown Kind>" {
		t.Fatalf("unknown kind: got %q", got)
	}
}

func TestValue_zeroIsKindAnyNil(t *testing.T) {
	var v Value
	if v.Kind() != KindAny {
		t.Fatalf("zero Value kind: got %v want KindAny", v.Kind())
	}
	if v.Any() != nil {
		t.Fatalf("zero Value Any: got %#v", v.Any())
	}
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
		{name: "bool-kind", args: Args{in: true}, expects: Expects{want: KindBool}},
		{name: "duration-kind", args: Args{in: time.Duration(1)}, expects: Expects{want: KindDuration}},
		{name: "time-kind", args: Args{in: time.Unix(1, 0).UTC()}, expects: Expects{want: KindTime}},
		{name: "float64-kind", args: Args{in: float64(1.5)}, expects: Expects{want: KindFloat64}},
		{name: "float32-kind", args: Args{in: float32(2)}, expects: Expects{want: KindFloat64}},
		{name: "group-from-attrs", args: Args{in: []Attr{String("k", "v")}}, expects: Expects{want: KindGroup}},
		{name: "int64-value-wrapper", args: Args{in: Int64Value(7)}, expects: Expects{want: KindInt64}},
		{name: "struct-fallback-any", args: Args{in: struct{ x int }{1}}, expects: Expects{want: KindAny}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := AnyValue(tc.args.in)
			if v.Kind() != tc.expects.want {
				t.Fatalf("AnyValue(%T %v): kind got %v want %v", tc.args.in, tc.args.in, v.Kind(), tc.expects.want)
			}
		})
	}
}

func TestGroupValue_dropsEmptyChildGroups(t *testing.T) {
	empty := Group("inner")
	nonEmpty := Group("inner", String("k", "v"))
	v := GroupValue(empty, nonEmpty)
	if v.Kind() != KindGroup {
		t.Fatal("expected KindGroup")
	}
	attrs := v.Group()
	if len(attrs) != 1 {
		t.Fatalf("want 1 attr after dropping empty, got %d", len(attrs))
	}
}

func TestValue_String_stringKindNoAllocPath(t *testing.T) {
	v := StringValue("plain")
	if v.String() != "plain" {
		t.Fatalf("want literal string, got %q", v.String())
	}
}

func TestValue_String_nonStringUsesTextFormat(t *testing.T) {
	v := Int64Value(42)
	if !strings.Contains(v.String(), "42") {
		t.Fatalf("want formatted int, got %q", v.String())
	}
}

func TestValue_Any(t *testing.T) {
	if AnyValue("x").Any() != "x" {
		t.Fatal("string Any")
	}
	if AnyValue(int64(3)).Any() != int64(3) {
		t.Fatal("int64 Any")
	}
	g := GroupValue(String("a", "b"))
	if AnyValue([]Attr{String("a", "b")}).Kind() != KindGroup {
		t.Fatal("group Any")
	}
	if len(g.Group()) != 1 {
		t.Fatal("group length")
	}
}

func TestValue_Equal(t *testing.T) {
	if !Int64Value(1).Equal(Int64Value(1)) {
		t.Fatal("int64 equal")
	}
	if Int64Value(1).Equal(Int64Value(2)) {
		t.Fatal("int64 not equal")
	}
	if Int64Value(1).Equal(StringValue("1")) {
		t.Fatal("kind mismatch")
	}
	ts := time.Unix(100, 0).UTC()
	if !TimeValue(ts).Equal(TimeValue(ts)) {
		t.Fatal("time equal")
	}
	if !GroupValue(String("k", "v")).Equal(GroupValue(String("k", "v"))) {
		t.Fatal("group equal")
	}
}

func TestValue_Equal_kindAnyComparable(t *testing.T) {
	a := AnyValue(1)
	b := AnyValue(1)
	if !a.Equal(b) {
		t.Fatal("KindAny int equal")
	}
}

func TestValue_GroupPanicsWrongKind(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
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
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tc.args.call()
		})
	}
}

func TestTimeValue_equalSelf(t *testing.T) {
	ts := time.Date(2026, 3, 28, 12, 0, 0, 123456789, time.UTC)
	v := TimeValue(ts)
	if v.Kind() != KindTime {
		t.Fatalf("kind: got %v", v.Kind())
	}
	if !v.Equal(TimeValue(ts)) {
		t.Fatal("TimeValue Equal self")
	}
}
