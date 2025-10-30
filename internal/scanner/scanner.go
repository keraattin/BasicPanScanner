// Package scanner handles file and directory scanning for credit cards
// This package orchestrates the scanning process using detector and filter packages
//
// UPDATED v3.0 - Pure GO Office Document Support:
//   - Uses ONLY GO standard library (no external dependencies!)
//   - Supports 17 office document formats!
//   - Microsoft Office: DOCX, XLSX, PPTX (+ macro/template variants)
//   - OpenDocument: ODT, ODS, ODP
//   - No external servers, No API calls
//   - 100% self-contained
//   - Simple ZIP+XML parsing
package scanner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"../detector"
	"../filter"
)

// ============================================================
// DATA STRUCTURES
// ============================================================

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

// ============================================================
// CONSTRUCTOR
// ============================================================

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

// ============================================================
// SCAN FILE FUNCTION (WITH PURE GO OFFICE SUPPORT)
// ============================================================

// ScanFile scans a single file for credit cards
//
// UPDATED v3.0:
//   - NOW SUPPORTS 17 OFFICE DOCUMENT FORMATS!
//   - Uses ONLY GO standard library!
//   - NO external dependencies
//   - NO external servers or API calls
//   - 100% self-contained
//
// HOW IT WORKS:
//  1. Check if file is an office document (17 supported formats)
//  2. If YES: Extract text using our pure GO parser (ZIP+XML)
//  3. If NO: Read directly with os.ReadFile (plain text)
//  4. Pass text to credit card detector
//  5. Convert results to Finding format
//
// SUPPORTED FILE TYPES:
//
//	✅ Plain text files (.txt, .log, .csv, .json, etc.)
//
//	✅ Microsoft Office 2007+ (14 formats):
//	   • Word: DOCX, DOCM, DOTX, DOTM
//	   • Excel: XLSX, XLSM, XLTX, XLTM
//	   • PowerPoint: PPTX, PPTM, POTX, POTM
//
//	✅ OpenDocument Format (3 formats):
//	   • Text: ODT
//	   • Spreadsheet: ODS
//	   • Presentation: ODP
//
// NOT SUPPORTED (would need external libraries):
//
//	❌ Old Office formats (.doc, .xls, .ppt) - binary format
//	❌ PDF - requires complex parsing
//
// Parameters:
//   - filePath: Path to the file to scan
//
// Returns:
//   - []Finding: List of all credit cards found in the file
//   - error: Error if file can't be read or other issues occur
//
// Example:
//
//	Scanning a text file (direct read)
//	findings, err := scanner.ScanFile("/var/log/app.log")
//
//	Scanning a Word document (ZIP+XML parsing)
//	findings, err := scanner.ScanFile("/documents/report.docx")
//
//	Scanning an Excel spreadsheet (ZIP+XML parsing)
//	findings, err := scanner.ScanFile("/data/customers.xlsx")
func (s *basicScanner) ScanFile(filePath string) ([]Finding, error) {
	// ============================================================
	// STEP 1: Read file content
	// ============================================================

	var text string // Will hold the file content as text
	var err error

	// Check if PDF file
	if isPDF, _ := isPDFFile(filePath); isPDF {
		text, err = readPDF(filePath)
		if err != nil {
			// Log the error but continue with empty text
			log.Printf("Warning: PDF read error for %s: %v", filePath, err)
			text = "" // Continue with empty text
		}

		// Check if this is an office document
	} else if isOfficeDocument(filePath) {
		// OFFICE DOCUMENT PATH
		// This handles: .docx, .xlsx, .pptx
		//
		// How it works:
		//  1. Open file as ZIP archive (archive/zip)
		//  2. Extract XML files inside
		//  3. Parse XML to get text (encoding/xml)
		//  4. Return plain text
		//
		// All using GO standard library!
		text, err = readOfficeDocument(filePath)
		if err != nil {
			// Return error with helpful message
			return nil, fmt.Errorf("failed to read office document: %w", err)
		}
	} else {
		// PLAIN TEXT FILE PATH
		// This handles: txt, log, csv, json, xml, html, code files, etc.
		//
		// Just read the file directly - it's already text!
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		// Convert bytes to string
		text = string(content)
	}

	// ============================================================
	// STEP 2: Detect credit cards in the text
	// ============================================================
	// This works the same for both office and plain text files
	// After extraction, all content is just text to be scanned
	//
	// The detector performs the complete pipeline:
	//   Phase 1: Find card-like patterns (fast regex)
	//   Phase 2: Match issuer (BIN database lookup)
	//   Phase 3: Validate Luhn (checksum)
	//   Phase 4: Calculate line numbers
	cardLocations := detector.DetectCardsInFile(text)

	// ============================================================
	// STEP 3: Convert CardLocation to Finding
	// ============================================================
	// CardLocation has: CardNumber, CardType, LineNumber, StartIndex, EndIndex
	// Finding has: FilePath, CardNumber, CardType, MaskedCard, Timestamp
	//
	// We need to convert and add file-specific information
	var findings []Finding

	for _, cardLoc := range cardLocations {
		// Create Finding with all necessary information
		finding := Finding{
			FilePath:   filePath,
			LineNumber: cardLoc.LineNumber,
			CardType:   cardLoc.CardType,
			CardNumber: cardLoc.CardNumber,
			MaskedCard: detector.MaskCardNumber(cardLoc.CardNumber),
			Timestamp:  time.Now(),
		}

		// Add to results
		findings = append(findings, finding)
	}

	return findings, nil
}

