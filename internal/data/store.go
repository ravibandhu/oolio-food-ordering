package data

import (
	"fmt"
	"sync"

	"github.com/ravibandhu/oolio-food-ordering/internal/config"
	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// Store represents the data store for products and coupons
type Store struct {
	products *ProductStore
	coupons  *CouponStoreConcurrent
	config   *config.Config
	mu       sync.RWMutex
}

// NewStore creates a new Store instance
func NewStore(cfg *config.Config) (*Store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create product store
	productStore := NewProductStore()
	if err := productStore.LoadProducts(cfg.Files.ProductsFile); err != nil {
		return nil, fmt.Errorf("failed to load products: %w", err)
	}

	// Get coupon store instance
	couponStore, err := Instance(cfg.Files.CouponsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize coupon store: %w", err)
	}

	// Create and initialize store
	store := &Store{
		products: productStore,
		coupons:  couponStore,
		config:   cfg,
	}

	return store, nil
}

// GetProduct retrieves a product by ID
func (s *Store) GetProduct(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.products.GetProduct(id)
}

// GetAllProducts returns all available products
func (s *Store) GetAllProducts() []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.products.GetAllProducts()
}

// ValidateCoupon checks if a coupon is valid
func (s *Store) ValidateCoupon(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coupons.GetCoupon(code)
}

// ReloadData reloads both products and coupons data
func (s *Store) ReloadData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset coupon store singleton state to force reload
	instance = nil
	once = sync.Once{}
	loadErr = nil
	loadDir = ""
	loaded = false

	// Reload products
	if err := s.products.LoadProducts(s.config.Files.ProductsFile); err != nil {
		return fmt.Errorf("failed to reload products: %w", err)
	}

	// Get a fresh instance of the coupon store with reloaded data
	couponStore, err := Instance(s.config.Files.CouponsDir)
	if err != nil {
		return fmt.Errorf("failed to reload coupons: %w", err)
	}
	s.coupons = couponStore

	return nil
}
