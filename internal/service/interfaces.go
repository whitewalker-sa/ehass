package service

import (
	"context"

	"github.com/whitewalker-sa/ehass/internal/model"
)

// UserService defines the business logic for user operations
type UserService interface {
	Register(ctx context.Context, user *model.User, password string) error
	Login(ctx context.Context, email, password string) (*model.User, string, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error
}

// DoctorService defines the business logic for doctor operations
type DoctorService interface {
	CreateDoctor(ctx context.Context, doctor *model.Doctor, user *model.User, password string) error
	GetDoctorByID(ctx context.Context, id uint) (*model.Doctor, error)
	GetDoctorByUserID(ctx context.Context, userID uint) (*model.Doctor, error)
	ListDoctors(ctx context.Context, page, pageSize int) ([]*model.Doctor, int64, error)
	ListDoctorsBySpecialty(ctx context.Context, specialty string, page, pageSize int) ([]*model.Doctor, int64, error)
	UpdateDoctor(ctx context.Context, doctor *model.Doctor) error
	DeleteDoctor(ctx context.Context, id uint) error

	// Availability management
	AddAvailability(ctx context.Context, availability *model.Availability) error
	GetAvailabilityByDoctorID(ctx context.Context, doctorID uint) ([]*model.Availability, error)
	UpdateAvailability(ctx context.Context, availability *model.Availability) error
	DeleteAvailability(ctx context.Context, id uint) error
}

// PatientService defines the business logic for patient operations
type PatientService interface {
	CreatePatient(ctx context.Context, patient *model.Patient, user *model.User, password string) error
	GetPatientByID(ctx context.Context, id uint) (*model.Patient, error)
	GetPatientByUserID(ctx context.Context, userID uint) (*model.Patient, error)
	UpdatePatient(ctx context.Context, patient *model.Patient) error
	DeletePatient(ctx context.Context, id uint) error

	// Medical record management
	AddMedicalRecord(ctx context.Context, record *model.MedicalRecord) error
	GetMedicalRecords(ctx context.Context, patientID uint, page, pageSize int) ([]*model.MedicalRecord, int64, error)
}

// AppointmentService defines the business logic for appointment operations
type AppointmentService interface {
	CreateAppointment(ctx context.Context, appointment *model.Appointment) error
	GetAppointmentByID(ctx context.Context, id uint) (*model.Appointment, error)
	GetPatientAppointments(ctx context.Context, patientID uint, page, pageSize int) ([]*model.Appointment, int64, error)
	GetDoctorAppointments(ctx context.Context, doctorID uint, page, pageSize int) ([]*model.Appointment, int64, error)
	GetDoctorSchedule(ctx context.Context, doctorID uint, startDate, endDate string, page, pageSize int) ([]*model.Appointment, int64, error)
	UpdateAppointment(ctx context.Context, appointment *model.Appointment) error
	CancelAppointment(ctx context.Context, id uint) error
	CompleteAppointment(ctx context.Context, id uint, notes string) error
}
