package users

import "github.com/Simon-Martens/caveman/models"

type User struct {
	ID       int
	Name     string
	Email    string
	Password []byte
	Created  int
	Modified int
	Expired  int
	Role     int
}

func (u User) TableName() string {
	return models.DEFAULT_USERS_TABLE_NAME
}
