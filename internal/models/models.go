package models

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Package models provides the data models for the Oolio Food Ordering system
// swagger:meta

// Product represents a food item available for ordering
type Product struct {
	// The unique identifier of the product
	// @required
	// @example 1
	ID string `json:"id" validate:"required"`

	// The name of the product
	// @required
	// @example Waffle with Berries
	Name string `json:"name" validate:"required"`

	// The price of the product in the default currency
	// @required
	// @minimum 0.01
	// @example 6.50
	Price float64 `json:"price" validate:"required,gt=0"`

	// The category of the product
	// @required
	// @example Waffle
	Category string `json:"category" validate:"required"`

	// The product images in different sizes
	// @required
	Image *ProductImage `json:"image" validate:"required"`

	// The timestamp when the product was created
	// @example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at,omitempty"`

	// The timestamp when the product was last updated
	// @example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ProductImage represents different sizes of a product image
type ProductImage struct {
	// Thumbnail version of the image
	// @example https://orderfoodonline.deno.dev/public/images/image-waffle-thumbnail.jpg
	Thumbnail string `json:"thumbnail" validate:"required,url"`

	// Mobile version of the image
	// @example https://orderfoodonline.deno.dev/public/images/image-waffle-mobile.jpg
	Mobile string `json:"mobile" validate:"required,url"`

	// Tablet version of the image
	// @example https://orderfoodonline.deno.dev/public/images/image-waffle-tablet.jpg
	Tablet string `json:"tablet" validate:"required,url"`

	// Desktop version of the image
	// @example https://orderfoodonline.deno.dev/public/images/image-waffle-desktop.jpg
	Desktop string `json:"desktop" validate:"required,url"`
}

// OrderItem represents a single item in an order with its quantity
type OrderItem struct {
	// The ID of the product being ordered
	// @required
	// @example 1
	ProductID string `json:"productId" validate:"required"`

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

// Order represents a complete order with its items and details
type Order struct {
	// The unique identifier of the order
	// @example order-0000-0000-0000-0000
	ID string `json:"id" validate:"required"`

	// List of ordered items
	// @required
	Items []OrderItem `json:"items" validate:"required"`

	// List of products in the order with their details
	// @required
	Products []Product `json:"products" validate:"required"`

	// The total amount of the order after any discounts
	// @required
	// @minimum 0
	// @example 19.99
	TotalAmount float64 `json:"total_amount" validate:"required,gte=0"`

	// The coupon code used for the order, if any
	// @example SAVE10
	CouponCode string `json:"coupon_code,omitempty"`

	// The timestamp when the order was created
	// @example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at,omitempty"`

	// The timestamp when the order was last updated
	// @example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at,omitempty"`
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

// OrderRequest represents the request body for placing an order
type OrderRequest struct {
	// Optional coupon code to apply to the order
	// @example SAVE20
	CouponCode string `json:"couponCode"`

	// List of items to order
	// @required
	Items []OrderItem `json:"items" validate:"required,min=1,dive"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	// Error code
	// @example INVALID_REQUEST
	Code string `json:"code" validate:"required"`

	// Error message
	// @example Invalid request data
	Message string `json:"message" validate:"required"`

	// Additional error details
	Details map[string]interface{} `json:"details,omitempty"`
}

// Validate uses the validator package to validate a struct
func Validate(i interface{}) error {
	validate := validator.New()
	return validate.Struct(i)
}

// NewProduct creates a new Product instance
func NewProduct(id, name string, price float64, category string, image *ProductImage) *Product {
	now := time.Now()
	return &Product{
		ID:          id,
		Name:        name,
		Price:       price,
		Category:    category,
		Image:       image,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewOrder creates a new Order instance
func NewOrder(items []OrderItem, products []Product, totalAmount float64, couponCode string) *Order {
	return &Order{
		ID:          fmt.Sprintf("order-%s", uuid.New().String()),
		Items:       items,
		Products:    products,
		TotalAmount: totalAmount,
		CouponCode:  couponCode,
		CreatedAt:   time.Now(),
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

// NewErrorResponse creates a new ErrorResponse
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// AddDetail adds a detail to the error response
func (e *ErrorResponse) AddDetail(key string, value interface{}) *ErrorResponse {
	e.Details[key] = value
	return e
}

// AddDetails adds multiple field-specific error details to the response
func (e *ErrorResponse) AddDetails(details map[string]string) *ErrorResponse {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for field, message := range details {
		e.Details[field] = interface{}(message)
	}
	return e
}
