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

type UserManager struct {
	db *db.DB

	stmtInsert *sql.Stmt
	stmtDelete *sql.Stmt
	stmtUpdate *sql.Stmt
	stmtSelect *sql.Stmt

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
			"email TEXT NOT NULL, " +
			"name TEXT NOT NULL, " +
			"avatar TEXT, " +
			"password BLOB NOT NULL, " +
			"role INTEGER DEFAULT 0, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0, " +
			"expires INTEGER DEFAULT 0, " +
			"active BOOLEAN DEFAULT TRUE, " +
			"verified BOOLEAN DEFAULT FALSE);",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	q = ncdb.NewQuery(
		"CREATE UNIQUE INDEX IF NOT EXISTS email_idx ON " +
			tn +
			"(email);")
	_, err = q.Execute()
	if err != nil {
		return err
	}

	q = ncdb.NewQuery(
		"CREATE UNIQUE INDEX IF NOT EXISTS " + idfield + "_idx ON " +
			tn +
			"(" + idfield + ");")

	_, err = q.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (s *UserManager) Select(id int) (*User, error) {
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
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	err = s.CheckPassword(user, pw)
	if err != nil {
		return nil, err
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

func (s *UserManager) Delete(id int) error {
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

	return user.ID > 0, nil
}
