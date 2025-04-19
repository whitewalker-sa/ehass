package repository

import (
	"context"

	"github.com/whitewalker-sa/ehass/internal/model"
)

// UserRepository defines operations for user data access
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
}

// DoctorRepository defines operations for doctor data access
type DoctorRepository interface {
	Create(ctx context.Context, doctor *model.Doctor) error
	FindByID(ctx context.Context, id uint) (*model.Doctor, error)
	FindByUserID(ctx context.Context, userID uint) (*model.Doctor, error)
	FindAll(ctx context.Context, limit, offset int) ([]*model.Doctor, int64, error)
	FindBySpecialty(ctx context.Context, specialty string, limit, offset int) ([]*model.Doctor, int64, error)
	Update(ctx context.Context, doctor *model.Doctor) error
	Delete(ctx context.Context, id uint) error
}

// AvailabilityRepository defines operations for doctor availability data access
type AvailabilityRepository interface {
	Create(ctx context.Context, availability *model.Availability) error
	FindByDoctorID(ctx context.Context, doctorID uint) ([]*model.Availability, error)
	Update(ctx context.Context, availability *model.Availability) error
	Delete(ctx context.Context, id uint) error
}

// PatientRepository defines operations for patient data access
type PatientRepository interface {
	Create(ctx context.Context, patient *model.Patient) error
	FindByID(ctx context.Context, id uint) (*model.Patient, error)
	FindByUserID(ctx context.Context, userID uint) (*model.Patient, error)
	Update(ctx context.Context, patient *model.Patient) error
	Delete(ctx context.Context, id uint) error
}

// AppointmentRepository defines the repository interface for appointment operations
type AppointmentRepository interface {
	Create(ctx context.Context, appointment *model.Appointment) error
	FindByID(ctx context.Context, id uint) (*model.Appointment, error)
	FindByPatientID(ctx context.Context, patientID uint, limit, offset int) ([]*model.Appointment, int64, error)
	FindByDoctorID(ctx context.Context, doctorID uint, limit, offset int) ([]*model.Appointment, int64, error)
	FindByDateRange(ctx context.Context, doctorID uint, startDate, endDate string, limit, offset int) ([]*model.Appointment, int64, error)
	Update(ctx context.Context, appointment *model.Appointment) error
	Delete(ctx context.Context, id uint) error
}

// SessionRepository defines operations for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	FindByToken(ctx context.Context, token string) (*model.Session, error)
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

// MedicalRecordRepository defines operations for medical record data access
type MedicalRecordRepository interface {
	Create(ctx context.Context, record *model.MedicalRecord) error
	FindByID(ctx context.Context, id uint) (*model.MedicalRecord, error)
	FindByPatientID(ctx context.Context, patientID uint, limit, offset int) ([]*model.MedicalRecord, int64, error)
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, id uint) error
}

// AuditLogRepository defines operations for audit log data access
type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.AuditLog, int64, error)
	FindByEntityTypeAndID(ctx context.Context, entityType string, entityID uint, limit, offset int) ([]*model.AuditLog, int64, error)
}
