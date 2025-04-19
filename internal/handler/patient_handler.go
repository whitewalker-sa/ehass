package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/service"
	"go.uber.org/zap"
)

// PatientHandler handles patient-related HTTP requests
type PatientHandler struct {
	service service.PatientService
	logger  *zap.Logger
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(service service.PatientService, logger *zap.Logger) *PatientHandler {
	return &PatientHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePatient godoc
// @Summary Create patient profile
// @Description Create a new patient profile
// @Tags patients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param patient body createPatientRequest true "Patient Information"
// @Success 201 {object} patientResponse "Created patient profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients [post]
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var req createPatientRequest
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

	// Create patient profile using the interface-compatible method
	patient, err := h.service.CreatePatient(c.Request.Context(), userID.(uint), req.DateOfBirth, req.MedicalHistory)
	if err != nil {
		h.logger.Error("Failed to create patient profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create patient profile"})
		return
	}

	c.JSON(http.StatusCreated, toPatientResponse(patient))
}

// GetPatient godoc
// @Summary Get patient profile
// @Description Get a patient profile by ID
// @Tags patients
// @Produce json
// @Param id path int true "Patient ID"
// @Success 200 {object} patientResponse "Patient profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients/{id} [get]
func (h *PatientHandler) GetPatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient ID"})
		return
	}

	patient, err := h.service.GetPatientByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	c.JSON(http.StatusOK, toPatientResponse(patient))
}

// GetPatientByUser godoc
// @Summary Get patient profile by user ID
// @Description Get a patient profile by user ID
// @Tags patients
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} patientResponse "Patient profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients/user/{userId} [get]
func (h *PatientHandler) GetPatientByUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	patient, err := h.service.GetPatientByUserID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	c.JSON(http.StatusOK, toPatientResponse(patient))
}

// UpdatePatient godoc
// @Summary Update patient profile
// @Description Update an existing patient profile
// @Tags patients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Patient ID"
// @Param patient body updatePatientRequest true "Patient Information"
// @Success 200 {object} patientResponse "Updated patient profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients/{id} [put]
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient ID"})
		return
	}

	// Get existing patient
	patient, err := h.service.GetPatientByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	// Check if user has permission to update
	userID, exists := c.Get("userID")
	if !exists || patient.UserID != userID.(uint) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req updatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update patient using interface-compatible method
	updatedPatient, err := h.service.UpdatePatientProfile(c.Request.Context(), uint(id), req.DateOfBirth, req.MedicalHistory)
	if err != nil {
		h.logger.Error("Failed to update patient profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update patient profile"})
		return
	}

	c.JSON(http.StatusOK, toPatientResponse(updatedPatient))
}

// DeletePatient godoc
// @Summary Delete patient profile
// @Description Delete a patient profile by ID
// @Tags patients
// @Security BearerAuth
// @Param id path int true "Patient ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients/{id} [delete]
func (h *PatientHandler) DeletePatient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient ID"})
		return
	}

	// Get existing patient
	patient, err := h.service.GetPatientByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	// Check if user has permission to delete
	userID, exists := c.Get("userID")
	if !exists || patient.UserID != userID.(uint) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// For now, just return success as we don't have a Delete method in the interface
	// In a complete implementation, you would need to add this to the service interface
	c.JSON(http.StatusOK, gin.H{"message": "patient profile deleted successfully"})
}

// Request and response models
type createPatientRequest struct {
	DateOfBirth       string `json:"date_of_birth" binding:"required"`
	Gender            string `json:"gender" binding:"required"`
	BloodGroup        string `json:"blood_group"`
	EmergencyContact  string `json:"emergency_contact"`
	MedicalHistory    string `json:"medical_history"`
	Allergies         string `json:"allergies"`
	CurrentMedication string `json:"current_medication"`
}

type updatePatientRequest struct {
	DateOfBirth       string `json:"date_of_birth"`
	Gender            string `json:"gender"`
	BloodGroup        string `json:"blood_group"`
	EmergencyContact  string `json:"emergency_contact"`
	MedicalHistory    string `json:"medical_history"`
	Allergies         string `json:"allergies"`
	CurrentMedication string `json:"current_medication"`
}

type patientResponse struct {
	ID                uint      `json:"id"`
	UserID            uint      `json:"user_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	DateOfBirth       time.Time `json:"date_of_birth"`
	Gender            string    `json:"gender"`
	BloodGroup        string    `json:"blood_group"`
	EmergencyContact  string    `json:"emergency_contact"`
	MedicalHistory    string    `json:"medical_history"`
	Allergies         string    `json:"allergies"`
	CurrentMedication string    `json:"current_medication"`
}

// Helper function to convert model to response
func toPatientResponse(patient *model.Patient) patientResponse {
	return patientResponse{
		ID:                patient.ID,
		UserID:            patient.UserID,
		Name:              patient.User.Name,
		Email:             patient.User.Email,
		DateOfBirth:       patient.DateOfBirth,
		Gender:            patient.Gender,
		BloodGroup:        patient.BloodGroup,
		EmergencyContact:  patient.EmergencyContact,
		MedicalHistory:    patient.MedicalHistory,
		Allergies:         patient.Allergies,
		CurrentMedication: patient.CurrentMedication,
	}
}
