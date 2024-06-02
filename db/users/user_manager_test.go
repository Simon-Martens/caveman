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

	_, err = um.Insert(&u, "password")
	if err != nil {
		t.Fatal(err)
	}

	hasa, err = um.HasAdmins()
	if hasa == false {
		t.Fatal("HasAdmins() should return true")
	}

	us, err := um.SelectByEmail("martens@tss-hd.de")
	if err != nil {
		t.Fatal(err)
	}

	if us.Name != "Simon" || us.Email != "martens@tss-hd.de" || us.Role != 3 || us.Active != true || us.Verified != true {
		t.Fatal("User data is not correct")
	}

	us, err = um.Select(u.ID)

	if err != nil {
		t.Fatal(err)
	}

	if us.Name != "Simon" || us.Email != "martens@tss-hd.de" || us.Role != 3 || us.Active != true || us.Verified != true {
		t.Fatal("User data is not correct")
	}

	us, err = um.CheckGetUser("martens@tss-hd.de", "password")
	if err != nil || us == nil {
		t.Fatal(err)
	}

	us.Name = "Hans Hohenstein"
	err = um.Update(us)
	if err != nil {
		t.Fatal(err)
	}

	us, err = um.Select(us.ID)
	if err != nil {
		t.Fatal(err)
	}

	if us.Name != "Hans Hohenstein" {
		t.Fatal("User data is not correct")
	}

	wus, err := um.CheckGetUser("martens@tss-hd.de", "wrongpassword")
	if err == nil || wus != nil {
		t.Fatal(err)
	}

	err = um.Delete(us.ID)
	if err != nil {
		t.Fatal(err)
	}

	hasa, err = um.HasAdmins()
	if hasa == true {
		t.Fatal("HasAdmins() should return false")
	}

	us, err = um.Select(u.ID)
	if us != nil || err == nil {
		t.Fatal("User should not exist")
	}

	db.Close()

}
