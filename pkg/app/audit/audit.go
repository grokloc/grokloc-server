// package audit provides mutation recording
package audit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/models"
	"go.uber.org/zap"
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
	note, source, source_id string,
	db *sql.DB) error {

	defer func() {
		_ = zap.L().Sync()
	}()

	q := fmt.Sprintf(`insert into %s
                          (id,
                           code,
                           note,
                           source,
                           source_id)
                          values
                          ($1,$2,$3,$4,$5)`,
		app.AuditTableName)

	result, err := db.ExecContext(ctx,
		q,
		uuid.NewString(),
		code,
		note,
		source,
		source_id)

	if err != nil {
		zap.L().Error("audit::Insert: Exec",
			zap.Error(err),
			zap.Int("code", code),
			zap.String("source", source),
			zap.String("source_id", source_id),
		)
		return err
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		// the db does not support a basic feature
		panic("cannot exec RowsAffected:" + err.Error())
	}
	if inserted != 1 {
		zap.L().Error("audit::Insert: rows affected",
			zap.Error(models.ErrRowsAffected),
			zap.Int("code", code),
			zap.String("source", source),
			zap.String("source_id", source_id),
		)
		return models.ErrRowsAffected
	}

	return nil
}
