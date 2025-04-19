package service

import (
	"context"
	"errors"
	"time"

	"github.com/whitewalker-sa/ehass/internal/config"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v4"
)

type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
	logger   *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, cfg *config.Config, logger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

// Register registers a new user
func (s *userService) Register(ctx context.Context, user *model.User, password string) error {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return errors.New("failed to hash password")
	}

	// Set password hash
	user.PasswordHash = string(hashedPassword)

	// Create user
	return s.userRepo.Create(ctx, user)
}

// Login authenticates a user and returns a JWT token
func (s *userService) Login(ctx context.Context, email, password string) (*model.User, string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Debug("Login failed: user not found", zap.String("email", email))
		return nil, "", errors.New("invalid email or password")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.logger.Debug("Login failed: invalid password", zap.String("email", email))
		return nil, "", errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
		return nil, "", errors.New("failed to generate authentication token")
	}

	return user, token, nil
}

// GetUserByID gets a user by ID
func (s *userService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, user *model.User) error {
	// Ensure password hash is not modified directly
	existingUser, err := s.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return err
	}

	user.PasswordHash = existingUser.PasswordHash
	return s.userRepo.Update(ctx, user)
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// Find user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return errors.New("failed to hash password")
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

// UpdateUserProfile updates a user's profile information
func (s *userService) UpdateUserProfile(ctx context.Context, id uint, name, phone, address string) (*model.User, error) {
	// Find user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if name != "" {
		user.Name = name
	}
	if phone != "" {
		user.Phone = phone
	}
	if address != "" {
		user.Address = address
	}

	// Save changes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateAvatar updates a user's avatar URL
func (s *userService) UpdateAvatar(ctx context.Context, id uint, avatarURL string) (*model.User, error) {
	// Find user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update avatar URL
	user.Avatar = avatarURL

	// Save changes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update avatar", zap.Error(err))
		return nil, errors.New("failed to update avatar")
	}

	return user, nil
}

// generateToken generates a JWT token for authentication
func (s *userService) generateToken(user *model.User) (string, error) {
	// Create claims
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(s.cfg.Auth.AccessTokenExpiry).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	signedToken, err := token.SignedString([]byte(s.cfg.Auth.AccessTokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
