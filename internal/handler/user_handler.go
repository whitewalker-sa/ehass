package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/service"
	"go.uber.org/zap"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body registerRequest true "User Registration"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if err := req.validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user model
	user := &model.User{
		Name:    req.Name,
		Email:   req.Email,
		Role:    model.Role(req.Role),
		Phone:   req.Phone,
		Address: req.Address,
	}

	// Register user
	if err := h.userService.Register(c.Request.Context(), user, req.Password); err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body loginRequest true "Login Credentials"
// @Success 200 {object} loginResponse "Successful login"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Login user
	user, token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Return user info and token
	c.JSON(http.StatusOK, loginResponse{
		User: userResponse{
			ID:      user.ID,
			Name:    user.Name,
			Email:   user.Email,
			Role:    string(user.Role),
			Phone:   user.Phone,
			Address: user.Address,
		},
		Token: token,
	})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} userResponse "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	// Return user info
	c.JSON(http.StatusOK, userResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Role:    string(user.Role),
		Phone:   user.Phone,
		Address: user.Address,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body updateProfileRequest true "User Profile"
// @Success 200 {object} userResponse "Updated user profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get current user
	user, err := h.userService.GetUserByID(c.Request.Context(), userID.(uint))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	// Update user fields
	user.Name = req.Name
	user.Phone = req.Phone
	user.Address = req.Address

	// Update user
	if err := h.userService.UpdateUser(c.Request.Context(), user); err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Return updated user info
	c.JSON(http.StatusOK, userResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Role:    string(user.Role),
		Phone:   user.Phone,
		Address: user.Address,
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change authenticated user's password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body changePasswordRequest true "Password Change"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/change-password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate request
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password and confirm password don't match"})
		return
	}

	// Change password
	if err := h.userService.ChangePassword(c.Request.Context(), userID.(uint), req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get user by ID (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} userResponse "User"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Parse user ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return user info
	c.JSON(http.StatusOK, userResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Role:    string(user.Role),
		Phone:   user.Phone,
		Address: user.Address,
	})
}

// Request and response types

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

func (r *registerRequest) validate() error {
	// Validate role
	if r.Role != string(model.RolePatient) && r.Role != string(model.RoleDoctor) && r.Role != string(model.RoleAdmin) {
		return errors.New("invalid role")
	}
	return nil
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	User  userResponse `json:"user"`
	Token string       `json:"token"`
}

type userResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

type updateProfileRequest struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}
