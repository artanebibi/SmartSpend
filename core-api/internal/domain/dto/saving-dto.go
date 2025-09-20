package dto

import "time"

type SavingDto struct {
	ID     int64      `json:"id"`
	Amount *float32   `json:"amount"`
	From   *time.Time `json:"from"`
	To     *time.Time `json:"to"`
}
