package service

import (
	"context"
	"fmt"
	"time"

	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"go.uber.org/zap"
)

type doctorService struct {
	repo   repository.DoctorRepository
	logger *zap.Logger
}

// NewDoctorService creates a new doctor service
func NewDoctorService(repo repository.DoctorRepository, logger *zap.Logger) DoctorService {
	return &doctorService{
		repo:   repo,
		logger: logger,
	}
}

// CreateDoctor creates a new doctor profile
func (s *doctorService) CreateDoctor(ctx context.Context, userID uint, specialty, education string, experience int) (*model.Doctor, error) {
	// Create doctor model
	doctor := &model.Doctor{
		UserID:     userID,
		Specialty:  specialty,
		Education:  education,
		Experience: experience,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Call repository to save doctor
	if err := s.repo.Create(ctx, doctor); err != nil {
		return nil, fmt.Errorf("failed to create doctor profile: %w", err)
	}

	return doctor, nil
}

// GetDoctorByID retrieves a doctor by ID
func (s *doctorService) GetDoctorByID(ctx context.Context, id uint) (*model.Doctor, error) {
	return s.repo.FindByID(ctx, id)
}

// GetDoctorByUserID retrieves a doctor by user ID
func (s *doctorService) GetDoctorByUserID(ctx context.Context, userID uint) (*model.Doctor, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// GetAllDoctors retrieves all doctors with pagination
func (s *doctorService) GetAllDoctors(ctx context.Context, page, pageSize int) ([]*model.Doctor, int64, error) {
	// Calculate offset for pagination
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindAll(ctx, pageSize, offset)
}

// GetDoctorsBySpecialty retrieves doctors by specialty with pagination
func (s *doctorService) GetDoctorsBySpecialty(ctx context.Context, specialty string, page, pageSize int) ([]*model.Doctor, int64, error) {
	// Calculate offset for pagination
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindBySpecialty(ctx, specialty, pageSize, offset)
}

// UpdateDoctorProfile updates doctor profile information
func (s *doctorService) UpdateDoctorProfile(ctx context.Context, id uint, specialty, bio string, experience int) (*model.Doctor, error) {
	doctor, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	doctor.Specialty = specialty
	doctor.Bio = bio
	doctor.Experience = experience

	err = s.repo.Update(ctx, doctor)
	if err != nil {
		return nil, err
	}

	return doctor, nil
}

// UpdateDoctor updates doctor information
func (s *doctorService) UpdateDoctor(ctx context.Context, doctor *model.Doctor) error {
	return s.repo.Update(ctx, doctor)
}

// DeleteDoctor deletes a doctor by ID
func (s *doctorService) DeleteDoctor(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
