package models

import (
	"testing"
	"time"
)

func TestNewProduct(t *testing.T) {
	id := "prod-1"
	name := "Burger"
	desc := "Delicious burger"
	price := 9.99
	category := "Main Course"

	p := NewProduct(id, name, desc, price, category)

	if p.ID != id {
		t.Errorf("expected ID %s, got %s", id, p.ID)
	}
	if p.Name != name {
		t.Errorf("expected Name %s, got %s", name, p.Name)
	}
	if p.Description != desc {
		t.Errorf("expected Description %s, got %s", desc, p.Description)
	}
	if p.Price != price {
		t.Errorf("expected Price %.2f, got %.2f", price, p.Price)
	}
	if p.Category != category {
		t.Errorf("expected Category %s, got %s", category, p.Category)
	}
	if p.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Test validation
	if err := Validate(p); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestNewOrder(t *testing.T) {
	id := "order-1"
	customerID := "cust-1"
	items := []OrderItem{
		{
			ProductID: "prod-1",
			Quantity:  2,
			Price:     9.99,
		},
	}
	total := 19.98
	couponCode := "SAVE10"

	o := NewOrder(id, customerID, items, total, &couponCode)

	if o.ID != id {
		t.Errorf("expected ID %s, got %s", id, o.ID)
	}
	if o.CustomerID != customerID {
		t.Errorf("expected CustomerID %s, got %s", customerID, o.CustomerID)
	}
	if len(o.Items) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(o.Items))
	}
	if o.TotalAmount != total {
		t.Errorf("expected TotalAmount %.2f, got %.2f", total, o.TotalAmount)
	}
	if o.Status != "pending" {
		t.Errorf("expected Status pending, got %s", o.Status)
	}
	if o.CouponCode == nil || *o.CouponCode != couponCode {
		t.Errorf("expected CouponCode %s, got %v", couponCode, o.CouponCode)
	}
	if o.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if o.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Test validation
	if err := Validate(o); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestNewCoupon(t *testing.T) {
	code := "SAVE10"
	discount := 10.0
	minAmount := 20.0
	expiry := time.Now().Add(24 * time.Hour)
	maxUsage := 1

	c := NewCoupon(code, discount, minAmount, expiry, maxUsage)

	if c.Code != code {
		t.Errorf("expected Code %s, got %s", code, c.Code)
	}
	if c.DiscountPercent != discount {
		t.Errorf("expected DiscountPercent %.2f, got %.2f", discount, c.DiscountPercent)
	}
	if c.MinOrderAmount != minAmount {
		t.Errorf("expected MinOrderAmount %.2f, got %.2f", minAmount, c.MinOrderAmount)
	}
	if !c.ExpiryDate.Equal(expiry) {
		t.Errorf("expected ExpiryDate %v, got %v", expiry, c.ExpiryDate)
	}
	if c.MaxUsagePerUser != maxUsage {
		t.Errorf("expected MaxUsagePerUser %d, got %d", maxUsage, c.MaxUsagePerUser)
	}
	if !c.IsActive {
		t.Error("expected IsActive to be true")
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if c.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Test validation
	if err := Validate(c); err != nil {
		t.Errorf("validation failed: %v", err)
	}
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

	if e1.Code != code {
		t.Errorf("expected Code %s, got %s", code, e1.Code)
	}
	if e1.Message != message {
		t.Errorf("expected Message %s, got %s", message, e1.Message)
	}
	if e1.Details == nil {
		t.Error("Details should not be nil")
	}
	if e1.Details["price"] != "must be greater than 0" {
		t.Errorf("expected detail message %s, got %s", "must be greater than 0", e1.Details["price"])
	}

	// Test creating error response with multiple details
	e2 := NewErrorResponse(code, message).AddDetails(details)

	if len(e2.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(e2.Details))
	}
	if e2.Details["price"] != details["price"] {
		t.Errorf("expected price detail %s, got %s", details["price"], e2.Details["price"])
	}
	if e2.Details["quantity"] != details["quantity"] {
		t.Errorf("expected quantity detail %s, got %s", details["quantity"], e2.Details["quantity"])
	}

	// Test method chaining
	e3 := NewErrorResponse(code, message).
		AddDetail("price", "must be greater than 0").
		AddDetail("quantity", "must be at least 1")

	if len(e3.Details) != 2 {
		t.Errorf("expected 2 details, got %d", len(e3.Details))
	}

	// Test validation
	if err := Validate(e1); err != nil {
		t.Errorf("validation failed: %v", err)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:  "invalid product - missing required fields",
			input: &Product{
				// Missing required fields
			},
			wantErr: true,
		},
		{
			name: "invalid order - negative total amount",
			input: &Order{
				ID:          "order-1",
				CustomerID:  "cust-1",
				Items:       []OrderItem{{ProductID: "prod-1", Quantity: 1, Price: 9.99}},
				TotalAmount: -1,
				Status:      "pending",
			},
			wantErr: true,
		},
		{
			name: "invalid coupon - discount over 100%",
			input: &Coupon{
				Code:            "SAVE200",
				DiscountPercent: 200,
				MinOrderAmount:  0,
				ExpiryDate:      time.Now(),
				MaxUsagePerUser: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid error response - missing code",
			input: &ErrorResponse{
				Message: "Error occurred",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
