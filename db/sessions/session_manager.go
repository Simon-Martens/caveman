package sessions

import (
	"database/sql"
	"errors"

	"github.com/Simon-Martens/caveman/db"
)

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
			"session BLOB, " +
			"session_data BLOB, " +
			"created_on INTEGER DEFAULT 0, " +
			"modified_on INTEGER DEFAULT 0, " +
			"expires_on INTEGER DEFAULT 0, " +
			"user INTEGER NOT NULL, " +
			"FOREIGN KEY(user) REFERENCES " + utn + "(" + idfield + "));",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	return nil
}
