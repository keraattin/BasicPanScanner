// Package scanner - Worker pool for concurrent scanning
// This file implements the concurrent scanning using goroutines and channels
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WorkerPool manages a pool of worker goroutines for concurrent file scanning
// This provides better performance on multi-core systems
type WorkerPool struct {
	numWorkers int     // Number of worker goroutines
	config     *Config // Scanner configuration
}

// NewWorkerPool creates a new worker pool
//
// Parameters:
//   - numWorkers: Number of concurrent workers (typically CPU cores / 2)
//   - config: Scanner configuration
//
// Returns:
//   - *WorkerPool: Configured worker pool
//
// Example:
//
//	pool := NewWorkerPool(4, scannerConfig)
//	result, err := pool.ScanDirectory("/var/log")
func NewWorkerPool(numWorkers int, config *Config) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		config:     config,
	}
}

// fileJob represents a file that needs to be scanned
// This is passed through channels to workers
type fileJob struct {
	path string      // Full path to the file
	info os.FileInfo // File information (size, etc.)
}

// scanResult represents the result of scanning a single file
// Workers send these back through a results channel
type scanResult struct {
	path     string    // File path
	findings []Finding // Cards found in this file
	err      error     // Error if scanning failed
}

// ScanDirectory scans a directory using concurrent workers
// This is the concurrent implementation of directory scanning
//
// Process:
//  1. Collect all files to scan
//  2. Start worker goroutines
//  3. Distribute files to workers via channel
//  4. Workers scan files and send results back
//  5. Aggregate all results
//
// Parameters:
//   - dirPath: Directory to scan
//
// Returns:
//   - *ScanResult: Complete scan results
//   - error: Error if directory can't be accessed
func (wp *WorkerPool) ScanDirectory(dirPath string) (*ScanResult, error) {
	startTime := time.Now()

	// Result accumulator
	result := &ScanResult{
		GroupedByFile: make(map[string][]Finding),
	}

	// Mutex to protect result from concurrent access
	// Multiple goroutines will update result, so we need synchronization
	var mu sync.Mutex

	// ============================================================
	// PHASE 1: Collect files to scan
	// ============================================================

	var filesToScan []fileJob

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Handle directories
		if info.IsDir() {
			// Check if we should skip this directory
			if wp.config.DirFilter.ShouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count this file
		mu.Lock()
		result.TotalFiles++
		mu.Unlock()

		// Check file size limit
		if wp.config.MaxFileSize > 0 && info.Size() > wp.config.MaxFileSize {
			mu.Lock()
			result.SkippedBySize++
			mu.Unlock()
			return nil
		}

		// Check extension filter
		if !wp.config.ExtFilter.ShouldScan(path) {
			mu.Lock()
			result.SkippedByExt++
			mu.Unlock()
			return nil
		}

		// Add to scan queue
		filesToScan = append(filesToScan, fileJob{
			path: path,
			info: info,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	// If no files to scan, return early
	if len(filesToScan) == 0 {
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// ============================================================
	// PHASE 2: Set up channels for worker communication
	// ============================================================

	// jobsChan: Send files to workers
	// Buffered channel improves performance by allowing producers to continue
	// even if consumers are temporarily busy
	jobsChan := make(chan fileJob, wp.numWorkers*2)

	// resultsChan: Receive scan results from workers
	// Buffered to prevent workers from blocking
	resultsChan := make(chan scanResult, wp.numWorkers*2)

	// ============================================================
	// PHASE 3: Start worker goroutines
	// ============================================================

	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < wp.numWorkers; i++ {
		wg.Add(1)

		// Each worker runs in its own goroutine
		go func(workerID int) {
			defer wg.Done()

			// Worker loop: process jobs until channel is closed
			for job := range jobsChan {
				// Scan the file
				findings, err := wp.scanFile(job.path)

				// Send result back
				resultsChan <- scanResult{
					path:     job.path,
					findings: findings,
					err:      err,
				}
			}
		}(i)
	}

	// ============================================================
	// PHASE 4: Send jobs to workers
	// ============================================================

	// Goroutine to send jobs
	// This runs concurrently so we don't block result collection
	go func() {
		for _, job := range filesToScan {
			jobsChan <- job
		}
		close(jobsChan) // Signal no more jobs
	}()

	// ============================================================
	// PHASE 5: Collect results from workers
	// ============================================================

	// Goroutine to close results channel after all workers finish
	go func() {
		wg.Wait()          // Wait for all workers
		close(resultsChan) // Close results channel
	}()

	// Collect all results
	for scanRes := range resultsChan {
		mu.Lock()

		result.ScannedFiles++

		// If scanning succeeded and found cards
		if scanRes.err == nil && len(scanRes.findings) > 0 {
			result.CardsFound += len(scanRes.findings)
			result.Findings = append(result.Findings, scanRes.findings...)
			result.GroupedByFile[scanRes.path] = scanRes.findings
		}

		// Call progress callback if provided
		if wp.config.ProgressCallback != nil {
			wp.config.ProgressCallback(result.ScannedFiles, len(filesToScan), result.CardsFound)
		}

		mu.Unlock()
	}

	// ============================================================
	// PHASE 6: Calculate final statistics
	// ============================================================

	result.Duration = time.Since(startTime)
	if result.Duration.Seconds() > 0 {
		result.ScanRate = float64(result.ScannedFiles) / result.Duration.Seconds()
	}

	return result, nil
}

// scanFile is a helper method for scanning individual files
// This wraps the Scanner.ScanFile method for use in the worker pool
func (wp *WorkerPool) scanFile(path string) ([]Finding, error) {
	// Create a temporary scanner for this file
	// We can't reuse the same scanner because it might not be thread-safe
	scanner := basicScanner{config: wp.config}
	return scanner.ScanFile(path)
}
