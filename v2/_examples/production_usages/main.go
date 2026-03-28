// Production-style logging: JSON lines on stdout, info-level default, compact durations.
//
// Run from the v2 module directory:
//
//	go run ./_examples/production_usages/
package main

import (
	"fmt"
	"os"
	"time"

	golog "github.com/jkaveri/golog/v2"
)

func main() {
	// Equivalent: golog.NewLogger(golog.ProductionConfig())
	log, err := golog.NewProductionLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	start := time.Now()
	log.Info("job started", golog.String("job_id", "batch-42"))
	// Debug is below min level and is not emitted.
	log.Debug("this line is dropped when Level is Info")
	log.Info("job finished",
		golog.String("job_id", "batch-42"),
		golog.Duration("elapsed", time.Since(start)),
	)
}
