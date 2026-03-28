package golog

import "context"

// Enricher adds or modifies attributes for log records before they reach the [Writer].
// Implementations may read [context.Context] (for example request ID or user) and call
// [RecordBuilder.AddAttr] / [RecordBuilder.AddAttrs]. Use [NewContextEnricher] with
// [FromContext] getters for typed context keys, or [NewSourceEnricher] for caller file/line.
type Enricher interface {
	// Enrich mutates the builder in place before the immutable record is built.
	// Implementations should use AddAttr/AddAttrs to append values.
	// It must not retain the builder or its attrs after return.
	Enrich(ctx context.Context, builder *RecordBuilder)
}

// EnricherFunc adapts a function to [Enricher], avoiding a new named type for one-off hooks.
//
// Example:
//
//	e := EnricherFunc(func(ctx context.Context, b *RecordBuilder) {
//		b.AddAttr(String("build", version))
//	})
type EnricherFunc func(context.Context, *RecordBuilder)

func (f EnricherFunc) Enrich(ctx context.Context, builder *RecordBuilder) {
	f(ctx, builder)
}
