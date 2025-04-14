package repository

import (
	"context"
	"errors"

	"github.com/whitewalker-sa/ehass/internal/model"
	"gorm.io/gorm"
)

type patientRepository struct {
	db *gorm.DB
}

// NewPatientRepository creates a new patient repository
func NewPatientRepository(db *gorm.DB) PatientRepository {
	return &patientRepository{
		db: db,
	}
}

// Create creates a new patient
func (r *patientRepository) Create(ctx context.Context, patient *model.Patient) error {
	return r.db.WithContext(ctx).Create(patient).Error
}

// FindByID finds a patient by ID with preloaded user data
func (r *patientRepository) FindByID(ctx context.Context, id uint) (*model.Patient, error) {
	var patient model.Patient
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}
	return &patient, nil
}

// FindByUserID finds a patient by user ID
func (r *patientRepository) FindByUserID(ctx context.Context, userID uint) (*model.Patient, error) {
	var patient model.Patient
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("patient not found")
		}
		return nil, err
	}
	return &patient, nil
}

// Update updates a patient
func (r *patientRepository) Update(ctx context.Context, patient *model.Patient) error {
	return r.db.WithContext(ctx).Save(patient).Error
}

// Delete soft deletes a patient
func (r *patientRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Patient{}, id).Error
}
