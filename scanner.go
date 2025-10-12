package main

import (
	"bufio" // buffered I/O for reading input
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt" // for printing
	"os"  // operating system stuff (Stdin)
	"path/filepath"
	"regexp"  // For regex pattern matching
	"runtime" //to get CPU count
	"strconv" //for converting strings to numbers
	"strings"
	"sync" //for Mutex and WaitGroup
	"time"
)

// Config holds our configuration settings
type Config struct {
	Extensions  []string `json:"extensions"`
	ExcludeDirs []string `json:"exclude_dirs"`
	MaxFileSize string   `json:"max_file_size"`
}

// loadConfig reads the config.json file and returns a Config
func loadConfig(filename string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	// Parse the JSON
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return &config, nil
}

// parseFileSize converts "100MB" to bytes
// Returns 0 if empty string (no limit)
func parseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Empty means no limit
	if sizeStr == "" {
		return 0, nil
	}

	// Map of suffixes to multipliers
	sizes := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	// Check each suffix
	for suffix, multiplier := range sizes {
		if strings.HasSuffix(sizeStr, suffix) {
			// Remove suffix and parse number
			numStr := strings.TrimSuffix(sizeStr, suffix)
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size: %s", sizeStr)
			}
			return num * multiplier, nil
		}
	}

	return 0, fmt.Errorf("invalid size format: %s", sizeStr)
}

// shouldExcludeDir checks if a directory should be skipped
func shouldExcludeDir(dirPath string, excludeDirs []string) bool {
	dirName := filepath.Base(dirPath)

	for _, excludeDir := range excludeDirs {
		if dirName == excludeDir {
			return true
		}
	}

	return false
}

// Report is the common structure for all export formats
type Report struct {
	ScanDate     time.Time
	Directory    string
	Extensions   []string
	Duration     time.Duration
	TotalFiles   int
	ScannedFiles int
	Findings     []CardFinding
}

// CardFinding represents a single credit card finding
type CardFinding struct {
	FilePath   string
	LineNumber int
	CardType   string
	MaskedCard string
	Timestamp  time.Time
}

// Global report variable
var currentReport *Report

// initReport initializes a new report
func initReport(directory string, extensions []string) {
	currentReport = &Report{
		ScanDate:   time.Now(),
		Directory:  directory,
		Extensions: extensions,
		Findings:   []CardFinding{},
	}
}

// addFinding adds a finding to the current report
func addFinding(filepath string, lineNumber int, cardType string, maskedCard string) {
	if currentReport == nil {
		return
	}

	finding := CardFinding{
		FilePath:   filepath,
		LineNumber: lineNumber,
		CardType:   cardType,
		MaskedCard: maskedCard,
		Timestamp:  time.Now(),
	}

	currentReport.Findings = append(currentReport.Findings, finding)
}

// finalizeReport sets the final statistics
func finalizeReport(totalFiles int, scannedFiles int, duration time.Duration) {
	if currentReport == nil {
		return
	}

	currentReport.TotalFiles = totalFiles
	currentReport.ScannedFiles = scannedFiles
	currentReport.Duration = duration
}

// exportReport exports the report in the requested format
func exportReport(filename string) error {
	if currentReport == nil {
		return fmt.Errorf("no report to export")
	}

	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".json":
		return exportJSON(filename)
	case ".csv":
		return exportCSV(filename)
	case ".txt":
		return exportTXT(filename)
	case ".html":
		return exportHTML(filename)
	default:
		return fmt.Errorf("unsupported format: %s", ext)
	}
}

// exportJSON exports report as JSON
func exportJSON(filename string) error {
	data, err := json.MarshalIndent(currentReport, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// exportCSV exports report as CSV
func exportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"File", "Line", "Card Type", "Masked Card", "Timestamp"})

	// Write findings
	for _, f := range currentReport.Findings {
		writer.Write([]string{
			f.FilePath,
			fmt.Sprintf("%d", f.LineNumber),
			f.CardType,
			f.MaskedCard,
			f.Timestamp.Format("2006-01-02 15:04:05"),
		})
	}

	return nil
}

