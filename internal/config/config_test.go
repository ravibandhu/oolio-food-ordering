package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test cases table
	tests := []struct {
		name        string
		envVars     map[string]string
		configFile  string
		wantErr     bool
		validateCfg func(*testing.T, *Config)
	}{
		{
			name: "valid config from env vars",
			envVars: map[string]string{
				"PRODUCTS_FILE":       "./testdata/products.json",
				"COUPONS_DIR":         "./testdata/coupons",
				"SERVER_PORT":         ":9090",
				"SERVER_READ_TIMEOUT": "20s",
				"LOG_LEVEL":           "debug",
				"LOG_FORMAT":          "text",
			},
			wantErr: false,
			validateCfg: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != ":9090" {
					t.Errorf("expected port :9090, got %s", cfg.Server.Port)
				}
				if cfg.Server.ReadTimeout != 20*time.Second {
					t.Errorf("expected read timeout 20s, got %v", cfg.Server.ReadTimeout)
				}
				if cfg.Files.ProductsFile != "./testdata/products.json" {
					t.Errorf("expected products file ./testdata/products.json, got %s", cfg.Files.ProductsFile)
				}
				if cfg.Files.CouponsDir != "./testdata/coupons" {
					t.Errorf("expected coupons dir ./testdata/coupons, got %s", cfg.Files.CouponsDir)
				}
				if cfg.Logging.Level != "debug" {
					t.Errorf("expected log level debug, got %s", cfg.Logging.Level)
				}
				if cfg.Logging.Format != "text" {
					t.Errorf("expected log format text, got %s", cfg.Logging.Format)
				}
			},
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"PRODUCTS_FILE": "./testdata/products.json",
				"COUPONS_DIR":   "./testdata/coupons",
				"LOG_LEVEL":     "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			envVars: map[string]string{
				"PRODUCTS_FILE": "./testdata/products.json",
				"COUPONS_DIR":   "./testdata/coupons",
				"LOG_FORMAT":    "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout duration",
			envVars: map[string]string{
				"PRODUCTS_FILE":       "./testdata/products.json",
				"COUPONS_DIR":         "./testdata/coupons",
				"SERVER_READ_TIMEOUT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid config from file with env var overrides",
			configFile: `server:
  port: ":8081"
  readtimeout: "30s"
  writetimeout: "30s"
  idletimeout: "120s"
files:
  productsfile: "./data/products.json"
  couponsdir: "./data/coupons"
logging:
  level: "info"
  format: "json"`,
			envVars: map[string]string{
				"SERVER_PORT":   ":9000",                  // Override port from config file
				"PRODUCTS_FILE": "./custom/products.json", // Override products file
				"COUPONS_DIR":   "./custom/coupons",       // Required field
			},
			wantErr: false,
			validateCfg: func(t *testing.T, cfg *Config) {
				// Environment variables should override config file
				if cfg.Server.Port != ":9000" {
					t.Errorf("expected port :9000 (from env), got %s", cfg.Server.Port)
				}
				if cfg.Files.ProductsFile != "./custom/products.json" {
					t.Errorf("expected products file ./custom/products.json (from env), got %s", cfg.Files.ProductsFile)
				}
				if cfg.Files.CouponsDir != "./custom/coupons" {
					t.Errorf("expected coupons dir ./custom/coupons (from env), got %s", cfg.Files.CouponsDir)
				}

				// Other values should be from config file
				if cfg.Server.ReadTimeout != 30*time.Second {
					t.Errorf("expected read timeout 15s, got %v", cfg.Server.ReadTimeout)
				}
				if cfg.Server.WriteTimeout != 30*time.Second {
					t.Errorf("expected write timeout 15s, got %v", cfg.Server.WriteTimeout)
				}
				if cfg.Server.IdleTimeout != 120*time.Second {
					t.Errorf("expected idle timeout 120s, got %v", cfg.Server.IdleTimeout)
				}
				if cfg.Logging.Level != "info" {
					t.Errorf("expected log level info, got %s", cfg.Logging.Level)
				}
				if cfg.Logging.Format != "json" {
					t.Errorf("expected log format json, got %s", cfg.Logging.Format)
				}
			},
		},
		{
			name: "valid config from file",
			configFile: `server:
  port: ":8081"
  readtimeout: "30s"
  writetimeout: "30s"
  idletimeout: "120s"
files:
  productsfile: "./data/products.json"
  couponsdir: "./data/coupons"
logging:
  level: "info"
  format: "json"`,
			envVars: map[string]string{
				"PRODUCTS_FILE": "./data/products.json", // Required field
				"COUPONS_DIR":   "./data/coupons",       // Required field
			},
			wantErr: false,
			validateCfg: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != ":8081" {
					t.Errorf("expected port :8081, got %s", cfg.Server.Port)
				}
				if cfg.Server.ReadTimeout != 30*time.Second {
					t.Errorf("expected read timeout 30s, got %v", cfg.Server.ReadTimeout)
				}
				if cfg.Server.WriteTimeout != 30*time.Second {
					t.Errorf("expected write timeout 30s, got %v", cfg.Server.WriteTimeout)
				}
				if cfg.Server.IdleTimeout != 120*time.Second {
					t.Errorf("expected idle timeout 120s, got %v", cfg.Server.IdleTimeout)
				}
				if cfg.Files.ProductsFile != "./data/products.json" {
					t.Errorf("expected products file ./data/products.json, got %s", cfg.Files.ProductsFile)
				}
				if cfg.Files.CouponsDir != "./data/coupons" {
					t.Errorf("expected coupons dir ./data/coupons, got %s", cfg.Files.CouponsDir)
				}
				if cfg.Logging.Level != "info" {
					t.Errorf("expected log level info, got %s", cfg.Logging.Level)
				}
				if cfg.Logging.Format != "json" {
					t.Errorf("expected log format json, got %s", cfg.Logging.Format)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				// Cleanup environment after test
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			// Create test config file if provided
			if tt.configFile != "" {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte(tt.configFile), 0644); err != nil {
					t.Fatal(err)
				}
				// Add the temp dir to viper's config path
				os.Setenv("CONFIG_PATH", tmpDir)
				defer os.Unsetenv("CONFIG_PATH")
			}

			// Test configuration loading
			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && tt.validateCfg != nil {
				tt.validateCfg(t, cfg)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	// Setup minimum required environment variables
	os.Setenv("PRODUCTS_FILE", "./testdata/products.json")
	os.Setenv("COUPONS_DIR", "./testdata/coupons")
	defer func() {
		os.Unsetenv("PRODUCTS_FILE")
		os.Unsetenv("COUPONS_DIR")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed with minimum config: %v", err)
	}

	// Test default values
	if cfg.Server.Port != ":8080" {
		t.Errorf("expected default port :8080, got %s", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("expected default read timeout 15s, got %v", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 15*time.Second {
		t.Errorf("expected default write timeout 15s, got %v", cfg.Server.WriteTimeout)
	}
	if cfg.Server.IdleTimeout != 60*time.Second {
		t.Errorf("expected default idle timeout 60s, got %v", cfg.Server.IdleTimeout)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("expected default log format json, got %s", cfg.Logging.Format)
	}
}

func TestGetServerTimeouts(t *testing.T) {
	cfg := &Config{
		Server: Server{
			ReadTimeout:  20 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	read, write, idle := cfg.GetServerTimeouts()
	if read != 20*time.Second {
		t.Errorf("expected read timeout 20s, got %v", read)
	}
	if write != 30*time.Second {
		t.Errorf("expected write timeout 30s, got %v", write)
	}
	if idle != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", idle)
	}
}
