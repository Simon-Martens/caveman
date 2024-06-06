package sessions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strconv"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/lcg"
	"github.com/pocketbase/dbx"
)

var ErrSessionExpired = errors.New("session expired")
var ErrSessionNotFound = errors.New("session not found")

type SessionManager struct {
	db      *db.DB
	table   string
	idfield string

	long_exp  int
	short_exp int

	HMACKey []byte
	lcg     *lcg.LCG
	seed    uint64
}

func New(db *db.DB, tablename, usertable, idfield string, l_exp, s_exp int, lcg_seed uint64) (*SessionManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if usertable == "" || idfield == "" {
		return nil, errors.New("user table or user id column name is empty")
	}

	// HMAC secret for CSRF token get lost if server is restarted
	hmacs, err := CreateHMACSecret()
	if err != nil {
		return nil, err
	}

	if lcg_seed == 0 {
		lcg_seed = lcg.GenRandomUIntNotPrime()
	}

	lcg := lcg.New(lcg_seed)

	s := &SessionManager{
		db:        db,
		table:     tablename,
		long_exp:  l_exp,
		short_exp: s_exp,
		HMACKey:   hmacs,
		lcg:       lcg,
	}

	err = s.createTable(usertable, idfield)

	c, _ := s.Count()
	if c > 0 {
		s.lcg.Skip(int64(c))
	}

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
			"session STRING NOT NULL, " +
			"session_data STRING, " +
			"ip STRING, " +
			"agent STRING, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"user_id INTEGER NOT NULL, " +
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

	return nil
}

func (s *SessionManager) InsertEternal(user int64, agent string, ip string) (*Session, error) {
	n := Session{
		Record: models.NewRecord(),
		User:   user,
		Agent:  agent,
		IP:     ip,
		ID:     int64(s.lcg.Next()),
	}

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

func (s *SessionManager) Insert(user int64, short bool, agent string, ip string) (*Session, error) {
	n := Session{
		Record: models.NewRecord(),
		User:   user,
		Agent:  agent,
		IP:     ip,
		ID:     int64(s.lcg.Next()),
	}

	var dexp time.Duration
	if short {
		dexp, _ = time.ParseDuration(strconv.Itoa(s.short_exp) + "s")
	} else {
		dexp, _ = time.ParseDuration(strconv.Itoa(s.long_exp) + "s")
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

	if !se.Expires.IsZero() && se.Expires.Time().Before(time.Now()) {
		s.DeleteBySession(se.Session)
		return nil, ErrSessionExpired
	}

	return &se, nil
}

func (s *SessionManager) Count() (int, error) {
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
	bas := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	return bas, nil
}

func CreateHMACSecret() ([]byte, error) {
	b := make([]byte, 2048)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *SessionManager) CreateCSRFToken(session *Session) string {
	t := session.Session + ":" + session.Created.String() + ":" + strconv.FormatInt(session.User, 10)
	mac := hmac.New(sha256.New, s.HMACKey)
	mac.Write([]byte(t))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(mac.Sum(nil))
}

func (s *SessionManager) ValidateCSRFToken(session *Session, token string) bool {
	t := session.Session + ":" + session.Created.String() + ":" + strconv.FormatInt(session.User, 10)
	mac := hmac.New(sha256.New, s.HMACKey)
	mac.Write([]byte(t))
	expected := mac.Sum(nil)
	actual, _ := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(token)
	return hmac.Equal(expected, actual)
}
