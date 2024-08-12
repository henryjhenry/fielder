package testdata

type GormModel struct {
	ID   int64  `gorm:"column:id"`
	Name string `gorm:"column:name;type:varchar;size:255" json:"name"`
}
