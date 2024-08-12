package testdata

type DefaultModel struct {
	ID   string `db:"id"`     // will use db tag
	Name string `json:"name"` // will use json tag
}
