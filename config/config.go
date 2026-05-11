package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// App
	AppPort string
	AppEnv  string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// JWT
	JWTSecret      string
	JWTExpireHours int

	// Midtrans
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string
}

// Load reads configuration from environment variables.
// It attempts to load a .env file first, but will not fail if the file
// does not exist (for production environments that inject env vars via the system).
func Load() (*Config, error) {
	// Load .env file if present; ignore error if file does not exist.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		// godotenv returns a plain string error, not os.ErrNotExist,
		// so we check the message to distinguish "file not found" from real errors.
		if !isFileNotFound(err) {
			return nil, fmt.Errorf("config: failed to load .env file: %w", err)
		}
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8181"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),

		JWTSecret: os.Getenv("JWT_SECRET"),

		MidtransServerKey: os.Getenv("MIDTRANS_SERVER_KEY"),
		MidtransClientKey: os.Getenv("MIDTRANS_CLIENT_KEY"),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
	}

	// Parse JWT expiry hours; default to 24 if not set or invalid.
	if raw := os.Getenv("JWT_EXPIRE_HOURS"); raw != "" {
		hours, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("config: JWT_EXPIRE_HOURS must be a valid integer, got %q", raw)
		}
		cfg.JWTExpireHours = hours
	} else {
		cfg.JWTExpireHours = 24
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required fields are present.
func (c *Config) validate() error {
	required := []struct {
		name  string
		value string
	}{
		{"DB_HOST", c.DBHost},
		{"DB_NAME", c.DBName},
		{"DB_USER", c.DBUser},
		{"DB_PASSWORD", c.DBPassword},
		{"JWT_SECRET", c.JWTSecret},
	}

	for _, f := range required {
		if f.value == "" {
			return fmt.Errorf("config: required environment variable %q is not set or empty", f.name)
		}
	}

	return nil
}

// MysqlDSN returns a MySQL DSN string suitable for sql.Open or gorm.Open.
func (c *Config) MysqlDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// RedisAddr returns the Redis address in host:port format.
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set or is empty.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// isFileNotFound checks whether a godotenv error indicates the .env file
// was simply not found (which is acceptable in production).
func isFileNotFound(err error) bool {
	if err == nil {
		return false
	}
	// godotenv wraps the underlying os error message in a plain string.
	return os.IsNotExist(err) || containsNotFound(err.Error())
}

func containsNotFound(msg string) bool {
	return strings.Contains(msg, "no such file") || strings.Contains(msg, "cannot find")
}
