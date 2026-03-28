// Custom logger: explicit [golog.Writer], enrichers, and scoped loggers.
//
// Run from the v2 module directory:
//
//	go run ./_examples/customize_logger/
package main

import (
	"context"
	"fmt"
	"os"

	golog "github.com/jkaveri/golog/v2"
)

type requestIDKey struct{}

func main() {
	log := golog.NewLoggerWriter(
		golog.NewTextWriter(os.Stdout),
		golog.LevelInfo,
		golog.NewContextEnricher(
			golog.FromContext.String(requestIDKey{}, "request_id"),
		),
		golog.NewSourceEnricher(golog.SourceEnricherOptions{
			FieldName: "caller",
			Format:    golog.SourceFormatFileLine,
			Skip:      2,
		}),
		golog.EnricherFunc(func(ctx context.Context, b *golog.RecordBuilder) {
			b.AddAttr(golog.String("service", "checkout"))
		}),
	)

	ctx := context.WithValue(context.Background(), requestIDKey{}, "req-abc")
	scope := log.With(golog.String("version", "1.0.0")).WithContext(ctx)

	scope.Info("order placed", golog.Int("items", 3))

	// Per-logger level without rebuilding the writer.
	scope.SetLevel(golog.LevelError)
	scope.Info("suppressed: min level is now Error")
	scope.Error("payment failed", golog.String("gateway", "stripe"))

	fmt.Fprintln(os.Stderr, "(stderr) done")
}
