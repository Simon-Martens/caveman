package models

import "github.com/Simon-Martens/caveman/tools/types"

type Record struct {
	Created  types.DateTime `db:"created"`
	Modified types.DateTime `db:"modified"`
}

func NewRecord() Record {
	r := Record{}

	r.Created = types.NowDateTime()
	r.Modified = r.Created

	return r
}
