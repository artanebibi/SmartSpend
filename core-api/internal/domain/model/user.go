package model

import (
	"SmartSpend/internal/domain/enum"
	"time"
)

type User struct {
	ID                     string        `gorm:"type:text;primaryKey"`
	FirstName              string        `gorm:"size:100" json:"first_name"`
	LastName               string        `gorm:"size:100" json:"last_name"`
	Username               string        `gorm:"size:50;uniqueIndex" json:"username"`
	GoogleEmail            string        `gorm:"size:255" json:"google_email"`
	AppleEmail             string        `gorm:"size:255" json:"apple_email"`
	RefreshToken           string        `gorm:"size:255"`
	RefreshTokenExpiryDate time.Time     // 30 days after each log in
	AvatarURL              string        `gorm:"size:512" json:"avatar_url"`
	CreatedAt              time.Time     `gorm:"autoCreateTime" json:"created_at"`
	Balance                float64       `gorm:"number" json:"balance"`
	MonthlySavingGoal      float64       `gorm:"number" json:"monthly_saving_goal"`
	PreferredCurrency      enum.Currency `gorm:"size:255" json:"preferred_currency"`
}
