package test

import "testing"

func TestAccessTokenManager(t *testing.T) {
	Clean()
	_ = TestNewDatabaseEnv(t)

}
