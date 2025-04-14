package repository

import (
	"context"
	"errors"

	"github.com/whitewalker-sa/ehass/internal/model"
	"gorm.io/gorm"
)

type doctorRepository struct {
	db *gorm.DB
}

// NewDoctorRepository creates a new doctor repository
func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return &doctorRepository{
		db: db,
	}
}

// Create creates a new doctor
func (r *doctorRepository) Create(ctx context.Context, doctor *model.Doctor) error {
	return r.db.WithContext(ctx).Create(doctor).Error
}

// FindByID finds a doctor by ID with preloaded user data
func (r *doctorRepository) FindByID(ctx context.Context, id uint) (*model.Doctor, error) {
	var doctor model.Doctor
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("doctor not found")
		}
		return nil, err
	}
	return &doctor, nil
}

// FindByUserID finds a doctor by user ID
func (r *doctorRepository) FindByUserID(ctx context.Context, userID uint) (*model.Doctor, error) {
	var doctor model.Doctor
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("doctor not found")
		}
		return nil, err
	}
	return &doctor, nil
}

// FindAll finds all doctors with pagination
func (r *doctorRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.Doctor, int64, error) {
	var doctors []*model.Doctor
	var count int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&model.Doctor{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Preload("User").Limit(limit).Offset(offset).Find(&doctors).Error; err != nil {
		return nil, 0, err
	}

	return doctors, count, nil
}

// FindBySpecialty finds doctors by specialty with pagination
func (r *doctorRepository) FindBySpecialty(ctx context.Context, specialty string, limit, offset int) ([]*model.Doctor, int64, error) {
	var doctors []*model.Doctor
	var count int64

	// Count total records with this specialty
	if err := r.db.WithContext(ctx).Model(&model.Doctor{}).Where("specialty = ?", specialty).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Preload("User").Where("specialty = ?", specialty).Limit(limit).Offset(offset).Find(&doctors).Error; err != nil {
		return nil, 0, err
	}

	return doctors, count, nil
}

// Update updates a doctor
func (r *doctorRepository) Update(ctx context.Context, doctor *model.Doctor) error {
	return r.db.WithContext(ctx).Save(doctor).Error
}

// Delete soft deletes a doctor
func (r *doctorRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Doctor{}, id).Error
}
