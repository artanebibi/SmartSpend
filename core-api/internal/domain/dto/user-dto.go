package dto

import (
	"SmartSpend/internal/domain/enum"
	"time"
)

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	MKD Currency = "MKD"
)

type UserDto struct {
	FirstName         string        `json:"first_name"`
	LastName          string        `json:"last_name"`
	Username          string        `json:"username"`
	GoogleEmail       string        `json:"google_email"`
	AppleEmail        string        `json:"apple_email"`
	AvatarURL         string        `json:"avatar_url"`
	CreatedAt         time.Time     `json:"created_at"`
	Balance           float64       `json:"balance"`
	MonthlySavingGoal float64       `json:"monthly_saving_goal"`
	PreferredCurrency enum.Currency `json:"preferred_currency"`
}

type UpdateUserDto struct {
	FirstName         *string        `json:"first_name"`
	LastName          *string        `json:"last_name"`
	Username          *string        `json:"username"`
	AvatarURL         *string        `json:"avatar_url"`
	Balance           *float64       `json:"balance"`
	MonthlySavingGoal *float64       `json:"monthly_saving_goal"`
	PreferredCurrency *enum.Currency `json:"preferred_currency"`
}
