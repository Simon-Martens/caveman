package users

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Simon-Martens/caveman/db"
	"github.com/Simon-Martens/caveman/models"
)

func TestUserManager(t *testing.T) {
	os.RemoveAll(models.DEFAULT_TEST_DATA_DIR_NAME)
	_ = os.MkdirAll(models.DEFAULT_TEST_DATA_DIR_NAME, os.ModePerm)
	p := filepath.Join(models.DEFAULT_TEST_DATA_DIR_NAME, models.DEFAULT_DATA_FILE_NAME)

	db, err := db.New(p, 120, 12)
	if err != nil {
		t.Fatal(err)
	}

	um, err := New(db, models.DEFAULT_USERS_TABLE_NAME, models.DEFAULT_ID_FIELD)
	if err != nil {
		t.Fatal(err)
	}

	hasa, err := um.HasAdmins()

	if hasa == true {
		t.Fatal("HasAdmins() should return false")
	}

	if err != nil {
		t.Fatal(err)
	}

	u := User{
		Name:     "Simon",
		Email:    "martens@tss-hd.de",
		Role:     3,
		Active:   true,
		Verified: true,
	}

	err = um.Insert(&u, "password")
	if err != nil {
		t.Fatal(err)
	}

	hasa, err = um.HasAdmins()
	if hasa == false {
		t.Fatal("HasAdmins() should return true")
	}

	db.Close()

}
