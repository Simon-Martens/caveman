package users

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type User struct {
	models.Record
	Name     string         `db:"name"`
	HID      string         `db:"hid"`
	Email    string         `db:"email"`
	Password string         `db:"password"`
	UserData types.JsonMap  `db:"user_data"`
	Avatar   string         `db:"avatar"`
	Expires  types.DateTime `db:"expires"`
	Role     int            `db:"role"`
	Active   bool           `db:"active"`
	Verified bool           `db:"verified"`
}

func (u User) TableName() string {
	return models.DEFAULT_USERS_TABLE_NAME
}
