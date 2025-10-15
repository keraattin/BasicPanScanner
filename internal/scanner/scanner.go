// Package scanner handles file and directory scanning for credit cards
// This package orchestrates the scanning process using detector and filter packages
package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"../detector"
	"../filter"
)

// Finding represents a single credit card finding
// This struct holds all information about where and what was found
type Finding struct {
	FilePath   string    // Full path to the file
	LineNumber int       // Line number where card was found
	CardType   string    // Card issuer (e.g., "Visa", "Mastercard")
	CardNumber string    // Full card number (digits only)
	MaskedCard string    // PCI-compliant masked version
	Timestamp  time.Time // When the finding was made
}

// ScanResult holds the results of a scanning operation
// This is returned after scanning completes
type ScanResult struct {
	TotalFiles    int                  // Total files found
	ScannedFiles  int                  // Files actually scanned
	SkippedBySize int                  // Files skipped due to size
	SkippedByExt  int                  // Files skipped by extension filter
	CardsFound    int                  // Total credit cards found
	Findings      []Finding            // All findings
	GroupedByFile map[string][]Finding // Findings grouped by file
	Duration      time.Duration        // How long the scan took
	ScanRate      float64              // Files per second
}

// Scanner interface defines the contract for file/directory scanning
// Different implementations can be created (single-threaded, concurrent, etc.)
type Scanner interface {
	// ScanFile scans a single file for credit cards
	ScanFile(filePath string) ([]Finding, error)

	// ScanDirectory scans a directory recursively
	ScanDirectory(dirPath string) (*ScanResult, error)

	// GetConfig returns the scanner configuration
	GetConfig() *Config
}

// Config holds scanner configuration
// This is passed to the scanner during creation
type Config struct {
	// Extension filter for determining which files to scan
	ExtFilter *filter.ExtensionFilter

	// Directory filter for determining which directories to skip
	DirFilter *filter.DirectoryFilter

	// Maximum file size to scan (in bytes)
	// Files larger than this will be skipped
	// 0 means no limit
	MaxFileSize int64

	// Number of worker goroutines for concurrent scanning
	// 1 means single-threaded, >1 means concurrent
	Workers int

	// Callback function for progress updates (optional)
	// Called after each file is scanned
	// Parameters: scannedCount, totalCount, cardsFound
	ProgressCallback func(int, int, int)
}

// basicScanner is the default scanner implementation
// It provides both single-threaded and concurrent scanning
type basicScanner struct {
	config *Config
}

// NewScanner creates a new scanner with the given configuration
//
// Parameters:
//   - config: Scanner configuration
//
// Returns:
//   - Scanner: Configured scanner ready to use
//
// Example:
//
//	extFilter := filter.NewExtensionFilter("blacklist", nil, []string{".exe", ".dll"})
//	dirFilter := filter.NewDirectoryFilter([]string{".git", "node_modules"})
//
//	scannerConfig := &scanner.Config{
//	    ExtFilter:   extFilter,
//	    DirFilter:   dirFilter,
//	    MaxFileSize: 52428800, // 50MB
//	    Workers:     2,
//	}
//
//	s := scanner.NewScanner(scannerConfig)
func NewScanner(config *Config) Scanner {
	return &basicScanner{
		config: config,
	}
}

// GetConfig returns the scanner configuration
func (s *basicScanner) GetConfig() *Config {
	return s.config
}

