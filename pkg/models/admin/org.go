package admin

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/models"
)

const OrgVersion = 0

type Org struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// Create instantiates a new owner, inserts it, and inserts a new org
// (should read org from db before returning to capture ctime, mtime)
func Create(
	ctx context.Context,
	name, ownerDisplayName, ownerEmail, ownerPassword string,
	key []byte,
	db *sql.DB) (*Org, error) {
	// generate org id
	id := uuid.NewString()

	// create the owner user
	ownerUser, err := EncryptedUser(ownerDisplayName, ownerEmail, id, ownerPassword, key)
	if err != nil {
		return nil, err
	}

	// insert owner user
	err = ownerUser.Insert(ctx, db)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
