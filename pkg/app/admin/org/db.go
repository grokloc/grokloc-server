package org

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user"
	"github.com/grokloc/grokloc-server/pkg/app/audit"
	"github.com/grokloc/grokloc-server/pkg/grokloc"
	"github.com/grokloc/grokloc-server/pkg/models"
	"go.uber.org/zap"
)

// Create instantiates a new owner, inserts it, and inserts a new org
// (read org to capture ctime, mtime)
func Create(
	ctx context.Context,
	name, ownerDisplayName, ownerEmail, ownerPassword string,
	key []byte,
	db *sql.DB) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	// generate org id
	id := uuid.NewString()

	// create the owner user
	// owner user is still unconfirmed, see update status below
	ownerUser, err := user.Encrypted(
		ctx,
		ownerDisplayName,
		ownerEmail,
		id,
		ownerPassword,
		key)
	if err != nil {
		zap.L().Error("org::Create: user::Encrypted",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	// insert owner user
	err = ownerUser.Insert(ctx, db)
	if err != nil {
		zap.L().Error("org::Create: user::Insert",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	// make active
	err = ownerUser.UpdateStatus(ctx, models.StatusActive, db)
	if err != nil {
		zap.L().Error("org::Create: UpdateStatus",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	// insert org
	q := fmt.Sprintf(`insert into %s
                          (id,
                           name,
                           owner,
                           status,
                           schema_version)
                          values
                          ($1,$2,$3,$4,$5)`,
		app.OrgsTableName)

	result, err := db.ExecContext(ctx,
		q,
		id,
		name,
		ownerUser.ID,
		models.StatusActive,
		Version)

	if err != nil {
		zap.L().Error("org::Create: Exec",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
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
		zap.L().Error("org::Create: rows affected",
			zap.Error(models.ErrRowsAffected),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, models.ErrRowsAffected
	}

	_ = audit.Insert(ctx, audit.ORG_INSERT, "", app.OrgsTableName, id, db)

	// read back to get ctime, mtime
	return Read(ctx, id, db)
}

func Read(ctx context.Context, id string, db *sql.DB) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	q := fmt.Sprintf(`select
                          name,
                          owner,
                          ctime,
                          mtime,
                          status,
                          schema_version
                          from %s
                          where id = $1`,
		app.OrgsTableName)

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
		zap.L().Error("org::Read: QueryRow",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	o.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}

	if o.Meta.SchemaVersion != Version {
		// handle migrating different versions, or err
		zap.L().Error("org::Read: schema version",
			zap.Error(models.ErrModelMigrate),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, models.ErrModelMigrate
	}

	return o, nil
}

// UpdateOwner sets the org owner
// TODO check owner exists, is in same org, is active
func (o *Org) UpdateOwner(ctx context.Context,
	owner string,
	db *sql.DB) error {

	defer func() {
		_ = zap.L().Sync()
	}()

	q := fmt.Sprintf(`select count(*)
                          from %s
                          where
                          id = $1
                           and
                          org = $2
                           and
                          status = $3`, app.UsersTableName)

	var count int
	err := db.QueryRowContext(ctx, q, owner, o.ID, models.StatusActive).Scan(&count)
	if err != nil {
		zap.L().Error("org::UpdateOwner: QueryRow",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return err
	}

	if count != 1 {
		zap.L().Error("org::UpdateOwner: owner check",
			zap.Error(models.ErrRelatedUser),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return models.ErrRelatedUser
	}

	err = models.Update(ctx, app.OrgsTableName, o.ID, "owner", owner, db)
	if err != nil {
		zap.L().Error("org::UpdateOwner: Update",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
	} else {
		o.Owner = owner
		_ = audit.Insert(ctx, audit.ORG_OWNER, "", app.OrgsTableName, o.ID, db)
	}

	return err
}

// UpdateStatus sets the org status
func (o *Org) UpdateStatus(ctx context.Context,
	status models.Status,
	db *sql.DB) error {

	defer func() {
		_ = zap.L().Sync()
	}()

	if status == models.StatusNone {
		zap.L().Error("org::UpdateStatus: status",
			zap.Error(models.ErrDisallowedValue),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return models.ErrDisallowedValue
	}

	err := models.Update(ctx, app.OrgsTableName, o.ID, "status", status, db)
	if err != nil {
		zap.L().Error("org::UpdateStatus: Update",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
	} else {
		o.Meta.Status = status
		_ = audit.Insert(ctx, audit.STATUS, "", app.OrgsTableName, o.ID, db)
	}

	return err
}
