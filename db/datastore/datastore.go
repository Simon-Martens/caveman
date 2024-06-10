package datastore

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type DataStore struct {
	models.Record
	ID   int64         `db:"pk,id"`
	Key  string        `db:"key"`
	Data types.JsonRaw `db:"data"`
}

func (s DataStore) TableName() string {
	return models.DEFAULT_DATASTORE_TABLE
}

type Data interface {
	Key() string
}
