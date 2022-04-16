// Package l is a global logger
package l

import (
	"log"

	"go.uber.org/zap"
)

// Og is a global logger that can be accessed as l.Og
var Og *zap.Logger

// init by default sets Og to the development logger, other
// contexts can just set it directly
func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	Og = logger
}
