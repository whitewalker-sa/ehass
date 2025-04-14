package repository

import (
	"context"
	"errors"

	"github.com/whitewalker-sa/ehass/internal/model"
	"gorm.io/gorm"
)

type appointmentRepository struct {
	db *gorm.DB
}

// NewAppointmentRepository creates a new appointment repository
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{
		db: db,
	}
}

// Create creates a new appointment
func (r *appointmentRepository) Create(ctx context.Context, appointment *model.Appointment) error {
	return r.db.WithContext(ctx).Create(appointment).Error
}

// FindByID finds an appointment by ID
func (r *appointmentRepository) FindByID(ctx context.Context, id uint) (*model.Appointment, error) {
	var appointment model.Appointment
	err := r.db.WithContext(ctx).
		Preload("Patient.User").
		Preload("Doctor.User").
		Where("id = ?", id).
		First(&appointment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("appointment not found")
		}
		return nil, err
	}
	return &appointment, nil
}

// FindByPatientID finds appointments by patient ID with pagination
func (r *appointmentRepository) FindByPatientID(ctx context.Context, patientID uint, limit, offset int) ([]*model.Appointment, int64, error) {
	var appointments []*model.Appointment
	var count int64

	// Count total records
	if err := r.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("patient_id = ?", patientID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Preload("Doctor.User").
		Where("patient_id = ?", patientID).
		Order("scheduled_start DESC").
		Limit(limit).
		Offset(offset).
		Find(&appointments).Error; err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByDoctorID finds appointments by doctor ID with pagination
func (r *appointmentRepository) FindByDoctorID(ctx context.Context, doctorID uint, limit, offset int) ([]*model.Appointment, int64, error) {
	var appointments []*model.Appointment
	var count int64

	// Count total records
	if err := r.db.WithContext(ctx).
		Model(&model.Appointment{}).
		Where("doctor_id = ?", doctorID).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Preload("Patient.User").
		Where("doctor_id = ?", doctorID).
		Order("scheduled_start DESC").
		Limit(limit).
		Offset(offset).
		Find(&appointments).Error; err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// FindByDateRange finds appointments by doctor ID and date range with pagination
func (r *appointmentRepository) FindByDateRange(ctx context.Context, doctorID uint, start, end string, limit, offset int) ([]*model.Appointment, int64, error) {
	var appointments []*model.Appointment
	var count int64

	query := r.db.WithContext(ctx).Model(&model.Appointment{}).Where("doctor_id = ?", doctorID)

	if start != "" {
		query = query.Where("scheduled_start >= ?", start)
	}

	if end != "" {
		query = query.Where("scheduled_start <= ?", end)
	}

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with preloaded associations
	queryPreloaded := r.db.WithContext(ctx).
		Preload("Patient.User").
		Where("doctor_id = ?", doctorID)

	if start != "" {
		queryPreloaded = queryPreloaded.Where("scheduled_start >= ?", start)
	}

	if end != "" {
		queryPreloaded = queryPreloaded.Where("scheduled_start <= ?", end)
	}

	if err := queryPreloaded.
		Order("scheduled_start ASC").
		Limit(limit).
		Offset(offset).
		Find(&appointments).Error; err != nil {
		return nil, 0, err
	}

	return appointments, count, nil
}

// Update updates an appointment
func (r *appointmentRepository) Update(ctx context.Context, appointment *model.Appointment) error {
	return r.db.WithContext(ctx).Save(appointment).Error
}

// Delete soft deletes an appointment
func (r *appointmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Appointment{}, id).Error
}
