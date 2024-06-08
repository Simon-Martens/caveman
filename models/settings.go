package models

import "github.com/Simon-Martens/caveman/tools/security"

type Settings struct {
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	URL     string `json:"url"`
	Edition string `json:"edition"`
	Contact string `json:"contact"`

	UserSeed    uint64 `json:"user_seed"`
	SessionSeed uint64 `json:"session_seed"`
}

func (s *Settings) Key() string {
	return DATASTORE_SETTINGS_KEY
}

func DefaultSettings() *Settings {
	return &Settings{
		URL:         "http://localhost:8080",
		UserSeed:    security.GenRandomUIntNotPrime(),
		SessionSeed: security.GenRandomUIntNotPrime(),
	}
}

type Config struct {
	*Settings
	Dev     bool   `json:"dev"`
	DataDir string `json:"data"`
}
