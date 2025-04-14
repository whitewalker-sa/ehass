package model

import (
	"time"
)

// Patient represents a patient in the system
type Patient struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	User              User      `json:"user" gorm:"foreignKey:UserID"`
	DateOfBirth       time.Time `json:"date_of_birth"`
	Gender            string    `json:"gender" gorm:"size:20"`
	BloodGroup        string    `json:"blood_group" gorm:"size:10"`
	EmergencyContact  string    `json:"emergency_contact" gorm:"size:100"`
	MedicalHistory    string    `json:"medical_history" gorm:"type:text"`
	Allergies         string    `json:"allergies" gorm:"type:text"`
	CurrentMedication string    `json:"current_medication" gorm:"type:text"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (Patient) TableName() string {
	return "patients"
}

// MedicalRecord represents a patient's medical record
type MedicalRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PatientID    uint      `json:"patient_id" gorm:"index;not null"`
	Patient      Patient   `json:"-" gorm:"foreignKey:PatientID"`
	DoctorID     uint      `json:"doctor_id" gorm:"index;not null"`
	Doctor       Doctor    `json:"-" gorm:"foreignKey:DoctorID"`
	Diagnosis    string    `json:"diagnosis" gorm:"type:text"`
	Prescription string    `json:"prescription" gorm:"type:text"`
	Notes        string    `json:"notes" gorm:"type:text"`
	VisitDate    time.Time `json:"visit_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (MedicalRecord) TableName() string {
	return "medical_records"
}
