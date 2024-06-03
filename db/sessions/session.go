package sessions

import (
	"database/sql"

	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type Session struct {
	models.Record
	Session     string         `db:"session"`
	SessionData types.JsonMap  `db:"session_data"`
	Resource    sql.NullString `db:"resource"`
	Expires     types.DateTime `db:"expires"`
	User        sql.NullInt64  `db:"user_id"`
}

func (s Session) TableName() string {
	return models.DEFAULT_SESSIONS_TABLE_NAME
}
