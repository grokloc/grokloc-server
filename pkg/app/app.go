// Package app contains shared cross-app definitions
package app

import (
	"database/sql"

	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/matthewhartstonge/argon2"
	"go.uber.org/zap"
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
	L                                    *zap.Logger
}
