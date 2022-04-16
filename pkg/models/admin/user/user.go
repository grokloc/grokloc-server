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
	"go.uber.org/zap"
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

	defer func() {
		_ = zap.L().Sync()
	}()

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
		zap.L().Error("user::Insert: Exec",
			zap.Error(err),
		)
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
		zap.L().Error("user::Insert: rows affected",
			zap.Error(models.ErrRowsAffected),
		)
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

	defer func() {
		_ = zap.L().Sync()
	}()

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
		zap.L().Error("user::Create: QueryRow",
			zap.Error(err),
		)
		return nil, err
	}

	if count != 1 {
		zap.L().Error("user::Create: org check failure",
			zap.Error(models.ErrRelatedOrg),
		)
		return nil, models.ErrRelatedOrg
	}

	// generate encrypted user
	u, err := Encrypted(displayName, email, org, password, key)
	if err != nil {
		zap.L().Error("user::Create: Encrypted",
			zap.Error(err),
		)
		return nil, err
	}

	// insert user
	err = u.Insert(ctx, db)
	if err != nil {
		zap.L().Error("user::Create: Insert",
			zap.Error(err),
		)
		return nil, err
	}

	user, err := Read(ctx, u.ID, key, db)
	if err != nil {
		zap.L().Error("user::Create: Read",
			zap.Error(err),
		)
		return nil, err
	}

	return user, nil
}

// Encrypted creates a new user that can be inserted
func Encrypted(displayName, email, org, password string, key []byte) (*User, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	if !security.SafeStr(displayName) {
		err := errors.New("display_name deemed unsafe")
		zap.L().Error("user::Encrypted: display name",
			zap.Error(err),
		)
		return nil, err
	}
	if !security.SafeStr(email) {
		err := errors.New("email deemed unsafe")
		zap.L().Error("user::Encrypted: email",
			zap.Error(err),
		)
		return nil, err
	}
	if !security.SafeStr(password) {
		err := errors.New("password deemed unsafe")
		zap.L().Error("user::Encrypted: password",
			zap.Error(err),
		)
		return nil, err
	}

	// org will be checked in the db, password assumed derived

	apiSecret := uuid.NewString()
	apiSecretEncrypted, err := security.Encrypt(apiSecret, key)
	if err != nil {
		zap.L().Error("user::Encrypted: api secret",
			zap.Error(err),
		)
		return nil, err
	}
	displayNameEncrypted, err := security.Encrypt(displayName, key)
	if err != nil {
		zap.L().Error("user::Encrypted: display name",
			zap.Error(err),
		)
		return nil, err
	}
	emailEncrypted, err := security.Encrypt(email, key)
	if err != nil {
		zap.L().Error("user::Encrypted: email",
			zap.Error(err),
		)
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

	defer func() {
		_ = zap.L().Sync()
	}()

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
		zap.L().Error("user::Read: QueryRow",
			zap.Error(err),
		)
		return nil, err
	}

	u.APISecret, err = security.Decrypt(encryptedAPISecret, u.APISecretDigest, key)
	if err != nil {
		zap.L().Error("user::Read: api secret",
			zap.Error(err),
		)
		return nil, err
	}

	u.DisplayName, err = security.Decrypt(encryptedDisplayName, u.DisplayNameDigest, key)
	if err != nil {
		zap.L().Error("user::Read: display name",
			zap.Error(err),
		)
		return nil, err
	}

	u.Email, err = security.Decrypt(encryptedEmail, u.EmailDigest, key)
	if err != nil {
		zap.L().Error("user::Read: email",
			zap.Error(err),
		)
		return nil, err
	}

	u.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		zap.L().Error("user::Read: status",
			zap.Error(err),
		)
		return nil, err
	}

	if u.Meta.SchemaVersion != Version {
		zap.L().Error("user::Read: schema version",
			zap.Error(models.ErrModelMigrate),
		)
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

	defer func() {
		_ = zap.L().Sync()
	}()

	if !security.SafeStr(displayName) {
		err := errors.New("display name malformed")
		zap.L().Error("user::UpdateDisplayName: display name",
			zap.Error(err),
		)
		return err
	}

	// both the display name and the digest must be reset
	encryptedDisplayName, err := security.Encrypt(displayName, key)
	if err != nil {
		zap.L().Error("user::UpdateDisplayName: display name",
			zap.Error(err),
		)
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
		zap.L().Error("user::UpdateDisplayName: Exec",
			zap.Error(err),
		)
		return err
	}

	updated, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}

	if updated != 1 {
		zap.L().Error("user::UpdateDisplayName: rows affected",
			zap.Error(err),
		)
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

	defer func() {
		_ = zap.L().Sync()
	}()

	if !security.SafeStr(password) {
		err := errors.New("password malformed")
		zap.L().Error("user::UpdatePassword: password",
			zap.Error(err),
		)
		return err
	}

	err := models.Update(ctx, schemas.UsersTableName, u.ID, "password", password, db)
	if err != nil {
		zap.L().Error("user::UpdatePassword: Update",
			zap.Error(err),
		)
	} else {
		u.Password = password
	}

	return err
}

// UpdateStatus sets the user status
func (u *User) UpdateStatus(ctx context.Context,
	status models.Status,
	db *sql.DB) error {

	defer func() {
		_ = zap.L().Sync()
	}()

	if status == models.StatusNone {
		err := errors.New("cannot use None as a stored status")
		zap.L().Error("user::UpdateStatus: status",
			zap.Error(err),
		)
		return err
	}

	err := models.Update(ctx, schemas.UsersTableName, u.ID, "status", status, db)
	if err != nil {
		zap.L().Error("user::UpdateStatus: status",
			zap.Error(err),
		)
	} else {
		u.Meta.Status = status
	}

	return err
}
