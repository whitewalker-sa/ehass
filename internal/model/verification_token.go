package model

import (
	"time"
)

// TokenType represents the type of verification token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
)

// VerificationToken represents tokens for email verification and password reset
type VerificationToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"not null"`
	Token     string    `json:"token" gorm:"size:255;uniqueIndex;not null"`
	Type      TokenType `json:"type" gorm:"size:50;not null"`
	ExpiresAt time.Time `json:"expiresAt" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
}

// TableName overrides the table name
func (VerificationToken) TableName() string {
	return "verification_tokens"
}
