package models

type Settings struct {
	Record
	Icon string `json:"-"`

	Name    string `json:"name"`
	Desc    string `json:"desc"`
	URL     string `json:"url"`
	Edition string `json:"edition"`
	Contact string `json:"contact"`

	DataDir string `json:"data_dir"`
	Dev     bool   `json:"dev"`
}
