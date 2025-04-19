package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/config"
	"github.com/whitewalker-sa/ehass/internal/handler"
	"github.com/whitewalker-sa/ehass/internal/middleware"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"github.com/whitewalker-sa/ehass/internal/service"
	"github.com/whitewalker-sa/ehass/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup initializes all dependencies and returns the router
func Setup(cfg *config.Config, logger *zap.Logger) (*gin.Engine, func(), error) {
	// Connect to database
	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	doctorRepo := repository.NewDoctorRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)
	// Implement or comment out the availability repository for now
	// availabilityRepo := repository.NewAvailabilityRepository(db)
	authRepo := repository.NewAuthRepository(db)

	// Setup services
	emailService := service.NewEmailService(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.FromEmail,
		cfg.Server.BaseURL,
	)

	oauthService := service.NewOAuthService(
		cfg.OAuth.GitHub.ClientID,
		cfg.OAuth.GitHub.ClientSecret,
		cfg.OAuth.Google.ClientID,
		cfg.OAuth.Google.ClientSecret,
	)

	authService := service.NewAuthService(
		authRepo,
		cfg.Auth.AccessTokenSecret,
		int(cfg.Auth.AccessTokenExpiry.Minutes()),
		emailService,
		oauthService,
	)

	userService := service.NewUserService(userRepo, cfg, logger)
	// Implement these services or use simpler constructors
	doctorService := service.NewDoctorService(doctorRepo, logger)
	patientService := service.NewPatientService(patientRepo, logger)
	appointmentService := service.NewAppointmentService(appointmentRepo, doctorRepo, patientRepo, logger)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Setup handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, logger)
	doctorHandler := handler.NewDoctorHandler(doctorService, logger)
	patientHandler := handler.NewPatientHandler(patientService, logger)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService, logger)

	// Setup router
	router := SetupRouter(
		authHandler,
		userHandler,
		doctorHandler,
		patientHandler,
		appointmentHandler,
		authMiddleware,
	)

	// Setup cleanup function
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get database connection", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}

	return router, cleanup, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	// Add models to be migrated
	return db.AutoMigrate(
		&model.User{},
		&model.Doctor{},
		&model.Patient{},
		&model.Appointment{},
		&model.VerificationToken{},
	)
}
