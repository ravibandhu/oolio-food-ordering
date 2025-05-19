package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment variables or files.
type Config struct {
	Server  ServerConfig  // Server-related settings
	Files   FilesConfig   // File paths and directories
	Logging LoggingConfig // Logging level and options
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port         string        // Port on which the server listens (e.g., ":8080")
	ReadTimeout  time.Duration // Maximum duration for reading the entire request
	WriteTimeout time.Duration // Maximum duration for writing the response
	IdleTimeout  time.Duration // Maximum duration to wait for the next request
}

// FilesConfig holds file and directory paths for data sources.
type FilesConfig struct {
	ProductsFile string // Path to the products JSON file
	CouponsDir   string // Directory containing gzipped coupon files
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string // Logging level (e.g., "info", "debug", "warn")
	Format string // Log format (e.g., "json", "text")
}

// Load reads configuration from environment variables and config files using Viper.
func Load() (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.AddConfigPath(configPath)
	}
	// v.AddConfigPath(".")
	v.AddConfigPath("../../config")

	// Configure environment variables
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map env vars to viper keys
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.readtimeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.writetimeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idletimeout", "SERVER_IDLE_TIMEOUT")
	v.BindEnv("files.productsfile", "PRODUCTS_FILE")
	v.BindEnv("files.couponsdir", "COUPONS_DIR")
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")

	// Set defaults
	v.SetDefault("server.port", ":8080")
	v.SetDefault("server.readtimeout", "15s")
	v.SetDefault("server.writetimeout", "15s")
	v.SetDefault("server.idletimeout", "60s")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Try to read config file (ignore error if not found)
	_ = v.ReadInConfig()

	// Parse durations
	readTimeout, err := time.ParseDuration(v.GetString("server.readtimeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid server.readtimeout: %w", err)
	}
	writeTimeout, err := time.ParseDuration(v.GetString("server.writetimeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid server.writetimeout: %w", err)
	}
	idleTimeout, err := time.ParseDuration(v.GetString("server.idletimeout"))
	if err != nil {
		return nil, fmt.Errorf("invalid server.idletimeout: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:         v.GetString("server.port"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		Files: FilesConfig{
			ProductsFile: v.GetString("files.productsfile"),
			CouponsDir:   v.GetString("files.couponsdir"),
		},
		Logging: LoggingConfig{
			Level:  v.GetString("logging.level"),
			Format: v.GetString("logging.format"),
		},
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if all required configuration fields are set and valid.
func (c *Config) validate() error {
	if c.Files.ProductsFile == "" {
		return fmt.Errorf("PRODUCTS_FILE is required")
	}
	if c.Files.CouponsDir == "" {
		return fmt.Errorf("COUPONS_DIR is required")
	}

	// Validate log level
	switch strings.ToLower(c.Logging.Level) {
	case "debug", "info", "warn", "error", "fatal", "panic":
		// Valid log levels
	default:
		return fmt.Errorf("invalid LOG_LEVEL: %s", c.Logging.Level)
	}

	// Validate log format
	switch strings.ToLower(c.Logging.Format) {
	case "json", "text":
		// Valid formats
	default:
		return fmt.Errorf("invalid LOG_FORMAT: %s", c.Logging.Format)
	}

	return nil
}

// GetServerTimeouts returns the server timeout configurations.
func (c *Config) GetServerTimeouts() (read, write, idle time.Duration) {
	return c.Server.ReadTimeout, c.Server.WriteTimeout, c.Server.IdleTimeout
}
