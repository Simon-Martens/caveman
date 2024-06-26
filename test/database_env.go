package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/db/accesstokens"
	"github.com/Simon-Martens/caveman/db/datastore"
	"github.com/Simon-Martens/caveman/db/sessions"
	"github.com/Simon-Martens/caveman/db/users"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/security"
)

type DatabaseEnv struct {
	DB  *db.DB
	UM  *users.UserManager
	SM  *sessions.SessionManager
	ATM *accesstokens.AccessTokenManager
	DSM *datastore.DataStoreManager
}

func TestNewDatabaseEnv(T *testing.T) *DatabaseEnv {

	db, err := db.New(Path(), 120, 12)
	db.ConnectLogger()
	if err != nil {
		T.Fatal(err)
	}

	um, err := users.New(db,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_USER_EXPIRATION,
		security.GenRandomUIntNotPrime())
	if err != nil {
		T.Fatal(err)
	}

	sm, err := sessions.New(db,
		models.DEFAULT_SESSIONS_TABLE,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_SESSION_EXPIRATION,
		security.GenRandomUIntNotPrime(),
	)
	if err != nil {
		T.Fatal(err)
	}

	atm, err := accesstokens.New(db,
		models.DEFAULT_ACCESS_TOKENS_TABLE,
		models.DEFAULT_USERS_TABLE,
		models.DEFAULT_ID_FIELD,
		models.DEFAULT_LONG_RESOURCE_SESSION_EXPIRATION,
		models.DEFAULT_SHORT_RESOURCE_SESSION_EXPIRATION)
	if err != nil {
		T.Fatal(err)
	}

	ds, err := datastore.New(db,
		models.DEFAULT_DATASTORE_TABLE,
		models.DEFAULT_ID_FIELD)
	if err != nil {
		T.Fatal(err)
	}

	return &DatabaseEnv{db, um, sm, atm, ds}
}

func Path() string {
	_ = os.MkdirAll(models.DEFAULT_TEST_DATA_DIR, os.ModePerm)
	return filepath.Join(models.DEFAULT_TEST_DATA_DIR, models.DEFAULT_DATA_FILE)
}

func (dbenv *DatabaseEnv) Close() {
	dbenv.UM = nil
	dbenv.SM = nil
	dbenv.DB.Close()
}

func Clean() {
	os.RemoveAll(models.DEFAULT_TEST_DATA_DIR)
}
