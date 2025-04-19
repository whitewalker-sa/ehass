package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"go.uber.org/zap"
)

type appointmentService struct {
	appointmentRepo repository.AppointmentRepository
	doctorRepo      repository.DoctorRepository
	patientRepo     repository.PatientRepository
	logger          *zap.Logger
}

// NewAppointmentService creates a new appointment service
func NewAppointmentService(
	appointmentRepo repository.AppointmentRepository,
	doctorRepo repository.DoctorRepository,
	patientRepo repository.PatientRepository,
	logger *zap.Logger,
) AppointmentService {
	return &appointmentService{
		appointmentRepo: appointmentRepo,
		doctorRepo:      doctorRepo,
		patientRepo:     patientRepo,
		logger:          logger,
	}
}

// CreateAppointment creates a new appointment
func (s *appointmentService) CreateAppointment(ctx context.Context, patientID, doctorID uint, date, timeStr string, reason string) (*model.Appointment, error) {
	// Parse date and time strings
	dateTime, err := parseDateTime(date, timeStr)
	if err != nil {
		return nil, errors.New("invalid date or time format")
	}

	// Create appointment model
	appointment := &model.Appointment{
		PatientID:      patientID,
		DoctorID:       doctorID,
		ScheduledStart: dateTime,
		ScheduledEnd:   dateTime.Add(30 * time.Minute),
		Reason:         reason,
		Status:         model.AppointmentStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Call repository to save appointment
	if err := s.appointmentRepo.Create(ctx, appointment); err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	return appointment, nil
}

// GetAppointmentByID gets an appointment by ID
func (s *appointmentService) GetAppointmentByID(ctx context.Context, id uint) (*model.Appointment, error) {
	return s.appointmentRepo.FindByID(ctx, id)
}

// GetPatientAppointments gets appointments for a patient with pagination
func (s *appointmentService) GetPatientAppointments(ctx context.Context, patientID uint, page, pageSize int) ([]*model.Appointment, int64, error) {
	offset := (page - 1) * pageSize
	return s.appointmentRepo.FindByPatientID(ctx, patientID, pageSize, offset)
}

// GetDoctorAppointments gets appointments for a doctor with pagination
func (s *appointmentService) GetDoctorAppointments(ctx context.Context, doctorID uint, page, pageSize int) ([]*model.Appointment, int64, error) {
	offset := (page - 1) * pageSize
	return s.appointmentRepo.FindByDoctorID(ctx, doctorID, pageSize, offset)
}

// GetDoctorAppointmentsByDateRange gets a doctor's appointments for a specific date range
func (s *appointmentService) GetDoctorAppointmentsByDateRange(ctx context.Context, doctorID uint, startDate, endDate string, page, pageSize int) ([]*model.Appointment, int64, error) {
	offset := (page - 1) * pageSize
	return s.appointmentRepo.FindByDateRange(ctx, doctorID, startDate, endDate, pageSize, offset)
}

// UpdateAppointment updates an appointment
func (s *appointmentService) UpdateAppointment(ctx context.Context, id uint, date, timeStr, status, reason string) (*model.Appointment, error) {
	// Get existing appointment
	existingAppointment, err := s.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if appointment can be modified
	if existingAppointment.Status == model.AppointmentStatusCompleted ||
		existingAppointment.Status == model.AppointmentStatusCancelled {
		return nil, errors.New("cannot update a completed or cancelled appointment")
	}

	// Update fields that were provided
	if date != "" && timeStr != "" {
		scheduledStart, err := parseDateTime(date, timeStr)
		if err != nil {
			return nil, errors.New("invalid date or time format")
		}

		// Validate appointment time
		if scheduledStart.Before(time.Now()) {
			return nil, errors.New("appointment cannot be scheduled in the past")
		}

		existingAppointment.ScheduledStart = scheduledStart
		existingAppointment.ScheduledEnd = scheduledStart.Add(30 * time.Minute)

		// Check for overlapping appointments
		overlappingAppointments, _, err := s.appointmentRepo.FindByDateRange(
			ctx,
			existingAppointment.DoctorID,
			existingAppointment.ScheduledStart.Format(time.RFC3339),
			existingAppointment.ScheduledEnd.Format(time.RFC3339),
			100, 0, // Fetch up to 100 appointments in this range
		)
		if err != nil {
			s.logger.Error("Failed to check overlapping appointments", zap.Error(err))
			return nil, errors.New("failed to check doctor's schedule")
		}

		for _, existing := range overlappingAppointments {
			if existing.ID != existingAppointment.ID &&
				existing.Status != model.AppointmentStatusCancelled &&
				((existingAppointment.ScheduledStart.Before(existing.ScheduledEnd) &&
					existingAppointment.ScheduledEnd.After(existing.ScheduledStart)) ||
					(existingAppointment.ScheduledStart.Equal(existing.ScheduledStart))) {
				return nil, errors.New("appointment time conflicts with an existing appointment")
			}
		}
	}

	if status != "" {
		existingAppointment.Status = model.AppointmentStatus(status)
	}

	if reason != "" {
		existingAppointment.Reason = reason
	}

	// Update appointment
	if err := s.appointmentRepo.Update(ctx, existingAppointment); err != nil {
		s.logger.Error("Failed to update appointment", zap.Error(err))
		return nil, errors.New("failed to update appointment")
	}

	return existingAppointment, nil
}

// CancelAppointment cancels an appointment
func (s *appointmentService) CancelAppointment(ctx context.Context, id uint) error {
	// Get appointment
	appointment, err := s.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if appointment can be cancelled
	if appointment.Status == model.AppointmentStatusCompleted ||
		appointment.Status == model.AppointmentStatusCancelled {
		return errors.New("appointment is already completed or cancelled")
	}

	// Check if it's too late to cancel
	if time.Until(appointment.ScheduledStart) < time.Hour {
		return errors.New("appointment cannot be cancelled less than 1 hour before the scheduled time")
	}

	// Update status
	appointment.Status = model.AppointmentStatusCancelled
	return s.appointmentRepo.Update(ctx, appointment)
}

// Helper function to parse date and time strings
func parseDateTime(date, timeStr string) (time.Time, error) {
	dateTimeStr := date + " " + timeStr
	return time.Parse("2006-01-02 15:04", dateTimeStr)
}

// CompleteAppointment marks an appointment as completed with notes
func (s *appointmentService) CompleteAppointment(ctx context.Context, id uint, notes string) error {
	// Get appointment
	appointment, err := s.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if appointment can be completed
	if appointment.Status == model.AppointmentStatusCancelled {
		return errors.New("cannot complete a cancelled appointment")
	}

	if appointment.Status == model.AppointmentStatusCompleted {
		return errors.New("appointment is already marked as completed")
	}

	// Check if appointment date has passed
	if time.Now().Before(appointment.ScheduledStart) {
		return errors.New("cannot complete an appointment before its scheduled time")
	}

	// Update status
	appointment.Status = model.AppointmentStatusCompleted
	appointment.Notes = notes

	return s.appointmentRepo.Update(ctx, appointment)
}
