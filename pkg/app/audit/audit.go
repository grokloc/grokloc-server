// package audit provides mutation recording
package audit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/models"
)

const (
	STATUS            int = 10
	ORG_INSERT        int = 100
	ORG_OWNER         int = 101
	USER_INSERT       int = 200
	USER_DISPLAY_NAME int = 201
	USER_PASSWORD     int = 202
)

func Insert(
	ctx context.Context,
	code int,
	source, source_id string,
	db *sql.DB) error {

	q := fmt.Sprintf(`insert into %s
                          (id,
                           code,
                           source,
                           source_id)
                          values
                          (?,?,?,?)`,
		app.AuditTableName)

	result, err := db.ExecContext(ctx,
		q,
		uuid.NewString(),
		code,
		source,
		source_id)

	if err != nil {
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
