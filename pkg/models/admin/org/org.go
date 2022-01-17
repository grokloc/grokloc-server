// Package org contains package methods for org support
package org

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/models/admin/user"
	"github.com/grokloc/grokloc-server/pkg/schemas"
)

type Org struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

const Version = 0

// Create instantiates a new owner, inserts it, and inserts a new org
// (read org to capture ctime, mtime)
func Create(
	ctx context.Context,
	name, ownerDisplayName, ownerEmail, ownerPassword string,
	key []byte,
	db *sql.DB) (*Org, error) {

	// generate org id
	id := uuid.NewString()

	// create the owner user
	// NOTE! this user is still unconfirmed
	ownerUser, err := user.Encrypted(
		ownerDisplayName,
		ownerEmail,
		id,
		ownerPassword,
		key)
	if err != nil {
		return nil, err
	}

	// insert owner user
	err = ownerUser.Insert(ctx, db)
	if err != nil {
		return nil, err
	}

	// TODO change owner status to active

	// insert org
	q := fmt.Sprintf(`insert into %s
                          (id,
                           name,
                           owner,
                           status,
                           schema_version)
                          values
                          ($1,$2,$3,$4,$5)`,
		schemas.OrgsTableName)

	result, err := db.ExecContext(ctx,
		q,
		id,
		name,
		ownerUser.ID,
		models.StatusActive,
		Version)

	if err != nil {
		if models.UniqueConstraint(err) {
			return nil, models.ErrConflict
		}
		return nil, err
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}
	if inserted != 1 {
		return nil, models.ErrRowsAffected
	}

	// read back to get ctime, mtime
	return Read(ctx, id, db)
}

func Read(ctx context.Context, id string, db *sql.DB) (*Org, error) {
	q := fmt.Sprintf(`select
                          name,
                          owner,
                          ctime,
                          mtime,
                          status,
                          schema_version
                          from %s
                          where id = $1`,
		schemas.OrgsTableName)

	var statusRaw int
	o := &Org{}
	o.ID = id

	err := db.QueryRowContext(ctx, q, id).Scan(
		&o.Name,
		&o.Owner,
		&o.Meta.Ctime,
		&o.Meta.Mtime,
		&statusRaw,
		&o.Meta.SchemaVersion)
	if err != nil {
		return nil, err
	}

	o.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}

	if o.Meta.SchemaVersion != Version {
		// handle migrating different versions, or err
		return nil, models.ErrModelMigrate
	}
	return o, nil
}
