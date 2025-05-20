package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct(t *testing.T) {
	id := "prod-1"
	name := "Burger"
	price := 9.99
	category := "Main Course"
	image := &ProductImage{
		Thumbnail: "https://example.com/images/burger-thumb.jpg",
		Mobile:    "https://example.com/images/burger-mobile.jpg",
		Tablet:    "https://example.com/images/burger-tablet.jpg",
		Desktop:   "https://example.com/images/burger-desktop.jpg",
	}

	p := NewProduct(id, name, price, category, image)

	assert.Equal(t, id, p.ID)
	assert.Equal(t, name, p.Name)
	assert.Equal(t, price, p.Price)
	assert.Equal(t, category, p.Category)
	require.NotNil(t, p.Image)
	assert.Equal(t, image.Thumbnail, p.Image.Thumbnail)
	assert.Equal(t, image.Mobile, p.Image.Mobile)
	assert.Equal(t, image.Tablet, p.Image.Tablet)
	assert.Equal(t, image.Desktop, p.Image.Desktop)
	assert.False(t, p.CreatedAt.IsZero())
	assert.False(t, p.UpdatedAt.IsZero())

	// Test validation
	err := Validate(p)
	assert.NoError(t, err)
}

func TestNewOrder(t *testing.T) {
	items := []OrderItem{
		{
			ProductID: "prod-1",
			Quantity:  2,
			Price:     9.99,
		},
	}
	products := []Product{
		{
			ID:          "prod-1",
			Name:        "Test Product",
			Price:       9.99,
			Category:    "Test Category",
			Image: &ProductImage{
				Thumbnail: "https://example.com/images/test-thumb.jpg",
				Mobile:    "https://example.com/images/test-mobile.jpg",
				Tablet:    "https://example.com/images/test-tablet.jpg",
				Desktop:   "https://example.com/images/test-desktop.jpg",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	totalAmount := 19.98
	couponCode := "TEST10"

	order := NewOrder(items, products, totalAmount, couponCode)

	assert.NotEmpty(t, order.ID)
	assert.Equal(t, len(items), len(order.Items))
	assert.Equal(t, len(products), len(order.Products))
	assert.Equal(t, totalAmount, order.TotalAmount)
	assert.Equal(t, couponCode, order.CouponCode)
	assert.False(t, order.CreatedAt.IsZero())

	// Test validation
	err := Validate(order)
	assert.NoError(t, err)
}

func TestNewCoupon(t *testing.T) {
	code := "SAVE10"
	discount := 10.0
	minAmount := 20.0
	expiry := time.Now().Add(24 * time.Hour)
	maxUsage := 1

	c := NewCoupon(code, discount, minAmount, expiry, maxUsage)

	assert.Equal(t, code, c.Code)
	assert.Equal(t, discount, c.DiscountPercent)
	assert.Equal(t, minAmount, c.MinOrderAmount)
	assert.True(t, expiry.Equal(c.ExpiryDate))
	assert.Equal(t, maxUsage, c.MaxUsagePerUser)
	assert.True(t, c.IsActive)
	assert.False(t, c.CreatedAt.IsZero())
	assert.False(t, c.UpdatedAt.IsZero())

	// Test validation
	err := Validate(c)
	assert.NoError(t, err)
}

func TestNewErrorResponse(t *testing.T) {
	code := "INVALID_INPUT"
	message := "Invalid input provided"
	details := map[string]string{
		"price":    "must be greater than 0",
		"quantity": "must be at least 1",
	}

	// Test creating error response with single detail
	e1 := NewErrorResponse(code, message).AddDetail("price", "must be greater than 0")

	assert.Equal(t, code, e1.Code)
	assert.Equal(t, message, e1.Message)
	require.NotNil(t, e1.Details)
	assert.Equal(t, "must be greater than 0", e1.Details["price"])

	// Test creating error response with multiple details
	e2 := NewErrorResponse(code, message).AddDetails(details)

	assert.Equal(t, 2, len(e2.Details))
	assert.Equal(t, details["price"], e2.Details["price"])
	assert.Equal(t, details["quantity"], e2.Details["quantity"])

	// Test method chaining
	e3 := NewErrorResponse(code, message).
		AddDetail("price", "must be greater than 0").
		AddDetail("quantity", "must be at least 1")

	assert.Equal(t, 2, len(e3.Details))

	// Test validation
	err := Validate(e1)
	assert.NoError(t, err)
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		// Product validation tests
		{
			name: "valid product",
			input: &Product{
				ID:          "prod-1",
				Name:        "Test Product",
				Price:       9.99,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "https://example.com/images/test-thumb.jpg",
					Mobile:    "https://example.com/images/test-mobile.jpg",
					Tablet:    "https://example.com/images/test-tablet.jpg",
					Desktop:   "https://example.com/images/test-desktop.jpg",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid product - missing ID",
			input: &Product{
				Name:        "Test Product",
				Price:       9.99,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "https://example.com/images/test-thumb.jpg",
					Mobile:    "https://example.com/images/test-mobile.jpg",
					Tablet:    "https://example.com/images/test-tablet.jpg",
					Desktop:   "https://example.com/images/test-desktop.jpg",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid product - missing name",
			input: &Product{
				ID:          "prod-1",
				Price:       9.99,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "https://example.com/images/test-thumb.jpg",
					Mobile:    "https://example.com/images/test-mobile.jpg",
					Tablet:    "https://example.com/images/test-tablet.jpg",
					Desktop:   "https://example.com/images/test-desktop.jpg",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid product - zero price",
			input: &Product{
				ID:          "prod-1",
				Name:        "Test Product",
				Price:       0,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "https://example.com/images/test-thumb.jpg",
					Mobile:    "https://example.com/images/test-mobile.jpg",
					Tablet:    "https://example.com/images/test-tablet.jpg",
					Desktop:   "https://example.com/images/test-desktop.jpg",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid product - negative price",
			input: &Product{
				ID:          "prod-1",
				Name:        "Test Product",
				Price:       -9.99,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "https://example.com/images/test-thumb.jpg",
					Mobile:    "https://example.com/images/test-mobile.jpg",
					Tablet:    "https://example.com/images/test-tablet.jpg",
					Desktop:   "https://example.com/images/test-desktop.jpg",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid product - missing image",
			input: &Product{
				ID:          "prod-1",
				Name:        "Test Product",
				Price:       9.99,
				Category:    "Test Category",
			},
			wantErr: true,
		},
		{
			name: "invalid product - invalid image URLs",
			input: &Product{
				ID:          "prod-1",
				Name:        "Test Product",
				Price:       9.99,
				Category:    "Test Category",
				Image: &ProductImage{
					Thumbnail: "invalid-url",
					Mobile:    "invalid-url",
					Tablet:    "invalid-url",
					Desktop:   "invalid-url",
				},
			},
			wantErr: true,
		},

		// Order validation tests
		{
			name: "valid order",
			input: &Order{
				ID: "order-1",
				Items: []OrderItem{
					{ProductID: "prod-1", Quantity: 1, Price: 9.99},
				},
				Products: []Product{
					{
						ID:          "prod-1",
						Name:        "Test Product",
						Price:       9.99,
						Category:    "Test Category",
						Image: &ProductImage{
							Thumbnail: "https://example.com/images/test-thumb.jpg",
							Mobile:    "https://example.com/images/test-mobile.jpg",
							Tablet:    "https://example.com/images/test-tablet.jpg",
							Desktop:   "https://example.com/images/test-desktop.jpg",
						},
					},
				},
				TotalAmount: 9.99,
			},
			wantErr: false,
		},
		{
			name: "invalid order - missing ID",
			input: &Order{
				Items: []OrderItem{
					{ProductID: "prod-1", Quantity: 1, Price: 9.99},
				},
				Products: []Product{
					{
						ID:          "prod-1",
						Name:        "Test Product",
						Price:       9.99,
						Category:    "Test Category",
						Image: &ProductImage{
							Thumbnail: "https://example.com/images/test-thumb.jpg",
							Mobile:    "https://example.com/images/test-mobile.jpg",
							Tablet:    "https://example.com/images/test-tablet.jpg",
							Desktop:   "https://example.com/images/test-desktop.jpg",
						},
					},
				},
				TotalAmount: 9.99,
			},
			wantErr: true,
		},
		{
			name: "invalid order - empty items",
			input: &Order{
				ID:          "order-1",
				Items:       []OrderItem{},
				Products:    []Product{},
				TotalAmount: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid order - negative total amount",
			input: &Order{
				ID: "order-1",
				Items: []OrderItem{
					{ProductID: "prod-1", Quantity: 1, Price: 9.99},
				},
				Products: []Product{
					{
						ID:          "prod-1",
						Name:        "Test Product",
						Price:       9.99,
						Category:    "Test Category",
						Image: &ProductImage{
							Thumbnail: "https://example.com/images/test-thumb.jpg",
							Mobile:    "https://example.com/images/test-mobile.jpg",
							Tablet:    "https://example.com/images/test-tablet.jpg",
							Desktop:   "https://example.com/images/test-desktop.jpg",
						},
					},
				},
				TotalAmount: -9.99,
			},
			wantErr: true,
		},

		// Coupon validation tests
		{
			name: "valid coupon",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: 10,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: false,
		},
		{
			name: "invalid coupon - missing code",
			input: &Coupon{
				DiscountPercent: 10,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - zero discount",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: 0,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - negative discount",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: -10,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - discount over 100%",
			input: &Coupon{
				Code:            "SAVE200",
				DiscountPercent: 200,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - negative min order amount",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: 10,
				MinOrderAmount:  -20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - zero max usage",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: 10,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - negative max usage",
			input: &Coupon{
				Code:            "SAVE10",
				DiscountPercent: 10,
				MinOrderAmount:  20,
				ExpiryDate:      time.Now().Add(24 * time.Hour),
				MaxUsagePerUser: -1,
			},
			wantErr: true,
		},

		// Error response validation tests
		{
			name: "valid error response",
			input: &ErrorResponse{
				Code:    "INVALID_INPUT",
				Message: "Invalid input provided",
			},
			wantErr: false,
		},
		{
			name: "invalid error response - missing code",
			input: &ErrorResponse{
				Message: "Invalid input provided",
			},
			wantErr: true,
		},
		{
			name: "invalid error response - missing message",
			input: &ErrorResponse{
				Code: "INVALID_INPUT",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
