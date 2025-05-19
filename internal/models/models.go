package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// Package models provides the data models for the Oolio Food Ordering system
// swagger:meta

// Product represents a food item available for ordering
type Product struct {
	// The unique identifier of the product
	// @required
	// @example prod-123
	ID string `json:"id" validate:"required"`

	// The name of the product
	// @required
	// @example Chicken Burger
	Name string `json:"name" validate:"required"`

	// A detailed description of the product
	// @required
	// @example A juicy chicken burger with fresh lettuce and tomatoes
	Description string `json:"description" validate:"required"`

	// The price of the product in the default currency
	// @required
	// @minimum 0.01
	// @example 9.99
	Price float64 `json:"price" validate:"required,gt=0"`

	// The category of the product
	// @required
	// @example Main Course
	Category string `json:"category" validate:"required"`

	// The timestamp when the product was created
	// @example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`

	// The timestamp when the product was last updated
	// @example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItem represents a single item in an order with its quantity
type OrderItem struct {
	// The ID of the product being ordered
	// @required
	// @example prod-123
	ProductID string `json:"product_id" validate:"required"`

	// The quantity of the product ordered
	// @required
	// @minimum 1
	// @example 2
	Quantity int `json:"quantity" validate:"required,gt=0"`

	// The price of the product at the time of ordering
	// @required
	// @minimum 0.01
	// @example 9.99
	Price float64 `json:"price" validate:"required,gt=0"`
}

// Order represents a customer's food order
type Order struct {
	// The unique identifier of the order
	// @required
	// @example order-123
	ID string `json:"id" validate:"required"`

	// The ID of the customer who placed the order
	// @required
	// @example cust-123
	CustomerID string `json:"customer_id" validate:"required"`

	// The list of items in the order
	// @required
	// @minItems 1
	Items []OrderItem `json:"items" validate:"required,dive"`

	// The total amount of the order
	// @required
	// @minimum 0.01
	// @example 29.99
	TotalAmount float64 `json:"total_amount" validate:"required,gt=0"`

	// The current status of the order
	// @required
	// @enum pending confirmed preparing ready delivered cancelled
	// @example pending
	Status string `json:"status" validate:"required,oneof=pending confirmed preparing ready delivered cancelled"`

	// The coupon code applied to the order, if any
	// @example SAVE10
	CouponCode *string `json:"coupon_code,omitempty"`

	// The timestamp when the order was created
	// @example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`

	// The timestamp when the order was last updated
	// @example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// Coupon represents a discount coupon that can be applied to orders
type Coupon struct {
	// The unique code of the coupon
	// @required
	// @example SAVE10
	Code string `json:"code" validate:"required"`

	// The percentage discount offered by the coupon
	// @required
	// @minimum 0.01
	// @maximum 100
	// @example 10
	DiscountPercent float64 `json:"discount_percent" validate:"required,gt=0,lte=100"`

	// The minimum order amount required to use the coupon
	// @required
	// @minimum 0
	// @example 20
	MinOrderAmount float64 `json:"min_order_amount" validate:"required,gte=0"`

	// The expiry date of the coupon
	// @required
	// @example 2024-12-31T23:59:59Z
	ExpiryDate time.Time `json:"expiry_date" validate:"required"`

	// The maximum number of times a user can use this coupon
	// @required
	// @minimum 1
	// @example 1
	MaxUsagePerUser int `json:"max_usage_per_user" validate:"required,gt=0"`

	// Whether the coupon is currently active
	// @example true
	IsActive bool `json:"is_active"`

	// The timestamp when the coupon was created
	// @example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`

	// The timestamp when the coupon was last updated
	// @example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorDetails represents additional error information
type ErrorDetails struct {
	// The field that caused the error
	// @example price
	Field string `json:"field,omitempty"`

	// The specific error for the field
	// @example must be greater than 0
	Error string `json:"error,omitempty"`
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	// The error code
	// @required
	// @example INVALID_INPUT
	Code string `json:"code" validate:"required"`

	// A human-readable error message
	// @required
	// @example Invalid input provided
	Message string `json:"message" validate:"required"`

	// Additional error details mapping field names to error messages
	// @example {"price": "must be greater than 0", "quantity": "must be at least 1"}
	Details map[string]string `json:"details,omitempty"`
}

// Validate uses the validator package to validate a struct
func Validate(i interface{}) error {
	validate := validator.New()
	return validate.Struct(i)
}

// NewProduct creates a new Product with the current timestamp
func NewProduct(id, name, description string, price float64, category string) *Product {
	now := time.Now()
	return &Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewOrder creates a new Order with the current timestamp
func NewOrder(id, customerID string, items []OrderItem, totalAmount float64, couponCode *string) *Order {
	now := time.Now()
	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      "pending",
		CouponCode:  couponCode,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewCoupon creates a new Coupon with the current timestamp
func NewCoupon(code string, discountPercent, minOrderAmount float64, expiryDate time.Time, maxUsagePerUser int) *Coupon {
	now := time.Now()
	return &Coupon{
		Code:            code,
		DiscountPercent: discountPercent,
		MinOrderAmount:  minOrderAmount,
		ExpiryDate:      expiryDate,
		MaxUsagePerUser: maxUsagePerUser,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// NewErrorResponse creates a new ErrorResponse with the given code and message
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: make(map[string]string),
	}
}

// AddDetail adds a field-specific error detail to the response
func (e *ErrorResponse) AddDetail(field, errorMessage string) *ErrorResponse {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[field] = errorMessage
	return e
}

// AddDetails adds multiple field-specific error details to the response
func (e *ErrorResponse) AddDetails(details map[string]string) *ErrorResponse {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	for field, message := range details {
		e.Details[field] = message
	}
	return e
}
