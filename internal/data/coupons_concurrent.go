package data

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"hash/fnv" // For a simple string hashing for sharding
	"io"
	"math/bits" // For bits.OnesCount32
	"os"
	"path/filepath"
	"runtime" // For runtime.NumCPU()
	"strings"
	"sync"
	// "sync/atomic" // No longer needed for sharedBitmaskMap values
	"time"
)

// CouponStoreConcurrent struct remains the same
type CouponStoreConcurrent struct {
	coupons map[string]struct{}
	mu      sync.RWMutex
}

// Singleton variables remain the same
var (
	once     sync.Once
	instance *CouponStoreConcurrent
	loadErr  error
	loadDir  string
	loaded   bool
)

// NewCouponStoreConcurrent and CouponStoreConcurrentInstance remain the same
func NewCouponStoreConcurrent() *CouponStoreConcurrent {
	return &CouponStoreConcurrent{
		coupons: make(map[string]struct{}),
	}
}

func CouponStoreConcurrentInstance(dir string) (*CouponStoreConcurrent, error) {
	once.Do(func() {
		instance = NewCouponStoreConcurrent()
		loadDir = dir
		loadErr = instance.LoadAndFindValidCoupons(dir)
		if loadErr == nil {
			loaded = true
		}
	})
	if loaded && loadDir != dir {
		fmt.Printf("[%s] Warning: CouponStore already loaded with directory '%s'. Requested directory '%s' is different. Returning existing instance.\n", time.Now().Format(time.RFC3339Nano), loadDir, dir)
	}
	return instance, loadErr
}

type couponData struct {
	couponString string
	fileBitmask  uint32
}

// --- Sharded Map Implementation ---
const numShards = 256 // Tunable. Power of 2 can be good for bitwise modulo.

type Shard struct {
	mu sync.Mutex
	m  map[string]uint32 // Stores uint32 directly for bitmasks
}

// Shards array for the globally shared bitmask data
var couponShards [numShards]Shard

// Initialize shards (call this once before workers start)
func initializeShards() {
	for i := range couponShards {
		couponShards[i].m = make(map[string]uint32)
	}
}

// getShardIndex calculates the shard for a given coupon string.
// Using FNV-1a hash, common and simple.
func getShardIndex(couponStr string) uint32 {
	hasher := fnv.New32a()
	hasher.Write([]byte(couponStr)) // This allocates a byte slice from string for Write.
	                               // For extreme performance, a non-allocating hash or maphash could be used.
	return hasher.Sum32() % numShards
}

// flushBatchSharded merges a worker's local batch into the sharded global map.
func flushBatchSharded(workerID int, localBatch map[string]uint32, sds []Shard) { // sds is couponShards
	if len(localBatch) == 0 {
		return
	}
	// fmt.Printf("[%s] Worker %d: Flushing batch of %d unique coupon strings to sharded map.\n", time.Now().Format(time.RFC3339Nano), workerID, len(localBatch))
	// startFlush := time.Now()

	for couponStr, batchAggregatedBitmask := range localBatch {
		if batchAggregatedBitmask == 0 {
			continue
		}
		shardIndex := getShardIndex(couponStr)

		sds[shardIndex].mu.Lock()
		sds[shardIndex].m[couponStr] |= batchAggregatedBitmask // Bitwise OR under shard lock
		sds[shardIndex].mu.Unlock()
	}
	// flushDuration := time.Since(startFlush)
	// if flushDuration.Milliseconds() > 100 {
	// 	fmt.Printf("[%s] Worker %d: Sharded batch flush of %d items took %s.\n", time.Now().Format(time.RFC3339Nano), workerID, len(localBatch), flushDuration)
	// }
}

// worker function for the worker pool using sharded map
func workerSharded(workerID int, assumeCleanLines bool, dataChan <-chan couponData, sds []Shard, wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Printf("[%s] Worker %d (sharded): Started.\n", time.Now().Format(time.RFC3339Nano), workerID)

	localBatchData := make(map[string]uint32)
	itemsProcessedForCurrentBatch := 0
	const flushTriggerCount = 8192 // Tunable

	for data := range dataChan {
		couponStr := data.couponString
		if !assumeCleanLines {
			couponStr = strings.TrimSpace(couponStr)
		}

		couponLen := len(couponStr)
		if couponLen >= 8 && couponLen <= 10 {
			localBatchData[couponStr] |= data.fileBitmask
		}

		itemsProcessedForCurrentBatch++
		if itemsProcessedForCurrentBatch >= flushTriggerCount {
			flushBatchSharded(workerID, localBatchData, sds) // Pass shards slice
			localBatchData = make(map[string]uint32)
			itemsProcessedForCurrentBatch = 0
		}
	}

	if len(localBatchData) > 0 {
		// fmt.Printf("[%s] Worker %d (sharded): Flushing final local batch of %d items.\n", time.Now().Format(time.RFC3339Nano), workerID, len(localBatchData))
		flushBatchSharded(workerID, localBatchData, sds) // Pass shards slice
	}
	// fmt.Printf("[%s] Worker %d (sharded): Exiting.\n", time.Now().Format(time.RFC3339Nano), workerID)
}


