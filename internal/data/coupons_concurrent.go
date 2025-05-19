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

	"golang.org/x/sync/errgroup"
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
		loadDir = dir // Store the directory used for loading
		loadErr = instance.LoadAndFindValidCoupons(dir)
		if loadErr == nil {
			loaded = true
		}
	})

    if loaded && loadDir != dir {
		fmt.Printf("Warning: CouponStore already loaded with directory '%s'. Requested directory '%s' is different.\n", loadDir, dir)
	}

	return instance, loadErr
}

// extractAndSort reads coupon codes from a file, preprocesses them (trimming, filtering empty),
// and pipes them to the system's `sort` command, which writes to a sorted output file.
func extractAndSort(inputPath, tempDir string, fileIndex int) (string, error) {
	outputPath := filepath.Join(tempDir, fmt.Sprintf("sorted_coupons_%d.txt", fileIndex))

	// Setup the sort command to read from stdin and write to outputPath
	cmd := exec.Command("sort", "-o", outputPath)
	cmd.Env = append(os.Environ(), "LC_ALL=C") // Use C locale for faster sorting

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdin pipe for sort (%s): %w", inputPath, err)
	}

	// Start the sort command. It will block waiting for input on stdinPipe.
	if err := cmd.Start(); err != nil {
		stdinPipe.Close() // Close pipe if command fails to start
		return "", fmt.Errorf("error starting sort command for %s: %w", inputPath, err)
	}

	// Use an errgroup to manage the goroutine writing to stdin and waiting for cmd.
	// The group's context will be cancelled if any of its goroutines return an error.
	var eg errgroup.Group

	// Goroutine to open the input file, preprocess lines, and write to sort's stdin.
	eg.Go(func() error {
		// IMPORTANT: Ensure stdinPipe is closed when this goroutine finishes,
		// either successfully or due to an error. This signals EOF to the sort command.
		defer stdinPipe.Close()

		inFile, err := os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("error opening input file %s: %w", inputPath, err)
		}
		defer inFile.Close()

		var currentReader io.Reader = inFile
		if strings.HasSuffix(inputPath, ".gz") {
			gzReader, err := gzip.NewReader(inFile) // gzip.NewReader handles BOM if present
			if err != nil {
				return fmt.Errorf("error creating gzip reader for %s: %w", inputPath, err)
			}
			defer gzReader.Close() // This closes the gzip stream, inFile is closed by its own defer
			currentReader = gzReader
		}

		// Use bufio.Writer for efficient writes to the pipe.
		writer := bufio.NewWriter(stdinPipe)
		// Use bufio.Scanner for efficient reads from the file/gzReader.
		scanner := bufio.NewScanner(currentReader)
        // Increase buffer size for scanner if lines can be very long (default is 64KB)
        // const maxCapacity = 1024 * 1024 // 1 MB for example
        // buf := make([]byte, maxCapacity)
        // scanner.Buffer(buf, maxCapacity)


		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" { // Filter out empty lines after trimming
				if _, err := writer.WriteString(line + "\n"); err != nil {
					// This error typically occurs if the sort command has exited prematurely.
					return fmt.Errorf("error writing to sort stdin for %s: %w", inputPath, err)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error scanning input file %s: %w", inputPath, err)
		}

		// Crucial: Flush any buffered data to the pipe.
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("error flushing to sort stdin for %s: %w", inputPath, err)
		}

		return nil // Success for this goroutine
	})

	// Wait for the sort command to finish and capture its error.
	cmdErr := cmd.Wait()

	// Wait for the goroutine that writes to stdinPipe to finish.
	// This will return the error from that goroutine, if any.
	pipeErr := eg.Wait()

	// Prioritize error from the piping goroutine as it might cause cmdErr.
	if pipeErr != nil {
		// If cmdErr is not nil, it might be a consequence of pipeErr (e.g., broken pipe).
		// We log cmdErr for diagnostics but return pipeErr as the primary cause.
		if cmdErr != nil {
			// Consider logging cmdErr here if needed: fmt.Fprintf(os.Stderr, "Sort command also failed for %s: %v\n", inputPath, cmdErr)
		}
		return "", fmt.Errorf("error during preprocessing/piping for %s: %w", inputPath, pipeErr)
	}
	// If piping was successful, but sort command itself failed.
	if cmdErr != nil {
		return "", fmt.Errorf("sort command failed for %s: %w", inputPath, cmdErr)
	}

	return outputPath, nil
}


