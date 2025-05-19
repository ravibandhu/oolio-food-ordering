package data

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// CouponStoreConcurrent struct to hold the loaded coupon codes.
type CouponStoreConcurrent struct {
	coupons map[string]struct{} // Set-like for efficient lookups
	mu      sync.RWMutex      // Mutex for concurrent access
}

var (
	once     sync.Once
	instance *CouponStoreConcurrent
	loadErr  error
	loadDir  string
	loaded   bool
)

// NewCouponStoreConcurrent creates and initializes a new CouponStoreConcurrent (private).
func NewCouponStoreConcurrent() *CouponStoreConcurrent {
	return &CouponStoreConcurrent{
		coupons: make(map[string]struct{}),
	}
}

// CouponStoreConcurrentInstance returns the singleton instance of CouponStoreConcurrent, loading if not already loaded.
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
		fmt.Println("Warning: CouponStore already loaded with a different directory.")
	}

	return instance, loadErr
}

// sortFile performs external sorting on a file using the system's `sort` command.
func sortFile(inputPath, outputPath string) error {
	cmd := exec.Command("sort", inputPath, "-o", outputPath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error sorting %s: %w", inputPath, err)
	}
	return nil
}

// extractAndSort reads coupon codes from a file, writes them to a temporary file, sorts it, and returns the sorted file path.
func extractAndSort(inputPath, tempDir string, fileIndex int) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("sorted_coupons_%d.txt", fileIndex))
	tempExtractPath := filepath.Join(tempDir, fmt.Sprintf("extracted_coupons_%d.txt", fileIndex))

	inFile, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("error opening input file %s: %w", inputPath, err)
	}
	defer inFile.Close()

	outFile, err := os.Create(tempExtractPath)
	if err != nil {
		return "", fmt.Errorf("error creating temporary file %s: %w", tempExtractPath, err)
	}
	defer outFile.Close()

	var reader *bufio.Reader
	var readCloser io.ReadCloser // to hold either the file or the gzip reader
	readCloser = inFile          // default to the file
	defer readCloser.Close()    // ensure closing

	if strings.HasSuffix(inputPath, ".gz") {
		gzReader, err := gzip.NewReader(inFile)
		if err != nil {
			return "", fmt.Errorf("error creating gzip reader for %s: %w", inputPath, err)
		}
		readCloser = gzReader // replace with gzip reader
		reader = bufio.NewReader(gzReader)
	} else {
		reader = bufio.NewReader(inFile)
	}

	// Determine the number of goroutines to use
	numCPU := getNumCPU()
	var wg sync.WaitGroup
	errChan := make(chan error, numCPU) // Channel for errors from goroutines
	linesChan := make(chan string, 1024)  // Buffered channel for sending lines

	// Function to process a chunk of lines
	processLines := func() {
		defer wg.Done()
		for line := range linesChan {
			if strings.TrimSpace(line) != "" {
				_, err := outFile.WriteString(line + "\n")
				if err != nil {
					errChan <- fmt.Errorf("error writing to temporary file %s: %w", tempExtractPath, err)
					return // Exit the goroutine on error
				}
			}
		}
	}

	// Start worker goroutines
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go processLines()
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		linesChan <- line // Send line to the channel
	}
	close(linesChan) // Close the channel after all lines are sent

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input file %s: %w", inputPath, err)
	}

	wg.Wait() // Wait for all goroutines to finish

	// Check for errors from goroutines
	select {
	case err := <-errChan:
		return "", err // Return the first error encountered
	default:
		// No error
	}

	if err := sortFile(tempExtractPath, outputPath); err != nil {
		return "", err
	}

	err = os.Remove(tempExtractPath) // Clean up the intermediate file.
	if err != nil {
		return "", fmt.Errorf("error removing temporary file %s: %w", tempExtractPath, err)
	}

	return outputPath, nil
}

