package model

import "github.com/google/uuid"

type Category struct {
	ID   uuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	Name string    `gorm:"size:100" json:"first_name"`
}
