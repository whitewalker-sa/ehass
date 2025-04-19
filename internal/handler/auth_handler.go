package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/service"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents request body for user registration
type RegisterRequest struct {
	Name     string     `json:"name" binding:"required"`
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required,min=8"`
	Role     model.Role `json:"role" binding:"required"`
	Phone    string     `json:"phone"`
	Address  string     `json:"address"`
}

// LoginRequest represents request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// OAuthLoginRequest represents request body for OAuth login
type OAuthLoginRequest struct {
	Provider      model.AuthProvider `json:"provider" binding:"required"`
	ProviderToken string             `json:"providerToken" binding:"required"`
}

// VerifyEmailRequest represents request body for email verification
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// RequestPasswordResetRequest represents request body for password reset request
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents request body for password reset
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// RefreshTokenRequest represents request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Setup2FAResponse represents response body for 2FA setup
type Setup2FAResponse struct {
	URI string `json:"uri"`
}

// Enable2FARequest represents request body for 2FA enablement
type Enable2FARequest struct {
	Secret string `json:"secret" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

// Disable2FARequest represents request body for 2FA disablement
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// Verify2FARequest represents request body for 2FA verification
type Verify2FARequest struct {
	UserID uint   `json:"userId" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

// LinkOAuthRequest represents request body for linking OAuth account
type LinkOAuthRequest struct {
	Provider      model.AuthProvider `json:"provider" binding:"required"`
	ProviderToken string             `json:"providerToken" binding:"required"`
}

// TokenResponse represents response body for token generation
type TokenResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         interface{} `json:"user"`
	Require2FA   bool        `json:"require2fa"`
	UserID       uint        `json:"userId,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration successful. Please check your email to verify your account.",
		"user":    model.SanitizeUser(*user),
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Check if 2FA is required
		if err.Error() == "two-factor authentication required" {
			c.JSON(http.StatusOK, TokenResponse{
				Require2FA: true,
				UserID:     user.ID,
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         model.SanitizeUser(*user),
		Require2FA:   false,
	})
}

// OAuthLogin handles OAuth login
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	var req OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.authService.OAuthLogin(c.Request.Context(), req.Provider, req.ProviderToken)
	if err != nil {
		// Check if 2FA is required
		if err.Error() == "two-factor authentication required" {
			c.JSON(http.StatusOK, TokenResponse{
				Require2FA: true,
				UserID:     user.ID,
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         model.SanitizeUser(*user),
		Require2FA:   false,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// RequestPasswordReset handles password reset request
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal if email exists, but log the error
		c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive password reset instructions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive password reset instructions"})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// Setup2FA handles 2FA setup
func (h *AuthHandler) Setup2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	uri, err := h.authService.Setup2FA(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, Setup2FAResponse{URI: uri})
}

// Enable2FA handles 2FA enablement
func (h *AuthHandler) Enable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.Enable2FA(c.Request.Context(), userID.(uint), req.Secret, req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Two-factor authentication enabled successfully"})
}

// Disable2FA handles 2FA disablement
func (h *AuthHandler) Disable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.Disable2FA(c.Request.Context(), userID.(uint), req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Two-factor authentication disabled successfully"})
}

// Verify2FA handles 2FA verification
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid, err := h.authService.Verify2FA(c.Request.Context(), req.UserID, req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA token"})
		return
	}

	// Get user for response
	user, err := h.authService.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         model.SanitizeUser(*user),
		Require2FA:   false,
	})
}

// LinkOAuth handles linking OAuth account
func (h *AuthHandler) LinkOAuth(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req LinkOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.LinkOAuthAccount(c.Request.Context(), userID.(uint), req.Provider, req.ProviderToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OAuth account linked successfully"})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization header"})
		return
	}

	// Remove Bearer prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
