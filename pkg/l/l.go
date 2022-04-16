// Package logging sets the default logger
package logging

import (
	"log"

	"go.uber.org/zap"
)

// init overrides the global zap logger with the dev logger;
// it can be further modified in different contexts
func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	_ = zap.ReplaceGlobals(logger)
}
