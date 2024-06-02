package users

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type User struct {
	ID       int            `db:"pk,id"`
	Name     string         `db:"name"`
	Email    string         `db:"email"`
	Password string         `db:"password"`
	Created  types.DateTime `db:"created"`
	Modified types.DateTime `db:"modified"`
	Expires  types.DateTime `db:"expires"`
	Role     int            `db:"role"`
	Active   bool           `db:"active"`
	Verified bool           `db:"verified"`
}

func (u User) TableName() string {
	return models.DEFAULT_USERS_TABLE_NAME
}
