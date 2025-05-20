package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupTestData(t *testing.T) (string, string, *config.Config, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "products-test")
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
				"thumbnail": "https://example.com/images/test1-thumb.jpg",
				"mobile": "https://example.com/images/test1-mobile.jpg",
				"tablet": "https://example.com/images/test1-tablet.jpg",
				"desktop": "https://example.com/images/test1-desktop.jpg"
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
				"thumbnail": "https://example.com/images/test2-thumb.jpg",
				"mobile": "https://example.com/images/test2-mobile.jpg",
				"tablet": "https://example.com/images/test2-tablet.jpg",
				"desktop": "https://example.com/images/test2-desktop.jpg"
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

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return productsFile, couponsDir, cfg, cleanup
}

func TestListProducts(t *testing.T) {
	// Setup test data
	_, _, cfg, cleanup := setupTestData(t)
	defer cleanup()

	// Create store
	ctx := context.Background()
	store, err := data.NewStore(ctx, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Create handler
	handler := NewProductHandler(store)

	// Create test request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	// Call handler
	handler.ListProducts(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check response body
	var got []models.Product
	err = json.NewDecoder(rec.Body).Decode(&got)
	assert.NoError(t, err)
	assert.Len(t, got, 2)

	// Create a map of products by ID to verify all expected products are present
	// without assuming a specific order
	productMap := make(map[string]models.Product)
	for _, p := range got {
		productMap[p.ID] = p
	}

	// Verify both expected products exist
	assert.Contains(t, productMap, "prod-1", "Product prod-1 should be present in the response")
	assert.Contains(t, productMap, "prod-2", "Product prod-2 should be present in the response")
}

func TestGetProduct(t *testing.T) {
	// Setup test data
	_, _, cfg, cleanup := setupTestData(t)
	defer cleanup()

	// Create store
	ctx := context.Background()
	store, err := data.NewStore(ctx, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Create handler
	handler := NewProductHandler(store)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid product",
			productID:      "prod-1",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got models.Product
				err := json.NewDecoder(rec.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, "prod-1", got.ID)
				assert.Equal(t, "Test Product 1", got.Name)
				assert.Equal(t, 9.99, got.Price)
			},
		},
		{
			name:           "product not found",
			productID:      "invalid-id",
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got models.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&got)
				assert.NoError(t, err)
				assert.Equal(t, "NOT_FOUND", got.Code)
				assert.Equal(t, "Product not found", got.Message)
				assert.Equal(t, "invalid-id", got.Details["productId"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request and response recorder
			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productID, nil)
			rec := httptest.NewRecorder()

			// Call handler
			handler.GetProduct(rec, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check response
			tt.checkResponse(t, rec)
		})
	}
}
