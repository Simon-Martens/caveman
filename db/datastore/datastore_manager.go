package datastore

import (
	"database/sql"
	"errors"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
	"github.com/Simon-Martens/caveman/tools/types"
	"github.com/pocketbase/dbx"
)

var ErrNotFound = errors.New("not found")

type DataStoreManager struct {
	db      *db.DB
	table   string
	idfield string
}

func New(db *db.DB, tablename, idfield string) (*DataStoreManager, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if tablename == "" {
		return nil, errors.New("table name is empty")
	}

	if idfield == "" {
		return nil, errors.New("id field name is empty")
	}

	s := &DataStoreManager{
		db:      db,
		table:   tablename,
		idfield: idfield,
	}

	err := s.createTable()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *DataStoreManager) createTable() error {
	ncdb := s.db.NonConcurrentDB()

	tn := ncdb.QuoteTableName(s.table)

	q := ncdb.NewQuery(
		"CREATE TABLE IF NOT EXISTS " +
			tn +
			" (" + s.idfield + " INTEGER PRIMARY KEY, " +
			"key TEXT NOT NULL, " +
			"data TEXT NOT NULL, " +
			"created INTEGER DEFAULT 0, " +
			"modified INTEGER DEFAULT 0);")

	_, err := q.Execute()
	if err != nil {
		return err
	}

	err = s.db.CreateIndex(s.table, "key")
	if err != nil {
		return err
	}

	q = ncdb.NewQuery("CREATE INDEX IF NOT EXISTS " +
		s.table +
		"_created_idx ON " +
		tn +
		" (created DESC);")
	_, err = q.Execute()
	if err != nil {
		return err
	}

	q = ncdb.NewQuery("CREATE INDEX IF NOT EXISTS " +
		s.table +
		"_modified_idx ON " +
		tn +
		" (modified DESC);")
	_, err = q.Execute()
	if err != nil {
		return err
	}

	return nil
}

func (s *DataStoreManager) Insert(data Data) (*DataStore, error) {
	ncdb := s.db.NonConcurrentDB()

	// Marshal data to json string
	d := types.JsonRaw{}
	if err := d.Scan(data); err != nil {
		return nil, err
	}

	sets := &DataStore{
		Record: models.NewRecord(),
		Data:   d,
		Key:    data.Key(),
	}

	if err := ncdb.Model(sets).Insert(); err != nil {
		return nil, err
	}

	return sets, nil
}

func (s *DataStoreManager) Update(id int64, data Data) error {
	ncdb := s.db.NonConcurrentDB()

	// Marshal data to json string
	d := types.JsonRaw{}
	if err := d.Scan(data); err != nil {
		return err
	}

	ds := &DataStore{
		ID:     id,
		Record: models.NewRecord(),
		Data:   d,
		Key:    data.Key(),
	}

	return ncdb.Model(ds).Update()
}

// TODO: all these functions prob cause heap allocs since we return a pointer
func (s *DataStoreManager) SelectLatest(key string) (*DataStore, error) {
	db := s.db.ConcurrentDB()

	ds := &[]DataStore{}
	err := db.NewQuery("SELECT * FROM " +
		s.table +
		" WHERE key = {:key} ORDER BY modified DESC LIMIT 1").
		Bind(dbx.Params{"key": key}).
		All(ds)
	if err != nil {
		return nil, err
	}
	if len(*ds) == 0 {
		return nil, ErrNotFound
	}

	first := (*ds)[0]

	return &first, nil
}

func (s *DataStoreManager) SelectAll(key string) ([]DataStore, error) {
	db := s.db.ConcurrentDB()

	ds := []DataStore{}
	err := db.NewQuery("SELECT * FROM " +
		s.table +
		" WHERE key = {:key} ORDER BY modified DESC").
		Bind(dbx.Params{"key": key}).
		All(&ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (s *DataStoreManager) Select(id int64) (*DataStore, error) {
	db := s.db.ConcurrentDB()

	ds := &DataStore{}
	err := db.NewQuery("SELECT * FROM " +
		s.table +
		" WHERE " + s.idfield + " = {:id}").
		Bind(dbx.Params{"id": id}).
		One(ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (s *DataStoreManager) Delete(id int64) error {
	db := s.db.NonConcurrentDB()

	tn := db.QuoteTableName(s.table)
	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE " + s.idfield + " = {:id}").
		Bind(dbx.Params{"id": id})

	_, err := q.Execute()
	return err
}

func (s *DataStoreManager) DeleteAll(key string) error {
	db := s.db.NonConcurrentDB()

	tn := db.QuoteTableName(s.table)
	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE key = {:key}").
		Bind(dbx.Params{"key": key})

	_, err := q.Execute()
	return err
}

func (s *DataStoreManager) DeleteOlderThan(unixtime int, key string) error {
	db := s.db.NonConcurrentDB()

	tn := db.QuoteTableName(s.table)
	q := db.NewQuery(
		"DELETE FROM " + tn + " WHERE modified < {:modified} AND key = {:key}").
		Bind(dbx.Params{"modified": unixtime, "key": key})

	_, err := q.Execute()
	return err
}

func (s *DataStoreManager) Count() (int, error) {
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
