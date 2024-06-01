package users

import (
	"database/sql"
	"errors"

	"github.com/Simon-Martens/caveman/db"
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
			"password BLOB NOT NULL, " +
			"role INTEGER DEFAULT 0, " +
			"created_on INTEGER DEFAULT 0, " +
			"modified_on INTEGER DEFAULT 0, " +
			"expires_on INTEGER DEFAULT 0);",
	)

	_, err := q.Execute()
	if err != nil {
		return err
	}

	return nil
}
