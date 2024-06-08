package test

import (
	"testing"

	"github.com/Simon-Martens/caveman/db/accesstokens"
)

func TestAccessTokenManager(t *testing.T) {
	Clean()
	d := TestNewDatabaseEnv(t)

	u, err := d.UM.Insert(&TestSuperAdmin, "password")
	if err != nil {
		t.Fatal(err)
	}

	at, err := d.ATM.InsertEternal(u.ID, "/")
	if err != nil || at == nil {
		t.Fatal(err)
	}

	at2, err := d.ATM.Insert(u.ID, 999, "/", true)
	if err != nil || at2 == nil {
		t.Fatal(err)
	}

	at3, err := d.ATM.Insert(u.ID, 1, "/", false)
	if err != nil || at3 == nil {
		t.Fatal(err)
	}

	c, err := d.ATM.Count()
	if err != nil || c != 3 {
		t.Fatal(err)
	}

	at_rec, err := d.ATM.SelectByAccessToken(at.Token, "/")
	if err != nil || at_rec == nil {
		t.Fatal(err)
	}

	if at_rec.ID != at.ID || at_rec.Uses != 0 {
		t.Fatal("at_rec: ", at_rec)
	}

	c, err = d.ATM.Count()
	if err != nil || c != 3 {
		t.Fatal("err: ", err, " c: ", c)
	}

	at_rec, err = d.ATM.SelectByAccessToken(at.Token, "/")
	if err == nil || at_rec != nil {
		t.Fatal("We cannot select a token that has been used up")
	}

	c, err = d.ATM.Count()
	if err != nil || c != 2 {
		t.Fatal(err)
	}

	iat := &accesstokens.AccessToken{
		Token: "jens",
		Uses:  999,
		Path:  "/",
	}
	err = d.ATM.InsertUnsafe(iat)
	if err == nil {
		t.Fatal(err)
	}

	iat.Creator = u.ID
	err = d.ATM.InsertUnsafe(iat)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", iat)

	c, err = d.ATM.Count()
	if err != nil || c != 3 {
		t.Fatal(err)
	}

	iat_rec, err := d.ATM.SelectByAccessToken(iat.Token, "/")
	if err != nil || iat_rec == nil {
		t.Fatal(err)
	}

}
