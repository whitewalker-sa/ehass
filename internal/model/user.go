package model

import (
	"time"
)

// Role represents user roles in the system
type Role string

const (
	RolePatient Role = "patient"
	RoleDoctor  Role = "doctor"
	RoleAdmin   Role = "admin"
)

// User represents a user in the system
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	Email        string    `json:"email" gorm:"size:100;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	Role         Role      `json:"role" gorm:"size:20;not null"`
	Phone        string    `json:"phone" gorm:"size:20"`
	Address      string    `json:"address" gorm:"size:255"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// SanitizeUser removes sensitive data from user for response
func SanitizeUser(user User) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"role":       user.Role,
		"phone":      user.Phone,
		"address":    user.Address,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
