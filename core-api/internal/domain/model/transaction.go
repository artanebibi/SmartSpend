package model

import (
	"SmartSpend/internal/domain/enum"
	"time"
)

type Transaction struct {
	ID         int64                `json:"id"`
	Title      string               `json:"title"`
	Price      float32              `json:"price"`
	DateMade   time.Time            `json:"date_made"`
	OwnerId    string               `json:"owner_id"`
	CategoryId *int64               `json:"category_id"`
	Type       enum.TransactionType `json:"type"`
}
