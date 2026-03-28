package golog

import (
	"context"
	"testing"
	"time"
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
			if !ok {
				t.Fatal("getter failed")
			}
			if attr.Key != tc.expects.wantKey {
				t.Fatalf("key: got %q want %q", attr.Key, tc.expects.wantKey)
			}
		})
	}
}

func TestContextEnricher_mismatchAndMissing(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "not-int")

	if _, ok := FromContext.Int64(key{}, "n")(ctx); ok {
		t.Fatal("want mismatch to skip")
	}
	if _, ok := FromContext.String(struct{}{}, "s")(ctx); ok {
		t.Fatal("want missing key to skip")
	}
	if _, ok := FromContext.Any(key{}, "")(ctx); ok {
		t.Fatal("want empty log key to skip")
	}
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
	if record.NumAttrs() != 2 {
		t.Fatalf("want 2 attrs, got %d", record.NumAttrs())
	}
	if record.Attr(0).Key != "a" || record.Attr(1).Key != "b" {
		t.Fatalf("attr order mismatch: %q, %q", record.Attr(0).Key, record.Attr(1).Key)
	}
}
