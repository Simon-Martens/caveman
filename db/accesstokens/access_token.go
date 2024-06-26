package accesstokens

import (
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
)

type AccessToken struct {
	models.Record
	ID        int64          `db:"pk,id"`
	Token     string         `db:"token"`
	TokenData types.JsonMap  `db:"token_data"`
	Path      string         `db:"path"`
	Creator   int64          `db:"creator_id"`
	Expires   types.DateTime `db:"expires"`
	Uses      int64          `db:"uses"`
}

func (a AccessToken) TableName() string {
	return models.DEFAULT_ACCESS_TOKENS_TABLE
}
