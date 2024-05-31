package db

import "github.com/jmoiron/sqlx"

type DB struct {
	ConcurrentDB    *sqlx.DB
	NonConcurrentDB *sqlx.DB
}

func New(filepath string) (*DB, error) {
	concurrentDB, err := ConnectDB(filepath)
	if err != nil {
		return nil, err
	}

	nonconcurrentDB, err := ConnectDB(filepath)
	if err != nil {
		return nil, err
	}

	return &DB{
		ConcurrentDB:    concurrentDB,
		NonConcurrentDB: nonconcurrentDB,
	}, nil
}

func (db *DB) Close() error {
	if db.ConcurrentDB != nil {
		if err := db.ConcurrentDB.Close(); err != nil {
			return err
		}
	}

	if db.NonConcurrentDB != nil {
		if err := db.NonConcurrentDB.Close(); err != nil {
			return err
		}
	}
	return nil
}
