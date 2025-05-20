package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestData holds test data and cleanup function
type TestData struct {
	TempDir      string
	ProductsFile string
	CouponsDir   string
	Config       *config.Config
	Cleanup      func()
}

// SetupTestData creates a temporary directory with test data
func SetupTestData(t *testing.T) *TestData {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "test-data")
	assert.NoError(t, err)

	// Create coupons directory
	couponsDir := filepath.Join(tempDir, "coupons")
	err = os.MkdirAll(couponsDir, 0755)
	assert.NoError(t, err)

	// Create test coupon files
	couponFiles := []string{"coupons1.txt", "coupons2.txt", "coupons3.txt"}
	for _, file := range couponFiles {
		couponFile := filepath.Join(couponsDir, file)
		err = os.WriteFile(couponFile, []byte("TEST10\nTEST20\n"), 0644)
		assert.NoError(t, err)
	}

	// Create products file
	productsFile := filepath.Join(tempDir, "products.json")
	err = os.WriteFile(productsFile, []byte(`[
		{
			"id": "prod-1",
			"name": "Test Product 1",
			"description": "Test Description 1",
			"price": 9.99,
			"category": "Test Category",
			"image": {
				"thumbnail": "https://example.com/images/prod-1-thumb.jpg",
				"mobile": "https://example.com/images/prod-1-mobile.jpg",
				"tablet": "https://example.com/images/prod-1-tablet.jpg",
				"desktop": "https://example.com/images/prod-1-desktop.jpg"
			},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		},
		{
			"id": "prod-2",
			"name": "Test Product 2",
			"description": "Test Description 2",
			"price": 19.99,
			"category": "Test Category",
			"image": {
				"thumbnail": "https://example.com/images/prod-2-thumb.jpg",
				"mobile": "https://example.com/images/prod-2-mobile.jpg",
				"tablet": "https://example.com/images/prod-2-tablet.jpg",
				"desktop": "https://example.com/images/prod-2-desktop.jpg"
			},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	]`), 0644)
	assert.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Server: config.Server{
			Port:         ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Files: config.Files{
			ProductsFile: productsFile,
			CouponsDir:   couponsDir,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}

	// Return test data with cleanup function
	return &TestData{
		TempDir:      tempDir,
		ProductsFile: productsFile,
		CouponsDir:   couponsDir,
		Config:       cfg,
		Cleanup: func() {
			os.RemoveAll(tempDir)
		},
	}
}

// GetTestProduct returns a test product with valid data
func GetTestProduct() *models.Product {
	return &models.Product{
		ID:          "test-prod-1",
		Name:        "Test Product",
		Price:       9.99,
		Category:    "Test Category",
		Image: &models.ProductImage{
			Thumbnail: "https://example.com/images/test-thumb.jpg",
			Mobile:    "https://example.com/images/test-mobile.jpg",
			Tablet:    "https://example.com/images/test-tablet.jpg",
			Desktop:   "https://example.com/images/test-desktop.jpg",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTestOrder returns a test order with valid data
func GetTestOrder() *models.Order {
	items := []models.OrderItem{
		{
			ProductID: "test-prod-1",
			Quantity:  2,
			Price:     9.99,
		},
	}
	products := []models.Product{*GetTestProduct()}
	totalAmount := 19.98
	couponCode := "TEST10"

	return models.NewOrder(items, products, totalAmount, couponCode)
}

// GetTestCoupon returns a test coupon with valid data
func GetTestCoupon() *models.Coupon {
	return &models.Coupon{
		Code:            "TEST10",
		DiscountPercent: 10.0,
		MinOrderAmount:  20.0,
		ExpiryDate:      time.Now().Add(24 * time.Hour),
		MaxUsagePerUser: 1,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
