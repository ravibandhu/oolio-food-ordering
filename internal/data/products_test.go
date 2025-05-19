package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
)

func TestProductStore(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "product-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test products
	now := time.Now()
	testProducts := []*models.Product{
		{
			ID:          "prod-1",
			Name:        "Test Product 1",
			Description: "Description 1",
			Price:       9.99,
			Category:    "Category 1",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "prod-2",
			Name:        "Test Product 2",
			Description: "Description 2",
			Price:       19.99,
			Category:    "Category 2",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Create a test JSON file
	testFile := filepath.Join(tempDir, "products.json")
	file, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write test products to file
	if err := json.NewEncoder(file).Encode(testProducts); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	file.Close()

	// Create a new ProductStore
	store := NewProductStore()

	// Test LoadProducts
	t.Run("LoadProducts", func(t *testing.T) {
		err := store.LoadProducts(testFile)
		if err != nil {
			t.Errorf("LoadProducts failed: %v", err)
		}

		// Check if all products were loaded
		products := store.GetAllProducts()
		if len(products) != len(testProducts) {
			t.Errorf("Expected %d products, got %d", len(testProducts), len(products))
		}
	})

	// Test GetProduct
	t.Run("GetProduct", func(t *testing.T) {
		product, err := store.GetProduct("prod-1")
		if err != nil {
			t.Errorf("GetProduct failed: %v", err)
		}
		if product == nil || product.ID != "prod-1" || product.Name != "Test Product 1" {
			t.Errorf("Got wrong product: %+v", product)
		}

		// Test non-existent product
		_, err = store.GetProduct("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent product")
		}
	})

	// Test GetAllProducts
	t.Run("GetAllProducts", func(t *testing.T) {
		products := store.GetAllProducts()
		if len(products) != len(testProducts) {
			t.Errorf("Expected %d products, got %d", len(testProducts), len(products))
		}

		// Verify each product
		for _, p := range products {
			found := false
			for _, tp := range testProducts {
				if p.ID == tp.ID {
					found = true
					if p.Name != tp.Name || p.Price != tp.Price {
						t.Errorf("Product data mismatch: %+v", p)
					}
					break
				}
			}
			if !found {
				t.Errorf("Unexpected product found: %+v", p)
			}
		}
	})
}
