package sessions

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type Session struct {
	models.Record
	ID          int64          `db:"pk,id"`
	Session     string         `db:"session"`
	SessionData types.JsonMap  `db:"session_data"`
	Expires     types.DateTime `db:"expires"`
	IP          string         `db:"ip"`
	Agent       string         `db:"agent"`
	User        int64          `db:"user_id"`
}

func (s Session) TableName() string {
	return models.DEFAULT_SESSIONS_TABLE_NAME
}

func (s Session) PrimaryKey() string {
	b := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(b, s.ID)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}
