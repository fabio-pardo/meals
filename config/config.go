package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
}

// ServerConfig holds all server related configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Environment  string
}

// DatabaseConfig holds all database related configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// RedisConfig holds all Redis related configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// AuthConfig holds all authentication related configuration
type AuthConfig struct {
	GoogleKey         string
	GoogleSecret      string
	GoogleRedirectURL string
	SessionSecret     string
}

// AppConfig is the global configuration instance
var AppConfig Config

// InitConfig initializes the application configuration
func InitConfig() {
	// Set defaults
	setDefaults()

	// Set environment and load env-specific config
	env := getEnvironment()
	log.Printf("Loading configuration for environment: %s", env)

	// Determine absolute config directory
	_, b, _, _ := runtime.Caller(0)
	configDir := filepath.Dir(b)

	// Load base config file
	baseFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(baseFile); err == nil {
		viper.SetConfigFile(baseFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("Error reading base config file %s: %v", baseFile, err)
		}
	} else {
		log.Printf("No base config file found at %s, using defaults", baseFile)
	}

	// Load environment-specific config file
	envFile := filepath.Join(configDir, fmt.Sprintf("config.%s.yaml", env))
	if _, err := os.Stat(envFile); err == nil {
		viper.SetConfigFile(envFile)
		if err := viper.MergeInConfig(); err != nil {
			log.Printf("Error merging environment config file %s: %v", envFile, err)
		}
	} else {
		log.Printf("No environment-specific config file found at %s, using base/defaults", envFile)
	}

	// Override with environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Unmarshal the configuration
	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	// Ensure environment is set
	AppConfig.Server.Environment = env

	// Print the configuration for debugging (not in production)
	if env != "production" {
		log.Printf("Loaded configuration: %+v", AppConfig)
	}
}

// getEnvironment determines the current environment
func getEnvironment() string {
	// Check environment variable first
	env := os.Getenv("APP_ENV")
	if env == "" {
		// Default to development
		env = "development"
	}

	// Validate environment
	switch env {
	case "development", "test", "production":
		return env
	default:
		log.Printf("Unknown environment: %s, defaulting to development", env)
		return "development"
	}
}

// setDefaults sets the default values for the configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.readTimeout", 10*time.Second)
	viper.SetDefault("server.writeTimeout", 10*time.Second)
	viper.SetDefault("server.environment", "development")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.name", "meals_db")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host, c.User, c.Password, c.Name, c.Port, c.SSLMode)
}

// GetServerAddress returns the server address in the format host:port
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
