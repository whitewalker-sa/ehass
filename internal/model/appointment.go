package model

import (
	"time"
)

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusNoShow    AppointmentStatus = "no_show"
)

// Appointment represents a medical appointment in the system
type Appointment struct {
	ID             uint              `json:"id" gorm:"primaryKey"`
	PatientID      uint              `json:"patient_id" gorm:"index;not null"`
	Patient        Patient           `json:"patient" gorm:"foreignKey:PatientID"`
	DoctorID       uint              `json:"doctor_id" gorm:"index;not null"`
	Doctor         Doctor            `json:"doctor" gorm:"foreignKey:DoctorID"`
	ScheduledStart time.Time         `json:"scheduled_start" gorm:"index;not null"`
	ScheduledEnd   time.Time         `json:"scheduled_end" gorm:"not null"`
	Status         AppointmentStatus `json:"status" gorm:"size:20;default:'pending'"`
	Notes          string            `json:"notes" gorm:"type:text"`
	Reason         string            `json:"reason" gorm:"size:255"`
	Type           string            `json:"type" gorm:"size:50;default:'in_person'"` // in_person, video, phone
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// TableName overrides the table name
func (Appointment) TableName() string {
	return "appointments"
}

// Session represents a user session
type Session struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Token     string    `json:"-" gorm:"size:500;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	UserAgent string    `json:"user_agent" gorm:"size:255"`
	IP        string    `json:"ip" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (Session) TableName() string {
	return "sessions"
}

// AuditLog represents system audit logs
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Action     string    `json:"action" gorm:"size:100;not null"`
	EntityID   uint      `json:"entity_id"`
	EntityType string    `json:"entity_type" gorm:"size:50"`
	OldValue   string    `json:"old_value" gorm:"type:text"`
	NewValue   string    `json:"new_value" gorm:"type:text"`
	IP         string    `json:"ip" gorm:"size:50"`
	UserAgent  string    `json:"user_agent" gorm:"size:255"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName overrides the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}
