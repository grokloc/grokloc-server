// Package state maintains handles to all external state mechanisms
package state

import (
	"errors"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/env"
)

// New creates a new state instance for the given level
func New(level env.Level) (*app.State, error) {
	if level == env.None {
		return nil, errors.New("no instance for None")
	}
	if level == env.Unit {
		return Unit(), nil
	}
	return nil, errors.New("no state constructor available")
}