// ScanFile scans a single file for credit cards
// This reads the file line by line and detects cards using the detector package
//
// Parameters:
//   - filePath: Full path to the file to scan
//
// Returns:
//   - []Finding: List of all credit cards found in the file
//   - error: Error if file can't be read or other issues occur
//
// Example:
//
//	findings, err := scanner.ScanFile("/var/log/app.log")
//	if err != nil {
//	    log.Printf("Error scanning file: %v", err)
//	    return
//	}
//
//	fmt.Printf("Found %d cards in file\n", len(findings))
//	for _, finding := range findings {
//	    fmt.Printf("  Line %d: %s - %s\n",
//	        finding.LineNumber, finding.CardType, finding.MaskedCard)
//	}
func (s *basicScanner) ScanFile(filePath string) ([]Finding, error) {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Use buffered scanner for efficient line-by-line reading
	// This is memory-efficient as it doesn't load entire file into memory
	scanner := bufio.NewScanner(file)

	var findings []Finding
	lineNumber := 0

	// Track cards we've already seen in this file to avoid duplicates
	// Some files might have the same card number on multiple lines
	seenCards := make(map[string]bool)

	// Read file line by line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Use detector package to find cards in this line
		cardsInLine := detector.FindCardsInText(line)

		// Process each card found
		for cardNumber, cardType := range cardsInLine {
			// Skip if we've already recorded this card in this file
			if seenCards[cardNumber] {
				continue
			}
			seenCards[cardNumber] = true

			// Create a finding record
			finding := Finding{
				FilePath:   filePath,
				LineNumber: lineNumber,
				CardType:   cardType,
				CardNumber: cardNumber,
				MaskedCard: detector.MaskCardNumber(cardNumber),
				Timestamp:  time.Now(),
			}

			findings = append(findings, finding)
		}
	}

	// Check if scanner encountered any errors
	if err := scanner.Err(); err != nil {
		return findings, fmt.Errorf("error reading file: %w", err)
	}

	return findings, nil
}

// ScanDirectory scans a directory recursively for credit cards
// This uses filepath.Walk to traverse the directory tree
//
// Parameters:
//   - dirPath: Path to the directory to scan
//
// Returns:
//   - *ScanResult: Complete scan results with statistics
//   - error: Error if directory can't be accessed
//
// Example:
//
//	result, err := scanner.ScanDirectory("/var/log")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Scanned: %d files\n", result.ScannedFiles)
//	fmt.Printf("Found: %d cards\n", result.CardsFound)
//	fmt.Printf("Duration: %s\n", result.Duration)
func (s *basicScanner) ScanDirectory(dirPath string) (*ScanResult, error) {
	// If workers > 1, use concurrent scanning
	if s.config.Workers > 1 {
		return s.scanDirectoryConcurrent(dirPath)
	}

	// Otherwise use single-threaded scanning
	return s.scanDirectorySingleThreaded(dirPath)
}

// scanDirectorySingleThreaded scans directory without concurrency
// This is simpler and uses less CPU/memory
func (s *basicScanner) scanDirectorySingleThreaded(dirPath string) (*ScanResult, error) {
	startTime := time.Now()

	result := &ScanResult{
		GroupedByFile: make(map[string][]Finding),
	}

	// Walk the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files/dirs we can't access
			return nil
		}

		// If it's a directory, check if we should skip it
		if info.IsDir() {
			if s.config.DirFilter.ShouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count this file
		result.TotalFiles++

		// Check file size limit
		if s.config.MaxFileSize > 0 && info.Size() > s.config.MaxFileSize {
			result.SkippedBySize++
			return nil
		}

		// Check extension filter
		if !s.config.ExtFilter.ShouldScan(path) {
			result.SkippedByExt++
			return nil
		}

		// Scan this file
		result.ScannedFiles++

		findings, err := s.ScanFile(path)
		if err != nil {
			// Log error but continue scanning
			// In a production system, you might want to collect these errors
			return nil
		}

		// Add findings to result
		if len(findings) > 0 {
			result.CardsFound += len(findings)
			result.Findings = append(result.Findings, findings...)
			result.GroupedByFile[path] = findings
		}

		// Call progress callback if provided
		if s.config.ProgressCallback != nil {
			s.config.ProgressCallback(result.ScannedFiles, result.TotalFiles, result.CardsFound)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("directory walk failed: %w", err)
	}

	// Calculate final statistics
	result.Duration = time.Since(startTime)
	if result.Duration.Seconds() > 0 {
		result.ScanRate = float64(result.ScannedFiles) / result.Duration.Seconds()
	}

	return result, nil
}

// scanDirectoryConcurrent scans directory using multiple worker goroutines
// This is faster but uses more CPU and memory
func (s *basicScanner) scanDirectoryConcurrent(dirPath string) (*ScanResult, error) {
	// This will be implemented in worker_pool.go
	// For now, delegate to the WorkerPool
	pool := NewWorkerPool(s.config.Workers, s.config)
	return pool.ScanDirectory(dirPath)
}
