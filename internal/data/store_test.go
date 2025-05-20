package data

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/ravibandhu/oolio-food-ordering/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCouponStore implements a simple test-only coupon store to avoid singleton issues
type MockCouponStore struct {
	validCoupons map[string]struct{}
}

func (m *MockCouponStore) GetCoupon(code string) bool {
	_, exists := m.validCoupons[code]
	return exists
}

func NewMockCouponStore(coupons []string) *MockCouponStore {
	store := &MockCouponStore{
		validCoupons: make(map[string]struct{}),
	}

	for _, coupon := range coupons {
		store.validCoupons[coupon] = struct{}{}
	}

	return store
}

// createTestStore creates a store with test data directly injected without file loading
func createTestStore(t *testing.T, ctx context.Context) *Store {
	// Initialize store components
	productStore := setupProductStore()

	// Create a simple mock coupon store with predefined valid coupons
	validCoupons := []string{"TEST10", "TEST20", "TEST30"}
	mockCouponStore := NewMockCouponStore(validCoupons)

	// Create test config
	cfg := &config.Config{
		Server: config.Server{
			Port: ":8080",
		},
		Files: config.Files{
			ProductsFile: "test_products.json",
			CouponsDir:   "test_coupons",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create a child context with cancellation
	storeCtx, cancel := context.WithCancel(ctx)

	// Create the store with our prepared components
	store := &Store{
		products: productStore,
		coupons:  mockCouponStore,
		config:   cfg,
		ctx:      storeCtx,
		cancel:   cancel,
	}

	return store
}

func TestNewStore_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	// Verify store was created correctly
	assert.NotNil(t, store)
	assert.NotNil(t, store.products)
	assert.NotNil(t, store.coupons)
	assert.NotNil(t, store.config)
	assert.NotNil(t, store.ctx)
	assert.NotNil(t, store.cancel)
}

func TestGetProduct_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	tests := []struct {
		name      string
		productID string
		wantErr   bool
	}{
		{
			name:      "existing product",
			productID: "prod-1",
			wantErr:   false,
		},
		{
			name:      "non-existent product",
			productID: "prod-999",
			wantErr:   true,
		},
		{
			name:      "empty product ID",
			productID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := store.GetProduct(tt.productID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.productID, product.ID)
			}
		})
	}
}

