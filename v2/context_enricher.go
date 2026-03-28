// Context enricher: compose typed [context.Context] lookups into a single [Enricher].
//
// Types and helpers here are the context-enricher feature ([ContextEnricher], [ContextValueGetter], [FromContext]).
package golog

import (
	"context"
	"time"
)

// ContextValueGetter extracts one attribute from [context.Context] for a [ContextEnricher].
// Return ok false to omit an attribute (missing value or wrong type).
type ContextValueGetter func(context.Context) (Attr, bool)

// ContextEnricher is an [Enricher] that runs [ContextValueGetter] hooks in order and
// appends each attribute when the getter returns ok=true. Build getters with [FromContext].
type ContextEnricher struct {
	getters []ContextValueGetter
}

// NewContextEnricher builds a [ContextEnricher] from one or more getters. Nil getters are skipped.
// Pass the result to [NewLogger] as a variadic [Enricher]; use [Logger.WithContext] so getters see your keys.
func NewContextEnricher(getters ...ContextValueGetter) *ContextEnricher {
	g := make([]ContextValueGetter, 0, len(getters))
	for _, getter := range getters {
		if getter != nil {
			g = append(g, getter)
		}
	}
	return &ContextEnricher{getters: g}
}

// Enrich runs each getter against ctx and adds attributes to builder.
func (e *ContextEnricher) Enrich(ctx context.Context, builder *RecordBuilder) {
	for i := 0; i < len(e.getters); i++ {
		if attr, ok := e.getters[i](ctx); ok {
			builder.AddAttr(attr)
		}
	}
}

// FromContext provides methods that build [ContextValueGetter] values for [NewContextEnricher].
// Each method maps ctx.Value(ctxKey) to a log attribute named logKey.
var FromContext = contextEnricherGetters{}

type contextEnricherGetters struct{}

// Any gets ctx.Value(ctxKey) and, if non-nil, appends [Any](logKey, value).
func (contextEnricherGetters) Any(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		v := ctx.Value(ctxKey)
		if v == nil {
			return Attr{}, false
		}
		return Any(logKey, v), true
	}
}

// String gets ctx.Value(ctxKey) as a string and appends [String](logKey, value).
func (contextEnricherGetters) String(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		v, ok := ctx.Value(ctxKey).(string)
		if !ok {
			return Attr{}, false
		}
		return String(logKey, v), true
	}
}

// Int64 gets ctx.Value(ctxKey) as a signed integer and appends [Int64](logKey, value).
func (contextEnricherGetters) Int64(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		switch v := ctx.Value(ctxKey).(type) {
		case int:
			return Int64(logKey, int64(v)), true
		case int8:
			return Int64(logKey, int64(v)), true
		case int16:
			return Int64(logKey, int64(v)), true
		case int32:
			return Int64(logKey, int64(v)), true
		case int64:
			return Int64(logKey, v), true
		default:
			return Attr{}, false
		}
	}
}

// Uint64 gets ctx.Value(ctxKey) as an unsigned integer and appends [Uint64](logKey, value).
func (contextEnricherGetters) Uint64(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		switch v := ctx.Value(ctxKey).(type) {
		case uint:
			return Uint64(logKey, uint64(v)), true
		case uint8:
			return Uint64(logKey, uint64(v)), true
		case uint16:
			return Uint64(logKey, uint64(v)), true
		case uint32:
			return Uint64(logKey, uint64(v)), true
		case uint64:
			return Uint64(logKey, v), true
		case uintptr:
			return Uint64(logKey, uint64(v)), true
		default:
			return Attr{}, false
		}
	}
}

// Bool gets ctx.Value(ctxKey) as a bool and appends [Bool](logKey, value).
func (contextEnricherGetters) Bool(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		v, ok := ctx.Value(ctxKey).(bool)
		if !ok {
			return Attr{}, false
		}
		return Bool(logKey, v), true
	}
}

// Duration gets ctx.Value(ctxKey) as a [time.Duration] and appends [Duration](logKey, value).
func (contextEnricherGetters) Duration(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		v, ok := ctx.Value(ctxKey).(time.Duration)
		if !ok {
			return Attr{}, false
		}
		return Duration(logKey, v), true
	}
}

// Time gets ctx.Value(ctxKey) as a [time.Time] and appends [Time](logKey, value).
func (contextEnricherGetters) Time(ctxKey any, logKey string) ContextValueGetter {
	return func(ctx context.Context) (Attr, bool) {
		if logKey == "" {
			return Attr{}, false
		}
		v, ok := ctx.Value(ctxKey).(time.Time)
		if !ok {
			return Attr{}, false
		}
		return Time(logKey, v), true
	}
}
