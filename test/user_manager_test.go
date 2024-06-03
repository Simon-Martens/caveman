package test

import (
	"testing"

	"github.com/Simon-Martens/caveman/db/users"
)

var TestSuperAdmin = users.User{
	Name:     "Mr. Test",
	Email:    "superadmin@test.com",
	Role:     3,
	Active:   true,
	Verified: true,
}

func TestUserManager(t *testing.T) {
	Clean()
	dbenv := TestNewDatabaseEnv(t)

	hasa, err := dbenv.UM.HasAdmins()

	if hasa == true {
		t.Fatal("HasAdmins() should return false")
	}

	if err != nil {
		t.Fatal(err)
	}

	_, err = dbenv.UM.Insert(&TestSuperAdmin, "password")
	if err != nil {
		t.Fatal(err)
	}

	hasa, err = dbenv.UM.HasAdmins()
	if hasa == false {
		t.Fatal("HasAdmins() should return true")
	}

	us, err := dbenv.UM.SelectByEmail(TestSuperAdmin.Email)
	if err != nil {
		t.Fatal(err)
	}

	if us.Name != TestSuperAdmin.Name || us.Email != TestSuperAdmin.Email || us.Role != 3 || us.Active != true || us.Verified != true {
		t.Fatal("User data is not correct")
	}

	if us.HID == "" || len(us.HID) != 20 {
		t.Fatal("HID is empty")
	}

	us, err = dbenv.UM.Select(TestSuperAdmin.ID)

	if err != nil {
		t.Fatal(err)
	}

	if us.Name != TestSuperAdmin.Name || us.Email != TestSuperAdmin.Email || us.Role != 3 || us.Active != true || us.Verified != true {
		t.Fatal("User data is not correct")
	}

	us, err = dbenv.UM.SelectByHID(us.HID)
	if err != nil {
		t.Fatal(err)
	}

	if us.Name != TestSuperAdmin.Name || us.Email != TestSuperAdmin.Email || us.Role != 3 || us.Active != true || us.Verified != true {
		t.Fatal("User data is not correct")
	}

	us, err = dbenv.UM.CheckGetUser(TestSuperAdmin.Email, "password")
	if err != nil || us == nil {
		t.Fatal(err)
	}

	us.Name = "Hans Hohenstein"
	err = dbenv.UM.Update(us)
	if err != nil {
		t.Fatal(err)
	}

	us, err = dbenv.UM.Select(us.ID)
	if err != nil {
		t.Fatal(err)
	}

	if us.Name != "Hans Hohenstein" {
		t.Fatal("User data is not correct")
	}

	wus, err := dbenv.UM.CheckGetUser(TestSuperAdmin.Email, "wrongpassword")
	if err == nil || wus != nil {
		t.Fatal(err)
	}

	err = dbenv.UM.Delete(us.ID)
	if err != nil {
		t.Fatal(err)
	}

	hasa, err = dbenv.UM.HasAdmins()
	if hasa == true {
		t.Fatal("HasAdmins() should return false")
	}

	us, err = dbenv.UM.Select(us.ID)
	if us != nil || err == nil {
		t.Fatal("User should not exist")
	}

	dbenv.Close()
}
