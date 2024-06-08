package test

import (
	"encoding/json"
	"testing"
)

type DataStoreTestData struct {
	Thing        string
	AnotherThing string
	AthirdThing  []string
}

func (d DataStoreTestData) Key() string {
	return "key"
}

func TestDataStoreManager(t *testing.T) {
	Clean()
	dbenv := TestNewDatabaseEnv(t)
	testdata := &DataStoreTestData{
		Thing:        "thing",
		AnotherThing: "another thing",
		AthirdThing:  []string{"a", "b", "c"},
	}

	testdata_later := &DataStoreTestData{
		Thing:        "later thing",
		AnotherThing: "later another thing",
		AthirdThing:  []string{"d", "e", "f"},
	}

	ds, err := dbenv.DSM.Insert(testdata)
	if err != nil || ds == nil {
		t.Fatal(err)
	}

	ds2, err := dbenv.DSM.Insert(testdata_later)
	if err != nil || ds2 == nil {
		t.Fatal(err)
	}

	testdate_rec, err := dbenv.DSM.SelectLatest(testdata.Key())
	if err != nil || testdate_rec == nil {
		t.Fatal(err)
	}

	testdata_copy := &DataStoreTestData{}

	err = json.Unmarshal(testdate_rec.Data, testdata_copy)
	if err != nil {
		t.Fatal(err)
	}

	if testdata_later.Thing != testdata_copy.Thing {
		t.Fatal("Thing is not correct")
	}

	if testdata_later.AnotherThing != testdata_copy.AnotherThing {
		t.Fatal("AnotherThing is not correct")
	}

	if len(testdata_later.AthirdThing) != len(testdata_copy.AthirdThing) {
		t.Fatal("AthirdThing is not correct")
	}
}
