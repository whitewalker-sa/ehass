package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"github.com/whitewalker-sa/ehass/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

// authService implements the AuthService interface
type authService struct {
	authRepo      repository.AuthRepository
	jwtSecret     string
	jwtExpiration int
	emailService  EmailService // Interface for sending emails
	oauthService  OAuthService // Interface for handling OAuth providers
}

// NewAuthService creates a new auth service
func NewAuthService(
	authRepo repository.AuthRepository,
	jwtSecret string,
	jwtExpiration int,
	emailService EmailService,
	oauthService OAuthService,
) AuthService {
	return &authService{
		authRepo:      authRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
		emailService:  emailService,
		oauthService:  oauthService,
	}
}

// Register implements the user registration flow
func (s *authService) Register(ctx context.Context, name, email, password string, role model.Role) (*model.User, error) {
	// Check if user exists
	existingUser, err := s.authRepo.FindUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Name:          name,
		Email:         email,
		PasswordHash:  string(hashedPassword),
		Role:          role,
		Provider:      model.AuthProviderLocal,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.authRepo.RegisterUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Generate verification token
	token := utils.GenerateRandomToken(32)
	verificationToken := &model.VerificationToken{
		UserID:    user.ID,
		Token:     token,
		Type:      model.TokenTypeEmailVerification,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token valid for 24 hours
		CreatedAt: time.Now(),
	}

	if err := s.authRepo.CreateVerificationToken(ctx, verificationToken); err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(ctx, user.Email, user.Name, token); err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return user, nil
}

// Login implements the login flow
func (s *authService) Login(ctx context.Context, email, password string) (string, string, *model.User, error) {
	// Find user by email
	user, err := s.authRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", "", nil, errors.New("invalid email or password")
	}

	// Check if user is using OAuth only
	if user.PasswordHash == "" && user.Provider != model.AuthProviderLocal {
		return "", "", nil, fmt.Errorf("please login with %s", user.Provider)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", nil, errors.New("invalid email or password")
	}

	// Check if email is verified
	if !user.EmailVerified {
		return "", "", nil, errors.New("email not verified, please verify your email first")
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update refresh token and last login
	if err := s.authRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return "", "", nil, fmt.Errorf("failed to update refresh token: %w", err)
	}

	if err := s.authRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Check if 2FA is enabled
	if user.TwoFactorAuth {
		return "", "", user, errors.New("two-factor authentication required")
	}

	return accessToken, refreshToken, user, nil
}

