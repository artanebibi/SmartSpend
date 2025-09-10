package dto

import (
	"SmartSpend/internal/domain/enum"
	"time"
)

type TransactionDto struct {
	ID         int64                 `json:"id"`
	Title      *string               `json:"title"`
	Price      *float32              `json:"price"`
	DateMade   *time.Time            `json:"date_made"`
	CategoryId *int64                `json:"category_id"`
	Type       *enum.TransactionType `json:"type"`
}
