package data

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupConcurrentTestData(t *testing.T) (string, func(), func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "coupon-concurrent-test")
	require.NoError(t, err)

	// Create test files
	files := map[string]string{
		"coupons1.txt":    "COUPON1\nCOUPON2\nCOUPON3\n",
		"coupons2.txt":    "COUPON4\nCOUPON5\n",
		"coupons3.txt.gz": "COUPON6\nCOUPON7\nCOUPON8\n",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if filepath.Ext(name) == ".gz" {
			file, err := os.Create(path)
			require.NoError(t, err)

			gw := gzip.NewWriter(file)
			_, err = gw.Write([]byte(content))
			require.NoError(t, err)

			err = gw.Close()
			require.NoError(t, err)
			err = file.Close()
			require.NoError(t, err)
		} else {
			err := os.WriteFile(path, []byte(content), 0644)
			require.NoError(t, err)
		}
	}

	// Reset singleton state
	resetState := func() {
		instance = nil
		once = sync.Once{}
		loadErr = nil
		loadDir = ""
		loaded = false
	}

	// Cleanup everything including directory
	cleanup := func() {
		resetState()
		os.RemoveAll(tempDir)
	}

	return tempDir, resetState, cleanup
}

func TestCouponStoreConcurrent_Singleton(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coupon-concurrent-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "coupons.txt")
	err = os.WriteFile(testFile, []byte("TEST10\nTEST20\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("singleton_pattern", func(t *testing.T) {
		// Get first instance
		store1, err := Instance(tempDir)
		if err != nil {
			t.Fatalf("Failed to get first instance: %v", err)
		}

		// Get second instance with same directory
		store2, err := Instance(tempDir)
		if err != nil {
			t.Fatalf("Failed to get second instance: %v", err)
		}

		// Verify it's the same instance
		if store1 != store2 {
			t.Error("Expected same instance, got different instances")
		}

		// Verify coupons were loaded
		if !store1.GetCoupon("TEST10") {
			t.Error("Expected TEST10 coupon to be loaded")
		}

		// Create a new directory
		newDir := filepath.Join(tempDir, "new")
		err = os.MkdirAll(newDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create new directory: %v", err)
		}

		// Create a new test file in the new directory
		newTestFile := filepath.Join(newDir, "new_coupons.txt")
		err = os.WriteFile(newTestFile, []byte("TEST30\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Get instance with new directory
		store3, err := Instance(newDir)
		if err != nil {
			t.Fatalf("Failed to get instance with new directory: %v", err)
		}

		// Verify new coupons were loaded
		if !store3.GetCoupon("TEST30") {
			t.Error("Expected TEST30 coupon to be loaded")
		}

		// Verify old coupons were cleared
		if store3.GetCoupon("TEST10") {
			t.Error("Expected TEST10 coupon to be cleared")
		}

		// Try with a non-existent directory
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		_, err = Instance(nonExistentDir)
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})
}

func TestCouponStoreConcurrent_LoadCoupons(t *testing.T) {
	tempDir, resetState, cleanup := setupConcurrentTestData(t)
	defer cleanup()

	tests := []struct {
		name    string
		dir     string
		wantErr bool
	}{
		{
			name:    "valid directory",
			dir:     tempDir,
			wantErr: false,
		},
		{
			name:    "empty directory path",
			dir:     "",
			wantErr: true,
		},
		{
			name:    "non-existent directory",
			dir:     "/nonexistent/path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetState() // Only reset state, keep the directory

			store, err := Instance(tt.dir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, store)

			// Verify coupons were loaded
			if tt.dir == tempDir {
				expectedCoupons := []string{"COUPON1", "COUPON2", "COUPON3", "COUPON4", "COUPON5", "COUPON6", "COUPON7", "COUPON8"}
				for _, coupon := range expectedCoupons {
					assert.True(t, store.GetCoupon(coupon), "Expected coupon %s to exist", coupon)
				}
			}
		})
	}
}

