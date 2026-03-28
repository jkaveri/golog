// Development-style logging: human-readable text on stdout, debug verbosity, optional source.
//
// Run from the v2 module directory:
//
//	go run ./_examples/development_usages/
package main

import (
	"fmt"
	"os"

	golog "github.com/jkaveri/golog/v2"
)

func main() {
	// Equivalent: golog.NewLogger(golog.DevelopmentConfig())
	log, err := golog.NewDevelopmentLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	log.Debug("verbose detail for local debugging", golog.String("component", "api"))
	log.Info("request handled", golog.Int("status", 200), golog.String("path", "/health"))
	log.Error("optional error line", golog.String("reason", "upstream timeout"))
}
