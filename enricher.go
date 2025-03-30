package golog

import "context"

// Enricher defines the interface for log entry enrichment.
// Enrichers can add additional fields to log entries based on context.
type Enricher interface {
	// Enrich adds additional fields to a log entry based on the context.
	// The fields map is modified in place to add the enriched fields.
	Enrich(ctx context.Context, level string, msg string, fields map[string]any)
}

// EnricherFunc is a function type that implements the Enricher interface.
// It allows functions to be used as enrichers without creating a new type.
type EnricherFunc func(ctx context.Context, level string, msg string, fields map[string]any)

// Enrich implements the Enricher interface for EnricherFunc.
// It simply calls the underlying function with the provided arguments.
func (f EnricherFunc) Enrich(ctx context.Context, level string, msg string, fields map[string]any) {
	f(ctx, level, msg, fields)
}
