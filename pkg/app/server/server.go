// Package server defines the ReST service
package server

import (
	"time"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/state"
	"github.com/grokloc/grokloc-server/pkg/env"
)

// Version is the current API version
const Version = "v0"

// API headers
// TokenRequest is formatted as security.EncodedSHA256(id+api-secret)
const (
	IDHeader           = "X-GrokLOC-ID"
	TokenRequestHeader = "X-GrokLOC-TokenRequest"
)

// Auth levels to be found in ctx with key authLevelCtxKey
const (
	AuthUser = iota
	AuthOrg
	AuthRoot
)

// contextKey is used to dismbiguate keys for vars put into request contexts
type contextKey struct {
	name string
}

// Context key instances for inserting and reading context vars
var (
	sessionCtxKey   = &contextKey{"session"}   // nolint
	authLevelCtxKey = &contextKey{"authlevel"} // nolint
)

// Instance is a single app server
type Instance struct {
	ST      *app.State
	Started time.Time
}

// New creates a new app server Instance
func New(level env.Level) (*Instance, error) {
	st, err := state.New(level)
	if err != nil {
		return nil, err
	}
	return &Instance{ST: st, Started: time.Now()}, nil
}
