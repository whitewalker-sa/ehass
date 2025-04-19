package middleware

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestIDKey is the context key for the request ID
const RequestIDKey = "request_id"

// LogSamplingConfig represents configuration for log sampling
type LogSamplingConfig struct {
	// Enabled indicates if sampling is enabled
	Enabled bool
	// Rate is the sampling rate (0.0-1.0), e.g., 0.1 = 10% of logs
	Rate float64
	// Always log errors regardless of sampling
	AlwaysLogErrors bool
}

// DefaultSamplingConfig provides reasonable defaults for log sampling
var DefaultSamplingConfig = LogSamplingConfig{
	Enabled:         false,
	Rate:            0.1,
	AlwaysLogErrors: true,
}

// Logger middleware attaches a structured logger with request context to the request
func Logger(log *zap.Logger) echo.MiddlewareFunc {
	return LoggerWithConfig(log, DefaultSamplingConfig)
}

// LoggerWithConfig creates a middleware with configurable log sampling
func LoggerWithConfig(log *zap.Logger, samplingConfig LogSamplingConfig) echo.MiddlewareFunc {
	// Initialize random seed for sampling
	rand.Seed(time.Now().UnixNano())

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			// Check for existing request ID from header
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				// Generate new request ID if not provided
				requestID = uuid.New().String()
				req.Header.Set("X-Request-ID", requestID)
			}

			// Set response header with request ID for traceability
			res.Header().Set("X-Request-ID", requestID)

			// Store request ID in context
			c.Set(RequestIDKey, requestID)

			// Create request-scoped logger with request ID and base data
			requestLogger := log.With(
				zap.String("request_id", requestID),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("remote_ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
				zap.String("referer", req.Referer()),
			)

			// Extract useful information from request headers for additional context
			if contentType := req.Header.Get("Content-Type"); contentType != "" {
				requestLogger = requestLogger.With(zap.String("content_type", contentType))
			}

			// Extract user identity if available in context
			if userID := GetUserID(c); userID != "" {
				requestLogger = requestLogger.With(zap.String("user_id", userID))
			}

			// Store logger in context
			c.Set("logger", requestLogger)

			// Create context with the same logger
			ctx := context.WithValue(req.Context(), "logger", requestLogger)
			c.SetRequest(req.WithContext(ctx))

			// Process request
			err := next(c)

			// Determine if we should log based on sampling configuration
			shouldLog := true
			if samplingConfig.Enabled {
				shouldLog = (rand.Float64() <= samplingConfig.Rate)

				// Always log errors if configured that way
				if !shouldLog && samplingConfig.AlwaysLogErrors && (err != nil || res.Status >= 400) {
					shouldLog = true
				}
			}

			// Log request completion if we should log
			if shouldLog {
				// Calculate request duration
				latency := time.Since(start)

				// Add response information to log
				requestLogger.Info("Request completed",
					zap.Int("status", res.Status),
					zap.Int64("size", res.Size),
					zap.Duration("latency", latency),
					zap.NamedError("error", err),
				)
			}

			return err
		}
	}
}

// GetRequestLogger extracts the logger with request context from the echo context
func GetRequestLogger(c echo.Context) *zap.Logger {
	if logger, ok := c.Get("logger").(*zap.Logger); ok {
		return logger
	}
	// Return a no-op logger if not found to avoid nil panic
	return zap.NewNop()
}

// GetRequestID extracts the request ID from the context
func GetRequestID(c echo.Context) string {
	if id, ok := c.Get(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID attempts to extract a user ID from the context
// This can be customized based on where/how user IDs are stored in your application
func GetUserID(c echo.Context) string {
	// Try to get from context - customize based on your auth implementation
	if userID, ok := c.Get("user_id").(string); ok {
		return userID
	}
	return ""
}
