package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Redis    RedisConfig
	OAuth    OAuthConfig
	Email    EmailConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BaseURL      string
}

// DatabaseConfig holds database connection details
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxOpen  int
	MaxIdle  int
	Lifetime time.Duration
}

// AuthConfig holds authentication related configuration
type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// RedisConfig holds Redis connection details
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	GitHub GitHubConfig
	Google GoogleConfig
}

// GitHubConfig holds GitHub OAuth configuration
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.readTimeout", time.Second*10)
	viper.SetDefault("server.writeTimeout", time.Second*10)
	viper.SetDefault("server.idleTimeout", time.Second*60)
	viper.SetDefault("server.baseURL", "http://localhost:8080")

	// Database defaults
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.maxOpen", 25)
	viper.SetDefault("database.maxIdle", 5)
	viper.SetDefault("database.lifetime", time.Minute*5)

	// Auth defaults
	viper.SetDefault("auth.accessTokenExpiry", time.Hour)
	viper.SetDefault("auth.refreshTokenExpiry", time.Hour*24*7)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)

	// Email defaults
	viper.SetDefault("email.smtpPort", 587)
	viper.SetDefault("email.fromEmail", "noreply@ehass.com")
}
