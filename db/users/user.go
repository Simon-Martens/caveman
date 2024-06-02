package users

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type User struct {
	models.Record
	Name     string         `db:"name"`
	Email    string         `db:"email"`
	Password string         `db:"password"`
	Expires  types.DateTime `db:"expires"`
	Role     int            `db:"role"`
	Active   bool           `db:"active"`
	Verified bool           `db:"verified"`
}

func (u User) TableName() string {
	return models.DEFAULT_USERS_TABLE_NAME
}
