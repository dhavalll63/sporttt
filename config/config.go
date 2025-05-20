package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	App struct {
		Env         string `env:"APP_ENV" envDefault:"development"`
		Port        string `env:"PORT"    envDefault:"8088"`
		FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
		UploadDir   string `env:"UPLOAD_DIR"   envDefault:"./public/uploads"`
	}
	DB struct {
		Host     string `env:"DB_HOST"     envDefault:"localhost"`
		Port     string `env:"DB_PORT"     envDefault:"5432"`
		User     string `env:"DB_USER"     envDefault:"postgres"`
		Password string `env:"DB_PASSWORD" envDefault:"password"`
		Name     string `env:"DB_NAME"     envDefault:"miow_db"`
		SSLMode  string `env:"DB_SSLMODE"  envDefault:"disable"`
	}
	JWT struct {
		AccessTokenSecret        string `env:"JWT_ACCESS_TOKEN_SECRET"  envDefault:"supersecret"`
		AccessTokenExpiryMinutes int    `env:"JWT_ACCESS_TOKEN_EXPIRY_MINUTES" envDefault:"15"`
		RefreshTokenSecret       string `env:"JWT_REFRESH_TOKEN_SECRET" envDefault:"supersecretrefresh"`
		RefreshTokenExpiryDays   int    `env:"JWT_REFRESH_TOKEN_EXPIRY_DAYS"   envDefault:"7"`
	}
	// Add other configurations like Email, SMS services if needed
	// Email struct { ... }
	// SMS struct { ... }
}

// Global DB instance, accessible after ConnectDB() is called via Initialize.
var DB *gorm.DB

// Global AppConfig instance, accessible after LoadConfig() is called via Initialize.
var appConfig *Config
var once sync.Once // Used for singleton pattern to load config only once

// LoadConfig loads configuration from environment variables into the Config struct.
// It's designed to be called once.
func LoadConfig() (*Config, error) {
	// Load .env file. It's okay if it doesn't exist, especially in production
	// where env vars are set directly.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading, relying on system environment variables.")
	}

	cfg := &Config{}

	// --- App Configuration ---
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Port = getEnv("PORT", "8088")
	cfg.App.FrontendURL = getEnv("FRONTEND_URL", "http://localhost:3000")
	cfg.App.UploadDir = getEnv("UPLOAD_DIR", "./public/uploads") // Ensure this path is writable

	// --- Database Configuration ---
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	cfg.DB.User = getEnv("DB_USER", "postgres")
	cfg.DB.Password = getEnv("DB_PASSWORD", "password")
	cfg.DB.Name = getEnv("DB_NAME", "miow_db")
	cfg.DB.SSLMode = getEnv("DB_SSLMODE", "disable")

	// --- JWT Configuration ---
	cfg.JWT.AccessTokenSecret = getEnv("JWT_ACCESS_TOKEN_SECRET", "your-very-strong-access-secret")
	cfg.JWT.RefreshTokenSecret = getEnv("JWT_REFRESH_TOKEN_SECRET", "your-very-strong-refresh-secret")

	var err error
	cfg.JWT.AccessTokenExpiryMinutes, err = getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRY_MINUTES", 15)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_EXPIRY_MINUTES: %w", err)
	}
	cfg.JWT.RefreshTokenExpiryDays, err = getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRY_DAYS", 7)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_EXPIRY_DAYS: %w", err)
	}

	// Basic validation for critical secrets
	if cfg.JWT.AccessTokenSecret == "your-very-strong-access-secret" || cfg.JWT.RefreshTokenSecret == "your-very-strong-refresh-secret" {
		log.Println("WARNING: Using default JWT secrets. Please set JWT_ACCESS_TOKEN_SECRET and JWT_REFRESH_TOKEN_SECRET environment variables for production.")
	}
	if cfg.DB.Password == "password" && cfg.App.Env == "production" {
		log.Println("WARNING: Using default DB password in production. Please set DB_PASSWORD environment variable.")
	}

	appConfig = cfg // Set the global instance
	return cfg, nil
}

// ConnectDB establishes a connection to the database using the provided configuration.
// It sets the global DB variable.
func ConnectDB(dbCfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai", // Example TimeZone
		dbCfg.DB.Host,
		dbCfg.DB.User,
		dbCfg.DB.Password,
		dbCfg.DB.Name,
		dbCfg.DB.Port,
		dbCfg.DB.SSLMode,
	)

	gormConfig := &gorm.Config{}
	if dbCfg.App.Env == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info) // Log SQL queries in development
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent) // Less verbose in production
	}

	var err error
	gormDB, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = gormDB // Set the global DB instance
	log.Println("Successfully connected to database!")
	return gormDB, nil
}

// Initialize loads all configurations and connects to the database.
// This should be called once at the start of your application (e.g., in main.go).
func Initialize() error {
	var loadErr error
	// Load configuration only once
	once.Do(func() {
		loadedCfg, err := LoadConfig()
		if err != nil {
			loadErr = fmt.Errorf("failed to load configuration: %w", err)
			return
		}
		appConfig = loadedCfg // Ensure global appConfig is set

		_, err = ConnectDB(*appConfig) // Use the loaded configuration
		if err != nil {
			loadErr = fmt.Errorf("failed to connect to database during initialization: %w", err)
			return
		}
	})
	return loadErr
}

// GetConfig returns the loaded application configuration.
// It panics if the configuration has not been loaded yet,
// ensuring that configuration is always available when requested after Initialize().
func GetConfig() *Config {
	if appConfig == nil {
		// This should ideally not happen if Initialize() is called correctly in main.
		log.Fatal("Configuration not loaded. Call config.Initialize() first.")
	}
	return appConfig
}

// Helper function to get an environment variable or return a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Helper function to get an environment variable as an integer or return a default value.
func getEnvAsInt(key string, fallback int) (int, error) {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback, fmt.Errorf("env var %s: expected integer, got '%s'", key, valueStr)
	}
	return value, nil
}
