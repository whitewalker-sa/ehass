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

// AppointmentHandler handles HTTP requests for appointments
type AppointmentHandler struct {
	appointmentService service.AppointmentService
	logger             *zap.Logger
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(appointmentService service.AppointmentService, logger *zap.Logger) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentService: appointmentService,
		logger:             logger,
	}
}

// CreateAppointment godoc
// @Summary Create a new appointment
// @Description Create a new appointment for a patient with a doctor
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appointment body createAppointmentRequest true "Appointment Details"
// @Success 201 {object} map[string]string "Appointment created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	var req createAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Parse appointment times
	scheduledStart, err := time.Parse(time.RFC3339, req.ScheduledStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled start time format"})
		return
	}

	scheduledEnd, err := time.Parse(time.RFC3339, req.ScheduledEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled end time format"})
		return
	}

	// Create appointment model
	appointment := &model.Appointment{
		PatientID:      req.PatientID,
		DoctorID:       req.DoctorID,
		ScheduledStart: scheduledStart,
		ScheduledEnd:   scheduledEnd,
		Reason:         req.Reason,
		Type:           req.Type,
		Notes:          req.Notes,
	}

	// Create appointment
	if err := h.appointmentService.CreateAppointment(c.Request.Context(), appointment); err != nil {
		h.logger.Error("Failed to create appointment", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Appointment created successfully",
		"id":      appointment.ID,
	})
}

// GetAppointmentByID godoc
// @Summary Get appointment by ID
// @Description Get appointment details by ID
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Appointment ID"
// @Success 200 {object} appointmentResponse "Appointment"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /appointments/{id} [get]
func (h *AppointmentHandler) GetAppointmentByID(c *gin.Context) {
	// Parse appointment ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get appointment
	appointment, err := h.appointmentService.GetAppointmentByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get appointment", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	// Return appointment
	c.JSON(http.StatusOK, formatAppointmentResponse(appointment))
}

