package service

import (
	"context"
	"errors"
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
func (s *appointmentService) CreateAppointment(ctx context.Context, appointment *model.Appointment) error {
	// Validate doctor exists
	doctor, err := s.doctorRepo.FindByID(ctx, appointment.DoctorID)
	if err != nil {
		return errors.New("doctor not found")
	}

	// Validate patient exists
	_, err = s.patientRepo.FindByID(ctx, appointment.PatientID)
	if err != nil {
		return errors.New("patient not found")
	}

	// Validate appointment time
	if appointment.ScheduledStart.Before(time.Now()) {
		return errors.New("appointment cannot be scheduled in the past")
	}

	if appointment.ScheduledEnd.Before(appointment.ScheduledStart) {
		return errors.New("appointment end time must be after start time")
	}

	// Check for overlapping appointments for doctor
	overlappingAppointments, _, err := s.appointmentRepo.FindByDateRange(
		ctx,
		doctor.ID,
		appointment.ScheduledStart.Format(time.RFC3339),
		appointment.ScheduledEnd.Format(time.RFC3339),
		100, 0, // Fetch up to 100 appointments in this range
	)
	if err != nil {
		s.logger.Error("Failed to check overlapping appointments", zap.Error(err))
		return errors.New("failed to check doctor's schedule")
	}

	for _, existing := range overlappingAppointments {
		if existing.Status != model.AppointmentStatusCancelled &&
			((appointment.ScheduledStart.Before(existing.ScheduledEnd) &&
				appointment.ScheduledEnd.After(existing.ScheduledStart)) ||
				(appointment.ScheduledStart.Equal(existing.ScheduledStart))) {
			return errors.New("appointment time conflicts with an existing appointment")
		}
	}

	// Set initial status
	appointment.Status = model.AppointmentStatusPending

	// Create appointment
	return s.appointmentRepo.Create(ctx, appointment)
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

// GetDoctorSchedule gets a doctor's schedule for a specific date range
func (s *appointmentService) GetDoctorSchedule(ctx context.Context, doctorID uint, startDate, endDate string, page, pageSize int) ([]*model.Appointment, int64, error) {
	offset := (page - 1) * pageSize
	return s.appointmentRepo.FindByDateRange(ctx, doctorID, startDate, endDate, pageSize, offset)
}

// UpdateAppointment updates an appointment
func (s *appointmentService) UpdateAppointment(ctx context.Context, appointment *model.Appointment) error {
	// Get existing appointment
	existingAppointment, err := s.appointmentRepo.FindByID(ctx, appointment.ID)
	if err != nil {
		return err
	}

	// Check if appointment can be modified
	if existingAppointment.Status == model.AppointmentStatusCompleted ||
		existingAppointment.Status == model.AppointmentStatusCancelled {
		return errors.New("cannot update a completed or cancelled appointment")
	}

	// Check for time conflicts if time is being updated
	if !appointment.ScheduledStart.Equal(existingAppointment.ScheduledStart) ||
		!appointment.ScheduledEnd.Equal(existingAppointment.ScheduledEnd) {
		// Validate appointment time
		if appointment.ScheduledStart.Before(time.Now()) {
			return errors.New("appointment cannot be scheduled in the past")
		}

		if appointment.ScheduledEnd.Before(appointment.ScheduledStart) {
			return errors.New("appointment end time must be after start time")
		}

		// Check for overlapping appointments
		overlappingAppointments, _, err := s.appointmentRepo.FindByDateRange(
			ctx,
			appointment.DoctorID,
			appointment.ScheduledStart.Format(time.RFC3339),
			appointment.ScheduledEnd.Format(time.RFC3339),
			100, 0, // Fetch up to 100 appointments in this range
		)
		if err != nil {
			s.logger.Error("Failed to check overlapping appointments", zap.Error(err))
			return errors.New("failed to check doctor's schedule")
		}

		for _, existing := range overlappingAppointments {
			if existing.ID != appointment.ID &&
				existing.Status != model.AppointmentStatusCancelled &&
				((appointment.ScheduledStart.Before(existing.ScheduledEnd) &&
					appointment.ScheduledEnd.After(existing.ScheduledStart)) ||
					(appointment.ScheduledStart.Equal(existing.ScheduledStart))) {
				return errors.New("appointment time conflicts with an existing appointment")
			}
		}
	}

	// Update appointment
	return s.appointmentRepo.Update(ctx, appointment)
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

// CompleteAppointment marks an appointment as completed
func (s *appointmentService) CompleteAppointment(ctx context.Context, id uint, notes string) error {
	// Get appointment
	appointment, err := s.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if appointment can be completed
	if appointment.Status != model.AppointmentStatusConfirmed {
		return errors.New("only confirmed appointments can be completed")
	}

	// Update status and notes
	appointment.Status = model.AppointmentStatusCompleted
	appointment.Notes = notes
	return s.appointmentRepo.Update(ctx, appointment)
}
