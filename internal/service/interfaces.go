package service

import (
	"context"

	"github.com/whitewalker-sa/ehass/internal/model"
)

// AuthService defines authentication service operations
type AuthService interface {
	Register(ctx context.Context, name, email, password string, role model.Role) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, string, *model.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	VerifyEmail(ctx context.Context, token string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// OAuth related
	OAuthLogin(ctx context.Context, provider model.AuthProvider, providerToken string) (string, string, *model.User, error)
	LinkOAuthAccount(ctx context.Context, userID uint, provider model.AuthProvider, providerToken string) error

	// 2FA related
	Setup2FA(ctx context.Context, userID uint) (string, error)
	Verify2FA(ctx context.Context, userID uint, token string) (bool, error)
	Enable2FA(ctx context.Context, userID uint, secret, token string) error
	Disable2FA(ctx context.Context, userID uint, password string) error

	// Session management
	Logout(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*model.User, error)
}

// UserService defines user management operations
type UserService interface {
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	UpdateUserProfile(ctx context.Context, id uint, name, phone, address string) (*model.User, error)
	ChangePassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	DeleteUser(ctx context.Context, id uint) error
	UpdateAvatar(ctx context.Context, id uint, avatarURL string) (*model.User, error)
}

// DoctorService defines doctor management operations
type DoctorService interface {
	CreateDoctor(ctx context.Context, userID uint, specialty, bio string, experience int) (*model.Doctor, error)
	GetDoctorByID(ctx context.Context, id uint) (*model.Doctor, error)
	GetDoctorByUserID(ctx context.Context, userID uint) (*model.Doctor, error)
	UpdateDoctorProfile(ctx context.Context, id uint, specialty, bio string, experience int) (*model.Doctor, error)
	GetAllDoctors(ctx context.Context, page, pageSize int) ([]*model.Doctor, int64, error)
	GetDoctorsBySpecialty(ctx context.Context, specialty string, page, pageSize int) ([]*model.Doctor, int64, error)
	DeleteDoctor(ctx context.Context, id uint) error
}

// PatientService defines patient management operations
type PatientService interface {
	CreatePatient(ctx context.Context, userID uint, dateOfBirth, medicalHistory string) (*model.Patient, error)
	GetPatientByID(ctx context.Context, id uint) (*model.Patient, error)
	GetPatientByUserID(ctx context.Context, userID uint) (*model.Patient, error)
	UpdatePatientProfile(ctx context.Context, id uint, dateOfBirth, medicalHistory string) (*model.Patient, error)
}

// AppointmentService defines appointment management operations
type AppointmentService interface {
	CreateAppointment(ctx context.Context, patientID, doctorID uint, date, time, reason string) (*model.Appointment, error)
	GetAppointmentByID(ctx context.Context, id uint) (*model.Appointment, error)
	GetPatientAppointments(ctx context.Context, patientID uint, page, pageSize int) ([]*model.Appointment, int64, error)
	GetDoctorAppointments(ctx context.Context, doctorID uint, page, pageSize int) ([]*model.Appointment, int64, error)
	GetDoctorAppointmentsByDateRange(ctx context.Context, doctorID uint, startDate, endDate string, page, pageSize int) ([]*model.Appointment, int64, error)
	UpdateAppointment(ctx context.Context, id uint, date, time, status, reason string) (*model.Appointment, error)
	CancelAppointment(ctx context.Context, id uint) error
	CompleteAppointment(ctx context.Context, id uint, notes string) error
}

// AvailabilityService defines availability management operations
type AvailabilityService interface {
	AddAvailability(ctx context.Context, doctorID uint, day string, startTime, endTime string) (*model.Availability, error)
	GetDoctorAvailability(ctx context.Context, doctorID uint) ([]*model.Availability, error)
	UpdateAvailability(ctx context.Context, id uint, day string, startTime, endTime string) (*model.Availability, error)
	RemoveAvailability(ctx context.Context, id uint) error
}

// MedicalRecordService defines medical record management operations
type MedicalRecordService interface {
	CreateMedicalRecord(ctx context.Context, patientID, doctorID uint, diagnosis, prescription, notes string) (*model.MedicalRecord, error)
	GetMedicalRecordByID(ctx context.Context, id uint) (*model.MedicalRecord, error)
	GetPatientMedicalRecords(ctx context.Context, patientID uint, page, pageSize int) ([]*model.MedicalRecord, int64, error)
	UpdateMedicalRecord(ctx context.Context, id uint, diagnosis, prescription, notes string) (*model.MedicalRecord, error)
	DeleteMedicalRecord(ctx context.Context, id uint) error
}
