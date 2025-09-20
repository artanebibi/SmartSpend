package model

import "time"

type Saving struct {
	ID      int64     `json:"id"`
	OwnerId string    `json:"owner_id"`
	Amount  float32   `json:"amount"`
	From    time.Time `json:"from"`
	To      time.Time `json:"to"`
}
