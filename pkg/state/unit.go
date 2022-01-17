package state

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/matthewhartstonge/argon2"
	_ "github.com/mattn/go-sqlite3" //
	"go.uber.org/zap"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/schemas"
	"github.com/grokloc/grokloc-server/pkg/security"
)

// Unit builds an instance for the Unit environment
func Unit() *app.State {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	// avoid concurrency bug with the sqlite library
	db.SetMaxOpenConns(1)
	_, err = db.Exec(schemas.App)
	if err != nil {
		log.Fatal(err)
	}
	dbKey, err := security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	tokenKey, err := security.MakeKey(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	// HERE: add call to Org.create, User.read using master,
	// assign to Root*
	return &app.State{
		Level:     env.Unit,
		Master:    db,
		Replicas:  []*sql.DB{db},
		DBKey:     dbKey,
		TokenKey:  tokenKey,
		Argon2Cfg: argon2.DefaultConfig(),
		L:         logger,
	}
}
