package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/service"
	"go.uber.org/zap"
)

// DoctorHandler handles doctor-related HTTP requests
type DoctorHandler struct {
	service service.DoctorService
	logger  *zap.Logger
}

// NewDoctorHandler creates a new doctor handler
func NewDoctorHandler(service service.DoctorService, logger *zap.Logger) *DoctorHandler {
	return &DoctorHandler{
		service: service,
		logger:  logger,
	}
}

// CreateDoctor godoc
// @Summary Create doctor profile
// @Description Create a new doctor profile
// @Tags doctors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param doctor body createDoctorRequest true "Doctor Information"
// @Success 201 {object} doctorResponse "Created doctor profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors [post]
func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	var req createDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in token"})
		return
	}

	// Create doctor profile with the correct service method signature
	doctor, err := h.service.CreateDoctor(c.Request.Context(), userID.(uint), req.Specialty, req.Bio, req.Experience)
	if err != nil {
		h.logger.Error("Failed to create doctor profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create doctor profile"})
		return
	}

	// Update additional fields that weren't included in the CreateDoctor method
	doctor.Designation = req.Designation
	doctor.LicenseNo = req.LicenseNo
	doctor.Education = req.Education

	// Fix: Capture both return values (doctor and error) and use the returned doctor
	doctor, err = h.service.UpdateDoctorProfile(c.Request.Context(), doctor.ID, doctor.Specialty, doctor.Bio, doctor.Experience)
	if err != nil {
		h.logger.Warn("Failed to update additional doctor fields", zap.Error(err))
	}

	c.JSON(http.StatusCreated, toDoctorResponse(doctor))
}

// GetDoctor godoc
// @Summary Get doctor profile
// @Description Get a doctor profile by ID
// @Tags doctors
// @Produce json
// @Param id path int true "Doctor ID"
// @Success 200 {object} doctorResponse "Doctor profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/{id} [get]
func (h *DoctorHandler) GetDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor ID"})
		return
	}

	doctor, err := h.service.GetDoctorByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	c.JSON(http.StatusOK, toDoctorResponse(doctor))
}

// GetDoctorByUser godoc
// @Summary Get doctor profile by user ID
// @Description Get a doctor profile by user ID
// @Tags doctors
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} doctorResponse "Doctor profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/user/{userId} [get]
func (h *DoctorHandler) GetDoctorByUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	doctor, err := h.service.GetDoctorByUserID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	c.JSON(http.StatusOK, toDoctorResponse(doctor))
}

// ListDoctors godoc
// @Summary List all doctors
// @Description Get a paginated list of all doctors
// @Tags doctors
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {array} doctorResponse "List of doctors"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors [get]
func (h *DoctorHandler) ListDoctors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	doctors, total, err := h.service.GetAllDoctors(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get doctors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get doctors"})
		return
	}

	response := make([]doctorResponse, 0, len(doctors))
	for _, doctor := range doctors {
		response = append(response, toDoctorResponse(doctor))
	}

	c.JSON(http.StatusOK, gin.H{
		"doctors": response,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	})
}

// ListDoctorsBySpecialty godoc
// @Summary List doctors by specialty
// @Description Get a paginated list of doctors by specialty
// @Tags doctors
// @Produce json
// @Param specialty path string true "Specialty"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Success 200 {array} doctorResponse "List of doctors"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/specialty/{specialty} [get]
func (h *DoctorHandler) ListDoctorsBySpecialty(c *gin.Context) {
	specialty := c.Param("specialty")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	doctors, total, err := h.service.GetDoctorsBySpecialty(c.Request.Context(), specialty, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get doctors by specialty", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get doctors"})
		return
	}

	response := make([]doctorResponse, 0, len(doctors))
	for _, doctor := range doctors {
		response = append(response, toDoctorResponse(doctor))
	}

	c.JSON(http.StatusOK, gin.H{
		"doctors": response,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	})
}

