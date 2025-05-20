package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of OrderService
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) PlaceOrder(req *models.OrderRequest) (*models.Order, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func TestPlaceOrder(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockOrderService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "valid order without coupon",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  2,
						Price:     9.99,
					},
				},
			},
			setupMock: func(m *MockOrderService) {
				m.On("PlaceOrder", mock.AnythingOfType("*models.OrderRequest")).Return(&models.Order{
					ID: "order-1",
					Items: []models.OrderItem{
						{
							ProductID: "prod-1",
							Quantity:  2,
							Price:     9.99,
						},
					},
					Products: []models.Product{
						{
							ID:          "prod-1",
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
						},
					},
					TotalAmount: 19.98,
					CreatedAt:   time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "valid order with coupon",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  2,
						Price:     9.99,
					},
				},
				CouponCode: "TEST10",
			},
			setupMock: func(m *MockOrderService) {
				m.On("PlaceOrder", mock.AnythingOfType("*models.OrderRequest")).Return(&models.Order{
					ID: "order-1",
					Items: []models.OrderItem{
						{
							ProductID: "prod-1",
							Quantity:  2,
							Price:     9.99,
						},
					},
					Products: []models.Product{
						{
							ID:          "prod-1",
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
						},
					},
					TotalAmount: 17.98, // 10% discount applied
					CouponCode:  "TEST10",
					CreatedAt:   time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			setupMock:      func(m *MockOrderService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{ // Use map instead of struct for flexible comparison
				"code":    "INVALID_REQUEST",
				"message": "Failed to parse request body",
				"details": map[string]interface{}{
					"error": "json: cannot unmarshal string into Go value of type models.OrderRequest",
				},
			},
		},
		{
			name: "validation error",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  0, // Invalid quantity
					},
				},
			},
			setupMock:      func(m *MockOrderService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error",
			requestBody: models.OrderRequest{
				Items: []models.OrderItem{
					{
						ProductID: "prod-1",
						Quantity:  2,
						Price:     9.99, // Add price to pass validation
					},
				},
			},
			setupMock: func(m *MockOrderService) {
				m.On("PlaceOrder", mock.AnythingOfType("*models.OrderRequest")).Return(nil,
					models.NewErrorResponse("PRODUCT_NOT_FOUND", "Product not found").
						AddDetail("productId", "prod-1"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: map[string]interface{}{ // Use map instead of struct for flexible comparison
				"code":    "PRODUCT_NOT_FOUND",
				"message": "Product not found",
				"details": map[string]interface{}{
					"productId": "prod-1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockOrderService)
			tt.setupMock(mockService)

			// Create handler
			handler := NewOrderHandler(mockService)

			// Create request
			var body bytes.Buffer
			if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
				t.Fatal(err)
			}

			// Create test request and response recorder
			req := httptest.NewRequest(http.MethodPost, "/order", &body)
			rec := httptest.NewRecorder()

			// Call handler
			handler.PlaceOrder(rec, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// If expected body is specified, check it
			if tt.expectedBody != nil {
				var got interface{}
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.expectedBody, got)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
