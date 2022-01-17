package admin

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

const UserVersion = 0

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

// EncryptedUser creates a new user that can be inserted
func EncryptedUser(displayName, email, org, password string, key []byte) (*User, error) {
	if !security.SafeStr(displayName) {
		return nil, errors.New("display_name deemed unsafe")
	}
	if !security.SafeStr(email) {
		return nil, errors.New("email deemed unsafe")
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
				SchemaVersion: UserVersion,
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
// (should read user from db before returning to capture ctime, mtime)

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

	u.APISecret, err = security.Decrypt(encryptedAPISecret, key)
	if err != nil {
		return nil, err
	}
	if security.EncodedSHA256(u.APISecret) != u.APISecretDigest {
		return nil, models.ErrDigest
	}

	u.DisplayName, err = security.Decrypt(encryptedDisplayName, key)
	if err != nil {
		return nil, err
	}
	if security.EncodedSHA256(u.DisplayName) != u.DisplayNameDigest {
		return nil, models.ErrDigest
	}

	u.Email, err = security.Decrypt(encryptedEmail, key)
	if err != nil {
		return nil, err
	}
	if security.EncodedSHA256(u.Email) != u.EmailDigest {
		return nil, models.ErrDigest
	}

	u.Meta.Status, err = models.NewStatus(statusRaw)
	if err != nil {
		return nil, err
	}

	if u.Meta.SchemaVersion != UserVersion {
		// handle migrating different versions, or err
		return nil, models.ErrModelMigrate
	}
	return u, nil
}
