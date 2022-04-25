// Package models provides shared model definitions
package models

import (
	"context"
	"database/sql"
	"fmt"
)

// Status is an int when stored
type Status int

// exported status values
const (
	StatusNone        = Status(0)
	StatusUnconfirmed = Status(1)
	StatusActive      = Status(2)
	StatusInactive    = Status(3)
)

// NewStatus creates a Status from an int
func NewStatus(status int) (Status, error) {
	switch status {
	case 1:
		return StatusUnconfirmed, nil
	case 2:
		return StatusActive, nil
	case 3:
		return StatusInactive, nil
	default:
		return StatusNone, ErrStatus
	}
}

// Meta models metadata common to all models
type Meta struct {
	Ctime         int64  `json:"ctime"`
	Mtime         int64  `json:"mtime"`
	SchemaVersion int    `json:"schema_version"`
	Status        Status `json:"status"`
}

// Base models core attributes common to all models
type Base struct {
	ID   string `json:"id"`
	Meta Meta   `json:"meta"`
}

// Update changes the value of a column given a tablename, column name and id
func Update(ctx context.Context,
	tableName,
	id,
	colName string,
	val interface{},
	db *sql.DB) error {

	q := fmt.Sprintf(`update %s
                          set %s = $1
                          where id = $2`,
		tableName, colName)

	result, err := db.ExecContext(ctx, q, val, id)
	if err != nil {
		return err
	}

	updated, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}

	if updated == 0 {
		return sql.ErrNoRows
	}

	if updated != 1 {
		return ErrRowsAffected
	}

	return nil
}
