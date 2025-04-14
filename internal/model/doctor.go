package model

import (
	"time"
)

// Doctor represents a doctor in the system
type Doctor struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	Specialty   string    `json:"specialty" gorm:"size:100;not null"`
	Designation string    `json:"designation" gorm:"size:100"`
	Education   string    `json:"education" gorm:"size:255"`
	Experience  int       `json:"experience" gorm:"default:0"`
	LicenseNo   string    `json:"license_no" gorm:"size:100"`
	Bio         string    `json:"bio" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (Doctor) TableName() string {
	return "doctors"
}

// Availability represents a doctor's available time slots
type Availability struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DoctorID  uint      `json:"doctor_id" gorm:"index"`
	Doctor    Doctor    `json:"-" gorm:"foreignKey:DoctorID"`
	DayOfWeek int       `json:"day_of_week" gorm:"type:smallint"` // 0-6 for Sunday-Saturday
	StartTime string    `json:"start_time" gorm:"type:time"`      // Format: HH:MM:SS
	EndTime   string    `json:"end_time" gorm:"type:time"`        // Format: HH:MM:SS
	Duration  int       `json:"duration" gorm:"default:30"`       // Duration in minutes
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the table name
func (Availability) TableName() string {
	return "availability"
}
