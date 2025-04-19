package repository

import (
	"context"

	"github.com/whitewalker-sa/ehass/internal/model"
)

// AuthRepository defines operations for authentication
type AuthRepository interface {
	RegisterUser(ctx context.Context, user *model.User) error
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUserByProviderID(ctx context.Context, provider model.AuthProvider, providerID string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	VerifyEmail(ctx context.Context, userID uint) error

	// OAuth related
	CreateOAuthUser(ctx context.Context, user *model.User) error
	LinkUserToProvider(ctx context.Context, userID uint, provider model.AuthProvider, providerID string) error

	// Token management
	CreateVerificationToken(ctx context.Context, token *model.VerificationToken) error
	FindVerificationToken(ctx context.Context, token string, tokenType model.TokenType) (*model.VerificationToken, error)
	DeleteVerificationToken(ctx context.Context, id uint) error
	DeleteExpiredTokens(ctx context.Context) error

	// 2FA related
	Enable2FA(ctx context.Context, userID uint, secret string) error
	Disable2FA(ctx context.Context, userID uint) error
	Update2FASecret(ctx context.Context, userID uint, secret string) error

	// Session management
	UpdateLastLogin(ctx context.Context, userID uint) error
	UpdateRefreshToken(ctx context.Context, userID uint, token string) error
}
