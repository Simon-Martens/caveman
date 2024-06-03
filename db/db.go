package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/fatih/color"
	"github.com/pocketbase/dbx"
)

type DB struct {
	concurrentDB    *dbx.DB
	nonConcurrentDB *dbx.DB
}

func New(filepath string, max_conn, max_idle_conn int) (*DB, error) {
	concurrentDB, err := ConnectDB(filepath)
	if err != nil {
		return nil, err
	}

	concurrentDB.DB().SetMaxOpenConns(max_conn)
	concurrentDB.DB().SetMaxIdleConns(max_idle_conn)
	concurrentDB.DB().SetConnMaxIdleTime(3 * time.Minute)

	nonconcurrentDB, err := ConnectDB(filepath)
	if err != nil {
		return nil, err
	}
	nonconcurrentDB.DB().SetMaxOpenConns(1)
	nonconcurrentDB.DB().SetMaxIdleConns(1)
	nonconcurrentDB.DB().SetConnMaxIdleTime(3 * time.Minute)

	return &DB{
		concurrentDB:    concurrentDB,
		nonConcurrentDB: nonconcurrentDB,
	}, nil
}

func (db *DB) Close() error {
	if db.concurrentDB != nil {
		if err := db.concurrentDB.Close(); err != nil {
			return err
		}
	}

	if db.nonConcurrentDB != nil {
		if err := db.nonConcurrentDB.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) ConnectLogger() {
	db.nonConcurrentDB.QueryLogFunc = queryLogFunc
	db.nonConcurrentDB.ExecLogFunc = execLogFunc
	db.concurrentDB.QueryLogFunc = queryLogFunc
	db.concurrentDB.ExecLogFunc = execLogFunc
}

func (db *DB) DisconnectLogger() {
	db.nonConcurrentDB.QueryLogFunc = nil
	db.nonConcurrentDB.ExecLogFunc = nil
	db.concurrentDB.QueryLogFunc = nil
	db.concurrentDB.ExecLogFunc = nil
}

func (db *DB) ConcurrentDB() *dbx.DB {
	return db.concurrentDB
}

func (db *DB) NonConcurrentDB() *dbx.DB {
	return db.nonConcurrentDB
}

func (db *DB) CreateUniqueIndex(table, field string) error {
	tb := db.nonConcurrentDB.QuoteTableName(table)
	q := db.nonConcurrentDB.NewQuery("CREATE UNIQUE INDEX IF NOT EXISTS " + table + "_" + field + "_idx ON " + tb + " (" + field + ")")
	_, err := q.Execute()
	return err
}

func (db *DB) CreateIndex(table, field string) error {
	tb := db.nonConcurrentDB.QuoteTableName(table)
	q := db.nonConcurrentDB.NewQuery("CREATE INDEX IF NOT EXISTS " + table + "_" + field + "_idx ON " + tb + " (" + field + ")")
	_, err := q.Execute()
	return err
}

func queryLogFunc(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
	color.HiBlue("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
}

func execLogFunc(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
	color.HiCyan("[%.2fms] %v\n", float64(t.Milliseconds()), sql)
}
