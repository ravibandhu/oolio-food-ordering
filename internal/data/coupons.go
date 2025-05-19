package data

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CouponStore struct to hold the loaded coupon codes.
type CouponStore struct {
	coupons map[string]struct{} // Set-like for efficient lookups
	mu      sync.RWMutex        // Mutex for concurrent access if needed (though LoadCoupons is typically done at startup)
}

// NewCouponStore creates and initializes a new CouponStore.
func NewCouponStore() *CouponStore {
	return &CouponStore{
		coupons: make(map[string]struct{}),
	}
}

// loadCouponsFromFile reads coupon codes from a single file.
func (s *CouponStore) loadCouponsFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	var reader *bufio.Reader
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("error creating gzip reader for %s: %w", filePath, err)
		}
		defer gzReader.Close()
		reader = bufio.NewReader(gzReader)
	} else {
		reader = bufio.NewReader(file)
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			s.coupons[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return nil
}

// LoadCoupons loads coupon codes from all files (including .gz) in the specified directory.
func (s *CouponStore) LoadCoupons(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}
		if !info.IsDir() {
			fmt.Printf("Loading coupons from file: %s\n", path)
			if err := s.loadCouponsFromFile(path); err != nil {
				fmt.Printf("Error loading coupons from %s: %v\n", path, err)
				// Decide if you want to continue loading from other files or stop here
				// For now, we'll continue. To stop, return the error.
			}
		}
		return nil
	})
}

// GetCoupon checks if a coupon code exists and returns a random discount percentage if it does.
func (s *CouponStore) GetCoupon(code string) (discountPercentage int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.coupons[code]; exists {
		// Generate a random discount percentage
		percentages := []int{10, 15, 20, 25, 30, 33, 40, 50}
		randomIndex := rand.Intn(len(percentages))
		return percentages[randomIndex], nil
	}
	return 0, fmt.Errorf("invalid coupon code: %s", code)
}
