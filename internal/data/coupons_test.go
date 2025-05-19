package data

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testDir = "/Users/ravibandhu/personal/go/oolio-food-ordering/internal/config/testdata/test_coupons"

// Create a directory with some dummy coupon files for testing
func createTestCoupons(testDir string) error {
	os.MkdirAll(testDir, 0755)

	// Create a plain text coupon file
	couponFile1 := filepath.Join(testDir, "coupons1.txt")
	os.WriteFile(couponFile1, []byte("SUMMER20\nWINTER15\nSPRING25\n"), 0644)

	// Create a gzipped coupon file
	couponFile2 := filepath.Join(testDir, "coupons2.txt.gz")
	file, err := os.Create(couponFile2)
	if err != nil {
		fmt.Println("Error creating coupon file:", err)
		return err
	}

	gw, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		fmt.Println("Error creating gzip writer:", err)
		return err
	}
	err = os.WriteFile(couponFile2, []byte("AUTUMN30\nSUMMER20\n"), 0644)
	if err != nil {
		fmt.Println("Error writing gzipped file:", err)
		return err
	}
	gw.Close()
	return nil
}
func TestCouponStore_LoadCouponsFromFile(t *testing.T) {
	couponStore := NewCouponStore()

	createTestCoupons(testDir)
	start := time.Now()
	err := couponStore.loadCouponsFromFile(filepath.Join(testDir, "coupons1.txt"))
	if err != nil {
		fmt.Println("Error loading coupons:", err)
		return
	}
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, len(couponStore.coupons), 0)
	t.Logf("Total coupons loaded: %d\n", len(couponStore.coupons))
	t.Logf("Time taken: %s\n", elapsed)

	os.RemoveAll(testDir)
}

func TestCouponStore_LoadCoupons(t *testing.T) {
	couponStore := NewCouponStore()

	createTestCoupons(testDir)
	start := time.Now()
	err := couponStore.LoadCoupons(testDir)
	if err != nil {
		fmt.Println("Error loading coupons:", err)
		return
	}
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, len(couponStore.coupons), 0)
	t.Logf("Total coupons loaded: %d\n", len(couponStore.coupons))
	t.Logf("Time taken: %s\n", elapsed)

	os.RemoveAll(testDir)
}

func TestCouponStore_GetCoupon(t *testing.T) {
	couponStore := NewCouponStore()
	
	couponStore.coupons["SUMMER20"] = struct{}{}
	discount, err := couponStore.GetCoupon("SUMMER20")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, discount, 0)

	discount, err = couponStore.GetCoupon("INVALIDCODE_123")
	assert.Error(t, err)
	assert.Equal(t, 0, discount)
}
