package accesstokens

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/security"
	"github.com/Simon-Martens/caveman/tools/types"
	"github.com/pocketbase/dbx"
)

var ErrAccessTokenExpired = errors.New("access token expired")
var ErrAccessTokenNotFound = errors.New("access token not found")
var ErrAccessTokenReused = errors.New("access token reuse")
var ErrAccessTokenInvalidPath = errors.New("wrong path for access token")
var ErrUserInvalid = errors.New("user invalid")
var PathInvalid = errors.New("path invalid")

type AccessTokenManager struct {
	db      *db.DB
	table   string
	idfield string

	long_exp  int
	short_exp int
}

func New(db *db.DB, tablename, usertable, idfield string, l_exp, s_exp int) (*AccessTokenManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if usertable == "" || idfield == "" {
		return nil, errors.New("user table or user id column name is empty")
	}

	s := &AccessTokenManager{
		db:        db,
		table:     tablename,
		long_exp:  l_exp,
		short_exp: s_exp,
	}

	err := s.createTable(usertable, idfield)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *AccessTokenManager) createTable(usertable, idfield string) error {
	ncdb := s.db.NonConcurrentDB()

	tn := ncdb.QuoteTableName(s.table)
	utn := ncdb.QuoteTableName(usertable)

	q := ncdb.NewQuery(
		"CREATE TABLE IF NOT EXISTS " +
			tn +
			" (" + idfield + " INTEGER PRIMARY KEY NOT NULL, " +
			"token TOKEN NOT NULL COLLATE BINARY, " +
			"token_data TEXT, " +
			"path STRING NOT NULL, " +
			"created INTEGER DEFAULT 0, " +
			"uses INTEGER DEFAULT 99999999, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"creator_id INTEGER NOT NULL, " +
			"FOREIGN KEY(creator_id) REFERENCES " + utn + "(" + idfield + "));",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	err = s.db.CreateUniqueIndex(s.table, "token")
	if err != nil {
		return err
	}

	err = s.db.CreateIndex(s.table, "creator_id")
	if err != nil {
		return err
	}

	err = s.db.CreateIndex(s.table, "path")
	if err != nil {
		return err
	}

	return nil
}

func (s *AccessTokenManager) Count() (int, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	c := models.Count{}

	err := db.NewQuery(
		"SELECT COUNT(*) AS count FROM " + tn).One(&c)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return c.Count, nil
}

// Creating an AT with user defined values is considered unsafe
func (s *AccessTokenManager) InsertUnsafe(at *AccessToken) error {
	if at == nil {
		return errors.New("at is nil")
	}

	if at.Creator == 0 {
		return ErrUserInvalid
	}

	if at.Path == "" {
		return PathInvalid
	}

	db := s.db.NonConcurrentDB()
	return db.Model(at).Insert()
}

// TODO: maybe eternal ats are a bad idea
// Eternal ATs can only used once, but they never expire
func (s *AccessTokenManager) InsertEternal(user int64, path string) (*AccessToken, error) {
	n := AccessToken{
		Record:  models.NewRecord(),
		Creator: user,
		Uses:    1,
		Path:    path,
	}

	tok, err := security.CreateRandomSHA256Token()
	if err != nil {
		return nil, err
	}

	n.Token = tok

	db := s.db.NonConcurrentDB()
	err = db.Model(&n).Insert()
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *AccessTokenManager) Insert(user int64, uses int64, path string, short bool) (*AccessToken, error) {
	n := AccessToken{
		Record:  models.NewRecord(),
		Creator: user,
		Uses:    uses,
		Path:    path,
	}

	var dexp time.Duration
	if short {
		dexp = time.Duration(s.short_exp) * time.Second
	} else {
		dexp = time.Duration(s.long_exp) * time.Second
	}

	n.Expires, _ = n.Created.Add(dexp)

	tok, err := security.CreateRandomSHA256Token()
	if err != nil {
		return nil, err
	}

	n.Token = tok

	db := s.db.NonConcurrentDB()
	err = db.Model(&n).Insert()
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *AccessTokenManager) DeleteByAccessToken(token string) error {
	db := s.db.NonConcurrentDB()
	tn := db.QuoteTableName(s.table)

	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE token = {:id}").
		Bind(dbx.Params{"id": token})

	_, err := q.Execute()
	return err
}

// We do not allow selection by AT without
//
//   - checking the path
//   - checking the expiration
//   - checking & decreasing use
func (s *AccessTokenManager) SelectByAccessToken(token string, path string) (*AccessToken, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	se := AccessToken{}

	err := db.NewQuery(
		"SELECT * FROM " + tn + " WHERE token = {:id} LIMIT 1").
		Bind(dbx.Params{"id": token}).
		One(&se)

	if err != nil {
		return nil, ErrAccessTokenNotFound
	}

	if !se.Expires.IsZero() && se.Expires.Time().Before(time.Now()) {
		s.DeleteByAccessToken(se.Token)
		return nil, ErrAccessTokenExpired
	}

	if se.Path != path {
		s.DeleteByAccessToken(se.Token)
		return nil, ErrAccessTokenInvalidPath
	}

	// TODO: This means I get notified on token reuse, but we we still will have a lot of
	// tokens in the database with Uses = 0. Maybe instead delete after decresing se.Uses?
	if se.Uses < 1 {
		s.DeleteByAccessToken(se.Token)
		return nil, ErrAccessTokenReused
	} else {
		se.Uses = se.Uses - 1
		s.Update(&se)
	}

	return &se, nil
}

func (atm *AccessTokenManager) Update(at *AccessToken) error {
	db := atm.db.NonConcurrentDB()
	at.Modified = types.NowDateTime()
	err := db.Model(at).Update()
	return err
}
