package models

// swagger:parameters placeOrder
type PlaceOrderRequest struct {
	// The customer ID placing the order
	// @required
	// @example cust-123
	CustomerID string `json:"customer_id" validate:"required"`

	// The list of items to order
	// @required
	// @minItems 1
	Items []OrderItemRequest `json:"items" validate:"required,dive"`

	// Optional coupon code to apply to the order
	// @example SAVE10
	CouponCode *string `json:"coupon_code,omitempty"`
}

// OrderItemRequest represents a product and quantity to order
type OrderItemRequest struct {
	// The ID of the product to order
	// @required
	// @example prod-123
	ProductID string `json:"product_id" validate:"required"`

	// The quantity of the product to order
	// @required
	// @minimum 1
	// @example 2
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

// swagger:response orderResponse
type PlaceOrderResponse struct {
	// The created order
	// @required
	Order *Order `json:"order" validate:"required"`

	// The amount saved by applying the coupon, if any
	// @example 5.99
	DiscountAmount *float64 `json:"discount_amount,omitempty"`

	// The original amount before discount
	// @required
	// @example 59.99
	OriginalAmount float64 `json:"original_amount" validate:"required"`
}

// NewPlaceOrderResponse creates a new PlaceOrderResponse
func NewPlaceOrderResponse(order *Order, discountAmount *float64, originalAmount float64) *PlaceOrderResponse {
	return &PlaceOrderResponse{
		Order:          order,
		DiscountAmount: discountAmount,
		OriginalAmount: originalAmount,
	}
}