func TestCouponStoreConcurrent_ConcurrentAccess(t *testing.T) {
	tempDir, resetState, cleanup := setupConcurrentTestData(t)
	defer cleanup()

	t.Run("concurrent reads", func(t *testing.T) {
		resetState() // Only reset state, keep the directory

		// Initialize store
		store, err := Instance(tempDir)
		require.NoError(t, err)

		var wg sync.WaitGroup
		numGoroutines := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.True(t, store.GetCoupon("COUPON1"))
				assert.False(t, store.GetCoupon("NONEXISTENT"))
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent initialization", func(t *testing.T) {
		resetState() // Only reset state, keep the directory

		var wg sync.WaitGroup
		numGoroutines := 10
		stores := make([]*CouponStoreConcurrent, numGoroutines)
		errors := make([]error, numGoroutines)

		// Launch multiple goroutines to initialize the store concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				store, err := Instance(tempDir)
				stores[index] = store
				errors[index] = err
			}(i)
		}

		wg.Wait()

		// Verify all instances are the same and no errors occurred
		for i := 1; i < numGoroutines; i++ {
			assert.NoError(t, errors[i])
			assert.Same(t, stores[0], stores[i], "Expected same instance for all goroutines")
		}
	})
}

func TestCouponStoreConcurrent_LoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir, resetState, cleanup := setupConcurrentTestData(t)
	defer cleanup()

	resetState() // Only reset state, keep the directory

	// Create many test files
	numFiles := 50
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf("COUPON%d\n", i)
		path := filepath.Join(tempDir, fmt.Sprintf("coupons%d.txt", i))
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Measure load time
	start := time.Now()
	store, err := Instance(tempDir)
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, store)

	t.Logf("Loaded %d files in %v", numFiles, duration)
	assert.Less(t, duration, 5*time.Second, "Loading should complete within reasonable time")
}

func TestCouponStoreConcurrent_EdgeCases(t *testing.T) {
	tempDir, resetState, cleanup := setupConcurrentTestData(t)
	defer cleanup()

	t.Run("empty files", func(t *testing.T) {
		resetState() // Only reset state, keep the directory

		// Create empty files
		emptyDir := filepath.Join(tempDir, "empty")
		require.NoError(t, os.MkdirAll(emptyDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(emptyDir, "empty.txt"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(emptyDir, "empty.txt.gz"), []byte(""), 0644))

		store, err := Instance(emptyDir)
		require.NoError(t, err)
		require.NotNil(t, store)
	})

	t.Run("corrupted gzip", func(t *testing.T) {
		resetState() // Only reset state, keep the directory

		// Create corrupted gzip file
		corruptDir := filepath.Join(tempDir, "corrupt")
		require.NoError(t, os.MkdirAll(corruptDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(corruptDir, "corrupt.txt.gz"), []byte("not a gzip file"), 0644))

		store, err := Instance(corruptDir)
		require.NoError(t, err) // Should not fail completely due to one bad file
		require.NotNil(t, store)
	})

	t.Run("mixed valid and invalid files", func(t *testing.T) {
		resetState() // Only reset state, keep the directory

		mixedDir := filepath.Join(tempDir, "mixed")
		require.NoError(t, os.MkdirAll(mixedDir, 0755))

		// Create valid file
		require.NoError(t, os.WriteFile(filepath.Join(mixedDir, "valid.txt"), []byte("VALIDCOUPON\n"), 0644))

		// Create unreadable file
		unreadableFile := filepath.Join(mixedDir, "unreadable.txt")
		require.NoError(t, os.WriteFile(unreadableFile, []byte("UNREADABLE\n"), 0000))

		store, err := Instance(mixedDir)
		require.NoError(t, err) // Should not fail completely
		require.NotNil(t, store)

		// Should have loaded the valid coupon
		assert.True(t, store.GetCoupon("VALIDCOUPON"))
	})
}