// findValidCoupons merges three sorted files and identifies coupons present in at least two of them.
func findValidCoupons(sortedFile1, sortedFile2, sortedFile3 string) ([]string, error) {
	files := [3]*os.File{nil, nil, nil}
	readers := [3]*bufio.Reader{nil, nil, nil}
	scanners := [3]*bufio.Scanner{nil, nil, nil}
	currentCodes := [3]string{"", "", ""}
	hasMore := [3]bool{true, true, true}
	err := error(nil)

	// Open files and create readers/scanners
	files[0], err = os.Open(sortedFile1)
	if err != nil {
		return nil, fmt.Errorf("error opening sorted file %s: %w", sortedFile1, err)
	}
	defer files[0].Close()
	readers[0] = bufio.NewReader(files[0])
	scanners[0] = bufio.NewScanner(readers[0])
	hasMore[0] = scanners[0].Scan()
	if hasMore[0] {
		currentCodes[0] = strings.TrimSpace(scanners[0].Text())
	}

	files[1], err = os.Open(sortedFile2)
	if err != nil {
		return nil, fmt.Errorf("error opening sorted file %s: %w", sortedFile2, err)
	}
	defer files[1].Close()
	readers[1] = bufio.NewReader(files[1])
	scanners[1] = bufio.NewScanner(readers[1])
	hasMore[1] = scanners[1].Scan()
	if hasMore[1] {
		currentCodes[1] = strings.TrimSpace(scanners[1].Text())
	}

	files[2], err = os.Open(sortedFile3)
	if err != nil {
		return nil, fmt.Errorf("error opening sorted file %s: %w", sortedFile3, err)
	}
	defer files[2].Close()
	readers[2] = bufio.NewReader(files[2])
	scanners[2] = bufio.NewScanner(readers[2])
	hasMore[2] = scanners[2].Scan()
	if hasMore[2] {
		currentCodes[2] = strings.TrimSpace(scanners[2].Text())
	}

	validCoupons := make([]string, 0)
	for hasMore[0] || hasMore[1] || hasMore[2] {
		// Determine the smallest current code
		minCode := ""
		for i := 0; i < 3; i++ {
			if hasMore[i] && (minCode == "" || currentCodes[i] < minCode) {
				minCode = currentCodes[i]
			}
		}

		// Count occurrences of the smallest code
		count := 0
		for i := 0; i < 3; i++ {
			if hasMore[i] && currentCodes[i] == minCode {
				count++
			}
		}

		// If the smallest code appears in at least two files, it's valid
		if count >= 2 {
			validCoupons = append(validCoupons, minCode)
		}

		// Advance the readers that match the smallest code
		for i := 0; i < 3; i++ {
			if hasMore[i] && currentCodes[i] == minCode {
				hasMore[i] = scanners[i].Scan()
				if hasMore[i] {
					currentCodes[i] = strings.TrimSpace(scanners[i].Text())
				} else {
					currentCodes[i] = ""
				}
			}
		}

		if minCode == "" {
			break // All files exhausted
		}
	}

	// Check for scanner errors
	for i := 0; i < 3; i++ {
		if err := scanners[i].Err(); err != nil && err != io.EOF {
			return nil, fmt.Errorf("error reading sorted file %d: %w", i+1, err)
		}
	}
	return validCoupons, nil
}

// LoadAndFindValidCoupons loads coupon codes from files, identifies valid ones, and populates the CouponStore.
func (s *CouponStoreConcurrent) LoadAndFindValidCoupons(dir string) error {
	if loaded && loadDir == dir {
		fmt.Println("CouponStore already loaded and validated from this directory.")
		return nil
	}
	s.coupons = make(map[string]struct{})
	loaded = false
	loadDir = ""

	// Check if directory exists and is accessible
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("error accessing coupon directory %s: %w", dir, err)
	}

	filePaths, err := filepath.Glob(filepath.Join(dir, "*")) // Get all files in the directory
	if err != nil {
		return fmt.Errorf("error listing files in directory %s: %w", dir, err)
	}

	if len(filePaths) != 3 {
		return fmt.Errorf("expected 3 coupon files in the directory, found %d", len(filePaths))
	}

	tempDir, err := os.MkdirTemp("", "coupon_sort")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up the temporary directory

	var sortedFiles [3]string
	var wg sync.WaitGroup
	errChan := make(chan error, 3) // Channel for errors from goroutines

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(fp string, index int) {
			defer wg.Done()
			sortedPath, sortErr := extractAndSort(fp, tempDir, index+1)
			if sortErr != nil {
				errChan <- sortErr
				return
			}
			sortedFiles[index] = sortedPath
		}(filePath, i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors from goroutines
	for err := range errChan {
		if err != nil {
			return err // Return the first error encountered
		}
	}

	validCoupons, err := findValidCoupons(sortedFiles[0], sortedFiles[1], sortedFiles[2])
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, code := range validCoupons {
		s.coupons[code] = struct{}{}
	}

	loaded = true
	loadDir = dir
	return nil
}

// GetCoupon checks if a coupon code exists.
func (s *CouponStoreConcurrent) GetCoupon(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.coupons[code]
	return exists
}

// Helper function to get the number of CPUs to use for parallel operations
func getNumCPU() int {
	numCPU := runtime.NumCPU()
	if numCPU > 8 { //limit the number of goroutines
		numCPU = 8
	}
	return numCPU
}
