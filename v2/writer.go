package golog

import (
	"context"
)

// Writer receives fully built [Record] values from a [Logger]. Implementations
// format and flush to an [io.Writer], remote sink, or discard. [TextWriter] and
// [JSONWriter]
// are the built-in implementations.
//
// Write must not retain record or its attributes after returning (the caller
// may reuse memory). Errors from Write are currently ignored by the standard
// [Logger] implementation; return
// them for custom callers that inspect errors.
type Writer interface {
	// Write serializes one log event. ctx is the logger’s context from
	// [Logger.WithContext].
	Write(ctx context.Context, record Record) error
}
