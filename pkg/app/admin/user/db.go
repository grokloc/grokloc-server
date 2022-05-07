// Package user contains package methods for user support
package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/audit"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
)

func (u User) Insert(ctx context.Context, db *sql.DB) error {

	q := fmt.Sprintf(`insert into %s
                          (id,
                           api_secret,
                           api_secret_digest,
                           display_name,
                           display_name_digest,
                           email,
                           email_digest,
                           org,
                           password,
                           status,
                           schema_version)
                          values
                          ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		app.UsersTableName)

	result, err := db.ExecContext(ctx,
		q,
		u.ID,
		u.APISecret,
		u.APISecretDigest,
		u.DisplayName,
		u.DisplayNameDigest,
		u.Email,
		u.EmailDigest,
		u.Org,
		u.Password,
		u.Meta.Status,
		u.Meta.SchemaVersion)

	if err != nil {
		if models.UniqueConstraint(err) {
			return models.ErrConflict
		}
		return err
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}
	if inserted != 1 {
		return models.ErrRowsAffected
	}

	_ = audit.Insert(ctx, audit.USER_INSERT, app.UsersTableName, u.ID, db)

	return nil
}

// Create creates an encrypted user, validates the org, then inserts the user
// (read user to capture ctime, mtime)
func Create(
	ctx context.Context,
	displayName,
	email,
	org,
	password string,
	key []byte,
	db *sql.DB) (*User, error) {

	// check that org exists and is active
	q := fmt.Sprintf(`select count(*)
                          from %s
                          where
                            id = $1
                          and
                            status = $2`,
		app.OrgsTableName)

	var count int
	err := db.QueryRowContext(ctx, q, org, models.StatusActive).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, models.ErrRelatedOrg
	}

	// generate encrypted user
	u, err := Encrypted(ctx, displayName, email, org, password, key)
	if err != nil {
		return nil, err
	}

	// insert user
	err = u.Insert(ctx, db)
	if err != nil {
		return nil, err
	}

	user, err := Read(ctx, u.ID, key, db)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Encrypted creates a new user that can be inserted
func Encrypted(
	ctx context.Context,
	displayName, email, org, password string,
	key []byte) (*User, error) {

	apiSecret := uuid.NewString()
	apiSecretEncrypted, err := security.Encrypt(apiSecret, key)
	if err != nil {
		return nil, err
	}
	displayNameEncrypted, err := security.Encrypt(displayName, key)
	if err != nil {
		return nil, err
	}
	emailEncrypted, err := security.Encrypt(email, key)
	if err != nil {
		return nil, err
	}

	return &User{
		Base: models.Base{
			ID: uuid.NewString(),
			Meta: models.Meta{
				SchemaVersion: Version,
				Status:        models.StatusUnconfirmed,
				// Ctime, Mtime remain 0
			},
		},
		APISecret:         apiSecretEncrypted,
		APISecretDigest:   security.EncodedSHA256(apiSecret),
		DisplayName:       displayNameEncrypted,
		DisplayNameDigest: security.EncodedSHA256(displayName),
		Email:             emailEncrypted,
		EmailDigest:       security.EncodedSHA256(email),
		Org:               org,
		Password:          password,
	}, nil
}

func Read(ctx context.Context, id string, key []byte, db *sql.DB) (*User, error) {

	q := fmt.Sprintf(`select
                          api_secret,
                          api_secret_digest,
                          display_name,
                          display_name_digest,
                          email,
                          email_digest,
                          org,
                          password,
                          ctime,
                          mtime,
                          status,
                          schema_version
                          from %s
                          where id = $1`,
		app.UsersTableName)

	var statusRaw int
	u := &User{}
	u.ID = id
	var encryptedAPISecret, encryptedDisplayName, encryptedEmail string

	err := db.QueryRowContext(ctx, q, id).Scan(
		&encryptedAPISecret,
		&u.APISecretDigest,
		&encryptedDisplayName,
		&u.DisplayNameDigest,
		&encryptedEmail,
		&u.EmailDigest,
		&u.Org,
		&u.Password,
		&u.Meta.Ctime,
		&u.Meta.Mtime,
		&statusRaw,
		&u.Meta.SchemaVersion)
	if err != nil {
		return nil, err
	}

	u.APISecret, err = security.Decrypt(encryptedAPISecret, u.APISecretDigest, key)
	if err != nil {
		return nil, err
	}

	u.DisplayName, err = security.Decrypt(encryptedDisplayName, u.DisplayNameDigest, key)
	if err != nil {
		return nil, err
	}

	u.Email, err = security.Decrypt(encryptedEmail, u.EmailDigest, key)
	if err != nil {
		return nil, err
	}

	u.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}

	if u.Meta.SchemaVersion != Version {
		// handle migrating different versions, or err
		return nil, models.ErrModelMigrate
	}

	return u, nil
}

// UpdateDisplayName sets the user display name
func (u *User) UpdateDisplayName(ctx context.Context,
	displayName string,
	key []byte,
	db *sql.DB) error {

	// both the display name and the digest must be reset
	encryptedDisplayName, err := security.Encrypt(displayName, key)
	if err != nil {
		return err
	}

	displayNameDigest := security.EncodedSHA256(displayName)

	q := `update users
              set display_name = $1,
              display_name_digest = $2
              where id = $3`

	result, err := db.ExecContext(ctx,
		q,
		encryptedDisplayName,
		displayNameDigest,
		u.ID)

	if err != nil {
		return err
	}

	updated, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}

	if updated != 1 {
		return models.ErrRowsAffected
	}

	u.DisplayName = displayName
	u.DisplayNameDigest = displayNameDigest

	_ = audit.Insert(ctx, audit.USER_DISPLAY_NAME, app.UsersTableName, u.ID, db)

	return nil
}

// UpdatePassword sets the user password
// password assumed derived
func (u *User) UpdatePassword(ctx context.Context,
	password string,
	db *sql.DB) error {

	err := models.Update(ctx, app.UsersTableName, u.ID, "password", password, db)
	if err == nil {
		u.Password = password
		_ = audit.Insert(ctx, audit.USER_PASSWORD, app.UsersTableName, u.ID, db)
	}

	return err
}

// UpdateStatus sets the user status
func (u *User) UpdateStatus(ctx context.Context,
	status models.Status,
	db *sql.DB) error {

	// unconfirmed can only be an initial state
	if status == models.StatusNone || status == models.StatusUnconfirmed {
		return models.ErrDisallowedValue
	}

	err := models.Update(ctx, app.UsersTableName, u.ID, "status", status, db)
	if err == nil {
		u.Meta.Status = status
		_ = audit.Insert(ctx, audit.STATUS, app.UsersTableName, u.ID, db)
	}

	return err
}
