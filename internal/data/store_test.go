package data

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/stretchr/testify/assert"
)

func setupTestData(t *testing.T) (string, string, *config.Config, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "store-test")
	assert.NoError(t, err)

	// Create coupons directory
	couponsDir := filepath.Join(tempDir, "coupons")
	err = os.MkdirAll(couponsDir, 0755)
	assert.NoError(t, err)

	// Create test coupon file
	couponFile := filepath.Join(couponsDir, "test_coupons.txt")
	err = os.WriteFile(couponFile, []byte("TEST10\nTEST20\n"), 0644)
	assert.NoError(t, err)

	// Create products file
	productsFile := filepath.Join(tempDir, "products.json")
	err = os.WriteFile(productsFile, []byte(`[
		{
			"id": "prod-1",
			"name": "Test Product 1",
			"description": "Description 1",
			"price": 9.99,
			"category": "Category 1",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	]`), 0644)
	assert.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Files: config.FilesConfig{
			ProductsFile: productsFile,
			CouponsDir:   couponsDir,
		},
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
		// Reset singleton state
		instance = nil
		once = sync.Once{}
		loadErr = nil
		loadDir = ""
		loaded = false
	}

	return tempDir, productsFile, cfg, cleanup
}

func TestNewStore(t *testing.T) {
	_, productsFile, validCfg, cleanup := setupTestData(t)
	defer cleanup()

	// Create a non-existent directory path for invalid tests
	invalidDir := filepath.Join(os.TempDir(), "nonexistent-"+time.Now().Format("20060102150405"))

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
		},
		{
			name: "invalid products file",
			cfg: &config.Config{
				Files: config.FilesConfig{
					ProductsFile: filepath.Join(invalidDir, "nonexistent.json"),
					CouponsDir:   validCfg.Files.CouponsDir,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid coupons directory",
			cfg: &config.Config{
				Files: config.FilesConfig{
					ProductsFile: productsFile,
					CouponsDir:   invalidDir,
				},
			},
			wantErr: true,
		},
		{
			name:    "valid config",
			cfg:     validCfg,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, store)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, store)
			}
		})
	}
}

func TestStore_GetProduct(t *testing.T) {
	_, _, cfg, cleanup := setupTestData(t)
	defer cleanup()

	store, err := NewStore(cfg)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		want    string
		wantErr bool
	}{
		{
			name:    "existing product",
			id:      "prod-1",
			want:    "Test Product 1",
			wantErr: false,
		},
		{
			name:    "non-existent product",
			id:      "prod-999",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := store.GetProduct(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.want, product.Name)
			}
		})
	}
}

func TestStore_GetAllProducts(t *testing.T) {
	_, _, cfg, cleanup := setupTestData(t)
	defer cleanup()

	store, err := NewStore(cfg)
	assert.NoError(t, err)

	products := store.GetAllProducts()
	assert.NotNil(t, products)
	assert.Len(t, products, 1)
	assert.Equal(t, "Test Product 1", products[0].Name)
}

func TestStore_ValidateCoupon(t *testing.T) {
	_, _, cfg, cleanup := setupTestData(t)
	defer cleanup()

	store, err := NewStore(cfg)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		code    string
		isValid bool
	}{
		{
			name:    "valid coupon",
			code:    "TEST10",
			isValid: true,
		},
		{
			name:    "invalid coupon",
			code:    "INVALID",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := store.ValidateCoupon(tt.code)
			assert.Equal(t, tt.isValid, valid)
		})
	}
}
