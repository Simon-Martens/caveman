package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/db/accesstokens"
	"github.com/Simon-Martens/caveman/db/sessions"
	"github.com/Simon-Martens/caveman/db/users"
	"github.com/Simon-Martens/caveman/models"
)

type DatabaseEnv struct {
	DB  *db.DB
	UM  *users.UserManager
	SM  *sessions.SessionManager
	ATM *accesstokens.AccessTokenManager
}

func TestNewDatabaseEnv(T *testing.T) *DatabaseEnv {

	db, err := db.New(Path(), 120, 12)
	if err != nil {
		T.Fatal(err)
	}

	um, err := users.New(db, models.DEFAULT_USERS_TABLE_NAME, models.DEFAULT_ID_FIELD)
	if err != nil {
		T.Fatal(err)
	}

	sm, err := sessions.New(db,
		models.DEFAULT_SESSIONS_TABLE_NAME,
		models.DEFAULT_USERS_TABLE_NAME,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_SESSION_EXPIRATION)
	if err != nil {
		T.Fatal(err)
	}

	atm, err := accesstokens.New(db,
		models.DEFAULT_ACCESS_TOKENS_TABLE_NAME,
		models.DEFAULT_USERS_TABLE_NAME,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_RESOURCE_SESSION_EXPIRATION)

	return &DatabaseEnv{db, um, sm, atm}
}

func Path() string {
	_ = os.MkdirAll(models.DEFAULT_TEST_DATA_DIR_NAME, os.ModePerm)
	return filepath.Join(models.DEFAULT_TEST_DATA_DIR_NAME, models.DEFAULT_DATA_FILE_NAME)
}

func (dbenv *DatabaseEnv) Close() {
	dbenv.UM = nil
	dbenv.SM = nil
	dbenv.DB.Close()
}

func Clean() {
	os.RemoveAll(models.DEFAULT_TEST_DATA_DIR_NAME)
}