// RefreshToken implements token refresh flow
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Find user by refresh token
	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	// Convert Subject from string to uint
	userID, err := utils.StringToUint(claims.Subject)
	if err != nil {
		return "", "", errors.New("invalid user ID in token")
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := s.generateTokens(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update refresh token
	if err := s.authRepo.UpdateRefreshToken(ctx, userID, newRefreshToken); err != nil {
		return "", "", fmt.Errorf("failed to update refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// VerifyEmail implements email verification flow
func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	// Find verification token
	verificationToken, err := s.authRepo.FindVerificationToken(ctx, token, model.TokenTypeEmailVerification)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	// Verify email
	if err := s.authRepo.VerifyEmail(ctx, verificationToken.UserID); err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	// Delete verification token
	if err := s.authRepo.DeleteVerificationToken(ctx, verificationToken.ID); err != nil {
		return fmt.Errorf("failed to delete verification token: %w", err)
	}

	return nil
}

// RequestPasswordReset implements password reset request flow
func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
	// Find user by email
	user, err := s.authRepo.FindUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	token := utils.GenerateRandomToken(32)
	resetToken := &model.VerificationToken{
		UserID:    user.ID,
		Token:     token,
		Type:      model.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
		CreatedAt: time.Now(),
	}

	if err := s.authRepo.CreateVerificationToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// Send password reset email
	if err := s.emailService.SendPasswordResetEmail(ctx, user.Email, user.Name, token); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword implements password reset flow
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Find reset token
	resetToken, err := s.authRepo.FindVerificationToken(ctx, token, model.TokenTypePasswordReset)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Find user
	user, err := s.authRepo.FindByID(ctx, resetToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := s.authRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	if err := s.authRepo.DeleteVerificationToken(ctx, resetToken.ID); err != nil {
		return fmt.Errorf("failed to delete reset token: %w", err)
	}

	return nil
}

// OAuthLogin implements OAuth login flow
func (s *authService) OAuthLogin(ctx context.Context, provider model.AuthProvider, providerToken string) (string, string, *model.User, error) {
	// Get user info from OAuth provider
	oauthUser, err := s.oauthService.GetUserInfo(ctx, provider, providerToken)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get user info from %s: %w", provider, err)
	}

	// Look for existing user with the provider ID
	user, err := s.authRepo.FindUserByProviderID(ctx, provider, oauthUser.ID)

	// If user doesn't exist, check if email exists
	if err != nil {
		existingUser, err := s.authRepo.FindUserByEmail(ctx, oauthUser.Email)
		if err == nil && existingUser != nil {
			// Link OAuth account to existing user
			if err := s.authRepo.LinkUserToProvider(ctx, existingUser.ID, provider, oauthUser.ID); err != nil {
				return "", "", nil, fmt.Errorf("failed to link %s account: %w", provider, err)
			}
			user = existingUser
		} else {
			// Create new user with OAuth provider
			user = &model.User{
				Name:          oauthUser.Name,
				Email:         oauthUser.Email,
				Provider:      provider,
				ProviderID:    oauthUser.ID,
				Role:          model.RolePatient, // Default role
				EmailVerified: true,              // OAuth email is considered verified
				Avatar:        oauthUser.Avatar,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := s.authRepo.CreateOAuthUser(ctx, user); err != nil {
				return "", "", nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokens(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update refresh token and last login
	if err := s.authRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return "", "", nil, fmt.Errorf("failed to update refresh token: %w", err)
	}

	if err := s.authRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Check if 2FA is enabled
	if user.TwoFactorAuth {
		return "", "", user, errors.New("two-factor authentication required")
	}

	return accessToken, refreshToken, user, nil
}

// LinkOAuthAccount implements linking OAuth account to existing user
func (s *authService) LinkOAuthAccount(ctx context.Context, userID uint, provider model.AuthProvider, providerToken string) error {
	// Get user info from OAuth provider
	oauthUser, err := s.oauthService.GetUserInfo(ctx, provider, providerToken)
	if err != nil {
		return fmt.Errorf("failed to get user info from %s: %w", provider, err)
	}

	// Link OAuth account to user
	if err := s.authRepo.LinkUserToProvider(ctx, userID, provider, oauthUser.ID); err != nil {
		return fmt.Errorf("failed to link %s account: %w", provider, err)
	}

	return nil
}

// Setup2FA implements 2FA setup flow
func (s *authService) Setup2FA(ctx context.Context, userID uint) (string, error) {
	// Get user
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	// Generate secret
	secret, err := generateTOTPSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	// Generate QR code URI
	uri, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "EHASS",
		AccountName: user.Email,
		Secret:      []byte(secret),
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate 2FA uri: %w", err)
	}

	return uri.String(), nil
}

// Verify2FA implements 2FA verification
func (s *authService) Verify2FA(ctx context.Context, userID uint, token string) (bool, error) {
	// Get user
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify token
	valid := totp.Validate(token, user.Secret2FA)
	return valid, nil
}

// Enable2FA implements 2FA enablement
func (s *authService) Enable2FA(ctx context.Context, userID uint, secret, token string) error {
	// Verify token
	valid := totp.Validate(token, secret)
	if !valid {
		return errors.New("invalid 2FA token")
	}

	// Enable 2FA
	if err := s.authRepo.Enable2FA(ctx, userID, secret); err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Disable2FA implements 2FA disablement
func (s *authService) Disable2FA(ctx context.Context, userID uint, password string) error {
	// Get user
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}

	// Disable 2FA
	if err := s.authRepo.Disable2FA(ctx, userID); err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// Logout implements logout flow
func (s *authService) Logout(ctx context.Context, token string) error {
	// Parse token
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return errors.New("invalid token")
	}

	// Convert Subject from string to uint
	userID, err := utils.StringToUint(claims.Subject)
	if err != nil {
		return errors.New("invalid user ID in token")
	}

	// Clear refresh token
	if err := s.authRepo.UpdateRefreshToken(ctx, userID, ""); err != nil {
		return fmt.Errorf("failed to clear refresh token: %w", err)
	}

	return nil
}

// ValidateToken implements token validation
func (s *authService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	// Parse token
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Convert Subject from string to uint
	userID, err := utils.StringToUint(claims.Subject)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	// Get user
	user, err := s.authRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// generateTokens generates access and refresh tokens
func (s *authService) generateTokens(userID uint) (string, string, error) {
	// Generate access token
	accessTokenClaims := jwt.StandardClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: time.Now().Add(time.Duration(s.jwtExpiration) * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshTokenClaims := jwt.StandardClaims{
		Subject:   fmt.Sprintf("%d", userID),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 days
		IssuedAt:  time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// generateTOTPSecret creates a cryptographically secure random secret for TOTP
func generateTOTPSecret() (string, error) {
	// Generate a 20-byte (160-bit) random secret
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	// Convert to base32 string (as required by the TOTP spec)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// OAuthUserInfo represents a user from an OAuth provider
type OAuthUserInfo struct {
	ID     string
	Email  string
	Name   string
	Avatar string
}

// EmailService defines operations for sending emails
type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, name, token string) error
	SendPasswordResetEmail(ctx context.Context, email, name, token string) error
}

// OAuthService defines operations for OAuth providers
type OAuthService interface {
	GetUserInfo(ctx context.Context, provider model.AuthProvider, token string) (*OAuthUserInfo, error)
}
