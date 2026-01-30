package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
	Proof    ProofConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	Environment string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// StorageConfig holds object storage configuration
type StorageConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// ProofConfig holds proof system configuration
type ProofConfig struct {
	EnableCommitment bool
	EnableGroth16    bool
	EnablePLONK      bool
	EnableSTARK      bool
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	FreeTier int
	ProTier  int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("API_PORT", "8080"),
			Environment:  getEnv("ENV", "development"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 15)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 15)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 60)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "zapiki"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "zapiki"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns: getEnvAsInt("POSTGRES_MAX_CONNS", 25),
			MinConns: getEnvAsInt("POSTGRES_MIN_CONNS", 5),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Storage: StorageConfig{
			Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
			Bucket:    getEnv("S3_BUCKET", "zapiki-proofs"),
			AccessKey: getEnv("S3_ACCESS_KEY", ""),
			SecretKey: getEnv("S3_SECRET_KEY", ""),
			UseSSL:    getEnvAsBool("S3_USE_SSL", false),
		},
		Proof: ProofConfig{
			EnableCommitment: getEnvAsBool("ENABLE_COMMITMENT", true),
			EnableGroth16:    getEnvAsBool("ENABLE_GROTH16", false),
			EnablePLONK:      getEnvAsBool("ENABLE_PLONK", false),
			EnableSTARK:      getEnvAsBool("ENABLE_STARK", false),
		},
		RateLimit: RateLimitConfig{
			FreeTier: getEnvAsInt("RATE_LIMIT_FREE_TIER", 10),
			ProTier:  getEnvAsInt("RATE_LIMIT_PRO_TIER", 1000),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	if !c.Proof.EnableCommitment && !c.Proof.EnableGroth16 && !c.Proof.EnablePLONK && !c.Proof.EnableSTARK {
		return fmt.Errorf("at least one proof system must be enabled")
	}

	return nil
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
