package data

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CouponStore struct to hold the loaded coupon codes.
type CouponStoreConcurrent struct {
	coupons map[string]struct{} // Set-like for efficient lookups
	mu      sync.RWMutex        // Mutex for concurrent access
}

var (
	once     sync.Once
	instance *CouponStoreConcurrent
	loadErr  error
	loadDir  string
	loaded   bool
)

// Instance returns the singleton instance of CouponStore, loading if not already loaded.
func Instance(dir string) (*CouponStoreConcurrent, error) {
	// Check if directory exists and is accessible
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("error accessing path %s: %w", dir, err)
	}

	// Reset singleton if directory changes
	if loaded && loadDir != dir {
		instance = nil
		once = sync.Once{}
		loadErr = nil
		loadDir = ""
		loaded = false
	}

	once.Do(func() {
		instance = &CouponStoreConcurrent{
			coupons: make(map[string]struct{}),
		}
		loadDir = dir
		loadErr = instance.LoadCouponsConcurrent(dir)
		if loadErr == nil {
			loaded = true
		}
	})

	return instance, loadErr
}

// LoadCouponsConcurrent loads coupons from all files in the directory concurrently
func (s *CouponStoreConcurrent) LoadCouponsConcurrent(dir string) error {
	// Read directory entries first, before taking any locks
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Create channel for results
	resultChan := make(chan map[string]struct{}, len(entries))

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup

	// Process each file concurrently
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		wg.Add(1)
		go func(e os.DirEntry) {
			defer wg.Done()

			filePath := filepath.Join(dir, e.Name())
			coupons, err := s.loadCouponsFromFile(filePath)
			if err != nil {
				// Log the error but don't fail completely
				fmt.Printf("Error loading coupons from %s: %v\n", filePath, err)
				return
			}
			if len(coupons) > 0 {
				resultChan <- coupons
			}
		}(entry)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resultChan)

	// Merge results under a single lock
	s.mu.Lock()
	s.coupons = make(map[string]struct{})
	for result := range resultChan {
		for code := range result {
			s.coupons[code] = struct{}{}
		}
	}
	s.mu.Unlock()

	return nil
}

// loadCouponsFromFile reads coupon codes from a single file
func (s *CouponStoreConcurrent) loadCouponsFromFile(filePath string) (map[string]struct{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", filePath, err)
	}
	defer file.Close()

	var reader *bufio.Reader
	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader for %s: %w", filePath, err)
		}
		defer gzReader.Close()
		reader = bufio.NewReader(gzReader)
	} else {
		reader = bufio.NewReader(file)
	}

	// Create a local map for this file's coupons
	coupons := make(map[string]struct{})

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			coupons[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return coupons, nil
}

// GetCoupon checks if a coupon code exists.
func (s *CouponStoreConcurrent) GetCoupon(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.coupons[code]
	return exists
}
