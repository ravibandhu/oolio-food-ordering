package data

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ravibandhu/oolio-food-ordering/internal/models"
	"github.com/stretchr/testify/require"
)

// Test variables for coupon testing
var (
	testOnce      sync.Once
	testSingleton = false
)

// resetForTest is a testing utility to reset the singleton state
// This helps tests to work with the singleton pattern without modifying it
func resetForTest() {
	// For test only: If we haven't touched the singleton yet, do nothing
	if !testSingleton {
		testSingleton = true
		return
	}

	// Reset the package variables used by CouponStoreConcurrentInstance
	once = sync.Once{}
	instance = nil
	loadErr = nil
	loadDir = ""
	loaded = false

	// Reset coupon shards
	for i := range couponShards {
		couponShards[i].m = make(map[string]uint32)
	}
}

// createGzipFile creates a gzipped file with the given content
func createGzipFile(t *testing.T, filepath, content string) {
	// Create and open the output file
	file, err := os.Create(filepath)
	require.NoError(t, err, "Failed to create file")
	defer file.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Write content to the gzip writer
	_, err = gzipWriter.Write([]byte(content))
	require.NoError(t, err, "Failed to write to gzip file")
}

// createTestProducts creates a set of test products
func createTestProducts() []models.Product {
	return []models.Product{
		{
			ID:          "prod-1",
			Name:    "Test Product 1",
			Price:   9.99,
			Category: "Test Category",
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
	}
}

// setupProductStore creates and initializes a ProductStore with test data
func setupProductStore() *ProductStore {
	store := NewProductStore()

	// Directly set products in the store
	for _, product := range createTestProducts() {
		store.products[product.ID] = &product
	}

	return store
}

// setupCouponStore creates and returns the CouponStoreConcurrent singleton
// with test coupons directly injected for proper validation
func setupCouponStore() *CouponStoreConcurrent {
	// Reset singleton state
	resetForTest()

	// Initialize shards
	initializeShards()

	// Define valid coupons - each appearing in at least 2 different "files" to be considered valid
	validCoupons := []string{"TEST10", "TEST20", "TEST30"}

	// Each coupon needs to have a bitmask value with at least 2 bits set
	// to indicate it appears in at least 2 files, per CouponStoreConcurrent logic
	for _, coupon := range validCoupons {
		shardIndex := getShardIndex(coupon)

		// Set bitmask to indicate the coupon appears in at least 2 files
		// We'll use bitmask values 3 (011 in binary) which means it appears in files 0 and 1
		couponShards[shardIndex].m[coupon] = 3 // 3 = 0b11 (binary) = appears in files 0 and 1
	}

	// Create and initialize the store
	store := NewCouponStoreConcurrent()

	// Directly populate the coupons map in the store with our valid coupons
	// This ensures the GetCoupon method will find them
	store.mu.Lock()
	store.coupons = make(map[string]struct{})
	for _, coupon := range validCoupons {
		store.coupons[coupon] = struct{}{}
	}
	store.mu.Unlock()

	return store
}

// countSetBits counts the number of bits set to 1 in a uint32 value
// This helper function mimics bits.OnesCount32 used in LoadAndFindValidCoupons
func countSetBits(n uint32) int {
	count := 0
	for n > 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

// setupCouponTestData creates test coupon files that will be used by tests
func setupCouponTestData(t *testing.T) string {
	// Reset the singleton state for tests
	resetForTest()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "coupon-test")
	require.NoError(t, err)

	// Set up cleanup on test finish
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Create test coupon files with the same content in each file to ensure they are valid
	// (each coupon must appear in at least 2 files to be valid)
	couponFiles := []string{"coupons1.txt.gz", "coupons2.txt.gz", "coupons3.txt.gz"}

	// Ensure every coupon appears in at least 2 files to be valid according to CouponStoreConcurrent
	coupons := []string{"TEST10", "TEST20", "TEST30"}

	// The content for each file will be the same to ensure all coupons are valid
	content := ""
	for _, coupon := range coupons {
		content += coupon + "\n"
	}

	for _, file := range couponFiles {
		couponFile := filepath.Join(tempDir, file)
		createGzipFile(t, couponFile, content)
	}

	return tempDir
}
