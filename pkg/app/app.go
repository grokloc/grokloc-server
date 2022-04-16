// Package app contains shared cross-app definitions
package app

import (
	"database/sql"
	"log"
	"math/rand"

	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/matthewhartstonge/argon2"
)

// State contains references to all external state mechanisms
type State struct {
	Level                                env.Level
	Master                               *sql.DB
	Replicas                             []*sql.DB
	DBKey                                []byte
	TokenKey                             []byte
	Argon2Cfg                            argon2.Config
	RootOrg, RootUser, RootUserAPISecret string
}

// RandomReplica selects a random replica
func (s *State) RandomReplica() *sql.DB {
	l := len(s.Replicas)
	if l == 0 {
		log.Fatal("there are no replicas")
	}
	return s.Replicas[rand.Intn(l)]
}
