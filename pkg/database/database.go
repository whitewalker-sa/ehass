package database

import (
	"fmt"
	"time"

	"github.com/whitewalker-sa/ehass/internal/config"
	"github.com/whitewalker-sa/ehass/internal/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetConnMaxLifetime(cfg.Database.Lifetime)

	log.Info("Connected to database",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.Name),
	)

	return db, nil
}

// AutoMigrate automatically migrates the database schema
func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	start := time.Now()
	log.Info("Running database migrations")

	// Add all models here for auto-migration
	err := db.AutoMigrate(
		&model.User{},
		&model.Doctor{},
		&model.Patient{},
		&model.Appointment{},
		&model.Session{},
		&model.Availability{},
		&model.MedicalRecord{},
		&model.AuditLog{},
	)

	if err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	log.Info("Database migrations completed", zap.Duration("duration", time.Since(start)))
	return nil
}
