package service

import (
	"context"
	"fmt"
	"time"

	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"go.uber.org/zap"
)

type patientService struct {
	repo   repository.PatientRepository
	logger *zap.Logger
}

// NewPatientService creates a new patient service
func NewPatientService(repo repository.PatientRepository, logger *zap.Logger) PatientService {
	return &patientService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePatient creates a new patient profile
func (s *patientService) CreatePatient(ctx context.Context, userID uint, dateOfBirth, medicalHistory string) (*model.Patient, error) {
	// Parse date of birth
	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return nil, fmt.Errorf("invalid date of birth format: %w", err)
	}

	// Create patient model
	patient := &model.Patient{
		UserID:         userID,
		DateOfBirth:    dob,
		MedicalHistory: medicalHistory,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Call repository to save patient
	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, fmt.Errorf("failed to create patient profile: %w", err)
	}

	return patient, nil
}

// GetPatientByID retrieves a patient by ID
func (s *patientService) GetPatientByID(ctx context.Context, id uint) (*model.Patient, error) {
	return s.repo.FindByID(ctx, id)
}

// GetPatientByUserID retrieves a patient by user ID
func (s *patientService) GetPatientByUserID(ctx context.Context, userID uint) (*model.Patient, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// UpdatePatient updates patient information
func (s *patientService) UpdatePatient(ctx context.Context, patient *model.Patient) error {
	return s.repo.Update(ctx, patient)
}

// UpdatePatientProfile updates patient profile information
func (s *patientService) UpdatePatientProfile(ctx context.Context, id uint, dateOfBirth, medicalHistory string) (*model.Patient, error) {
	patient, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Parse date of birth if provided
	if dateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", dateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date of birth format: %w", err)
		}
		patient.DateOfBirth = dob
	}

	// Update medical history if provided
	if medicalHistory != "" {
		patient.MedicalHistory = medicalHistory
	}

	patient.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, patient)
	if err != nil {
		return nil, err
	}

	return patient, nil
}

// DeletePatient deletes a patient by ID
func (s *patientService) DeletePatient(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