// GetPatientAppointments godoc
// @Summary Get patient appointments
// @Description Get appointments for the specified patient
// @Tags appointments,patients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param patient_id path int true "Patient ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} paginatedAppointmentsResponse "Patient appointments"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /patients/{patient_id}/appointments [get]
func (h *AppointmentHandler) GetPatientAppointments(c *gin.Context) {
	// Parse patient ID
	patientIDStr := c.Param("patient_id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	// Parse pagination params
	page, pageSize := h.getPaginationParams(c)

	// Get appointments
	appointments, totalCount, err := h.appointmentService.GetPatientAppointments(c.Request.Context(), uint(patientID), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get patient appointments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get appointments"})
		return
	}

	// Format response
	responseItems := make([]appointmentResponse, 0, len(appointments))
	for _, appt := range appointments {
		responseItems = append(responseItems, formatAppointmentResponse(appt))
	}

	c.JSON(http.StatusOK, paginatedAppointmentsResponse{
		Items:      responseItems,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetDoctorAppointments godoc
// @Summary Get doctor appointments
// @Description Get appointments for the specified doctor
// @Tags appointments,doctors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param doctor_id path int true "Doctor ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} paginatedAppointmentsResponse "Doctor appointments"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/{doctor_id}/appointments [get]
func (h *AppointmentHandler) GetDoctorAppointments(c *gin.Context) {
	// Parse doctor ID
	doctorIDStr := c.Param("doctor_id")
	doctorID, err := strconv.ParseUint(doctorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	// Parse pagination params
	page, pageSize := h.getPaginationParams(c)

	// Get appointments
	appointments, totalCount, err := h.appointmentService.GetDoctorAppointments(c.Request.Context(), uint(doctorID), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get doctor appointments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get appointments"})
		return
	}

	// Format response
	responseItems := make([]appointmentResponse, 0, len(appointments))
	for _, appt := range appointments {
		responseItems = append(responseItems, formatAppointmentResponse(appt))
	}

	c.JSON(http.StatusOK, paginatedAppointmentsResponse{
		Items:      responseItems,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// GetDoctorSchedule godoc
// @Summary Get doctor schedule
// @Description Get doctor's schedule for a date range
// @Tags appointments,doctors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param doctor_id path int true "Doctor ID"
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} paginatedAppointmentsResponse "Doctor schedule"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /doctors/{doctor_id}/schedule [get]
func (h *AppointmentHandler) GetDoctorSchedule(c *gin.Context) {
	// Parse doctor ID
	doctorIDStr := c.Param("doctor_id")
	doctorID, err := strconv.ParseUint(doctorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	// Get date range params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Parse pagination params
	page, pageSize := h.getPaginationParams(c)

	// Get appointments
	appointments, totalCount, err := h.appointmentService.GetDoctorSchedule(
		c.Request.Context(),
		uint(doctorID),
		startDate,
		endDate,
		page,
		pageSize,
	)
	if err != nil {
		h.logger.Error("Failed to get doctor schedule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get schedule"})
		return
	}

	// Format response
	responseItems := make([]appointmentResponse, 0, len(appointments))
	for _, appt := range appointments {
		responseItems = append(responseItems, formatAppointmentResponse(appt))
	}

	c.JSON(http.StatusOK, paginatedAppointmentsResponse{
		Items:      responseItems,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}

// UpdateAppointment godoc
// @Summary Update appointment
// @Description Update an existing appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Appointment ID"
// @Param appointment body updateAppointmentRequest true "Appointment Details"
// @Success 200 {object} map[string]string "Appointment updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /appointments/{id} [put]
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	// Parse appointment ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Get existing appointment
	appointment, err := h.appointmentService.GetAppointmentByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	var req updateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update fields if provided
	if req.ScheduledStart != "" {
		scheduledStart, err := time.Parse(time.RFC3339, req.ScheduledStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled start time format"})
			return
		}
		appointment.ScheduledStart = scheduledStart
	}

	if req.ScheduledEnd != "" {
		scheduledEnd, err := time.Parse(time.RFC3339, req.ScheduledEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled end time format"})
			return
		}
		appointment.ScheduledEnd = scheduledEnd
	}

	if req.Reason != "" {
		appointment.Reason = req.Reason
	}

	if req.Notes != "" {
		appointment.Notes = req.Notes
	}

	if req.Type != "" {
		appointment.Type = req.Type
	}

	if req.Status != "" {
		appointment.Status = model.AppointmentStatus(req.Status)
	}

	// Update appointment
	if err := h.appointmentService.UpdateAppointment(c.Request.Context(), appointment); err != nil {
		h.logger.Error("Failed to update appointment", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment updated successfully"})
}

// CancelAppointment godoc
// @Summary Cancel appointment
// @Description Cancel an existing appointment
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Appointment ID"
// @Success 200 {object} map[string]string "Appointment cancelled successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /appointments/{id}/cancel [post]
func (h *AppointmentHandler) CancelAppointment(c *gin.Context) {
	// Parse appointment ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	// Cancel appointment
	if err := h.appointmentService.CancelAppointment(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to cancel appointment", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled successfully"})
}

// CompleteAppointment godoc
// @Summary Complete appointment
// @Description Mark an appointment as completed
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Appointment ID"
// @Param data body completeAppointmentRequest true "Completion Details"
// @Success 200 {object} map[string]string "Appointment completed successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /appointments/{id}/complete [post]
func (h *AppointmentHandler) CompleteAppointment(c *gin.Context) {
	// Parse appointment ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var req completeAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Complete appointment
	if err := h.appointmentService.CompleteAppointment(c.Request.Context(), uint(id), req.Notes); err != nil {
		h.logger.Error("Failed to complete appointment", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment completed successfully"})
}

// Helper methods

func (h *AppointmentHandler) getPaginationParams(c *gin.Context) (page, pageSize int) {
	// Get page param
	pageStr := c.Query("page")
	page = 1
	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	// Get page size param
	pageSizeStr := c.Query("page_size")
	pageSize = 10
	if pageSizeStr != "" {
		pageSizeVal, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	return page, pageSize
}

func formatAppointmentResponse(appointment *model.Appointment) appointmentResponse {
	var patientName, doctorName string

	if appointment.Patient.User.ID > 0 {
		patientName = appointment.Patient.User.Name
	}

	if appointment.Doctor.User.ID > 0 {
		doctorName = appointment.Doctor.User.Name
	}

	return appointmentResponse{
		ID:             appointment.ID,
		PatientID:      appointment.PatientID,
		PatientName:    patientName,
		DoctorID:       appointment.DoctorID,
		DoctorName:     doctorName,
		ScheduledStart: appointment.ScheduledStart.Format(time.RFC3339),
		ScheduledEnd:   appointment.ScheduledEnd.Format(time.RFC3339),
		Status:         string(appointment.Status),
		Type:           appointment.Type,
		Reason:         appointment.Reason,
		Notes:          appointment.Notes,
		CreatedAt:      appointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      appointment.UpdatedAt.Format(time.RFC3339),
	}
}

// Request and response types

type createAppointmentRequest struct {
	PatientID      uint   `json:"patient_id" binding:"required"`
	DoctorID       uint   `json:"doctor_id" binding:"required"`
	ScheduledStart string `json:"scheduled_start" binding:"required"` // RFC3339 format
	ScheduledEnd   string `json:"scheduled_end" binding:"required"`   // RFC3339 format
	Reason         string `json:"reason"`
	Type           string `json:"type"` // in_person, video, phone
	Notes          string `json:"notes"`
}

type updateAppointmentRequest struct {
	ScheduledStart string `json:"scheduled_start,omitempty"` // RFC3339 format
	ScheduledEnd   string `json:"scheduled_end,omitempty"`   // RFC3339 format
	Status         string `json:"status,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Type           string `json:"type,omitempty"` // in_person, video, phone
	Notes          string `json:"notes,omitempty"`
}

type completeAppointmentRequest struct {
	Notes string `json:"notes"`
}

type appointmentResponse struct {
	ID             uint   `json:"id"`
	PatientID      uint   `json:"patient_id"`
	PatientName    string `json:"patient_name,omitempty"`
	DoctorID       uint   `json:"doctor_id"`
	DoctorName     string `json:"doctor_name,omitempty"`
	ScheduledStart string `json:"scheduled_start"`
	ScheduledEnd   string `json:"scheduled_end"`
	Status         string `json:"status"`
	Type           string `json:"type,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Notes          string `json:"notes,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type paginatedAppointmentsResponse struct {
	Items      []appointmentResponse `json:"items"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}
