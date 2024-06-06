package users

import (
	"encoding/base64"
	"encoding/binary"

	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type User struct {
	models.Record
	ID       int64          `db:"pk,id"`
	Name     string         `db:"name"`
	Email    string         `db:"email"`
	Password string         `db:"password"`
	UserData types.JsonMap  `db:"user_data"`
	Avatar   string         `db:"avatar"`
	Expires  types.DateTime `db:"expires"`
	LastSeen types.DateTime `db:"last_seen"`
	Role     int            `db:"role"`
	Active   bool           `db:"active"`
	Verified bool           `db:"verified"`
}

func (u User) TableName() string {
	return models.DEFAULT_USERS_TABLE_NAME
}

func (s User) PrimaryKey() string {
	b := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(b, s.ID)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}
