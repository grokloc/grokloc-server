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
	"github.com/grokloc/grokloc-server/pkg/app/admin/org"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/env"
	"github.com/grokloc/grokloc-server/pkg/security"
)

// Unit builds an instance for the Unit environment
func Unit() *app.State {
	// set the global logger for the unit env
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	_ = zap.ReplaceGlobals(logger)

	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		zap.L().Fatal("unit db",
			zap.Error(err),
		)
	}
	// avoid concurrency bug with the sqlite library
	db.SetMaxOpenConns(1)
	_, err = db.Exec(app.Schema)
	if err != nil {
		zap.L().Fatal("unit schema",
			zap.Error(err),
		)
	}
	dbKey, err := security.MakeKey(uuid.NewString())
	if err != nil {
		zap.L().Fatal("db key",
			zap.Error(err),
		)
	}
	tokenKey, err := security.MakeKey(uuid.NewString())
	if err != nil {
		zap.L().Fatal("token key",
			zap.Error(err),
		)
	}

	argon2Cfg := argon2.DefaultConfig()

	rootUserPassword, err := security.DerivePassword(uuid.NewString(), argon2Cfg)
	if err != nil {
		zap.L().Fatal("root password",
			zap.Error(err),
		)
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
		zap.L().Fatal("root org create",
			zap.Error(err),
		)
	}

	rootUser, err := user.Read(
		context.Background(),
		rootOrg.Owner,
		dbKey,
		db,
	)
	if err != nil {
		zap.L().Fatal("root user create",
			zap.Error(err),
		)
	}

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
