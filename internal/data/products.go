package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

// ProductStore represents a file-based store for products
type ProductStore struct {
	products map[string]*models.Product
	mu       sync.RWMutex
}

// NewProductStore creates a new ProductStore instance
func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]*models.Product),
	}
}

// LoadProducts reads product data from JSON files in the specified directory
func (s *ProductStore) LoadProducts(dir string) error {
	// Lock for writing
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing products
	s.products = make(map[string]*models.Product)

	// Get all JSON files in the directory
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("error finding product files: %w", err)
	}

	// Process each file
	for _, file := range files {
		if err := s.loadProductFile(file); err != nil {
			return fmt.Errorf("error loading file %s: %w", file, err)
		}
	}

	return nil
}

// loadProductFile reads and parses a single product file
func (s *ProductStore) loadProductFile(filename string) error {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a buffered reader
	reader := bufio.NewReader(file)

	// Create a decoder for JSON
	decoder := json.NewDecoder(reader)

	// Read the opening array bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error reading opening bracket: %w", err)
	}

	// Read products
	for decoder.More() {
		var product models.Product
		if err := decoder.Decode(&product); err != nil {
			return fmt.Errorf("error decoding product: %w", err)
		}

		// Validate the product
		if err := models.Validate(&product); err != nil {
			return fmt.Errorf("invalid product data: %w", err)
		}

		// Store the product
		s.products[product.ID] = &product
	}

	return nil
}

// GetProduct retrieves a product by ID
func (s *ProductStore) GetProduct(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[id]
	if !exists {
		return nil, fmt.Errorf("product not found: %s", id)
	}

	return product, nil
}

// GetAllProducts returns all products
func (s *ProductStore) GetAllProducts() []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]*models.Product, 0, len(s.products))
	for _, product := range s.products {
		products = append(products, product)
	}

	return products
}
