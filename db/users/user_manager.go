package users

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
	"github.com/pocketbase/dbx"
	"golang.org/x/crypto/bcrypt"
)

// TODO: bcrypt has a size limit of 72 bytes for the password. This should be checked and handled.
// Otherwise, switch to argon2id, and save the salt along with the hash in the table.
var ErrUserNotFound = errors.New("user not found")
var ErrWrongPassword = errors.New("wrong password")
var ErrHIDChanged = errors.New("HID is not allowed to be changed")

type UserManager struct {
	db      *db.DB
	table   string
	idfield string
}

func New(db *db.DB, tablename, idfield string) (*UserManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if idfield == "" {
		return nil, errors.New("user table or user id column name is empty")
	}

	s := &UserManager{
		db:    db,
		table: tablename,
	}

	err := s.createTable(idfield)
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
			" (" + idfield + " INTEGER PRIMARY KEY AUTOINCREMENT, " +
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

// TODO: Implement peppering of passwords
// So hackers cant't do anything if the table is stolen but the process space is intact
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

	pusexp, _ := time.ParseDuration(strconv.Itoa(models.DEFAULT_USER_EXPIRATION) + "s")
	user.Expires = user.Created.Add(pusexp)

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
