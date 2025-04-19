package repository

import (
	"context"
	"time"

	"github.com/whitewalker-sa/ehass/internal/model"
	"gorm.io/gorm"
)

// authRepository implements AuthRepository interface
type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) RegisterUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindUserByProviderID(ctx context.Context, provider model.AuthProvider, providerID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("provider = ? AND provider_id = ?", provider, providerID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *authRepository) VerifyEmail(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("email_verified", true).Error
}

func (r *authRepository) CreateOAuthUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *authRepository) LinkUserToProvider(ctx context.Context, userID uint, provider model.AuthProvider, providerID string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"provider":    provider,
			"provider_id": providerID,
		}).Error
}

func (r *authRepository) CreateVerificationToken(ctx context.Context, token *model.VerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) FindVerificationToken(ctx context.Context, token string, tokenType model.TokenType) (*model.VerificationToken, error) {
	var verificationToken model.VerificationToken
	err := r.db.WithContext(ctx).Where("token = ? AND type = ? AND expires_at > ?", token, tokenType, time.Now()).First(&verificationToken).Error
	if err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *authRepository) DeleteVerificationToken(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.VerificationToken{}, id).Error
}

func (r *authRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&model.VerificationToken{}).Error
}

func (r *authRepository) Enable2FA(ctx context.Context, userID uint, secret string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"two_factor_auth": true,
			"secret2fa":       secret,
		}).Error
}

func (r *authRepository) Disable2FA(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"two_factor_auth": false,
			"secret2fa":       "",
		}).Error
}

func (r *authRepository) Update2FASecret(ctx context.Context, userID uint, secret string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("secret2fa", secret).Error
}

func (r *authRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("last_login", &now).Error
}

func (r *authRepository) UpdateRefreshToken(ctx context.Context, userID uint, token string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("refresh_token", token).Error
}

// FindByID finds a user by their ID
func (r *authRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
