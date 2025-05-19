package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// OrderService handles order-related business logic
type OrderService struct {
	store *data.Store
}

// NewOrderService creates a new OrderService instance
func NewOrderService(store *data.Store) *OrderService {
	return &OrderService{
		store: store,
	}
}

// PlaceOrder processes a new order request
func (s *OrderService) PlaceOrder(req *models.PlaceOrderRequest) (*models.PlaceOrderResponse, error) {
	// Validate products and calculate total
	orderItems := make([]models.OrderItem, 0, len(req.Items))
	var totalAmount float64

	for _, item := range req.Items {
		// Get product details
		product, err := s.store.GetProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID %s: %w", item.ProductID, err)
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
		orderItems = append(orderItems, orderItem)

		// Add to total
		totalAmount += product.Price * float64(item.Quantity)
	}

	// Store original amount before discount
	originalAmount := totalAmount

	// Apply coupon if provided
	var discountAmount *float64
	if req.CouponCode != nil {
		// Validate coupon
		if !s.store.ValidateCoupon(*req.CouponCode) {
			return nil, fmt.Errorf("invalid coupon code: %s", *req.CouponCode)
		}

		// Calculate discount (10% for now - could be made configurable)
		discount := totalAmount * 0.10
		discountAmount = &discount
		totalAmount -= discount
	}

	// Create order
	order := models.NewOrder(
		fmt.Sprintf("order-%s", uuid.New().String()),
		req.CustomerID,
		orderItems,
		totalAmount,
		req.CouponCode,
	)

	// Create response
	response := models.NewPlaceOrderResponse(
		order,
		discountAmount,
		originalAmount,
	)

	return response, nil
}
