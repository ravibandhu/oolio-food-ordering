package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Server represents server configuration
type Server struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Files represents file paths configuration
type Files struct {
	ProductsFile string `mapstructure:"products_file"`
	CouponsDir   string `mapstructure:"coupons_dir"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // Logging level (e.g., "info", "debug", "warn")
	Format string `mapstructure:"format"` // Log format (e.g., "json", "text")
}

// Config represents the application configuration
type Config struct {
	Server  Server        `mapstructure:"server"`
	Files   Files         `mapstructure:"files"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// Load loads the configuration from the specified file and environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		log.Printf("Config path added successfully, %s", configPath)
		v.AddConfigPath(configPath)
	}
	// v.AddConfigPath(".")
	v.AddConfigPath("../../config")
	log.Printf("Config path added successfully, %s %s", v.GetString("files.productsfile"), v.GetString("files.couponsdir"))
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
		Server: Server{
			Port:         v.GetString("server.port"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		Files: Files{
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
