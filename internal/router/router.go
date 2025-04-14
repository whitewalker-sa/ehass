package router

import (
	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/config"
	"github.com/whitewalker-sa/ehass/internal/handler"
	"github.com/whitewalker-sa/ehass/internal/middleware"
	"github.com/whitewalker-sa/ehass/internal/model"
	"github.com/whitewalker-sa/ehass/internal/repository"
	"github.com/whitewalker-sa/ehass/internal/service"
	"github.com/whitewalker-sa/ehass/pkg/database"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup initializes the router and sets up all routes and middleware
func Setup(cfg *config.Config, logger *zap.Logger) (*gin.Engine, func(), error) {
	// Create router with default middleware
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Initialize database
	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		return nil, nil, err
	}

	// Auto migrate database schema
	if err := database.AutoMigrate(db, logger); err != nil {
		return nil, nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	doctorRepo := repository.NewDoctorRepository(db)
	patientRepo := repository.NewPatientRepository(db)
	appointmentRepo := repository.NewAppointmentRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, cfg, logger)
	appointmentService := service.NewAppointmentService(appointmentRepo, doctorRepo, patientRepo, logger)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, logger)
	appointmentHandler := handler.NewAppointmentHandler(appointmentService, logger)

	// Root API group
	api := r.Group("/api")

	// Health check endpoint at root API level
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 group
	v1 := api.Group("/v1")

	// Authentication routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// Protected routes - all require authentication
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/profile", userHandler.GetProfile)
		users.PUT("/profile", userHandler.UpdateProfile)
		users.PUT("/change-password", userHandler.ChangePassword)
	}

	// Appointment routes
	appointments := protected.Group("/appointments")
	{
		appointments.POST("", appointmentHandler.CreateAppointment)
		appointments.GET("/:id", appointmentHandler.GetAppointmentByID)
		appointments.PUT("/:id", appointmentHandler.UpdateAppointment)
		appointments.POST("/:id/cancel", appointmentHandler.CancelAppointment)
		appointments.POST("/:id/complete", appointmentHandler.CompleteAppointment)
	}

	// Patient routes
	patients := protected.Group("/patients")
	{
		patients.GET("/:patient_id/appointments", appointmentHandler.GetPatientAppointments)
	}

	// Doctor routes
	doctors := protected.Group("/doctors")
	{
		doctors.GET("/:doctor_id/appointments", appointmentHandler.GetDoctorAppointments)
		doctors.GET("/:doctor_id/schedule", appointmentHandler.GetDoctorSchedule)
	}

	// Admin routes (restricted to admin users)
	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware(model.RoleAdmin))
	{
		admin.GET("/users/:id", userHandler.GetUserByID)
	}

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Cleanup function
	cleanup := func() {
		logger.Info("Cleaning up resources")
		// Any cleanup logic here
	}

	return r, cleanup, nil
}
