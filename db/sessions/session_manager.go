package sessions

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/pocketbase/dbx"
)

type SessionExpiredErr struct{}

func (e SessionExpiredErr) Error() string {
	return "session expired"
}

type SessionManager struct {
	db *db.DB

	stmtInsert *sql.Stmt
	stmtDelete *sql.Stmt
	stmtUpdate *sql.Stmt
	stmtSelect *sql.Stmt

	table   string
	idfield string
}

func New(db *db.DB, tablename, usertable, idfield string) (*SessionManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if usertable == "" || idfield == "" {
		return nil, errors.New("user table or user id column name is empty")
	}

	s := &SessionManager{
		db:    db,
		table: tablename,
	}

	err := s.createTable(usertable, idfield)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SessionManager) createTable(usertable, idfield string) error {
	ncdb := s.db.NonConcurrentDB()

	tn := ncdb.QuoteTableName(s.table)
	utn := ncdb.QuoteTableName(usertable)

	q := ncdb.NewQuery(
		"CREATE TABLE IF NOT EXISTS " +
			tn +
			" (" + idfield + " INTEGER PRIMARY KEY NOT NULL, " +
			"session BLOB NOT NULL, " +
			"session_data STRING, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"user_id INTEGER NOT NULL, " +
			"FOREIGN KEY(user) REFERENCES " + utn + "(" + idfield + "));",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	q = ncdb.NewQuery(
		"CREATE UNIQUE INDEX IF NOT EXISTS session_idx ON " + tn + "(session);")
	_, err = q.Execute()
	if err != nil {
		return err
	}

	q = ncdb.NewQuery(
		"CREATE INDEX IF NOT EXISTS user_session_idx ON " + tn + "(user_id);")
	_, err = q.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionManager) Insert(user int, short bool) (*Session, error) {
	n := Session{
		Record: models.NewRecord(),
	}

	n.User = user

	var dexp time.Duration
	if short {
		dexp, _ = time.ParseDuration(strconv.Itoa(models.DEFAULT_SHORT_SESSION_EXPIRATION) + "s")
	} else {
		dexp, _ = time.ParseDuration(strconv.Itoa(models.DEFAULT_LONG_SESSION_EXPIRATION) + "s")
	}

	n.Expires = n.Created.Add(dexp)

	tok, err := createRandomToken()
	if err != nil {
		return nil, err
	}

	n.Session = tok

	db := s.db.NonConcurrentDB()
	err = db.Model(&n).Insert()
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *SessionManager) Delete(session string) error {
	db := s.db.NonConcurrentDB()
	tn := db.QuoteTableName(s.table)

	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE session = {:id}").
		Bind(dbx.Params{"id": session})

	_, err := q.Execute()
	return err
}

func (s *SessionManager) Select(session string) (*Session, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	se := Session{}

	err := db.NewQuery(
		"SELECT * FROM " + tn + " WHERE session = {:id} LIMIT 1").
		Bind(dbx.Params{"id": session}).
		One(&se)

	if err != nil {
		return nil, err
	}

	return &se, nil
}

func createRandomToken() (string, error) {
	c := 256
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	bas := base64.URLEncoding.EncodeToString(b)
	return bas, nil
}
