package sessions

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type Session struct {
	models.Record
	ID          int64          `db:"pk,id"`
	Session     string         `db:"session"`
	SessionData types.JsonMap  `db:"session_data"`
	Expires     types.DateTime `db:"expires"`
	User        int64          `db:"user_id"`
}

func (s Session) TableName() string {
	return models.DEFAULT_SESSIONS_TABLE_NAME
}
