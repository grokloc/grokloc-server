// Package user contains package methods for user support
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/schemas"
	"github.com/grokloc/grokloc-server/pkg/security"
)

const Version = 0

type User struct {
	models.Base
	APISecret         string `json:"api_secret"`
	APISecretDigest   string `json:"api_secret_digest"`
	DisplayName       string `json:"display_name"`
	DisplayNameDigest string `json:"display_name_digest"`
	Email             string `json:"email"`
	EmailDigest       string `json:"email_digest"`
	Org               string `json:"org"`
	// Password is assumed initialized as derived
	Password string `json:"-"`
}

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
		schemas.UsersTableName)

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
		schemas.OrgsTableName)

	var count int
	err := db.QueryRowContext(ctx, q, org, models.StatusActive).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, models.ErrRelatedOrg
	}

	// generate encrypted user
	u, err := Encrypted(displayName, email, org, password, key)
	if err != nil {
		return nil, err
	}

	// insert user
	err = u.Insert(ctx, db)
	if err != nil {
		return nil, err
	}

	return Read(ctx, u.ID, key, db)
}

// Encrypted creates a new user that can be inserted
func Encrypted(displayName, email, org, password string, key []byte) (*User, error) {
	if !security.SafeStr(displayName) {
		return nil, errors.New("display_name deemed unsafe")
	}
	if !security.SafeStr(email) {
		return nil, errors.New("email deemed unsafe")
	}
	if !security.SafeStr(password) {
		return nil, errors.New("password deemed unsafe")
	}

	// org will be checked in the db, password assumed derived

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
		schemas.UsersTableName)

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

	if !security.SafeStr(displayName) {
		return errors.New("display name malformed")
	}

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

	return nil
}

// UpdatePassword sets the user password
// password assumed derived
func (u *User) UpdatePassword(ctx context.Context,
	password string,
	db *sql.DB) error {

	if !security.SafeStr(password) {
		return errors.New("password malformed")
	}
	err := models.Update(ctx, schemas.UsersTableName, u.ID, "password", password, db)
	if err == nil {
		u.Password = password
	}
	return err
}

// UpdateStatus sets the user status
func (u *User) UpdateStatus(ctx context.Context,
	status models.Status,
	db *sql.DB) error {

	if status == models.StatusNone {
		return errors.New("cannot use None as a stored status")
	}
	err := models.Update(ctx, schemas.UsersTableName, u.ID, "status", status, db)
	if err == nil {
		u.Meta.Status = status
	}
	return err
}
