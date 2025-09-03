package model

type Category struct {
	ID   int8   `gorm:"type:text;primaryKey" json:"id"`
	Name string `gorm:"size:100" json:"first_name"`
}