// LoadAndFindValidCoupons processes coupon files.
func (s *CouponStoreConcurrent) LoadAndFindValidCoupons(dir string) (errFinal error) {
	startTime := time.Now()
	fmt.Printf("[%s] LoadAndFindValidCoupons: Initiating for directory '%s' (using sharded map).\n", startTime.Format(time.RFC3339Nano), dir)
	defer func() { /* ... (same defer for timing and panic recovery as before) ... */ 
		duration := time.Since(startTime)
		if r := recover(); r != nil {
			errFinal = fmt.Errorf("recovered panic in LoadAndFindValidCoupons: %v", r)
			fmt.Printf("[%s] LoadAndFindValidCoupons: CRITICAL PANIC after %s - %v\n", time.Now().Format(time.RFC3339Nano), duration, r)
		}
		if errFinal != nil {
			fmt.Printf("[%s] LoadAndFindValidCoupons: FAILED after %s. Error: %v\n", time.Now().Format(time.RFC3339Nano), duration, errFinal)
		} else {
			fmt.Printf("[%s] LoadAndFindValidCoupons: Successfully completed in %s.\n", time.Now().Format(time.RFC3339Nano), duration)
		}
	}()


	s.mu.Lock()
	s.coupons = make(map[string]struct{})
	s.mu.Unlock()

	// Initialize shards (do this once per application run, or ensure it's safe if called multiple times for tests)
	// For simplicity in this function, we initialize it here. If LoadAndFindValidCoupons is called multiple times
	// by different tests without resetting package state, this could be an issue. The singleton `once.Do`
	// ensures LoadAndFindValidCoupons itself is called once for the instance.
	initializeShards() // Ensure shard maps are created

	// ... (file path globbing, validation, etc. as before) ...
	if _, statErr := os.Stat(dir); statErr != nil {
		if os.IsNotExist(statErr) {return fmt.Errorf("coupon directory '%s' does not exist: %w", dir, statErr)}
		return fmt.Errorf("error accessing coupon directory '%s': %w", dir, statErr)
	}
	globPaths, globErr := filepath.Glob(filepath.Join(dir, "*"))
	if globErr != nil {return fmt.Errorf("error listing files in directory '%s': %w", dir, globErr)}
	var filePaths []string
	for _, fp := range globPaths {
		info, statErr := os.Stat(fp)
		if statErr != nil {
			fmt.Fprintf(os.Stderr, "[%s] Warning: Could not stat path '%s', skipping: %v\n", time.Now().Format(time.RFC3339Nano), fp, statErr)
			continue
		}
		if info.Mode().IsRegular() {filePaths = append(filePaths, fp)}
	}
	if len(filePaths) != 3 {
		return fmt.Errorf("expected 3 coupon files in directory '%s', found %d regular files: %v", dir, len(filePaths), filePaths)
	}
	fmt.Printf("[%s] LoadAndFindValidCoupons: Found %d files to process: %v\n", time.Now().Format(time.RFC3339Nano), len(filePaths), filePaths)


	dataChan := make(chan couponData, 2048*len(filePaths)) // Increased buffer slightly
	var readerWg sync.WaitGroup
	readerErrChan := make(chan error, len(filePaths))
	assumeCleanLines := true

	fmt.Printf("[%s] LoadAndFindValidCoupons: Starting %d file reader goroutines (assumeCleanLines=%t)...\n", time.Now().Format(time.RFC3339Nano), len(filePaths), assumeCleanLines)
	for i, filePath := range filePaths {
		readerWg.Add(1)
		go func(fp string, fileIndex int, readerLogIndex int) { // File reader goroutine (same as before)
			defer readerWg.Done()
			readerStartTime := time.Now()
			fileBitmask := uint32(1 << fileIndex)
			inFile, fileOpenErr := os.Open(fp)
			if fileOpenErr != nil {
				errMsg := fmt.Errorf("reader %d failed to open file '%s': %w", readerLogIndex, fp, fileOpenErr)
				fmt.Fprintln(os.Stderr, "["+time.Now().Format(time.RFC3339Nano)+"] "+errMsg.Error())
				readerErrChan <- errMsg
				return
			}
			defer inFile.Close()
			var currentReader io.Reader = inFile
			if strings.HasSuffix(strings.ToLower(fp), ".gz") {
				gzReader, gzErr := gzip.NewReader(inFile)
				if gzErr != nil {
					errMsg := fmt.Errorf("reader %d failed to create gzip reader for '%s': %w", readerLogIndex, fp, gzErr)
					fmt.Fprintln(os.Stderr, "["+time.Now().Format(time.RFC3339Nano)+"] "+errMsg.Error())
					readerErrChan <- errMsg
					return
				}
				defer gzReader.Close()
				currentReader = gzReader
			}
			scanner := bufio.NewScanner(currentReader)
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				dataChan <- couponData{couponString: scanner.Text(), fileBitmask: fileBitmask}
			}
			if scanErr := scanner.Err(); scanErr != nil {
				fmt.Fprintf(os.Stderr, "[%s] Reader %d (%s): Error during scan (at line ~%d): %v\n", time.Now().Format(time.RFC3339Nano), readerLogIndex, filepath.Base(fp), lineNum, scanErr)
			}
			fmt.Printf("[%s] Reader %d (%s): Finished. Processed %d lines in %s.\n", time.Now().Format(time.RFC3339Nano), readerLogIndex, filepath.Base(fp), lineNum, time.Since(readerStartTime))
		}(filePath, i, i+1)
	}

	go func() { // Goroutine to close channels once readers are done
		readerWg.Wait()
		close(dataChan)
		close(readerErrChan)
		fmt.Printf("[%s] LoadAndFindValidCoupons: All file readers completed. dataChan and readerErrChan closed.\n", time.Now().Format(time.RFC3339Nano))
	}()

	var workerWg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	if numWorkers < 2 && runtime.GOMAXPROCS(0) > 1 { numWorkers = 2 } else if numWorkers < 1 { numWorkers = 1 }
	// if numWorkers > 8 { numWorkers = 8 } // Example cap on workers

	fmt.Printf("[%s] LoadAndFindValidCoupons: Starting %d worker goroutines (batch flush trigger: %d items)...\n", time.Now().Format(time.RFC3339Nano), numWorkers, 8192) // 8192 is flushTriggerCount from worker
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go workerSharded(i+1, assumeCleanLines, dataChan, couponShards[:], &workerWg) // Pass slice of shards
	}

	workerWg.Wait()
	fmt.Printf("[%s] LoadAndFindValidCoupons: All worker goroutines completed.\n", time.Now().Format(time.RFC3339Nano))

	fmt.Printf("[%s] LoadAndFindValidCoupons: Checking for critical errors from file readers...\n", time.Now().Format(time.RFC3339Nano))
	for errFromReader := range readerErrChan {
		if errFromReader != nil {
			return fmt.Errorf("critical error during file reading phase: %w", errFromReader)
		}
	}
	fmt.Printf("[%s] LoadAndFindValidCoupons: No critical reader errors found.\n", time.Now().Format(time.RFC3339Nano))

	s.mu.Lock() // Lock for final write to s.coupons
	defer s.mu.Unlock()
	finalCouponCount := 0
	globallyUniqueCouponCount := 0
	fmt.Printf("[%s] LoadAndFindValidCoupons: Populating final coupon store from sharded map (%d shards)...\n", time.Now().Format(time.RFC3339Nano), numShards)
	
	iterationStartTime := time.Now()
	for i := 0; i < numShards; i++ {
		couponShards[i].mu.Lock() // Lock each shard for reading its map
		for coupon, mask := range couponShards[i].m {
			globallyUniqueCouponCount++ // This will count some coupons multiple times if not careful;
			                            // better to count unique keys only once globally.
			                            // For now, this counts total entries across all shard maps.
			if bits.OnesCount32(mask) >= 2 {
				s.coupons[coupon] = struct{}{}
				// finalCouponCount++ // This is correctly incremented below from len(s.coupons)
			}
		}
		couponShards[i].mu.Unlock()
	}
	finalCouponCount = len(s.coupons) // Get the accurate count after populating
	// The globallyUniqueCouponCount calculated above by summing len(shard.m) is more accurate.
	// Let's refine globallyUniqueCouponCount calculation after the loop.
	// Actually, we can just sum len(shards[i].m) to get an idea of total items stored in shards.
	var totalItemsInShards int
	for i := 0; i < numShards; i++ {
		couponShards[i].mu.Lock()
		totalItemsInShards += len(couponShards[i].m)
		couponShards[i].mu.Unlock()
	}

	fmt.Printf("[%s] LoadAndFindValidCoupons: Iterated sharded map (approx. %d total items) in %s.\n", time.Now().Format(time.RFC3339Nano), totalItemsInShards, time.Since(iterationStartTime))
	fmt.Printf("[%s] LoadAndFindValidCoupons: Stored %d valid coupons.\n", time.Now().Format(time.RFC3339Nano), finalCouponCount)
	return nil
}

// GetCoupon method remains the same
func (s *CouponStoreConcurrent) GetCoupon(code string) bool {
	codeLen := len(code)
	if codeLen < 8 || codeLen > 10 {return false}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.coupons[code]
	return exists
}