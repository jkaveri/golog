package golog

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContextEnricher_typedGetters(t *testing.T) {
	type sKey struct{}
	type iKey struct{}
	type uKey struct{}
	type bKey struct{}
	type dKey struct{}
	type tKey struct{}

	now := time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC)
	ctx := context.Background()
	ctx = context.WithValue(ctx, sKey{}, "req-1")
	ctx = context.WithValue(ctx, iKey{}, int32(7))
	ctx = context.WithValue(ctx, uKey{}, uint16(9))
	ctx = context.WithValue(ctx, bKey{}, true)
	ctx = context.WithValue(ctx, dKey{}, 2*time.Second)
	ctx = context.WithValue(ctx, tKey{}, now)

	type Args struct {
		getter ContextValueGetter
	}
	type Expects struct {
		wantKey string
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "string-from-context", args: Args{getter: FromContext.String(sKey{}, "request_id")}, expects: Expects{wantKey: "request_id"}},
		{name: "int64-from-context", args: Args{getter: FromContext.Int64(iKey{}, "attempt")}, expects: Expects{wantKey: "attempt"}},
		{name: "uint64-from-context", args: Args{getter: FromContext.Uint64(uKey{}, "shard")}, expects: Expects{wantKey: "shard"}},
		{name: "bool-from-context", args: Args{getter: FromContext.Bool(bKey{}, "ok")}, expects: Expects{wantKey: "ok"}},
		{name: "duration-from-context", args: Args{getter: FromContext.Duration(dKey{}, "latency")}, expects: Expects{wantKey: "latency"}},
		{name: "time-from-context", args: Args{getter: FromContext.Time(tKey{}, "when")}, expects: Expects{wantKey: "when"}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attr, ok := tc.args.getter(ctx)
			require.True(t, ok, "getter failed")
			require.Equal(t, tc.expects.wantKey, attr.Key)
		})
	}
}

func TestContextEnricher_mismatchAndMissing(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "not-int")

	_, ok := FromContext.Int64(key{}, "n")(ctx)
	require.False(t, ok, "want mismatch to skip")
	_, ok = FromContext.String(struct{}{}, "s")(ctx)
	require.False(t, ok, "want missing key to skip")
	_, ok = FromContext.Any(key{}, "")(ctx)
	require.False(t, ok, "want empty log key to skip")
}

func TestContextEnricher_appendsInOrder(t *testing.T) {
	type aKey struct{}
	type bKey struct{}

	ctx := context.Background()
	ctx = context.WithValue(ctx, aKey{}, "A")
	ctx = context.WithValue(ctx, bKey{}, "B")

	e := NewContextEnricher(
		FromContext.String(aKey{}, "a"),
		FromContext.String(bKey{}, "b"),
	)

	var b RecordBuilder
	e.Enrich(ctx, &b)
	record := b.Build()
	require.Equal(t, 2, record.NumAttrs())
	require.Equal(t, "a", record.Attr(0).Key)
	require.Equal(t, "b", record.Attr(1).Key)
}

func TestContextEnricher_FromContextAny(t *testing.T) {
	type k struct{}

	type Args struct {
		ctx     context.Context
		getter  ContextValueGetter
		wantOK  bool
		wantKey string
	}

	ctxWith := context.WithValue(context.Background(), k{}, map[string]int{"n": 1})
	ctxNil := context.Background()

	testCases := []struct {
		name string
		args Args
	}{
		{
			name: "value-present",
			args: Args{
				ctx:     ctxWith,
				getter:  FromContext.Any(k{}, "payload"),
				wantOK:  true,
				wantKey: "payload",
			},
		},
		{
			name: "value-absent",
			args: Args{
				ctx:    context.WithValue(context.Background(), k{}, nil),
				getter: FromContext.Any(k{}, "x"),
				wantOK: false,
			},
		},
		{
			name: "missing-context-key",
			args: Args{
				ctx:    ctxNil,
				getter: FromContext.Any(k{}, "x"),
				wantOK: false,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attr, ok := tc.args.getter(tc.args.ctx)
			require.Equal(t, tc.args.wantOK, ok)
			if tc.args.wantOK {
				require.Equal(t, tc.args.wantKey, attr.Key)
			}
		})
	}
}