func TestValidateCoupon_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	tests := []struct {
		name       string
		couponCode string
		want       bool
	}{
		{
			name:       "valid coupon",
			couponCode: "TEST10",
			want:       true,
		},
		{
			name:       "valid coupon 2",
			couponCode: "TEST20",
			want:       true,
		},
		{
			name:       "valid coupon 3",
			couponCode: "TEST30",
			want:       true,
		},
		{
			name:       "invalid coupon",
			couponCode: "INVALID",
			want:       false,
		},
		{
			name:       "empty coupon",
			couponCode: "",
			want:       false,
		},
		{
			name:       "case sensitive check",
			couponCode: "test10",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.ValidateCoupon(tt.couponCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClose_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	// Test closing the store
	err := store.Close()
	assert.NoError(t, err)

	// Test that operations fail after closing
	_, err = store.GetProduct("prod-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestGetAllProducts_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	products := store.GetAllProducts()
	assert.Equal(t, 2, len(products))

	// Verify products in the result
	foundProd1 := false
	foundProd2 := false

	for _, p := range products {
		switch p.ID {
		case "prod-1":
			foundProd1 = true
			assert.Equal(t, "Test Product 1", p.Name)
		case "prod-2":
			foundProd2 = true
			assert.Equal(t, "Test Product 2", p.Name)
		}
	}

	assert.True(t, foundProd1, "Product 1 should be in the results")
	assert.True(t, foundProd2, "Product 2 should be in the results")
}

func TestStore_ConcurrentAccess_WithMockData(t *testing.T) {
	ctx := context.Background()
	store := createTestStore(t, ctx)

	// Test concurrent product access
	t.Run("concurrent product access", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				product, err := store.GetProduct("prod-1")
				assert.NoError(t, err)
				assert.NotNil(t, product)
				done <- true
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	// Test concurrent coupon validation
	t.Run("concurrent coupon validation", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				valid := store.ValidateCoupon("TEST10")
				assert.True(t, valid)
				done <- true
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

func TestNewStore(t *testing.T) {
	// Reset the singleton for this test
	resetForTest()

	testData := testutil.SetupTestData(t)
	defer testData.Cleanup()

	// Use our coupon directory with consistent test data
	couponDir := setupCouponTestData(t)

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			cfg: &config.Config{
				Server: testData.Config.Server,
				Files: config.Files{
					ProductsFile: testData.ProductsFile,
					CouponsDir:   couponDir,
				},
				Logging: testData.Config.Logging,
			},
			wantErr: false,
		},
		{
			name: "invalid products file",
			cfg: &config.Config{
				Server: testData.Config.Server,
				Files: config.Files{
					ProductsFile: "nonexistent.json",
					CouponsDir:   couponDir,
				},
				Logging: testData.Config.Logging,
			},
			wantErr: true,
		},
		{
			name: "invalid coupons directory",
			cfg: &config.Config{
				Server: testData.Config.Server,
				Files: config.Files{
					ProductsFile: testData.ProductsFile,
					CouponsDir:   "nonexistent",
				},
				Logging: testData.Config.Logging,
			},
			wantErr: true,
		},
		{
			name: "invalid products file format",
			cfg: func() *config.Config {
				// Create invalid JSON file
				invalidFile := filepath.Join(testData.TempDir, "invalid.json")
				err := os.WriteFile(invalidFile, []byte(`{invalid json`), 0644)
				require.NoError(t, err)

				return &config.Config{
					Server: testData.Config.Server,
					Files: config.Files{
						ProductsFile: invalidFile,
						CouponsDir:   couponDir,
					},
					Logging: testData.Config.Logging,
				}
			}(),
			wantErr: true,
		},
		{
			name: "empty products file",
			cfg: func() *config.Config {
				// Create truly empty file (not a valid JSON empty array)
				emptyFile := filepath.Join(testData.TempDir, "empty.json")
				err := os.WriteFile(emptyFile, []byte(``), 0644)
				require.NoError(t, err)

				return &config.Config{
					Server: testData.Config.Server,
					Files: config.Files{
						ProductsFile: emptyFile,
						CouponsDir:   couponDir,
					},
					Logging: testData.Config.Logging,
				}
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton for each test case
			resetForTest()

			ctx := context.Background()
			store, err := NewStore(ctx, tt.cfg)
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
	// Reset the singleton for this test
	resetForTest()

	testData := testutil.SetupTestData(t)
	defer testData.Cleanup()

	// Use our coupon directory with consistent test data
	couponDir := setupCouponTestData(t)

	ctx := context.Background()
	cfg := &config.Config{
		Server: testData.Config.Server,
		Files: config.Files{
			ProductsFile: testData.ProductsFile,
			CouponsDir:   couponDir,
		},
		Logging: testData.Config.Logging,
	}

	store, err := NewStore(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, store)

	tests := []struct {
		name      string
		productID string
		want      *models.Product
		wantErr   bool
	}{
		{
			name:      "existing product",
			productID: "prod-1",
			want: &models.Product{
				ID:          "prod-1",
				Name:    "Test Product 1",
				Price:   9.99,
				Category: "Test Category",
				Image: &models.ProductImage{
					Thumbnail: "https://example.com/images/prod-1-thumb.jpg",
					Mobile:    "https://example.com/images/prod-1-mobile.jpg",
					Tablet:    "https://example.com/images/prod-1-tablet.jpg",
					Desktop:   "https://example.com/images/prod-1-desktop.jpg",
				},
			},
			wantErr: false,
		},
		{
			name:      "non-existent product",
			productID: "prod-999",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "empty product ID",
			productID: "",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := store.GetProduct(tt.productID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.want.ID, product.ID)
				assert.Equal(t, tt.want.Name, product.Name)
				assert.Equal(t, tt.want.Price, product.Price)
				assert.Equal(t, tt.want.Category, product.Category)
				assert.Equal(t, tt.want.Image.Thumbnail, product.Image.Thumbnail)
				assert.Equal(t, tt.want.Image.Mobile, product.Image.Mobile)
				assert.Equal(t, tt.want.Image.Tablet, product.Image.Tablet)
				assert.Equal(t, tt.want.Image.Desktop, product.Image.Desktop)
				assert.False(t, product.CreatedAt.IsZero())
				assert.False(t, product.UpdatedAt.IsZero())
			}
		})
	}
}

func TestStore_ValidateCoupon(t *testing.T) {
	// Create a store with mock coupon data
	ctx := context.Background()

	// Create test config
	cfg := &config.Config{
		Server: config.Server{
			Port: ":8080",
		},
		Files: config.Files{
			ProductsFile: "test_products.json",
			CouponsDir:   "test_coupons",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create mock coupon store with valid coupons
	validCoupons := []string{"TEST10", "TEST20", "TEST30"}
	mockCouponStore := NewMockCouponStore(validCoupons)

	// Create a product store
	productStore := setupProductStore()

	// Create a child context with cancellation
	storeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create the store with our prepared components
	store := &Store{
		products: productStore,
		coupons:  mockCouponStore,
		config:   cfg,
		ctx:      storeCtx,
		cancel:   cancel,
	}

	tests := []struct {
		name       string
		couponCode string
		want       bool
	}{
		{
			name:       "valid coupon",
			couponCode: "TEST10",
			want:       true,
		},
		{
			name:       "invalid coupon",
			couponCode: "INVALID",
			want:       false,
		},
		{
			name:       "empty coupon",
			couponCode: "",
			want:       false,
		},
		{
			name:       "case sensitive check",
			couponCode: "test10",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.ValidateCoupon(tt.couponCode)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStore_Close(t *testing.T) {
	// Reset the singleton for this test
	resetForTest()

	testData := testutil.SetupTestData(t)
	defer testData.Cleanup()

	// Use our coupon directory with consistent test data
	couponDir := setupCouponTestData(t)

	ctx := context.Background()
	cfg := &config.Config{
		Server: testData.Config.Server,
		Files: config.Files{
			ProductsFile: testData.ProductsFile,
			CouponsDir:   couponDir,
		},
		Logging: testData.Config.Logging,
	}

	store, err := NewStore(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Test closing the store
	err = store.Close()
	assert.NoError(t, err)

	// Test that operations fail after closing
	_, err = store.GetProduct("prod-1")
	assert.Error(t, err)
	assert.False(t, store.ValidateCoupon("TEST10"))
}

func TestStore_ConcurrentAccess(t *testing.T) {
	// Create a store with mock coupon data
	ctx := context.Background()

	// Create test config
	cfg := &config.Config{
		Server: config.Server{
			Port: ":8080",
		},
		Files: config.Files{
			ProductsFile: "test_products.json",
			CouponsDir:   "test_coupons",
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create mock coupon store with valid coupons
	validCoupons := []string{"TEST10", "TEST20", "TEST30"}
	mockCouponStore := NewMockCouponStore(validCoupons)

	// Create a product store
	productStore := setupProductStore()

	// Create a child context with cancellation
	storeCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create the store with our prepared components
	store := &Store{
		products: productStore,
		coupons:  mockCouponStore,
		config:   cfg,
		ctx:      storeCtx,
		cancel:   cancel,
	}

	// Test concurrent product access
	t.Run("concurrent product access", func(t *testing.T) {
		const numGoroutines = 10 // Reduced from 100 for faster tests
		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				product, err := store.GetProduct("prod-1")
				assert.NoError(t, err)
				assert.NotNil(t, product)
				done <- true
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	// Test concurrent coupon validation
	t.Run("concurrent coupon validation", func(t *testing.T) {
		const numGoroutines = 10 // Reduced from 100 for faster tests
		done := make(chan bool)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				valid := store.ValidateCoupon("TEST10")
				assert.True(t, valid)
				done <- true
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}