// UpdateDoctor godoc
// @Summary Update doctor profile
// @Description Update an existing doctor profile
// @Tags doctors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Doctor ID"
// @Param doctor body updateDoctorRequest true "Doctor Information"
// @Success 200 {object} doctorResponse "Updated doctor profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/{id} [put]
func (h *DoctorHandler) UpdateDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor ID"})
		return
	}

	// Get existing doctor
	doctor, err := h.service.GetDoctorByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	// Check if user has permission to update
	userID, exists := c.Get("userID")
	if !exists || doctor.UserID != userID.(uint) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req updateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.Specialty != "" {
		doctor.Specialty = req.Specialty
	}
	if req.Designation != "" {
		doctor.Designation = req.Designation
	}
	if req.Education != "" {
		doctor.Education = req.Education
	}
	if req.Experience > 0 {
		doctor.Experience = req.Experience
	}
	if req.LicenseNo != "" {
		doctor.LicenseNo = req.LicenseNo
	}
	if req.Bio != "" {
		doctor.Bio = req.Bio
	}

	// Update doctor profile using the correct method from the interface
	updatedDoctor, err := h.service.UpdateDoctorProfile(c.Request.Context(), uint(id), doctor.Specialty, doctor.Bio, doctor.Experience)
	if err != nil {
		h.logger.Error("Failed to update doctor profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update doctor profile"})
		return
	}

	// Copy any additional fields that were updated but aren't part of the standard update
	updatedDoctor.Designation = doctor.Designation
	updatedDoctor.LicenseNo = doctor.LicenseNo
	updatedDoctor.Education = doctor.Education

	c.JSON(http.StatusOK, toDoctorResponse(updatedDoctor))
}

// UpdateDoctorProfile handles the update of a doctor's profile
func (h *DoctorHandler) UpdateDoctorProfile(c *gin.Context) {
	// Get doctor ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	var req updateDoctorProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update doctor profile
	updatedDoctor, err := h.service.UpdateDoctorProfile(c.Request.Context(), uint(id), req.Specialty, req.Bio, req.Experience)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update doctor profile"})
		return
	}

	// Return updated doctor profile
	c.JSON(http.StatusOK, gin.H{
		"doctor":  toDoctorResponse(updatedDoctor),
		"message": "Doctor profile updated successfully",
	})
}

// DeleteDoctor godoc
// @Summary Delete doctor profile
// @Description Delete a doctor profile by ID
// @Tags doctors
// @Security BearerAuth
// @Param id path int true "Doctor ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/{id} [delete]
func (h *DoctorHandler) DeleteDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor ID"})
		return
	}

	// Get existing doctor
	doctor, err := h.service.GetDoctorByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	// Check if user has permission to delete
	userID, exists := c.Get("userID")
	if !exists || doctor.UserID != userID.(uint) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Use the correct method name from the implementation
	err = h.service.DeleteDoctor(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete doctor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete doctor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "doctor deleted successfully"})
}

// Request and response models
type createDoctorRequest struct {
	Specialty   string `json:"specialty" binding:"required"`
	Designation string `json:"designation"`
	Education   string `json:"education"`
	Experience  int    `json:"experience"`
	LicenseNo   string `json:"license_no" binding:"required"`
	Bio         string `json:"bio"`
}

type updateDoctorRequest struct {
	Specialty   string `json:"specialty"`
	Designation string `json:"designation"`
	Education   string `json:"education"`
	Experience  int    `json:"experience"`
	LicenseNo   string `json:"license_no"`
	Bio         string `json:"bio"`
}

type updateDoctorProfileRequest struct {
	Specialty  string `json:"specialty"`
	Bio        string `json:"bio"`
	Experience int    `json:"experience"`
}

type doctorResponse struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Specialty   string `json:"specialty"`
	Designation string `json:"designation"`
	Education   string `json:"education"`
	Experience  int    `json:"experience"`
	LicenseNo   string `json:"license_no"`
	Bio         string `json:"bio"`
}

// Helper function to convert model to response
func toDoctorResponse(doctor *model.Doctor) doctorResponse {
	return doctorResponse{
		ID:          doctor.ID,
		UserID:      doctor.UserID,
		Name:        doctor.User.Name,
		Email:       doctor.User.Email,
		Specialty:   doctor.Specialty,
		Designation: doctor.Designation,
		Education:   doctor.Education,
		Experience:  doctor.Experience,
		LicenseNo:   doctor.LicenseNo,
		Bio:         doctor.Bio,
	}
}