// findValidCoupons merges three sorted files and identifies coupons present in at least two of them.
// This function remains largely the same as it's already quite efficient for sorted inputs.
func findValidCoupons(sortedFile1, sortedFile2, sortedFile3 string) ([]string, error) {
    files := [3]*os.File{nil, nil, nil}
    readers := [3]*bufio.Reader{nil, nil, nil}
    scanners := [3]*bufio.Scanner{nil, nil, nil}
    currentCodes := [3]string{"", "", ""} // Store current code from each file
    hasMore := [3]bool{true, true, true} // Tracks if each file still has lines
    var setupErr error

    // Defer closing all files that are successfully opened
    defer func() {
        for i := 0; i < 3; i++ {
            if files[i] != nil {
                files[i].Close()
            }
        }
    }()

    // Open files and initialize scanners
    filePaths := [3]string{sortedFile1, sortedFile2, sortedFile3}
    for i := 0; i < 3; i++ {
        files[i], setupErr = os.Open(filePaths[i])
        if setupErr != nil {
            return nil, fmt.Errorf("error opening sorted file %s: %w", filePaths[i], setupErr)
        }
        readers[i] = bufio.NewReader(files[i])
        scanners[i] = bufio.NewScanner(readers[i])
        if scanners[i].Scan() {
            currentCodes[i] = strings.TrimSpace(scanners[i].Text()) // Already trimmed during sort prep, but good for safety
        } else {
            hasMore[i] = false // File is empty or only had error
            if err := scanners[i].Err(); err != nil {
                 return nil, fmt.Errorf("error reading initial line from sorted file %d (%s): %w", i+1, filePaths[i], err)
            }
        }
    }

    var validCoupons []string // Using dynamic slice; pre-allocation could be a micro-optimization if size is predictable

    for hasMore[0] || hasMore[1] || hasMore[2] { // Continue if at least one file has more lines
        minCode := ""
        // Determine the smallest current code among the files that still have lines
        for i := 0; i < 3; i++ {
            if hasMore[i] {
                if minCode == "" || currentCodes[i] < minCode {
                    minCode = currentCodes[i]
                }
            }
        }

        if minCode == "" { // Should only happen if all files are exhausted
            break
        }

        count := 0
        // Count occurrences of the smallest code
        for i := 0; i < 3; i++ {
            if hasMore[i] && currentCodes[i] == minCode {
                count++
            }
        }

        // If the smallest code appears in at least two files, it's valid
        if count >= 2 {
            validCoupons = append(validCoupons, minCode)
        }

        // Advance the scanners for all files that matched the smallest code
        for i := 0; i < 3; i++ {
            if hasMore[i] && currentCodes[i] == minCode {
                if scanners[i].Scan() {
                    currentCodes[i] = strings.TrimSpace(scanners[i].Text())
                } else {
                    hasMore[i] = false // No more lines in this file
                    if err := scanners[i].Err(); err != nil {
                        return nil, fmt.Errorf("error reading sorted file %d (%s): %w", i+1, filePaths[i], err)
                    }
                }
            }
        }
    }
    return validCoupons, nil
}


// LoadAndFindValidCoupons loads coupon codes from files, identifies valid ones, and populates the CouponStore.
func (s *CouponStoreConcurrent) LoadAndFindValidCoupons(dir string) error {
	// This check is now primarily handled by the CouponStoreConcurrentInstance logic with `once.Do`.
	// If this method were to be called directly multiple times for re-loading, this check would be important.
	// For the singleton pattern, `loaded` and `loadDir` are managed by `CouponStoreConcurrentInstance`.
	// However, resetting internal state if a direct reload is intended:
	// if loaded && loadDir == dir {
	// 	fmt.Println("CouponStore already loaded and validated from this directory.")
	// 	return nil
	// }

	s.mu.Lock() // Lock before modifying shared state like s.coupons
	s.coupons = make(map[string]struct{}) // Clear any previous coupons for a fresh load
	s.mu.Unlock()
	// 'loaded' and 'loadDir' global vars are set by CouponStoreConcurrentInstance after this function returns.


	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("coupon directory %s does not exist: %w", dir, err)
		}
		return fmt.Errorf("error accessing coupon directory %s: %w", dir, err)
	}

	filePaths, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return fmt.Errorf("error listing files in directory %s: %w", dir, err)
	}

    // Filter out directories from filePaths, only consider regular files
    var actualFiles []string
    for _, fp := range filePaths {
        info, err := os.Stat(fp)
        if err != nil {
            // Could log this error or decide how to handle unstat-able paths
            continue 
        }
        if info.Mode().IsRegular() {
            actualFiles = append(actualFiles, fp)
        }
    }
    filePaths = actualFiles


	if len(filePaths) != 3 {
		return fmt.Errorf("expected 3 coupon files in the directory '%s', found %d regular files", dir, len(filePaths))
	}

	tempDir, err := os.MkdirTemp("", "coupon_sort_") // Suffix for clarity
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	var sortedFiles [3]string
	var wg sync.WaitGroup
	errChan := make(chan error, len(filePaths)) // Buffer for one error per goroutine

	for i, filePath := range filePaths {
		wg.Add(1)
		go func(fp string, index int) {
			defer wg.Done()
			// Pass the actual file path (fp) to extractAndSort
			sortedPath, sortErr := extractAndSort(fp, tempDir, index+1)
			if sortErr != nil {
				errChan <- fmt.Errorf("failed to extract and sort file %s: %w", fp, sortErr)
				return
			}
			sortedFiles[index] = sortedPath
		}(filePath, i) // Pass filePath and i to the goroutine
	}

	wg.Wait()
	close(errChan) // Close errChan after all goroutines are done

	// Check for errors from goroutines
	for sortErr := range errChan {
		if sortErr != nil {
			return sortErr // Return the first error encountered
		}
	}

	validCoupons, err := findValidCoupons(sortedFiles[0], sortedFiles[1], sortedFiles[2])
	if err != nil {
		return fmt.Errorf("error finding valid coupons from sorted files: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, code := range validCoupons {
		s.coupons[code] = struct{}{}
	}
    // global `loaded` and `loadDir` will be set by the caller `CouponStoreConcurrentInstance`
	return nil
}

// GetCoupon checks if a coupon code exists.
func (s *CouponStoreConcurrent) GetCoupon(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.coupons[code]
	return exists
}

// Helper function to get the number of CPUs to use for parallel operations.
// Note: With the refactoring of `extractAndSort`, this function is not directly
// used in the coupon loading path as it was before (for the internal worker pool).
// It's kept here in case other parts of the package use it, or for future use.
// The number of concurrent `extractAndSort` operations is fixed at 3 (one per file).
func getNumCPU() int {
	numCPU := runtime.NumCPU()
	if numCPU > 8 { // limit the number of goroutines for CPU-bound tasks if it were used
		numCPU = 8
	}
    if numCPU <= 0 { // Ensure at least 1
        numCPU = 1
    }
	return numCPU
}