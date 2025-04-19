package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whitewalker-sa/ehass/internal/config"
	"github.com/whitewalker-sa/ehass/internal/router"
	"github.com/whitewalker-sa/ehass/pkg/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Check if running migration commands
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		handleMigrations(cfg, logger, os.Args)
		return
	}

	// Setup router with all dependencies
	r, cleanup, err := router.Setup(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to setup router", zap.Error(err))
	}
	defer cleanup()

	// Configure server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

// handleMigrations runs database migrations based on command line arguments
func handleMigrations(cfg *config.Config, logger *zap.Logger, args []string) {
	logger.Info("Setting up database connection for migrations")
	db, err := database.NewDatabase(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database connection", zap.Error(err))
		return
	}
	defer sqlDB.Close()

	// Determine migration action
	isRollback := len(args) > 2 && args[2] == "rollback"

	if isRollback {
		logger.Info("Rolling back the last migration")
		// For simplicity, we don't implement actual rollback logic here
		// In a real app, you would track migrations in a migrations table
		logger.Info("Migration rollback is not implemented")
	} else {
		logger.Info("Running migrations")
		if err := runMigrations(db, logger); err != nil {
			logger.Fatal("Migration failed", zap.Error(err))
			return
		}
		logger.Info("Migrations completed successfully")
	}
}

// runMigrations performs the actual database migrations
func runMigrations(db *gorm.DB, logger *zap.Logger) error {
	// Auto-migrate all models
	logger.Info("Running auto-migrations for all models")
	return database.AutoMigrate(db, logger)
}
