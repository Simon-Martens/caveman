package users

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/lcg"
	"github.com/Simon-Martens/caveman/tools/security"
	"github.com/Simon-Martens/caveman/tools/types"
	"github.com/pocketbase/dbx"
	"golang.org/x/crypto/bcrypt"
)

// INFO: bcrypt has a size limit of 72 bytes for the password. This should be checked and handled.
// Apple strong passwords contain 71 bits of entropy, we should be fine with this approach.
// ALT: switch to argon2id, and save the salt along with the hash in the table.
var ErrUserNotFound = errors.New("user not found")
var ErrWrongPassword = errors.New("wrong password")
var ErrHIDChanged = errors.New("HID is not allowed to be changed")

type UserManager struct {
	db      *db.DB
	table   string
	idfield string

	user_exp int
	lcg      *lcg.LCG
}

func New(db *db.DB, tablename, idfield string, user_exp int, lcg_seed uint64) (*UserManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if idfield == "" {
		return nil, errors.New("user table or user id column name is empty")
	}

	if lcg_seed == 0 {
		lcg_seed = security.GenRandomUIntNotPrime()
	}

	lcg := lcg.New(lcg_seed)

	s := &UserManager{
		db:    db,
		table: tablename,
		lcg:   lcg,
	}

	err := s.createTable(idfield)

	c, _ := s.Count()
	if c > 0 {
		s.lcg.Skip(int64(c))
	}

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *UserManager) createTable(idfield string) error {
	ncdb := s.db.NonConcurrentDB()

	tn := ncdb.QuoteTableName(s.table)

	q := ncdb.NewQuery(
		"CREATE TABLE IF NOT EXISTS " +
			tn +
			" (" + idfield + " INTEGER PRIMARY KEY, " +
			"email TEXT, " +
			"name TEXT, " +
			"user_data BLOB, " +
			"avatar TEXT, " +
			"password BLOB, " +
			"role INTEGER DEFAULT 0, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"last_seen INTEGER DEFAULT 0, " +
			"active BOOLEAN DEFAULT TRUE, " +
			"verified BOOLEAN DEFAULT FALSE);",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	err = s.db.CreateUniqueIndex(s.table, "email")
	if err != nil {
		return err
	}

	err = s.db.CreateUniqueIndex(s.table, idfield)
	if err != nil {
		return err
	}

	return err
}

func (s *UserManager) Select(id int64) (*User, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	user := User{}
	err := db.
		NewQuery("SELECT * FROM " + tn + " WHERE id = {:id} LIMIT 1").
		Bind(dbx.Params{"id": id}).
		One(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserManager) SelectByEmail(email string) (*User, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	user := User{}
	err := db.
		NewQuery("SELECT * FROM " + tn + " WHERE email = {:mail} LIMIT 1").
		Bind(dbx.Params{"mail": email}).
		One(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserManager) CheckPassword(user *User, pw string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pw))
	return err
}

func (s *UserManager) CheckGetUser(email string, pw string) (*User, error) {
	user, err := s.SelectByEmail(email)

	if user == nil || err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}

	err = s.CheckPassword(user, pw)
	if err != nil {
		return nil, ErrWrongPassword
	}

	return user, nil
}

func (s *UserManager) Insert(user *User, pw string) (*User, error) {
	db := s.db.NonConcurrentDB()
	hpw, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	if err != nil {
		return nil, err
	}
	user.Password = string(hpw)
	user.Record = models.NewRecord()
	user.ID = int64(s.lcg.Next())

	pusexp := time.Duration(s.user_exp) * time.Second
	user.Expires, _ = user.Created.Add(pusexp)

	err = db.Model(user).Insert()

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserManager) Update(user *User) error {
	db := s.db.NonConcurrentDB()
	user.Modified = types.NowDateTime()
	err := db.Model(user).Update()
	return err
}

func (s *UserManager) Delete(id int64) error {
	db := s.db.NonConcurrentDB()
	q := db.
		NewQuery("DELETE FROM " + db.QuoteTableName(s.table) + " WHERE id = {:id}").
		Bind(dbx.Params{"id": id})

	_, err := q.Execute()
	return err
}

func (s *UserManager) Count() (int, error) {
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

func (s *UserManager) HasAdmins() (bool, error) {
	db := s.db.ConcurrentDB()
	tn := db.QuoteTableName(s.table)

	user := User{}
	err := db.
		NewQuery("SELECT id, role FROM " + tn + " WHERE role = 3 LIMIT 1").
		One(&user)

	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
