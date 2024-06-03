package sessions

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strconv"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/pocketbase/dbx"
)

var ErrSessionExpired = errors.New("session expired")
var ErrSessionNotFound = errors.New("session not found")

type SessionManager struct {
	db      *db.DB
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
			"resource STRING, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"user_id INTEGER, " +
			"FOREIGN KEY(user_id) REFERENCES " + utn + "(" + idfield + "));",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	err = s.db.CreateUniqueIndex(s.table, "session")
	if err != nil {
		return err
	}

	err = s.db.CreateIndex(s.table, "user_id")
	if err != nil {
		return err
	}

	err = s.db.CreateIndex(s.table, "resource")
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionManager) Insert(user int64, short bool) (*Session, error) {
	n := Session{
		Record: models.NewRecord(),
	}

	u := sql.NullInt64{}
	err := u.Scan(user)
	if err != nil {
		return nil, err
	}
	n.User = u

	var dexp time.Duration
	if short {
		dexp, _ = time.ParseDuration(strconv.Itoa(models.DEFAULT_SHORT_SESSION_EXPIRATION) + "s")
	} else {
		dexp, _ = time.ParseDuration(strconv.Itoa(models.DEFAULT_LONG_SESSION_EXPIRATION) + "s")
	}

	n.Expires = n.Created.Add(dexp)

	tok, err := CreateRandomToken()
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

func (s *SessionManager) InsertResource(res string) (*Session, error) {
	n := Session{
		Record: models.NewRecord(),
	}

	dexp, _ := time.ParseDuration(strconv.Itoa(models.DEFAULT_RESOURCE_SESSION_EXPIRATION) + "s")
	n.Expires = n.Created.Add(dexp)

	tok, err := CreateRandomToken()
	if err != nil {
		return nil, err
	}

	n.Session = tok
	r := sql.NullString{}
	err = r.Scan(res)
	if err != nil {
		return nil, err
	}
	n.Resource = r

	db := s.db.NonConcurrentDB()
	err = db.Model(&n).Insert()
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *SessionManager) DeleteBySession(session string) error {
	db := s.db.NonConcurrentDB()
	tn := db.QuoteTableName(s.table)

	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE session = {:id}").
		Bind(dbx.Params{"id": session})

	_, err := q.Execute()
	return err
}

func (s *SessionManager) SelectBySession(session string) (*Session, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	se := Session{}

	err := db.NewQuery(
		"SELECT * FROM " + tn + " WHERE session = {:id} LIMIT 1").
		Bind(dbx.Params{"id": session}).
		One(&se)

	if err != nil {
		return nil, ErrSessionNotFound
	}

	if se.Expires.Time().Before(time.Now()) {
		s.DeleteBySession(se.Session)
		return nil, ErrSessionExpired
	}

	return &se, nil
}

func CreateRandomToken() (string, error) {
	// We use 256 bits of crypto/rand to generate a random token
	// We append the timestamp to make sure our seed is unique
	// Then we hash the result with sha512
	t := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(t, time.Now().Unix())

	c := 256
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	all := append(b, t...)
	hash := sha512.Sum512(all)

	// Sadly, 64 bits dont align to 6 bits, so there will be some padding
	bas := base64.URLEncoding.EncodeToString(hash[:])
	return bas, nil
}
