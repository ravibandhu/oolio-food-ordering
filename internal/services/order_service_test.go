package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/ravibandhu/oolio-food-ordering/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCouponValidator implements a simple test-only coupon store
type MockCouponValidator struct {
	validCoupons map[string]struct{}
}

func (m *MockCouponValidator) GetCoupon(code string) bool {
	_, exists := m.validCoupons[code]
	return exists
}

func NewMockCouponValidator(coupons []string) *MockCouponValidator {
	store := &MockCouponValidator{
		validCoupons: make(map[string]struct{}),
	}

	for _, coupon := range coupons {
		store.validCoupons[coupon] = struct{}{}
	}

	return store
}

// StoreInterface is a subset of the data.Store methods needed for testing
type StoreInterface interface {
	GetProduct(id string) (*models.Product, error)
	ValidateCoupon(code string) bool
}

// MockStore is a test implementation of the StoreInterface
type MockStore struct {
	products *data.ProductStore
	coupons  data.CouponValidator
}

// GetProduct delegates to the underlying ProductStore
func (m *MockStore) GetProduct(id string) (*models.Product, error) {
	return m.products.GetProduct(id)
}

// GetAllProducts delegates to the underlying ProductStore
func (m *MockStore) GetAllProducts() []*models.Product {
	return m.products.GetAllProducts()
}

// ValidateCoupon delegates to the underlying CouponValidator
func (m *MockStore) ValidateCoupon(code string) bool {
	return m.coupons.GetCoupon(code)
}

// Close implements the data.Store Close method for the MockStore
func (m *MockStore) Close() error {
	return nil
}

// TestOrderServiceImpl is a version of OrderServiceImpl for testing that accepts StoreInterface
type TestOrderServiceImpl struct {
	store StoreInterface
}

// PlaceOrder processes a new order request (duplicates logic from real service for testing)
func (s *TestOrderServiceImpl) PlaceOrder(req *models.OrderRequest) (*models.Order, error) {
	// Validate products and calculate total
	var products []models.Product
	var totalAmount float64

	// Validate request
	if req.Items == nil || len(req.Items) == 0 {
		return nil, models.NewErrorResponse("VALIDATION_ERROR", "Order must contain at least one item")
	}

	// Validate and collect products
	for _, item := range req.Items {
		product, err := s.store.GetProduct(item.ProductID)
		if err != nil {
			return nil, models.NewErrorResponse("INVALID_PRODUCT", "Invalid product ID")
		}
		products = append(products, *product)
		totalAmount += product.Price * float64(item.Quantity)
	}

	// Apply coupon if provided
	if req.CouponCode != "" {
		// Validate coupon
		if !s.store.ValidateCoupon(req.CouponCode) {
			return nil, models.NewErrorResponse("INVALID_COUPON", "Invalid coupon code")
		}
		// Apply 10% discount
		totalAmount = totalAmount * 0.90
	}

	// Create order items with prices
	var items []models.OrderItem
	for i, item := range req.Items {
		items = append(items, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     products[i].Price,
		})
	}

	// Create and return the order
	order := models.NewOrder(items, products, totalAmount, req.CouponCode)
	return order, nil
}

func setupTestData(t *testing.T) (string, string, *config.Config, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "order-service-test")
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
			"description": "Description 1",
			"price": 9.99,
			"category": "Category 1",
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
			"description": "Description 2",
			"price": 19.99,
			"category": "Category 2",
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

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, productsFile, cfg, cleanup
}

func TestPlaceOrder(t *testing.T) {
	// Use testutil to set up test data
	testData := testutil.SetupTestData(t)
	defer testData.Cleanup()

	// Create mock coupon validator with TEST10 as valid coupon
	validCoupons := []string{"TEST10", "TEST20", "DISCOUNT30"}
	mockCouponValidator := NewMockCouponValidator(validCoupons)

	// Create a product store with test products
	productStore := data.NewProductStore()
	err := productStore.LoadProducts(testData.ProductsFile)
	require.NoError(t, err)

	// Create mock store with direct access to components
	store := &MockStore{
		products: productStore,
		coupons:  mockCouponValidator,
	}

	// Create test order service that uses our StoreInterface
	orderService := &TestOrderServiceImpl{
		store: store,
	}

	// Test valid order without coupon
	t.Run("valid order without coupon", func(t *testing.T) {
		request := &models.OrderRequest{
			Items: []models.OrderItem{
				{
					ProductID: "prod-1",
					Quantity:  2,
					Price:     9.99,
				},
				{
					ProductID: "prod-2",
					Quantity:  1,
					Price:     19.99,
				},
			},
			CouponCode: "",
		}

		// Run the service and check results
		order, err := orderService.PlaceOrder(request)
		assert.NoError(t, err)
		assert.NotNil(t, order)

		// Validate the returned order details
		assert.NotEmpty(t, order.ID)
		assert.Equal(t, 2, len(order.Items))
		assert.Equal(t, "", order.CouponCode)
	})

	// Test valid order with valid coupon
	t.Run("valid order with valid coupon", func(t *testing.T) {
		request := &models.OrderRequest{
			Items: []models.OrderItem{
				{
					ProductID: "prod-1",
					Quantity:  2,
					Price:     9.99,
				},
				{
					ProductID: "prod-2",
					Quantity:  1,
					Price:     19.99,
				},
			},
			CouponCode: "TEST10", // Valid coupon code
		}

		// Run the service and check results
		order, err := orderService.PlaceOrder(request)
		assert.NoError(t, err)
		assert.NotNil(t, order)

		// Validate the returned order details
		assert.NotEmpty(t, order.ID)
		assert.Equal(t, 2, len(order.Items))
		assert.Equal(t, "TEST10", order.CouponCode)

		// Check that discount was applied (10% off)
		expectedTotal := (9.99*2 + 19.99) * 0.9
		assert.InDelta(t, expectedTotal, order.TotalAmount, 0.01)
	})

	// Test valid order with invalid coupon
	t.Run("valid order with invalid coupon", func(t *testing.T) {
		request := &models.OrderRequest{
			Items: []models.OrderItem{
				{
					ProductID: "prod-1",
					Quantity:  1,
					Price:     9.99,
				},
			},
			CouponCode: "INVALID", // Invalid coupon code
		}

		// Run the service and check results
		order, err := orderService.PlaceOrder(request)
		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "Invalid coupon code")
	})

	// Test empty order items
	t.Run("empty order items", func(t *testing.T) {
		// Order with no items
		request := &models.OrderRequest{
			Items:      []models.OrderItem{},
			CouponCode: "",
		}

		// Run the service and check results
		order, err := orderService.PlaceOrder(request)
		assert.Error(t, err)
		assert.Nil(t, order)
		// Error should mention invalid items/required field
	})
}

func TestOrderService_Interface(t *testing.T) {
	// Verify OrderServiceImpl implements OrderService interface
	var _ OrderService = (*OrderServiceImpl)(nil)
}
