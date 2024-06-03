package models

import "github.com/Simon-Martens/caveman/tools/types"

type Record struct {
	ID       int64          `db:"pk,id"`
	Created  types.DateTime `db:"created"`
	Modified types.DateTime `db:"modified"`
}

func NewRecord() Record {
	r := Record{}

	r.Created = types.NowDateTime()
	r.Modified = types.NowDateTime()

	return r
}