func TestContextEnricher_Int64_integerWidths(t *testing.T) {
	type ik struct{}

	type Args struct {
		value any
	}
	type Expects struct {
		want int64
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "int", args: Args{value: int(11)}, expects: Expects{want: 11}},
		{name: "int8", args: Args{value: int8(12)}, expects: Expects{want: 12}},
		{name: "int16", args: Args{value: int16(13)}, expects: Expects{want: 13}},
		{name: "int32", args: Args{value: int32(14)}, expects: Expects{want: 14}},
		{name: "int64", args: Args{value: int64(15)}, expects: Expects{want: 15}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), ik{}, tc.args.value)
			attr, ok := FromContext.Int64(ik{}, "v")(ctx)
			require.True(t, ok, "getter returned false")
			require.Equal(t, tc.expects.want, attr.Value.Int64())
		})
	}
}

func TestContextEnricher_Uint64_integerWidths(t *testing.T) {
	type uk struct{}

	type Args struct {
		value any
	}
	type Expects struct {
		want uint64
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "uint", args: Args{value: uint(21)}, expects: Expects{want: 21}},
		{name: "uint8", args: Args{value: uint8(22)}, expects: Expects{want: 22}},
		{name: "uint16", args: Args{value: uint16(23)}, expects: Expects{want: 23}},
		{name: "uint32", args: Args{value: uint32(24)}, expects: Expects{want: 24}},
		{name: "uint64", args: Args{value: uint64(25)}, expects: Expects{want: 25}},
		{name: "uintptr", args: Args{value: uintptr(26)}, expects: Expects{want: 26}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), uk{}, tc.args.value)
			attr, ok := FromContext.Uint64(uk{}, "u")(ctx)
			require.True(t, ok, "getter returned false")
			require.Equal(t, tc.expects.want, attr.Value.Uint64())
		})
	}
}

func TestContextEnricher_NewContextEnricher_skipsNilGetter(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, "v")

	e := NewContextEnricher(
		nil,
		FromContext.String(k{}, "id"),
	)

	var b RecordBuilder
	e.Enrich(ctx, &b)
	record := b.Build()
	require.Equal(t, 1, record.NumAttrs())
	require.Equal(t, "id", record.Attr(0).Key)
}

func TestContextEnricher_emptyLogKey_skipped(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, "x")

	type Args struct {
		getter ContextValueGetter
	}
	type Expects struct {
		wantOK bool
	}

	testCases := []struct {
		name    string
		args    Args
		expects Expects
	}{
		{name: "string", args: Args{getter: FromContext.String(k{}, "")}, expects: Expects{wantOK: false}},
		{name: "bool", args: Args{getter: FromContext.Bool(k{}, "")}, expects: Expects{wantOK: false}},
		{name: "duration", args: Args{getter: FromContext.Duration(k{}, "")}, expects: Expects{wantOK: false}},
		{name: "time", args: Args{getter: FromContext.Time(k{}, "")}, expects: Expects{wantOK: false}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := tc.args.getter(ctx)
			require.Equal(t, tc.expects.wantOK, ok)
		})
	}
}

func TestContextEnricher_String_typeMismatch(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, 404)
	_, ok := FromContext.String(k{}, "s")(ctx)
	require.False(t, ok, "want type mismatch")
}

func TestContextEnricher_Bool_wrongType(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, "not-bool")
	_, ok := FromContext.Bool(k{}, "b")(ctx)
	require.False(t, ok, "want wrong type")
}

func TestContextEnricher_Duration_wrongType(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, "not-dur")
	_, ok := FromContext.Duration(k{}, "d")(ctx)
	require.False(t, ok, "want wrong type")
}

func TestContextEnricher_Time_wrongType(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, "not-time")
	_, ok := FromContext.Time(k{}, "t")(ctx)
	require.False(t, ok, "want wrong type")
}

func TestContextEnricher_Int64_unsupportedType(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, float64(3))
	_, ok := FromContext.Int64(k{}, "n")(ctx)
	require.False(t, ok, "want unsupported type to skip")
}

func TestContextEnricher_Uint64_unsupportedType(t *testing.T) {
	type k struct{}
	ctx := context.WithValue(context.Background(), k{}, float64(3))
	_, ok := FromContext.Uint64(k{}, "n")(ctx)
	require.False(t, ok, "want unsupported type to skip")
}
