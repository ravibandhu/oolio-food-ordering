package services

import (
	"fmt"

	"github.com/ravibandhu/oolio-food-ordering/internal/data"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// OrderService defines the interface for order operations
type OrderService interface {
	PlaceOrder(req *models.OrderRequest) (*models.Order, error)
}

// OrderServiceImpl implements the OrderService interface
type OrderServiceImpl struct {
	store *data.Store
}

// NewOrderService creates a new OrderService instance
func NewOrderService(store *data.Store) OrderService {
	return &OrderServiceImpl{
		store: store,
	}
}

// PlaceOrder processes a new order request
func (s *OrderServiceImpl) PlaceOrder(req *models.OrderRequest) (*models.Order, error) {
	// Validate products and calculate total
	var products []models.Product
	var totalAmount float64

	// Validate and collect products
	for _, item := range req.Items {
		product, err := s.store.GetProduct(item.ProductID)
		if err != nil {
			return nil, models.NewErrorResponse("INVALID_PRODUCT", fmt.Sprintf("Invalid product ID: %s", item.ProductID))
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