// ============================================================
// SCAN DIRECTORY FUNCTION
// ============================================================

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

// ============================================================
// SINGLE-THREADED SCAN
// ============================================================

// scanDirectorySingleThreaded scans directory without concurrency
// This is simpler and uses less CPU/memory
//
// Process:
//  1. Walk directory tree
//  2. Filter by extension and directory
//  3. Scan each file
//  4. Collect results
//  5. Generate statistics
func (s *basicScanner) scanDirectorySingleThreaded(dirPath string) (*ScanResult, error) {
	startTime := time.Now()

	// Initialize result structure
	result := &ScanResult{
		Findings:      make([]Finding, 0),
		GroupedByFile: make(map[string][]Finding),
	}

	// Collect all files first
	var filesToScan []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be excluded
			// ShouldSkip returns true if we should skip this directory
			if s.config.DirFilter != nil && s.config.DirFilter.ShouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count total files
		result.TotalFiles++

		// Check extension filter
		if s.config.ExtFilter != nil && !s.config.ExtFilter.ShouldScan(path) {
			result.SkippedByExt++
			return nil
		}

		// Check file size
		if s.config.MaxFileSize > 0 && info.Size() > s.config.MaxFileSize {
			result.SkippedBySize++
			return nil
		}

		filesToScan = append(filesToScan, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Scan each file
	for i, filePath := range filesToScan {
		// Scan the file
		findings, err := s.ScanFile(filePath)
		if err != nil {
			// Log error but continue scanning
			// For office documents, this might be because file is corrupted
			fmt.Printf("Warning: failed to scan %s: %v\n", filePath, err)
			continue
		}

		// Update statistics
		result.ScannedFiles++
		result.CardsFound += len(findings)

		// Store findings
		if len(findings) > 0 {
			result.Findings = append(result.Findings, findings...)
			result.GroupedByFile[filePath] = findings
		}

		// Progress callback
		if s.config.ProgressCallback != nil {
			s.config.ProgressCallback(i+1, len(filesToScan), result.CardsFound)
		}
	}

	// Calculate statistics
	result.Duration = time.Since(startTime)
	if result.Duration.Seconds() > 0 {
		result.ScanRate = float64(result.ScannedFiles) / result.Duration.Seconds()
	}

	return result, nil
}

// ============================================================
// CONCURRENT SCAN (SIMPLIFIED)
// ============================================================

// scanDirectoryConcurrent scans directory with multiple workers
// This is faster but uses more CPU/memory
func (s *basicScanner) scanDirectoryConcurrent(dirPath string) (*ScanResult, error) {
	// For now, just use single-threaded
	// Concurrent implementation would require worker pools and channels
	// We'll keep it simple for learning
	return s.scanDirectorySingleThreaded(dirPath)
}
