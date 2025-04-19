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

// AuthProvider represents the authentication provider for OAuth
type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGithub AuthProvider = "github"
	AuthProviderGoogle AuthProvider = "google"
)

// User represents a user in the system
type User struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	Name          string       `json:"name" gorm:"size:100;not null"`
	Email         string       `json:"email" gorm:"size:100;uniqueIndex;not null"`
	EmailVerified bool         `json:"emailVerified" gorm:"default:false"`
	PasswordHash  string       `json:"-" gorm:"size:255"`
	Role          Role         `json:"role" gorm:"size:20;not null"`
	Phone         string       `json:"phone" gorm:"size:20"`
	Address       string       `json:"address" gorm:"size:255"`
	Provider      AuthProvider `json:"provider" gorm:"size:20;default:'local'"`
	ProviderID    string       `json:"providerId" gorm:"size:100"`
	RefreshToken  string       `json:"-" gorm:"size:255"`
	Avatar        string       `json:"avatar" gorm:"size:255"`
	TwoFactorAuth bool         `json:"twoFactorAuth" gorm:"default:false"`
	Secret2FA     string       `json:"-" gorm:"size:100"`
	LastLogin     *time.Time   `json:"lastLogin"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// SanitizeUser removes sensitive data from user for response
func SanitizeUser(user User) map[string]interface{} {
	return map[string]interface{}{
		"id":            user.ID,
		"name":          user.Name,
		"email":         user.Email,
		"emailVerified": user.EmailVerified,
		"role":          user.Role,
		"phone":         user.Phone,
		"address":       user.Address,
		"provider":      user.Provider,
		"avatar":        user.Avatar,
		"twoFactorAuth": user.TwoFactorAuth,
		"lastLogin":     user.LastLogin,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
	}
}