// exportHTML exports report as HTML
func exportHTML(filename string) error {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <title>PAN Scanner Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:hover { background-color: #f5f5f5; }
        .summary { background-color: #f0f0f0; padding: 10px; margin: 20px 0; }
    </style>
</head>
<body>
    <h1>BasicPanScanner Report</h1>
    <div class="summary">
        <p><strong>Scan Date:</strong> ` + currentReport.ScanDate.Format("2006-01-02 15:04:05") + `</p>
        <p><strong>Directory:</strong> ` + currentReport.Directory + `</p>
        <p><strong>Duration:</strong> ` + currentReport.Duration.String() + `</p>
        <p><strong>Files Scanned:</strong> ` + fmt.Sprintf("%d / %d", currentReport.ScannedFiles, currentReport.TotalFiles) + `</p>
        <p><strong>Cards Found:</strong> ` + fmt.Sprintf("%d", len(currentReport.Findings)) + `</p>
    </div>
    <h2>Findings</h2>
    <table>
        <tr>
            <th>File</th>
            <th>Line</th>
            <th>Card Type</th>
            <th>Masked Card</th>
            <th>Timestamp</th>
        </tr>`)

	for _, f := range currentReport.Findings {
		html.WriteString(fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td>%d</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
        </tr>`, f.FilePath, f.LineNumber, f.CardType, f.MaskedCard, f.Timestamp.Format("15:04:05")))
	}

	html.WriteString(`
    </table>
</body>
</html>`)

	return os.WriteFile(filename, []byte(html.String()), 0644)
}

