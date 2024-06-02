package sessions

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type Session struct {
	models.Record
	Session     string         `db:"session"`
	SessionData types.JsonRaw  `db:"session_data"`
	Expires     types.DateTime `db:"expires"`
	User        int            `db:"user_id"`
}

func (s Session) TableName() string {
	return models.DEFAULT_SESSIONS_TABLE_NAME
}
