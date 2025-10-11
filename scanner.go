package main

import (
	"bufio" // buffered I/O for reading input
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt" // for printing
	"os"  // operating system stuff (Stdin)
	"path/filepath"
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

// findCardNumber searches for credit card patterns (13-19 consecutive digits)
// Different card types have different lengths:
// - Visa: 13, 16, or 19 digits
// - MasterCard: 16 digits
// - Amex: 15 digits
// - Discover: 16 digits
// - Diners: 14 digits
// It handles common formats like spaces and dashes between digit groups
func findCardNumber(text string) string {
	consecutiveDigits := ""

	// Iterate through each character in the text
	for i := 0; i < len(text); i++ {
		char := text[i]

		if char >= '0' && char <= '9' {
			// Found a digit, add it to our collection
			consecutiveDigits = consecutiveDigits + string(char)

			// Check if we have a valid card length (13-19 digits)
			length := len(consecutiveDigits)

			// If we have between 13-19 digits, check what comes next
			if length >= 13 && length <= 19 {
				// Look ahead to see if there are more digits
				if i+1 < len(text) && text[i+1] >= '0' && text[i+1] <= '9' {
					// More digits coming, keep collecting
					continue
				}

				// No more digits, return what we have if it's valid length
				return consecutiveDigits
			}

		} else if char == ' ' || char == '-' {
			// Space or dash - could be formatting in a card number
			// Only reset if we haven't started collecting digits
			if len(consecutiveDigits) == 0 {
				consecutiveDigits = ""
			}
			// Otherwise, continue collecting (skip the separator)

		} else {
			// Any other character breaks the sequence
			// Check if we had a valid length before resetting
			length := len(consecutiveDigits)
			if length >= 13 && length <= 19 {
				return consecutiveDigits
			}

			// Reset for next potential card
			consecutiveDigits = ""
		}
	}

	// Check final collection at end of string
	length := len(consecutiveDigits)
	if length >= 13 && length <= 19 {
		return consecutiveDigits
	}

	// No valid card number found
	return ""
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

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		cardNumber := findCardNumber(line)
		if cardNumber != "" && validateLuhn(cardNumber) {
			validCount++

			cardType := getCardType(cardNumber)
			maskedCard := maskCardNumber(cardNumber)

			// Add to report (no console output - controlled by caller)
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

// getCardType identifies the card issuer based on the card number
func getCardType(cardNumber string) string {
	length := len(cardNumber)

	// Check length is valid for cards
	if length < 13 || length > 19 {
		return "Unknown"
	}

	// Check first digits for card type
	firstDigit := cardNumber[0]
	firstTwo := cardNumber[0:2]
	firstFour := ""
	if len(cardNumber) >= 4 {
		firstFour = cardNumber[0:4]
	}

	// Visa: Starts with 4 (13, 16, or 19 digits)
	if firstDigit == '4' {
		if length == 13 || length == 16 || length == 19 {
			return "Visa"
		}
		return "Unknown" // Wrong length for Visa
	}

	// MasterCard: Starts with 51-55 or 2221-2720 (16 digits only)
	if (firstTwo >= "51" && firstTwo <= "55") ||
		(firstFour >= "2221" && firstFour <= "2720") {
		if length == 16 {
			return "MasterCard"
		}
		return "Unknown" // Wrong length for MasterCard
	}

	// American Express: Starts with 34 or 37 (15 digits only)
	if firstTwo == "34" || firstTwo == "37" {
		if length == 15 {
			return "Amex"
		}
		return "Unknown" // Wrong length for Amex
	}

	// Discover: Starts with 6011, 644-649, 65 (16 digits only)
	if firstFour == "6011" ||
		(firstTwo >= "64" && firstTwo <= "65") {
		if length == 16 {
			return "Discover"
		}
		return "Unknown"
	}

	// Diners Club: Starts with 36, 38, or 300-305 (14 digits)
	if firstTwo == "36" || firstTwo == "38" ||
		(firstFour >= "3000" && firstFour <= "3059") {
		if length == 14 {
			return "Diners"
		}
		return "Unknown"
	}

	// JCB: Starts with 3528-3589 (16 digits)
	if firstFour >= "3528" && firstFour <= "3589" {
		if length == 16 {
			return "JCB"
		}
		return "Unknown"
	}

	return "Unknown"
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

	// Load config file (use defaults if it fails)
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