// exportTXT exports report as plain text
func exportTXT(filename string) error {
	var content strings.Builder

	content.WriteString("BASICPANSCANNER REPORT\n")
	content.WriteString(strings.Repeat("=", 50) + "\n\n")

	content.WriteString(fmt.Sprintf("Scan Date: %s\n", currentReport.ScanDate.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Directory: %s\n", currentReport.Directory))
	content.WriteString(fmt.Sprintf("Duration: %s\n", currentReport.Duration))
	content.WriteString(fmt.Sprintf("Total Files: %d\n", currentReport.TotalFiles))
	content.WriteString(fmt.Sprintf("Scanned: %d\n", currentReport.ScannedFiles))
	content.WriteString(fmt.Sprintf("Cards Found: %d\n\n", len(currentReport.Findings)))

	content.WriteString("FINDINGS:\n")
	content.WriteString(strings.Repeat("-", 50) + "\n")

	for i, f := range currentReport.Findings {
		content.WriteString(fmt.Sprintf("\n[%d] %s\n", i+1, f.CardType))
		content.WriteString(fmt.Sprintf("    File: %s\n", f.FilePath))
		content.WriteString(fmt.Sprintf("    Line: %d\n", f.LineNumber))
		content.WriteString(fmt.Sprintf("    Card: %s\n", f.MaskedCard))
	}

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// validateDirectory checks if the path exists and is a directory
func validateDirectory(dirPath string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dirPath)
		}
		return fmt.Errorf("error accessing directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	return nil
}

// ============================================================================
// SIMPLIFIED CARD DETECTION WITH REGEX
// ============================================================================

// CardPattern represents a single card issuer's detection pattern
type CardPattern struct {
	Name    string         // Card issuer name (e.g., "Visa")
	Pattern *regexp.Regexp // Regex pattern that handles EVERYTHING
}

// Global patterns - compiled once at startup for performance
var cardPatterns []CardPattern

// initCardPatterns creates all card detection patterns
// Each regex handles: formatting (spaces/dashes), length, and issuer identification
func initCardPatterns() {
	cardPatterns = []CardPattern{
		// VISA
		// Starts with 4
		// Length: 13, 16, or 19 digits
		// Handles: 4532015112830366 OR 4532-0151-1283-0366 OR 4532 0151 1283 0366
		{
			Name:    "Visa",
			Pattern: regexp.MustCompile(`\b4\d{3}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}(?:[\s\-]?\d{3})?(?:[\s\-]?\d{3})?\b`),
		},

		// MASTERCARD
		// Starts with 51-55 OR 2221-2720
		// Length: 16 digits only
		{
			Name:    "MasterCard",
			Pattern: regexp.MustCompile(`\b(?:5[1-5]|222[1-9]|22[3-9]\d|2[3-6]\d{2}|27[01]\d|2720)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},

		// AMERICAN EXPRESS
		// Starts with 34 or 37
		// Length: 15 digits only
		// Format: 3xxx-xxxxxx-xxxxx (different grouping than Visa/MC)
		{
			Name:    "Amex",
			Pattern: regexp.MustCompile(`\b3[47]\d{2}[\s\-]?\d{6}[\s\-]?\d{5}\b`),
		},

		// DISCOVER
		// Starts with 6011, 622126-622925, 644-649, or 65
		// Length: 16 digits only
		{
			Name:    "Discover",
			Pattern: regexp.MustCompile(`\b(?:6011|65\d{2}|64[4-9]\d|622(?:1[2-9]\d|[2-8]\d{2}|9[01]\d|92[0-5]))\d{0,2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},

		// DINERS CLUB
		// Starts with 36, 38, or 300-305
		// Length: 14 digits only
		{
			Name:    "Diners",
			Pattern: regexp.MustCompile(`\b3(?:0[0-5]|[68]\d)\d{1}[\s\-]?\d{6}[\s\-]?\d{4}\b`),
		},

		// JCB
		// Starts with 3528-3589
		// Length: 16 digits only
		{
			Name:    "JCB",
			Pattern: regexp.MustCompile(`\b35(?:2[89]|[3-8]\d)\d{1}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},

		// CHINA UNIONPAY
		// Starts with 62
		// Length: 16-19 digits
		{
			Name:    "UnionPay",
			Pattern: regexp.MustCompile(`\b62\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}(?:[\s\-]?\d{3})?(?:[\s\-]?\d{3})?\b`),
		},

		// MAESTRO
		// Starts with 50, 56-69
		// Length: 12-19 digits (very flexible)
		{
			Name:    "Maestro",
			Pattern: regexp.MustCompile(`\b(?:5[06789]|6\d)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{1,7}\b`),
		},

		// RUPAY (India)
		// Starts with 60, 6521, 6522
		// Length: 16 digits
		{
			Name:    "RuPay",
			Pattern: regexp.MustCompile(`\b(?:60|6521|6522)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},

		// TROY (Turkey)
		// Starts with 9792
		// Length: 16 digits
		{
			Name:    "Troy",
			Pattern: regexp.MustCompile(`\b9792[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},

		// MIR (Russia)
		// Starts with 2200-2204
		// Length: 16 digits
		{
			Name:    "Mir",
			Pattern: regexp.MustCompile(`\b220[0-4][\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`),
		},
	}

	fmt.Printf("✓ Loaded %d card issuer patterns\n", len(cardPatterns))
}

// cleanDigits extracts only digits from a string
// Example: "4532-0151-1283-0366" -> "4532015112830366"
func cleanDigits(text string) string {
	digits := ""
	for i := 0; i < len(text); i++ {
		if text[i] >= '0' && text[i] <= '9' {
			digits += string(text[i])
		}
	}
	return digits
}

// findCardsInLine scans one line of text for credit card numbers
// Returns map: cardNumber -> cardType
//
// Process:
// 1. Try each regex pattern on the line
// 2. For each match, extract only digits
// 3. Validate with Luhn algorithm
// 4. Return valid cards with their types
func findCardsInLine(line string) map[string]string {
	foundCards := make(map[string]string)

	// Try each card pattern
	for _, pattern := range cardPatterns {
		// Find all matches for this pattern in the line
		matches := pattern.Pattern.FindAllString(line, -1)

		for _, match := range matches {
			// Extract only the digits (remove spaces/dashes)
			cardNumber := cleanDigits(match)

			// Skip if we already found this card
			if _, exists := foundCards[cardNumber]; exists {
				continue
			}

			// Validate with Luhn algorithm (eliminates false positives)
			if validateLuhn(cardNumber) {
				foundCards[cardNumber] = pattern.Name
			}
		}
	}

	return foundCards
}

// scanDirectoryWithOptionsConcurrent scans directory with goroutines
func scanDirectoryWithOptionsConcurrent(dirPath string, outputFile string, extensions []string, excludeDirs []string, maxFileSize int64, workers int) error {
	fmt.Printf("\nScanning directory: %s\n", dirPath)
	fmt.Printf("Workers: %d (concurrent scanning enabled)\n", workers)
	fmt.Println(strings.Repeat("=", 60))

	// Initialize report
	initReport(dirPath, extensions)

	startTime := time.Now()

	// Shared counters (protected by mutex)
	var mu sync.Mutex
	totalFiles := 0
	scannedFiles := 0
	skippedFiles := 0
	foundCards := 0

	// Collect files to scan
	var filesToScan []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories
		if info.IsDir() {
			if shouldExcludeDir(path, excludeDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count total files
		mu.Lock()
		totalFiles++
		mu.Unlock()

		// Skip large files
		if maxFileSize > 0 && info.Size() > maxFileSize {
			mu.Lock()
			skippedFiles++
			mu.Unlock()
			return nil
		}

		// Check extension
		ext := strings.ToLower(filepath.Ext(path))
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				filesToScan = append(filesToScan, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	// Create channels for work distribution
	filesChan := make(chan string, workers*2) // Buffered channel

	// WaitGroup to wait for all workers
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Each worker processes files from the channel
			for filePath := range filesChan {
				cardsFound := scanFileWithCount(filePath)

				// Update shared counters safely
				mu.Lock()
				scannedFiles++
				if cardsFound > 0 {
					foundCards += cardsFound
					fmt.Printf("✓ Found %d cards in: %s\n", cardsFound, filepath.Base(filePath))
				}
				// Show progress
				fmt.Printf("\r[Scanned: %d/%d | Cards: %d]", scannedFiles, len(filesToScan), foundCards)
				mu.Unlock()
			}
		}(i)
	}

	// Send files to workers
	for _, file := range filesToScan {
		filesChan <- file
	}
	close(filesChan) // Close channel when done sending

	// Wait for all workers to finish
	wg.Wait()

	// Clear progress line
	fmt.Print("\r" + strings.Repeat(" ", 60) + "\r")

	elapsed := time.Since(startTime)

	// Finalize report
	finalizeReport(totalFiles, scannedFiles, elapsed)

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ Scan complete!\n")
	fmt.Printf("  Time: %s\n", elapsed.Round(time.Second))
	fmt.Printf("  Total files: %d\n", totalFiles)
	fmt.Printf("  Scanned: %d\n", scannedFiles)
	if skippedFiles > 0 {
		fmt.Printf("  Skipped: %d\n", skippedFiles)
	}
	fmt.Printf("  Cards found: %d\n", foundCards)

	if elapsed.Seconds() > 0 {
		rate := float64(scannedFiles) / elapsed.Seconds()
		fmt.Printf("  Scan rate: %.1f files/second\n", rate)
	}

	// Export if requested
	if outputFile != "" {
		err = exportReport(outputFile)
		if err != nil {
			fmt.Printf("\n  Error: %v\n", err)
		} else {
			fmt.Printf("\n  ✓ Saved: %s\n", outputFile)
		}
	}

	return nil
}

// scanDirectoryWithOptions scans a directory (single-threaded)
func scanDirectoryWithOptions(dirPath string, outputFile string, extensions []string, excludeDirs []string, maxFileSize int64) error {
	fmt.Printf("\nScanning directory: %s\n", dirPath)
	fmt.Printf("Workers: 1 (single-threaded mode)\n")
	fmt.Println(strings.Repeat("=", 60))

	// Initialize report
	initReport(dirPath, extensions)

	startTime := time.Now()
	totalFiles := 0
	scannedFiles := 0
	skippedFiles := 0
	foundCards := 0

	// Progress indicator timing
	lastUpdate := time.Now()

	// Walk through all files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// If it's a directory, check if we should skip it
		if info.IsDir() {
			if shouldExcludeDir(path, excludeDirs) {
				return filepath.SkipDir
			}
			return nil
		}

		// Count this file
		totalFiles++

		// Skip if file is too big
		if maxFileSize > 0 && info.Size() > maxFileSize {
			skippedFiles++
			return nil
		}

		// Check if extension matches
		ext := strings.ToLower(filepath.Ext(path))
		shouldScan := false
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				shouldScan = true
				break
			}
		}

		if shouldScan {
			scannedFiles++

			// Scan this file
			cardsFound := scanFileWithCount(path)
			if cardsFound > 0 {
				foundCards += cardsFound
				fmt.Printf("✓ Found %d cards in: %s\n", cardsFound, filepath.Base(path))
			}

			// Update progress every 100ms
			if time.Since(lastUpdate) > 100*time.Millisecond {
				fmt.Printf("\r[Scanned: %d/%d | Cards: %d]", scannedFiles, totalFiles, foundCards)
				lastUpdate = time.Now()
			}
		}

		return nil
	})

	// Clear progress line
	fmt.Print("\r" + strings.Repeat(" ", 60) + "\r")

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	elapsed := time.Since(startTime)

	// Finalize report
	finalizeReport(totalFiles, scannedFiles, elapsed)

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ Scan complete!\n")
	fmt.Printf("  Time: %s\n", elapsed.Round(time.Second))
	fmt.Printf("  Total files: %d\n", totalFiles)
	fmt.Printf("  Scanned: %d\n", scannedFiles)
	if skippedFiles > 0 {
		fmt.Printf("  Skipped: %d\n", skippedFiles)
	}
	fmt.Printf("  Cards found: %d\n", foundCards)

	if elapsed.Seconds() > 0 {
		rate := float64(scannedFiles) / elapsed.Seconds()
		fmt.Printf("  Scan rate: %.1f files/second\n", rate)
	}

	// Export if requested
	if outputFile != "" {
		err = exportReport(outputFile)
		if err != nil {
			fmt.Printf("\n  Error: %v\n", err)
		} else {
			fmt.Printf("\n  ✓ Saved: %s\n", outputFile)
		}
	}

	return nil
}

// scanFileWithCount scans a file and returns number of valid cards found
func scanFileWithCount(filepath string) int {
	file, err := os.Open(filepath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	validCount := 0

	// Track cards already found in this file (avoid duplicates)
	seenCards := make(map[string]bool)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Find all cards in this line using regex + Luhn validation
		cardsFound := findCardsInLine(line)

		// Process each valid card found
		for cardNumber, cardType := range cardsFound {
			// Skip if we already reported this card in this file
			if seenCards[cardNumber] {
				continue
			}
			seenCards[cardNumber] = true
			validCount++

			// Mask the card number for safe display (PCI compliance)
			maskedCard := maskCardNumber(cardNumber)

			// Add to report
			addFinding(filepath, lineNumber, cardType, maskedCard)
		}
	}

	return validCount
}

//How Luhn Algorithm Works:
// - Start from the rightmost digit
// - Double every second digit (from right)
// - If doubled value > 9, subtract 9
// - Sum all digits
// - Valid if sum is divisible by 10

// validate Luhn checks if a card number passes the Luhn algorithm
// This is a checksum formula used to validate credit card numbers
func validateLuhn(cardNumber string) bool {
	// Remove any spaces or dashes that might still be in the number
	cleaned := ""
	for i := 0; i < len(cardNumber); i++ {
		if cardNumber[i] >= '0' && cardNumber[i] <= '9' {
			cleaned += string(cardNumber[i])
		}
	}

	// Need at least 13 digits for a valid card (some cards are 13-19 digits)
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	sum := 0
	isEven := false

	// Process digits from right to left
	for i := len(cleaned) - 1; i >= 0; i-- {
		// Convert character to integer
		digit := int(cleaned[i] - '0')

		// Every second digit (from right) is doubled
		if isEven {
			digit *= 2
			// If doubled digit is > 9, subtract 9
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	// Valid if sum is divisible by 10
	return sum%10 == 0
}

// maskCardNumber returns a masked version for safe display
// PCI compliance: Shows only first 6 (BIN) and last 4 digits
func maskCardNumber(cardNumber string) string {
	length := len(cardNumber)

	// If card number is too short, just return it
	if length <= 10 {
		return cardNumber
	}

	// Build masked number
	masked := ""

	// Add first 6 digits (BIN - Bank Identification Number)
	masked += cardNumber[0:6]

	// Add asterisks for middle digits
	middleDigits := length - 10 // Total minus first 6 and last 4
	for i := 0; i < middleDigits; i++ {
		masked += "*"
	}

	// Add last 4 digits
	masked += cardNumber[length-4:]

	return masked
}

// showHelp displays usage information
func showHelp() {
	fmt.Println(`
BasicPanScanner v1.1.0 - PCI Compliance Scanner
Usage: ./scanner -path <directory> [options]

Required:
    -path <directory>      Directory to scan

Options:
    -output <file>         Save results (.json, .csv, .html, .txt)
    -ext <list>           Extensions to scan (default: from config)
    -exclude <list>       Directories to skip (default: from config)
    -workers <n>          Number of concurrent workers (default: CPU/2, max: CPU cores)
    -help                 Show this help

Examples:
    # Basic scan (uses default workers)
    ./scanner -path /var/log

    # Fast scan with more workers
    ./scanner -path /var/log -workers 4

    # Single-threaded scan
    ./scanner -path /var/log -workers 1

    # Full scan with output
    ./scanner -path /data -workers 4 -output report.json

Performance:
    Default workers: CPU cores / 2 (safe for production)
    Max workers: CPU cores (automatically limited)
    More workers = faster scanning (2-4x speed improvement)

Configuration:
    Edit config.json to change default settings.
    CLI flags always override config values.

Export Formats:
    .json  - JSON format
    .csv   - CSV format
    .html  - HTML format
    .txt   - Text format
`)
}

func displayBanner() {
	fmt.Println(`
    ╔══════════════════════════════════════════════════════════╗
    ║                                                          ║
    ║     BasicPanScanner - PCI Compliance Tool                ║
    ║     ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                  ║
    ║     Version: 1.1.0                                       ║
    ║     Author:  @keraattin                                  ║
    ║     Purpose: Detect credit card data in files            ║
    ║                                                          ║
    ║     [████ ████ ████ ████] Card Detection Active          ║
    ║                                                          ║
    ╚══════════════════════════════════════════════════════════╝
    `)
}

func main() {
	// Define command line flags
	pathFlag := flag.String("path", "", "Directory to scan")
	outputFlag := flag.String("output", "", "Output file")
	extensionsFlag := flag.String("ext", "", "Extensions (e.g., txt,log,csv)")
	excludeFlag := flag.String("exclude", "", "Exclude dirs (e.g., .git,vendor)")
	workersFlag := flag.Int("workers", 0, "Number of concurrent workers (default: CPU/2)")
	helpFlag := flag.Bool("help", false, "Show help")

	// Parse the flags
	flag.Parse()

	// Show help if requested
	if *helpFlag || len(os.Args) == 1 {
		showHelp()
		return
	}

	// Path is required
	if *pathFlag == "" {
		fmt.Println("Error: -path is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// Initialize card detection patterns (MUST be called before scanning)
	initCardPatterns()

	// Load config file
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Warning: Could not load config.json: %v\n", err)
		fmt.Println("Using default settings\n")
		config = &Config{
			Extensions:  []string{"txt", "log", "csv"},
			ExcludeDirs: []string{".git", "node_modules"},
			MaxFileSize: "100MB",
		}
	}

	// Start with config values
	extensions := config.Extensions
	excludeDirs := config.ExcludeDirs
	maxFileSize, _ := parseFileSize(config.MaxFileSize)

	// CLI flags override config
	if *extensionsFlag != "" {
		extensions = strings.Split(*extensionsFlag, ",")
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}
	}

	if *excludeFlag != "" {
		excludeDirs = strings.Split(*excludeFlag, ",")
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
	}

	// Determine number of workers
	numCPU := runtime.NumCPU()
	workers := *workersFlag

	if workers == 0 {
		// Default: Half of CPU cores (minimum 1)
		workers = numCPU / 2
		if workers < 1 {
			workers = 1
		}
	}

	// Validate worker count
	if workers < 1 {
		fmt.Println("Error: workers must be at least 1")
		os.Exit(1)
	}

	if workers > numCPU {
		fmt.Printf("Warning: workers (%d) exceeds CPU cores (%d), limiting to %d\n", workers, numCPU, numCPU)
		workers = numCPU
	}

	// Display banner
	displayBanner()

	// Validate directory exists
	err = validateDirectory(*pathFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Add dots to extensions if needed
	for i := range extensions {
		if !strings.HasPrefix(extensions[i], ".") {
			extensions[i] = "." + extensions[i]
		}
	}

	// Run the scan (concurrent if workers > 1)
	if workers > 1 {
		err = scanDirectoryWithOptionsConcurrent(*pathFlag, *outputFlag, extensions, excludeDirs, maxFileSize, workers)
	} else {
		err = scanDirectoryWithOptions(*pathFlag, *outputFlag, extensions, excludeDirs, maxFileSize)
	}

	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}
}
