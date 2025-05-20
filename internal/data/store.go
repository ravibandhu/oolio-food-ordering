package data

import (
	"context"
	"fmt"
	"sync"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// CouponValidator defines the interface for coupon validation
type CouponValidator interface {
	GetCoupon(code string) bool
}

// Store represents the data store for products and coupons
type Store struct {
	products *ProductStore
	coupons  CouponValidator
	config   *config.Config
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStore creates a new Store instance
func NewStore(ctx context.Context, cfg *config.Config) (*Store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create a child context with cancellation
	storeCtx, cancel := context.WithCancel(ctx)

	// Create product store
	productStore := NewProductStore()
	if err := productStore.LoadProducts(cfg.Files.ProductsFile); err != nil {
		cancel() // Clean up context if product loading fails
		return nil, fmt.Errorf("failed to load products: %w", err)
	}

	// Get coupon store instance
	couponStore, err := CouponStoreConcurrentInstance(cfg.Files.CouponsDir)
	if err != nil {
		cancel() // Clean up context if coupon store initialization fails
		return nil, fmt.Errorf("failed to initialize coupon store: %w", err)
	}

	// Create and initialize store
	store := &Store{
		products: productStore,
		coupons:  couponStore,
		config:   cfg,
		ctx:      storeCtx,
		cancel:   cancel,
	}

	return store, nil
}

// Close performs cleanup of the store resources
func (s *Store) Close() error {
	s.cancel() // Cancel the store's context
	// Add any additional cleanup needed for products and coupons
	return nil
}

// GetProduct retrieves a product by ID
func (s *Store) GetProduct(id string) (*models.Product, error) {
	// Check if context is cancelled
	if err := s.ctx.Err(); err != nil {
		return nil, fmt.Errorf("store is closed: %w", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.products.GetProduct(id)
}

// GetAllProducts returns all available products
func (s *Store) GetAllProducts() []*models.Product {
	// Check if context is cancelled
	if err := s.ctx.Err(); err != nil {
		return nil // Return empty slice if store is closed
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.products.GetAllProducts()
}

// ValidateCoupon checks if a coupon is valid
func (s *Store) ValidateCoupon(code string) bool {
	// Check if context is cancelled
	if err := s.ctx.Err(); err != nil {
		return false // Return invalid if store is closed
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coupons.GetCoupon(code)
}
