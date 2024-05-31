package models

type Settings struct {
	Record
	Icon string `json:"-"`

	Name    string `json:"name"`
	Desc    string `json:"desc"`
	URL     string `json:"url"`
	Edition string `json:"edition"`
	Contact string `json:"contact"`
}

type Config struct {
	*Settings
	Dev     bool   `json:"dev"`
	DataDir string `json:"data"`
}
