package data

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to create three gzipped coupon files with different coupon sets in a temporary directory.
func createTestCouponFiles(t *testing.T) (string, func()) {
	t.Helper() // Mark this as a helper function for better error reporting

	// Create a temporary directory
	testDir, err := os.MkdirTemp("", "coupon_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Helper function to create a gzipped file
	createGzipFile := func(filename string, coupons []string) {
		filePath := filepath.Join(testDir, filename)
		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
		defer file.Close()

		gw, err := gzip.NewWriterLevel(file, gzip.BestCompression)
		if err != nil {
			t.Fatalf("Failed to create gzip writer for %s: %v", filePath, err)
		}
		defer gw.Close()

		_, err = gw.Write([]byte(strings.Join(coupons, "\n")))
		if err != nil {
			t.Fatalf("Failed to write to gzipped file %s: %v", filePath, err)
		}
	}

	// Create three gzipped files with different coupon sets
	createGzipFile("coupons1.txt.gz", []string{"COUPONA1", "COUPONA2", "COUPONA3", "COMMONA1"})
	createGzipFile("coupons2.txt.gz", []string{"COUPONA2", "COUPONA4", "COUPONA5", "COMMONA1", "COMMONA2"})
	createGzipFile("coupons3.txt.gz", []string{"COUPONA3", "COUPONA5", "COUPONA6", "COMMONA1", "COMMONA2", "COMMONA3"})

	// Return the directory path and a cleanup function
	cleanup := func() {
		os.RemoveAll(testDir) // Remove the temporary directory and its contents
	}
	return testDir, cleanup
}

func TestCouponStore_LoadAndFindValidCoupons(t *testing.T) {
	// Create test files and get the directory
	testDir, cleanup := createTestCouponFiles(t)
	defer cleanup() // Clean up the files and directory when the test finishes

	// Initialize CouponStore using the Instance method (Singleton)
	store, err := CouponStoreConcurrentInstance(testDir)
	if err != nil {
		t.Fatalf("Failed to get CouponStoreConcurrent instance: %v", err)
	}

	// Test cases
	testCases := []struct {
		name          string
		couponCode    string
		expectedValid bool
	}{
		{
			name:          "Valid Coupon (Present in 2 files - COMMONA2)",
			couponCode:    "COMMONA2",
			expectedValid: true,
		},
		{
			name:          "Valid Coupon (Present in 2 files - COUPONA2)",
			couponCode:    "COUPONA2",
			expectedValid: true,
		},
		{
			name:          "Invalid Coupon (Present in only 1 file - COUPONA1)",
			couponCode:    "COUPONA1",
			expectedValid: false,
		},
		{
			name:          "Invalid Coupon (Not present in any file)",
			couponCode:    "INVALID",
			expectedValid: false,
		},
		{
			name:          "Valid Coupon (Present in 2 files - COUPONA3)",
			couponCode:    "COUPONA3",
			expectedValid: true,
		},
		{
			name:          "Valid Coupon (Present in 2 files - COUPONA5)",
			couponCode:    "COUPONA5",
			expectedValid: true,
		},
		{
			name:          "Invalid Coupon (Present in only 1 file - COUPONA6)",
			couponCode:    "COUPONA6",
			expectedValid: false,
		},
		{
			name:          "Valid Coupon (Present in 3 files - COMMONA1)",
			couponCode:    "COMMONA1",
			expectedValid: true,
		},
		{
			name:          "Invalid Coupon (Present in 1 files - COMMONA3)",
			couponCode:    "COMMONA3",
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := store.GetCoupon(tc.couponCode)
			if isValid != tc.expectedValid {
				t.Errorf("GetCoupon(%q) should return %v, but got %v", tc.couponCode, tc.expectedValid, isValid)
			}
		})
	}

	// Test with an empty directory.  This should not cause a panic.
	emptyDir, err := os.MkdirTemp("", "empty_coupons")
	if err != nil {
		t.Fatalf("Failed to create empty test directory: %v", err)
	}
	defer os.RemoveAll(emptyDir)

	store, err = CouponStoreConcurrentInstance(emptyDir) // re-use the instance, singleton
	if err != nil {
		t.Fatalf("Failed to get CouponStoreConcurrent instance for empty dir: %v", err)
	}

	isValid1 := store.GetCoupon("COUPONA1")
	if isValid1 {
		t.Errorf("GetCoupon should return true for valid coupon in mixed dir")
	}
	isValid2 := store.GetCoupon("NONEXIST")
	if isValid2 {
		t.Errorf("GetCoupon should return false for non-existent coupon in mixed dir")
	}
}
