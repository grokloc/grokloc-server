package state

import (
	"context"
	"database/sql"
	"log"

	"github.com/google/uuid"
	"github.com/matthewhartstonge/argon2"
	_ "github.com/mattn/go-sqlite3" //
	"go.uber.org/zap"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/models/admin/org"
	"github.com/grokloc/grokloc-server/pkg/models/admin/user"
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

	argon2Cfg := argon2.DefaultConfig()

	rootUserPassword, err := security.DerivePassword(uuid.NewString(), argon2Cfg)
	if err != nil {
		log.Fatal(err)
	}

	rootOrg, err := org.Create(
		context.Background(),
		uuid.NewString(), // org name
		uuid.NewString(), // org owner display name
		uuid.NewString(), // org owner email
		rootUserPassword,
		dbKey,
		db,
	)
	if err != nil {
		log.Fatal(err)
	}

	rootUser, err := user.Read(
		context.Background(),
		rootOrg.Owner,
		dbKey,
		db,
	)
	if err != nil {
		log.Fatal(err)
	}

	// set the global logger for the unit env
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	_ = zap.ReplaceGlobals(logger)

	return &app.State{
		Level:             env.Unit,
		Master:            db,
		Replicas:          []*sql.DB{db},
		DBKey:             dbKey,
		TokenKey:          tokenKey,
		Argon2Cfg:         argon2Cfg,
		RootOrg:           rootOrg.ID,
		RootUser:          rootUser.ID,
		RootUserAPISecret: rootUser.APISecret,
	}
}
