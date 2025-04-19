package router

import (
	"github.com/gin-gonic/gin"
	"github.com/whitewalker-sa/ehass/internal/handler"
)

// SetupRouter sets up the API routes
func SetupRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	doctorHandler *handler.DoctorHandler,
	patientHandler *handler.PatientHandler,
	appointmentHandler *handler.AppointmentHandler,
	authMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	// Public routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/oauth/login", authHandler.OAuthLogin)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/request-password-reset", authHandler.RequestPasswordReset)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/refresh-token", authHandler.RefreshToken)
			auth.POST("/verify-2fa", authHandler.Verify2FA)
		}

		// Protected routes
		protected := v1.Group("/", authMiddleware)
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/:id", userHandler.GetUserByID) // Changed to match actual implementation
				users.PUT("/:id", userHandler.UpdateProfile)
				users.PUT("/:id/change-password", userHandler.ChangePassword)
			}

			// Authentication management routes
			authManagement := protected.Group("/auth")
			{
				authManagement.POST("/logout", authHandler.Logout)
				authManagement.POST("/setup-2fa", authHandler.Setup2FA)
				authManagement.POST("/enable-2fa", authHandler.Enable2FA)
				authManagement.POST("/disable-2fa", authHandler.Disable2FA)
				authManagement.POST("/link-oauth", authHandler.LinkOAuth)
			}

			// Doctor routes
			doctors := protected.Group("/doctors")
			{
				doctors.POST("", doctorHandler.CreateDoctor)
				doctors.GET("", doctorHandler.ListDoctors)
				doctors.GET("/:id", doctorHandler.GetDoctor)
				doctors.PUT("/:id", doctorHandler.UpdateDoctor)
				doctors.GET("/specialty/:specialty", doctorHandler.ListDoctorsBySpecialty)
				doctors.GET("/user/:userID", doctorHandler.GetDoctorByUser)
			}

			// Patient routes
			patients := protected.Group("/patients")
			{
				patients.POST("", patientHandler.CreatePatient)
				patients.GET("/:id", patientHandler.GetPatient)
				patients.PUT("/:id", patientHandler.UpdatePatient)
				patients.GET("/user/:userID", patientHandler.GetPatientByUser)
			}

			// Appointment routes
			appointments := protected.Group("/appointments")
			{
				appointments.POST("", appointmentHandler.CreateAppointment)
				appointments.GET("/:id", appointmentHandler.GetAppointmentByID)
				appointments.PUT("/:id", appointmentHandler.UpdateAppointment)
				appointments.GET("/patient/:patientID", appointmentHandler.GetPatientAppointments)
				appointments.GET("/doctor/:doctorID", appointmentHandler.GetDoctorAppointments)
				appointments.GET("/doctor/:doctorID/schedule", appointmentHandler.GetDoctorSchedule)
			}
		}
	}

	return r
}
