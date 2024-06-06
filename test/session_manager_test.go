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
	sess, err := dbenv.SM.Insert(user.ID, true, "User-Agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	sess, err = dbenv.SM.SelectBySession(sess.Session)
	if err != nil {
		t.Fatal(err)
	}

	if sess.User != user.ID {
		t.Fatal("User ID is not correct")
	}

	// Test CSRF Tokens
	csrf := dbenv.SM.CreateCSRFToken(sess)
	if csrf == "" {
		t.Fatal("CSRF Token is empty")
	}

	if !dbenv.SM.ValidateCSRFToken(sess, csrf) {
		t.Fatal("CSRF Token is invalid")
	}

	sess2, err := dbenv.SM.Insert(sess.User, false, "User-Agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	if dbenv.SM.ValidateCSRFToken(sess2, csrf) {
		t.Fatal("CSRF Token is valid for another session")
	}

	// TODO: session expiration test
	sess3, err := dbenv.SM.InsertEternal(sess.User, "User-Agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	sess3_select, err := dbenv.SM.SelectBySession(sess3.Session)
	if err != nil {
		t.Fatal(err)
	}

	if !sess3_select.Expires.IsZero() {
		t.Fatal("Session should be eternal")
	}

	err = dbenv.SM.DeleteBySession(sess.Session)
	if err != nil {
		t.Fatal(err)
	}

	sess, err = dbenv.SM.SelectBySession(sess.Session)
	if err == nil || sessions.ErrSessionNotFound != err {
		t.Fatal("Session should not exist")
	}

	dbenv.Close()
}
