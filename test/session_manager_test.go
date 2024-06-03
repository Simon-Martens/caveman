package test

import (
	"testing"

	"github.com/Simon-Martens/caveman/db/sessions"
)

func TestSessionManager(t *testing.T) {
	Clean()
	dbenv := TestNewDatabaseEnv(t)

	// Test session manager here
	_, err := dbenv.UM.Insert(&TestSuperAdmin, "password")
	if err != nil {
		t.Fatal(err)
	}

	// Simulate Login
	user, err := dbenv.UM.CheckGetUser(TestSuperAdmin.Email, "password")
	if err != nil {
		t.Fatal(err)
	}

	// Test session manager here
	sess, err := dbenv.SM.Insert(user.ID, true)
	if err != nil {
		t.Fatal(err)
	}

	sess, err = dbenv.SM.SelectBySession(sess.Session)
	if err != nil {
		t.Fatal(err)
	}

	if !sess.User.Valid || sess.User.Int64 != user.ID {
		t.Fatal("User ID is not correct")
	}

	err = dbenv.SM.DeleteBySession(sess.Session)
	if err != nil {
		t.Fatal(err)
	}

	sess, err = dbenv.SM.SelectBySession(sess.Session)
	if err == nil || sessions.ErrSessionNotFound != err {
		t.Fatal("Session should not exist")
	}

	sess, err = dbenv.SM.InsertResource("/hello")

	dbenv.Close()
}
