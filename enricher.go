package golog

import "context"

// Enricher defines the interface for log entry enrichment.
// Enrichers add additional fields to log entries based on context.
// The fields map is modified in place; add keys to enrich the log entry.
type Enricher interface {
	// Enrich adds additional fields to a log entry based on the context.
	// Modify the fields map in place; do not replace it.
	Enrich(ctx context.Context, level string, msg string, fields map[string]any)
}

// EnricherFunc is a function type that implements the Enricher interface.
// Use it to create enrichers without defining a new type:
//
//	golog.RegisterEnricher(golog.EnricherFunc(func(ctx context.Context, level, msg string, fields map[string]any) {
//	    fields["trace_id"] = traceIDFromContext(ctx)
//	}))
type EnricherFunc func(ctx context.Context, level string, msg string, fields map[string]any)

// Enrich implements the Enricher interface for EnricherFunc.
// It simply calls the underlying function with the provided arguments.
func (f EnricherFunc) Enrich(ctx context.Context, level string, msg string, fields map[string]any) {
	f(ctx, level, msg, fields)
}
