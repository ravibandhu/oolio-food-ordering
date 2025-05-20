package data

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/ravibandhu/oolio-food-ordering/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadProducts loads products from a JSON file for testing
func LoadProducts(ctx context.Context, filePath string) ([]models.Product, error) {
	// Check if the context is cancelled
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Read the file contents
	var products []models.Product
	if err := json.NewDecoder(file).Decode(&products); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	// Check if the file contains products
	if len(products) == 0 {
		return nil, fmt.Errorf("no products found in file")
	}

	// Validate each product
	for i, product := range products {
		if err := models.Validate(&product); err != nil {
			return nil, fmt.Errorf("invalid product at index %d: %w", i, err)
		}
	}

	return products, nil
}

func TestProductStore(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "product-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test products file
	productsFile := filepath.Join(tempDir, "products.json")
	err = os.WriteFile(productsFile, []byte(`[
		{
			"id": "prod-1",
			"name": "Test Product 1",
			"description": "Test Description 1",
			"price": 9.99,
			"category": "Test Category",
			"image": {
				"thumbnail": "https://example.com/images/prod-1-thumb.jpg",
				"mobile": "https://example.com/images/prod-1-mobile.jpg",
				"tablet": "https://example.com/images/prod-1-tablet.jpg",
				"desktop": "https://example.com/images/prod-1-desktop.jpg"
			},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		},
		{
			"id": "prod-2",
			"name": "Test Product 2",
			"description": "Test Description 2",
			"price": 19.99,
			"category": "Test Category",
			"image": {
				"thumbnail": "https://example.com/images/prod-2-thumb.jpg",
				"mobile": "https://example.com/images/prod-2-mobile.jpg",
				"tablet": "https://example.com/images/prod-2-tablet.jpg",
				"desktop": "https://example.com/images/prod-2-desktop.jpg"
			},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	]`), 0644)
	assert.NoError(t, err)

	t.Run("LoadProducts", func(t *testing.T) {
		store := NewProductStore()
		err := store.LoadProducts(productsFile)
		assert.NoError(t, err)

		products := store.GetAllProducts()
		assert.Len(t, products, 2)
	})

	t.Run("GetProduct", func(t *testing.T) {
		store := NewProductStore()
		err := store.LoadProducts(productsFile)
		assert.NoError(t, err)

		product, err := store.GetProduct("prod-1")
		assert.NoError(t, err)
		assert.Equal(t, "Test Product 1", product.Name)
	})

	t.Run("GetAllProducts", func(t *testing.T) {
		store := NewProductStore()
		err := store.LoadProducts(productsFile)
		assert.NoError(t, err)

		products := store.GetAllProducts()
		assert.Len(t, products, 2)
		assert.Equal(t, "Test Product 1", products[0].Name)
		assert.Equal(t, "Test Product 2", products[1].Name)
	})
}

func TestLoadProducts(t *testing.T) {
	testData := testutil.SetupTestData(t)
	defer testData.Cleanup()

	tests := []struct {
		name    string
		setup   func() string
		want    []models.Product
		wantErr bool
	}{
		{
			name: "valid products file",
			setup: func() string {
				return testData.ProductsFile
			},
			want: []models.Product{
				{
					ID:          "prod-1",
					Name:        "Test Product 1",
					Price:       9.99,
					Category:    "Test Category",
					Image: &models.ProductImage{
						Thumbnail: "https://example.com/images/prod-1-thumb.jpg",
						Mobile:    "https://example.com/images/prod-1-mobile.jpg",
						Tablet:    "https://example.com/images/prod-1-tablet.jpg",
						Desktop:   "https://example.com/images/prod-1-desktop.jpg",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:          "prod-2",
					Name:        "Test Product 2",
					Price:       19.99,
					Category:    "Test Category",
					Image: &models.ProductImage{
						Thumbnail: "https://example.com/images/prod-2-thumb.jpg",
						Mobile:    "https://example.com/images/prod-2-mobile.jpg",
						Tablet:    "https://example.com/images/prod-2-tablet.jpg",
						Desktop:   "https://example.com/images/prod-2-desktop.jpg",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid JSON format",
			setup: func() string {
				file := filepath.Join(testData.TempDir, "invalid.json")
				err := os.WriteFile(file, []byte(`{invalid json`), 0644)
				require.NoError(t, err)
				return file
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty file",
			setup: func() string {
				file := filepath.Join(testData.TempDir, "empty.json")
				err := os.WriteFile(file, []byte(`[]`), 0644)
				require.NoError(t, err)
				return file
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing required fields",
			setup: func() string {
				file := filepath.Join(testData.TempDir, "missing_fields.json")
				products := []models.Product{
					{
						ID:   "prod-1",
						Name: "Test Product 1",
						// Missing required fields
					},
				}
				data, err := json.Marshal(products)
				require.NoError(t, err)
				err = os.WriteFile(file, data, 0644)
				require.NoError(t, err)
				return file
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid price",
			setup: func() string {
				file := filepath.Join(testData.TempDir, "invalid_price.json")
				products := []models.Product{
					{
						ID:          "prod-1",
						Name:        "Test Product 1",
						Price:       -9.99, // Negative price
						Category:    "Test Category",
						Image: &models.ProductImage{
							Thumbnail: "https://example.com/images/prod-1-thumb.jpg",
							Mobile:    "https://example.com/images/prod-1-mobile.jpg",
							Tablet:    "https://example.com/images/prod-1-tablet.jpg",
							Desktop:   "https://example.com/images/prod-1-desktop.jpg",
						},
					},
				}
				data, err := json.Marshal(products)
				require.NoError(t, err)
				err = os.WriteFile(file, data, 0644)
				require.NoError(t, err)
				return file
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid image URLs",
			setup: func() string {
				file := filepath.Join(testData.TempDir, "invalid_urls.json")
				products := []models.Product{
					{
						ID:          "prod-1",
						Name:        "Test Product 1",
						Price:       9.99,
						Category:    "Test Category",
						Image: &models.ProductImage{
							Thumbnail: "invalid-url",
							Mobile:    "invalid-url",
							Tablet:    "invalid-url",
							Desktop:   "invalid-url",
						},
					},
				}
				data, err := json.Marshal(products)
				require.NoError(t, err)
				err = os.WriteFile(file, data, 0644)
				require.NoError(t, err)
				return file
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(testData.TempDir, "nonexistent.json")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			file := tt.setup()
			got, err := LoadProducts(ctx, file)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.want), len(got))
				for i, want := range tt.want {
					assert.Equal(t, want.ID, got[i].ID)
					assert.Equal(t, want.Name, got[i].Name)
					assert.Equal(t, want.Price, got[i].Price)
					assert.Equal(t, want.Category, got[i].Category)
					assert.Equal(t, want.Image.Thumbnail, got[i].Image.Thumbnail)
					assert.Equal(t, want.Image.Mobile, got[i].Image.Mobile)
					assert.Equal(t, want.Image.Tablet, got[i].Image.Tablet)
					assert.Equal(t, want.Image.Desktop, got[i].Image.Desktop)
					assert.Equal(t, want.CreatedAt, got[i].CreatedAt)
					assert.Equal(t, want.UpdatedAt, got[i].UpdatedAt)
				}
			}
		})
	}
}
