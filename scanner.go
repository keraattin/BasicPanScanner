package main

import (
	"bufio" // buffered I/O for reading input
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
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

// ============================================================================
// CONFIGURATION STRUCTURES
// ============================================================================

// Config holds our configuration settings
type Config struct {
	Extensions  []string `json:"extensions"`
	ExcludeDirs []string `json:"exclude_dirs"`
	MaxFileSize string   `json:"max_file_size"`
}

// Report is the common structure for all export formats
type Report struct {
	ScanDate        time.Time                `json:"scan_date" xml:"ScanDate"`
	Directory       string                   `json:"directory" xml:"Directory"`
	Extensions      []string                 `json:"extensions" xml:"Extensions>Extension"`
	Duration        time.Duration            `json:"duration" xml:"Duration"`
	TotalFiles      int                      `json:"total_files" xml:"TotalFiles"`
	ScannedFiles    int                      `json:"scanned_files" xml:"ScannedFiles"`
	GroupedFindings map[string][]CardFinding `json:"grouped_findings" xml:"GroupedFindings"`
	Statistics      Statistics               `json:"statistics" xml:"Statistics"`
}

// CardFinding represents a single credit card finding
type CardFinding struct {
	FilePath   string    `json:"file_path" xml:"FilePath"`
	LineNumber int       `json:"line_number" xml:"LineNumber"`
	CardType   string    `json:"card_type" xml:"CardType"`
	MaskedCard string    `json:"masked_card" xml:"MaskedCard"`
	Timestamp  time.Time `json:"timestamp" xml:"Timestamp"`
}

// CardPattern represents a single card issuer's detection pattern
type CardPattern struct {
	Name    string         // Card issuer name (e.g., "Visa")
	Pattern *regexp.Regexp // Regex pattern that handles everything
}

// Statistics holds scan statistics and insights
type Statistics struct {
	// Card distribution
	CardsByType map[string]int // {"Visa": 15, "Mastercard": 8}

	// File distribution
	FilesByType    map[string]int // {".txt": 50, ".log": 30}
	FilesWithCards int            // Files that contained cards

	// Top findings
	TopFiles []FileStats // Files with most cards

	// Risk assessment
	HighRiskFiles   int // Files with 5+ cards
	MediumRiskFiles int // Files with 2-4 cards
	LowRiskFiles    int // Files with 1 card
}

// FileStats holds statistics for a single file
type FileStats struct {
	FilePath  string
	CardCount int
	CardTypes map[string]int // Card type distribution in this file
}

// ============================================================================
// GLOBAL VARIABLES
// ============================================================================

// Global report variable
var currentReport *Report

// Global patterns - compiled once at startup for performance
var cardPatterns []CardPattern

// ============================================================================
// CONFIGURATION AND VALIDATION
// ============================================================================

// validateConfig checks if config values are valid
func validateConfig(config *Config) error {
	// Check if extensions list is empty
	if len(config.Extensions) == 0 {
		return fmt.Errorf("config error: extensions list cannot be empty")
	}

	// Check if exclude_dirs list is empty (warning, not error)
	if len(config.ExcludeDirs) == 0 {
		fmt.Println("Warning: exclude_dirs is empty - will scan all directories")
	}

	// Validate max_file_size format
	if config.MaxFileSize != "" {
		_, err := parseFileSize(config.MaxFileSize)
		if err != nil {
			return fmt.Errorf("config error: invalid max_file_size '%s': %v", config.MaxFileSize, err)
		}
	}

	// Check for duplicate extensions
	extMap := make(map[string]bool)
	for _, ext := range config.Extensions {
		cleanExt := strings.ToLower(strings.TrimSpace(ext))
		if extMap[cleanExt] {
			fmt.Printf("Warning: duplicate extension '%s' in config\n", ext)
		}
		extMap[cleanExt] = true
	}

	// Check for duplicate exclude_dirs
	dirMap := make(map[string]bool)
	for _, dir := range config.ExcludeDirs {
		cleanDir := strings.TrimSpace(dir)
		if dirMap[cleanDir] {
			fmt.Printf("Warning: duplicate exclude_dir '%s' in config\n", dir)
		}
		dirMap[cleanDir] = true
	}

	return nil
}

// loadConfig reads and validates the config.json file
func loadConfig(filename string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	// Check if file is empty
	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}

	// Parse the JSON
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config (invalid JSON): %w", err)
	}

	// Validate the config
	err = validateConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// parseFileSize converts "100MB" to bytes
// Returns 0 if empty string (no limit)
// FIXED: Now checks suffixes in correct order (longest first)
func parseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Empty means no limit
	if sizeStr == "" {
		return 0, nil
	}

	// CRITICAL FIX: Check suffixes in order from LONGEST to SHORTEST
	// This prevents "100MB" from matching "B" before "MB"
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"GB", 1024 * 1024 * 1024}, // Check GB first
		{"MB", 1024 * 1024},        // Then MB
		{"KB", 1024},               // Then KB
		{"B", 1},                   // Then B last
	}

	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			// Remove suffix and parse number
			numStr := strings.TrimSuffix(sizeStr, s.suffix)
			numStr = strings.TrimSpace(numStr) // Remove any spaces

			// Parse the number
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size format '%s': cannot parse number", sizeStr)
			}

			// Validate reasonable range
			if num < 0 {
				return 0, fmt.Errorf("invalid size '%s': size cannot be negative", sizeStr)
			}
			if num == 0 {
				return 0, nil // "0MB" means no limit
			}

			result := num * s.multiplier

			// Check for overflow
			if result < 0 {
				return 0, fmt.Errorf("invalid size '%s': value too large", sizeStr)
			}

			return result, nil
		}
	}

	// No valid suffix found
	return 0, fmt.Errorf("invalid size format '%s': must end with B, KB, MB, or GB", sizeStr)
}

// formatBytes converts bytes to human-readable format
// Example: 52428800 -> "50.00 MB"
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), sizes[exp])
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
// CARD DETECTION WITH REGEX
// ============================================================================

// initCardPatterns creates all card detection patterns
// Each regex handles: formatting (spaces/dashes), length, and issuer identification
func initCardPatterns() {
	cardPatterns = []CardPattern{
		// VISA
		// Starts with 4
		// Length: 13, 16, or 19 digits
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

// ============================================================================
// LUHN ALGORITHM VALIDATION
// ============================================================================

// How Luhn Algorithm Works:
// - Start from the rightmost digit
// - Double every second digit (from right)
// - If doubled value > 9, subtract 9
// - Sum all digits
// - Valid if sum is divisible by 10

// validateLuhn checks if a card number passes the Luhn algorithm
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

// ============================================================================
// REPORTING FUNCTIONS
// ============================================================================

// initReport initializes a new report
func initReport(directory string, extensions []string) {
	currentReport = &Report{
		ScanDate:        time.Now(),
		Directory:       directory,
		Extensions:      extensions,
		GroupedFindings: make(map[string][]CardFinding), // Initialize the map
	}
}

// addFinding adds a finding to the current report
// Findings are automatically grouped by file path
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

	// Initialize GroupedFindings map if needed
	if currentReport.GroupedFindings == nil {
		currentReport.GroupedFindings = make(map[string][]CardFinding)
	}

	// Add directly to grouped findings
	currentReport.GroupedFindings[filepath] = append(currentReport.GroupedFindings[filepath], finding)
}

// finalizeReport sets the final statistics
func finalizeReport(totalFiles int, scannedFiles int, duration time.Duration) {
	if currentReport == nil {
		return
	}

	currentReport.TotalFiles = totalFiles
	currentReport.ScannedFiles = scannedFiles
	currentReport.Duration = duration

	// Sort findings within each file by line number
	for filePath := range currentReport.GroupedFindings {
		findings := currentReport.GroupedFindings[filePath]
		// Simple bubble sort by line number
		for i := 0; i < len(findings); i++ {
			for j := i + 1; j < len(findings); j++ {
				if findings[j].LineNumber < findings[i].LineNumber {
					findings[i], findings[j] = findings[j], findings[i]
				}
			}
		}
		currentReport.GroupedFindings[filePath] = findings
	}

	// Calculate statistics
	calculateStatistics()
}

// calculateStatistics generates statistics from grouped findings
func calculateStatistics() {
	if currentReport == nil || len(currentReport.GroupedFindings) == 0 {
		return
	}

	stats := Statistics{
		CardsByType: make(map[string]int),
		FilesByType: make(map[string]int),
		TopFiles:    []FileStats{},
	}

	stats.FilesWithCards = len(currentReport.GroupedFindings)

	// Process each file with findings
	fileStatsList := []FileStats{}
	totalCards := 0

	for filePath, findings := range currentReport.GroupedFindings {
		fileExt := strings.ToLower(filepath.Ext(filePath))
		stats.FilesByType[fileExt]++

		// Count cards by type in this file
		cardTypes := make(map[string]int)
		for _, finding := range findings {
			stats.CardsByType[finding.CardType]++
			cardTypes[finding.CardType]++
			totalCards++
		}

		// Create file stats
		fs := FileStats{
			FilePath:  filePath,
			CardCount: len(findings),
			CardTypes: cardTypes,
		}
		fileStatsList = append(fileStatsList, fs)

		// Risk assessment based on number of cards
		if len(findings) >= 5 {
			stats.HighRiskFiles++
		} else if len(findings) >= 2 {
			stats.MediumRiskFiles++
		} else {
			stats.LowRiskFiles++
		}
	}

	// Sort files by card count (descending)
	for i := 0; i < len(fileStatsList); i++ {
		for j := i + 1; j < len(fileStatsList); j++ {
			if fileStatsList[j].CardCount > fileStatsList[i].CardCount {
				fileStatsList[i], fileStatsList[j] = fileStatsList[j], fileStatsList[i]
			}
		}
	}

	// Keep top 10 files
	if len(fileStatsList) > 10 {
		stats.TopFiles = fileStatsList[:10]
	} else {
		stats.TopFiles = fileStatsList
	}

	currentReport.Statistics = stats
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
	case ".xml":
		return exportXML(filename)
	default:
		return fmt.Errorf("unsupported format: %s", ext)
	}
}

// exportJSON exports report as JSON with grouped findings and statistics
func exportJSON(filename string) error {
	// Helper function to count total cards
	totalCards := 0
	for _, findings := range currentReport.GroupedFindings {
		totalCards += len(findings)
	}

	// Create clean JSON structure
	type JSONReport struct {
		Version  string `json:"version"`
		ScanInfo struct {
			ScanDate     string `json:"scan_date"`
			Directory    string `json:"directory"`
			Duration     string `json:"duration"`
			TotalFiles   int    `json:"total_files"`
			ScannedFiles int    `json:"scanned_files"`
		} `json:"scan_info"`
		Summary struct {
			TotalCards      int `json:"total_cards"`
			FilesWithCards  int `json:"files_with_cards"`
			HighRiskFiles   int `json:"high_risk_files"`
			MediumRiskFiles int `json:"medium_risk_files"`
			LowRiskFiles    int `json:"low_risk_files"`
		} `json:"summary"`
		Statistics struct {
			CardsByType map[string]int `json:"cards_by_type"`
			FilesByType map[string]int `json:"files_by_type"`
			TopFiles    []FileStats    `json:"top_files"`
		} `json:"statistics"`
		Findings map[string][]CardFinding `json:"findings"`
	}

	report := JSONReport{}
	report.Version = "1.1.0"
	report.ScanInfo.ScanDate = currentReport.ScanDate.Format(time.RFC3339)
	report.ScanInfo.Directory = currentReport.Directory
	report.ScanInfo.Duration = currentReport.Duration.String()
	report.ScanInfo.TotalFiles = currentReport.TotalFiles
	report.ScanInfo.ScannedFiles = currentReport.ScannedFiles

	report.Summary.TotalCards = totalCards
	report.Summary.FilesWithCards = currentReport.Statistics.FilesWithCards
	report.Summary.HighRiskFiles = currentReport.Statistics.HighRiskFiles
	report.Summary.MediumRiskFiles = currentReport.Statistics.MediumRiskFiles
	report.Summary.LowRiskFiles = currentReport.Statistics.LowRiskFiles

	report.Statistics.CardsByType = currentReport.Statistics.CardsByType
	report.Statistics.FilesByType = currentReport.Statistics.FilesByType
	report.Statistics.TopFiles = currentReport.Statistics.TopFiles

	report.Findings = currentReport.GroupedFindings

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// exportXML exports report as XML with statistics and grouped findings
func exportXML(filename string) error {
	// Calculate total cards
	totalCards := 0
	for _, findings := range currentReport.GroupedFindings {
		totalCards += len(findings)
	}

	// XML-friendly structures (maps don't work in XML)
	type XMLCardType struct {
		Name  string `xml:"name,attr"`
		Count int    `xml:"count,attr"`
	}

	type XMLFileType struct {
		Extension string `xml:"extension,attr"`
		Count     int    `xml:"count,attr"`
	}

	// XML-friendly version of FileStats (no maps allowed)
	type XMLFileStats struct {
		FilePath  string        `xml:"FilePath"`
		CardCount int           `xml:"CardCount"`
		CardTypes []XMLCardType `xml:"CardTypes>CardType"` // Convert map to slice
	}

	type XMLStatistics struct {
		FilesWithCards  int            `xml:"FilesWithCards"`
		HighRiskFiles   int            `xml:"HighRiskFiles"`
		MediumRiskFiles int            `xml:"MediumRiskFiles"`
		LowRiskFiles    int            `xml:"LowRiskFiles"`
		CardsByType     []XMLCardType  `xml:"CardsByType>CardType"`
		FilesByType     []XMLFileType  `xml:"FilesByType>FileType"`
		TopFiles        []XMLFileStats `xml:"TopFiles>File"` // Use XML-friendly version
	}

	type XMLFileGroup struct {
		FilePath string        `xml:"path,attr"`
		Count    int           `xml:"count,attr"`
		Findings []CardFinding `xml:"Finding"`
	}

	type XMLReport struct {
		XMLName      xml.Name `xml:"ScanReport"`
		Version      string   `xml:"version,attr"`
		ScanDate     string   `xml:"ScanInfo>ScanDate"`
		Directory    string   `xml:"ScanInfo>Directory"`
		Duration     string   `xml:"ScanInfo>Duration"`
		TotalFiles   int      `xml:"ScanInfo>TotalFiles"`
		ScannedFiles int      `xml:"ScanInfo>ScannedFiles"`
		TotalCards   int      `xml:"Summary>TotalCards"`
		Statistics   XMLStatistics
		FileGroups   []XMLFileGroup `xml:"Findings>FileGroup"`
	}

	// Convert CardsByType map to slice
	cardTypes := []XMLCardType{}
	for cardType, count := range currentReport.Statistics.CardsByType {
		cardTypes = append(cardTypes, XMLCardType{
			Name:  cardType,
			Count: count,
		})
	}

	// Convert FilesByType map to slice
	fileTypes := []XMLFileType{}
	for fileExt, count := range currentReport.Statistics.FilesByType {
		fileTypes = append(fileTypes, XMLFileType{
			Extension: fileExt,
			Count:     count,
		})
	}

	// Convert TopFiles (with their CardTypes maps) to XML-friendly version
	topFiles := []XMLFileStats{}
	for _, fs := range currentReport.Statistics.TopFiles {
		// Convert the CardTypes map in each FileStats
		cardTypesInFile := []XMLCardType{}
		for cardType, count := range fs.CardTypes {
			cardTypesInFile = append(cardTypesInFile, XMLCardType{
				Name:  cardType,
				Count: count,
			})
		}

		topFiles = append(topFiles, XMLFileStats{
			FilePath:  fs.FilePath,
			CardCount: fs.CardCount,
			CardTypes: cardTypesInFile,
		})
	}

	// Build file groups
	fileGroups := []XMLFileGroup{}
	filePaths := make([]string, 0, len(currentReport.GroupedFindings))
	for filePath := range currentReport.GroupedFindings {
		filePaths = append(filePaths, filePath)
	}

	// Sort file paths
	for i := 0; i < len(filePaths); i++ {
		for j := i + 1; j < len(filePaths); j++ {
			if filePaths[j] < filePaths[i] {
				filePaths[i], filePaths[j] = filePaths[j], filePaths[i]
			}
		}
	}

	for _, filePath := range filePaths {
		findings := currentReport.GroupedFindings[filePath]
		fileGroups = append(fileGroups, XMLFileGroup{
			FilePath: filePath,
			Count:    len(findings),
			Findings: findings,
		})
	}

	// Create report
	xmlReport := XMLReport{
		Version:      "1.1.0",
		ScanDate:     currentReport.ScanDate.Format(time.RFC3339),
		Directory:    currentReport.Directory,
		Duration:     currentReport.Duration.String(),
		TotalFiles:   currentReport.TotalFiles,
		ScannedFiles: currentReport.ScannedFiles,
		TotalCards:   totalCards,
		Statistics: XMLStatistics{
			CardsByType:     cardTypes,
			FilesByType:     fileTypes,
			TopFiles:        topFiles, // Use converted version
			FilesWithCards:  currentReport.Statistics.FilesWithCards,
			HighRiskFiles:   currentReport.Statistics.HighRiskFiles,
			MediumRiskFiles: currentReport.Statistics.MediumRiskFiles,
			LowRiskFiles:    currentReport.Statistics.LowRiskFiles,
		},
		FileGroups: fileGroups,
	}

	// Create file and write
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write XML header
	_, err = file.WriteString(xml.Header)
	if err != nil {
		return err
	}

	// Create encoder with indentation
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	// Encode
	err = encoder.Encode(xmlReport)
	if err != nil {
		return err
	}

	// Final newline
	_, err = file.WriteString("\n")
	return err
}

// exportCSV exports report as CSV with summary and grouped findings
func exportCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Calculate total cards
	totalCards := 0
	for _, findings := range currentReport.GroupedFindings {
		totalCards += len(findings)
	}

	// Summary section
	writer.Write([]string{"BasicPanScanner Report - Version 1.1.0"})
	writer.Write([]string{""})
	writer.Write([]string{"SCAN INFORMATION"})
	writer.Write([]string{"Scan Date", currentReport.ScanDate.Format("2006-01-02 15:04:05")})
	writer.Write([]string{"Directory", currentReport.Directory})
	writer.Write([]string{"Duration", currentReport.Duration.String()})
	writer.Write([]string{"Total Files", fmt.Sprintf("%d", currentReport.TotalFiles)})
	writer.Write([]string{"Scanned Files", fmt.Sprintf("%d", currentReport.ScannedFiles)})
	writer.Write([]string{""})

	writer.Write([]string{"SUMMARY"})
	writer.Write([]string{"Total Cards Found", fmt.Sprintf("%d", totalCards)})
	writer.Write([]string{"Files with Cards", fmt.Sprintf("%d", currentReport.Statistics.FilesWithCards)})
	writer.Write([]string{"High Risk Files (5+ cards)", fmt.Sprintf("%d", currentReport.Statistics.HighRiskFiles)})
	writer.Write([]string{"Medium Risk Files (2-4 cards)", fmt.Sprintf("%d", currentReport.Statistics.MediumRiskFiles)})
	writer.Write([]string{"Low Risk Files (1 card)", fmt.Sprintf("%d", currentReport.Statistics.LowRiskFiles)})
	writer.Write([]string{""})

	// Card type distribution
	if len(currentReport.Statistics.CardsByType) > 0 {
		writer.Write([]string{"CARD TYPE DISTRIBUTION"})
		writer.Write([]string{"Card Type", "Count", "Percentage"})
		for cardType, count := range currentReport.Statistics.CardsByType {
			percentage := float64(count) / float64(totalCards) * 100
			writer.Write([]string{cardType, fmt.Sprintf("%d", count), fmt.Sprintf("%.1f%%", percentage)})
		}
		writer.Write([]string{""})
	}

	// Top files
	if len(currentReport.Statistics.TopFiles) > 0 {
		writer.Write([]string{"TOP FILES BY CARD COUNT"})
		writer.Write([]string{"Rank", "File Path", "Card Count", "Risk Level"})
		for i, fs := range currentReport.Statistics.TopFiles {
			if i >= 10 {
				break
			}
			risk := "Low"
			if fs.CardCount >= 5 {
				risk = "High"
			} else if fs.CardCount >= 2 {
				risk = "Medium"
			}
			writer.Write([]string{
				fmt.Sprintf("%d", i+1),
				fs.FilePath,
				fmt.Sprintf("%d", fs.CardCount),
				risk,
			})
		}
		writer.Write([]string{""})
	}

	// Detailed findings (grouped by file)
	writer.Write([]string{"DETAILED FINDINGS"})
	writer.Write([]string{""})

	// Sort file paths
	filePaths := make([]string, 0, len(currentReport.GroupedFindings))
	for filePath := range currentReport.GroupedFindings {
		filePaths = append(filePaths, filePath)
	}
	for i := 0; i < len(filePaths); i++ {
		for j := i + 1; j < len(filePaths); j++ {
			if filePaths[j] < filePaths[i] {
				filePaths[i], filePaths[j] = filePaths[j], filePaths[i]
			}
		}
	}

	// Write grouped findings
	for _, filePath := range filePaths {
		findings := currentReport.GroupedFindings[filePath]

		// File header
		writer.Write([]string{fmt.Sprintf("FILE: %s", filePath)})
		writer.Write([]string{"Cards Found", fmt.Sprintf("%d", len(findings))})
		writer.Write([]string{""})
		writer.Write([]string{"Line Number", "Card Type", "Masked Card", "Timestamp"})

		// Findings for this file
		for _, f := range findings {
			writer.Write([]string{
				fmt.Sprintf("%d", f.LineNumber),
				f.CardType,
				f.MaskedCard,
				f.Timestamp.Format("2006-01-02 15:04:05"),
			})
		}
		writer.Write([]string{""})
	}

	return nil
}

// exportHTML exports report as HTML with accordion, statistics, and card icons
func exportHTML(filename string) error {
	var html strings.Builder

	// Calculate total cards
	totalCards := 0
	for _, findings := range currentReport.GroupedFindings {
		totalCards += len(findings)
	}

	// Helper function to get card icon
	getCardIcon := func(cardType string) string {
		icons := map[string]string{
			"Visa":       "💳", // Credit card emoji
			"MasterCard": "💳",
			"Amex":       "💳",
			"Discover":   "💳",
			"Diners":     "💳",
			"JCB":        "💳",
			"UnionPay":   "💳",
			"Maestro":    "💳",
			"RuPay":      "💳",
			"Troy":       "💳",
			"Mir":        "💳",
		}
		if icon, exists := icons[cardType]; exists {
			return icon
		}
		return "💳"
	}

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PAN Scanner Report - BasicPanScanner v1.1.0</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 {
            font-size: 36px;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .header .version {
            font-size: 14px;
            opacity: 0.9;
        }
        .content {
            padding: 40px;
        }
        
        /* Executive Summary */
        .executive-summary {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
        }
        .executive-summary h2 {
            font-size: 24px;
            margin-bottom: 20px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .executive-summary p {
            font-size: 16px;
            line-height: 1.6;
            margin-bottom: 10px;
        }
        .executive-summary .highlight {
            font-weight: bold;
            font-size: 18px;
        }
        
        /* Summary Cards */
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
            transition: transform 0.3s ease;
        }
        .summary-card:hover {
            transform: translateY(-5px);
        }
        .summary-label {
            font-size: 13px;
            opacity: 0.9;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .summary-value {
            font-size: 32px;
            font-weight: bold;
            margin-top: 10px;
        }
        
        /* Statistics Section */
        .stats-section {
            margin: 30px 0;
        }
        .stats-section h2 {
            color: #2c3e50;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid #667eea;
            font-size: 24px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
            gap: 25px;
            margin-bottom: 30px;
        }
        .stats-card {
            background: #f8f9fa;
            padding: 25px;
            border-radius: 12px;
            border-left: 4px solid #667eea;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .stats-card h3 {
            color: #2c3e50;
            margin-bottom: 20px;
            font-size: 18px;
        }
        
        /* Risk Badges */
        .risk-item {
            margin: 12px 0;
            display: flex;
            align-items: center;
            gap: 15px;
        }
        .badge {
            display: inline-block;
            padding: 6px 16px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
            min-width: 100px;
            text-align: center;
        }
        .badge-high { 
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
        }
        .badge-medium { 
            background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
            color: #8b4513;
        }
        .badge-low { 
            background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
            color: #2d7a6e;
        }
        
        /* Bar Chart */
        .bar-chart {
            margin-top: 15px;
        }
        .bar-item {
            margin: 12px 0;
        }
        .bar-label {
            display: flex;
            justify-content: space-between;
            margin-bottom: 5px;
            font-size: 14px;
            color: #555;
        }
        .bar-bg {
            width: 100%;
            height: 24px;
            background: #e0e0e0;
            border-radius: 12px;
            overflow: hidden;
            position: relative;
        }
        .bar-fill {
            height: 100%;
            background: linear-gradient(90deg, #667eea, #764ba2);
            border-radius: 12px;
            transition: width 0.6s ease;
            display: flex;
            align-items: center;
            justify-content: flex-end;
            padding-right: 10px;
            color: white;
            font-size: 12px;
            font-weight: bold;
        }
        
        /* Top Files List */
        .top-files-list {
            margin-top: 15px;
        }
        .top-file-item {
            background: white;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            display: flex;
            justify-content: space-between;
            align-items: center;
            transition: transform 0.2s ease;
        }
        .top-file-item:hover {
            transform: translateX(5px);
        }
        .file-name {
            font-weight: 500;
            color: #2c3e50;
        }
        
        /* Accordion Findings */
        .findings-section {
            margin-top: 40px;
        }
        .accordion-item {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            margin: 12px 0;
            overflow: hidden;
            transition: all 0.3s ease;
        }
        .accordion-item:hover {
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .accordion-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 18px 25px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
            user-select: none;
            transition: background 0.3s ease;
        }
        .accordion-header:hover {
            background: linear-gradient(135deg, #5568d3 0%, #6a4291 100%);
        }
        .accordion-header .file-path {
            font-weight: 500;
            font-size: 15px;
            flex: 1;
        }
        .accordion-header .card-count {
            margin: 0 15px;
        }
        .accordion-header .toggle-icon {
            font-size: 20px;
            transition: transform 0.3s ease;
        }
        .accordion-item.active .toggle-icon {
            transform: rotate(180deg);
        }
        .accordion-body {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.4s ease;
            background: #f8f9fa;
        }
        .accordion-item.active .accordion-body {
            max-height: 2000px;
        }
        .accordion-content {
            padding: 20px 25px;
        }
        
        /* Finding Item */
        .finding-item {
            background: white;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            display: grid;
            grid-template-columns: 80px 120px 1fr;
            gap: 20px;
            align-items: center;
            transition: transform 0.2s ease;
        }
        .finding-item:hover {
            transform: translateX(5px);
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .finding-line {
            color: #7f8c8d;
            font-family: 'Courier New', monospace;
            font-weight: bold;
        }
        .finding-type {
            display: flex;
            align-items: center;
            gap: 8px;
            font-weight: 600;
            color: #2c3e50;
        }
        .finding-type .icon {
            font-size: 20px;
        }
        .finding-card {
            font-family: 'Courier New', monospace;
            color: #e74c3c;
            font-weight: bold;
            font-size: 15px;
        }
        
        /* No Findings */
        .no-findings {
            text-align: center;
            padding: 60px 20px;
            color: #27ae60;
            font-size: 20px;
            background: #eafaf1;
            border-radius: 12px;
            margin: 30px 0;
        }
        .no-findings .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        
        /* Footer */
        footer {
            background: #2c3e50;
            color: white;
            padding: 30px;
            text-align: center;
            margin-top: 40px;
        }
        footer p {
            margin: 5px 0;
            opacity: 0.9;
        }
        
        /* Responsive */
        @media (max-width: 768px) {
            .content {
                padding: 20px;
            }
            .summary-grid {
                grid-template-columns: 1fr;
            }
            .stats-grid {
                grid-template-columns: 1fr;
            }
            .finding-item {
                grid-template-columns: 1fr;
                gap: 10px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 BasicPanScanner Security Report</h1>
            <div class="version">Version 1.1.0 | PCI Compliance Scanner</div>
        </div>
        
        <div class="content">`)

	// Executive Summary
	riskLevel := "Low"
	riskColor := "#27ae60"
	if currentReport.Statistics.HighRiskFiles > 0 {
		riskLevel = "High"
		riskColor = "#e74c3c"
	} else if currentReport.Statistics.MediumRiskFiles > 0 {
		riskLevel = "Medium"
		riskColor = "#f39c12"
	}

	html.WriteString(fmt.Sprintf(`
            <div class="executive-summary">
                <h2>📋 Executive Summary</h2>
                <p>Scan completed on <span class="highlight">%s</span> covering <span class="highlight">%d files</span> in directory <strong>%s</strong>.</p>
                <p>Found <span class="highlight" style="font-size: 22px;">%d credit card numbers</span> across <span class="highlight">%d files</span>.</p>
                <p>Overall Risk Level: <span class="highlight" style="color: %s;">%s</span></p>`,
		currentReport.ScanDate.Format("January 2, 2006"),
		currentReport.ScannedFiles,
		currentReport.Directory,
		totalCards,
		currentReport.Statistics.FilesWithCards,
		riskColor,
		riskLevel))

	if totalCards > 0 {
		html.WriteString(`
                <p style="margin-top: 15px; padding-top: 15px; border-top: 1px solid rgba(255,255,255,0.3);">
                    <strong>Recommendation:</strong> Immediate remediation required. Review all flagged files and remove or encrypt sensitive card data according to PCI DSS guidelines.
                </p>`)
	}

	html.WriteString(`
            </div>`)

	// Summary Cards
	html.WriteString(`
            <div class="summary-grid">
                <div class="summary-card">
                    <div class="summary-label">Scan Duration</div>
                    <div class="summary-value">` + currentReport.Duration.Round(time.Second).String() + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Files Scanned</div>
                    <div class="summary-value">` + fmt.Sprintf("%d", currentReport.ScannedFiles) + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Cards Found</div>
                    <div class="summary-value">` + fmt.Sprintf("%d", totalCards) + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Affected Files</div>
                    <div class="summary-value">` + fmt.Sprintf("%d", currentReport.Statistics.FilesWithCards) + `</div>
                </div>
            </div>`)

	// Statistics Section
	if totalCards > 0 {
		html.WriteString(`
            <div class="stats-section">
                <h2>📊 Detailed Statistics</h2>
                <div class="stats-grid">`)

		// Card Type Distribution
		if len(currentReport.Statistics.CardsByType) > 0 {
			html.WriteString(`
                    <div class="stats-card">
                        <h3>Card Type Distribution</h3>
                        <div class="bar-chart">`)

			for cardType, count := range currentReport.Statistics.CardsByType {
				percentage := float64(count) / float64(totalCards) * 100
				html.WriteString(fmt.Sprintf(`
                            <div class="bar-item">
                                <div class="bar-label">
                                    <span>%s %s</span>
                                    <span>%d (%.1f%%)</span>
                                </div>
                                <div class="bar-bg">
                                    <div class="bar-fill" style="width: %.1f%%"></div>
                                </div>
                            </div>`, getCardIcon(cardType), cardType, count, percentage, percentage))
			}

			html.WriteString(`
                        </div>
                    </div>`)
		}

		// Risk Assessment
		html.WriteString(`
                    <div class="stats-card">
                        <h3>Risk Assessment</h3>
                        <div class="risk-item">
                            <span class="badge badge-high">HIGH RISK</span>
                            <span>` + fmt.Sprintf("%d files with 5+ cards", currentReport.Statistics.HighRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-medium">MEDIUM RISK</span>
                            <span>` + fmt.Sprintf("%d files with 2-4 cards", currentReport.Statistics.MediumRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-low">LOW RISK</span>
                            <span>` + fmt.Sprintf("%d files with 1 card", currentReport.Statistics.LowRiskFiles) + `</span>
                        </div>
                    </div>`)

		html.WriteString(`
                </div>`)

		// Top Files
		if len(currentReport.Statistics.TopFiles) > 0 {
			html.WriteString(`
                <div class="stats-card">
                    <h3>🎯 Top Files by Card Count</h3>
                    <div class="top-files-list">`)

			for i, fs := range currentReport.Statistics.TopFiles {
				if i >= 5 {
					break
				}

				badgeClass := "badge-low"
				if fs.CardCount >= 5 {
					badgeClass = "badge-high"
				} else if fs.CardCount >= 2 {
					badgeClass = "badge-medium"
				}

				html.WriteString(fmt.Sprintf(`
                        <div class="top-file-item">
                            <span class="file-name">%d. %s</span>
                            <span class="badge %s">%d cards</span>
                        </div>`, i+1, filepath.Base(fs.FilePath), badgeClass, fs.CardCount))
			}

			html.WriteString(`
                    </div>
                </div>`)
		}

		html.WriteString(`
            </div>`)

		// Detailed Findings (Accordion)
		html.WriteString(`
            <div class="findings-section">
                <h2>🔎 Detailed Findings (Click to Expand)</h2>`)

		// Sort file paths
		filePaths := make([]string, 0, len(currentReport.GroupedFindings))
		for filePath := range currentReport.GroupedFindings {
			filePaths = append(filePaths, filePath)
		}
		for i := 0; i < len(filePaths); i++ {
			for j := i + 1; j < len(filePaths); j++ {
				if filePaths[j] < filePaths[i] {
					filePaths[i], filePaths[j] = filePaths[j], filePaths[i]
				}
			}
		}

		// Create accordion items
		for _, filePath := range filePaths {
			findings := currentReport.GroupedFindings[filePath]

			badgeClass := "badge-low"
			if len(findings) >= 5 {
				badgeClass = "badge-high"
			} else if len(findings) >= 2 {
				badgeClass = "badge-medium"
			}

			html.WriteString(fmt.Sprintf(`
                <div class="accordion-item">
                    <div class="accordion-header" onclick="toggleAccordion(this)">
                        <span class="file-path">%s</span>
                        <span class="badge %s card-count">%d cards</span>
                        <span class="toggle-icon">▼</span>
                    </div>
                    <div class="accordion-body">
                        <div class="accordion-content">`, filePath, badgeClass, len(findings)))

			for _, finding := range findings {
				html.WriteString(fmt.Sprintf(`
                            <div class="finding-item">
                                <div class="finding-line">Line %d</div>
                                <div class="finding-type">
                                    <span class="icon">%s</span>
                                    <span>%s</span>
                                </div>
                                <div class="finding-card">%s</div>
                            </div>`, finding.LineNumber, getCardIcon(finding.CardType), finding.CardType, finding.MaskedCard))
			}

			html.WriteString(`
                        </div>
                    </div>
                </div>`)
		}

		html.WriteString(`
            </div>`)
	} else {
		html.WriteString(`
            <div class="no-findings">
                <div class="icon">✅</div>
                <div>No credit card numbers found in scanned files</div>
                <p style="font-size: 16px; margin-top: 10px; opacity: 0.8;">Your files are compliant with PCI DSS requirements.</p>
            </div>`)
	}

	html.WriteString(`
        </div>
        
        <footer>
            <p><strong>BasicPanScanner v1.1.0</strong> | PCI Compliance Tool</p>
            <p>Report generated on ` + time.Now().Format("January 2, 2006 at 15:04:05") + `</p>
            <p style="margin-top: 10px; font-size: 12px;">Supports: Visa, Mastercard, Amex, Discover, Diners Club, JCB, UnionPay, Maestro, RuPay, Troy, Mir</p>
        </footer>
    </div>
    
    <script>
        // Accordion toggle function
        function toggleAccordion(header) {
            const item = header.parentElement;
            const wasActive = item.classList.contains('active');
            
            // Close all accordion items
            document.querySelectorAll('.accordion-item').forEach(accordionItem => {
                accordionItem.classList.remove('active');
            });
            
            // Open clicked item if it wasn't active
            if (!wasActive) {
                item.classList.add('active');
            }
        }
        
        // Optional: Add keyboard navigation
        document.addEventListener('DOMContentLoaded', function() {
            const headers = document.querySelectorAll('.accordion-header');
            headers.forEach(header => {
                header.setAttribute('tabindex', '0');
                header.addEventListener('keypress', function(e) {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        toggleAccordion(this);
                    }
                });
            });
        });
    </script>
</body>
</html>`)

	return os.WriteFile(filename, []byte(html.String()), 0644)
}

// exportTXT exports report as plain text with improved formatting
func exportTXT(filename string) error {
	var content strings.Builder

	// Calculate total cards from grouped findings
	totalCards := 0
	for _, findings := range currentReport.GroupedFindings {
		totalCards += len(findings)
	}

	// Header
	content.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	content.WriteString("║          BASICPANSCANNER SECURITY REPORT                   ║\n")
	content.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	// Scan Information
	content.WriteString("SCAN INFORMATION\n")
	content.WriteString(strings.Repeat("─", 60) + "\n")
	content.WriteString(fmt.Sprintf("Date:           %s\n", currentReport.ScanDate.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Directory:      %s\n", currentReport.Directory))
	content.WriteString(fmt.Sprintf("Duration:       %s\n", currentReport.Duration.Round(time.Second)))
	content.WriteString(fmt.Sprintf("Files Scanned:  %d / %d (%.1f%%)\n",
		currentReport.ScannedFiles,
		currentReport.TotalFiles,
		float64(currentReport.ScannedFiles)/float64(currentReport.TotalFiles)*100))
	content.WriteString("\n")

	// Summary Statistics
	content.WriteString("SUMMARY\n")
	content.WriteString(strings.Repeat("─", 60) + "\n")
	content.WriteString(fmt.Sprintf("Total Cards Found:     %d\n", totalCards))
	content.WriteString(fmt.Sprintf("Files with Cards:      %d\n", currentReport.Statistics.FilesWithCards))
	content.WriteString(fmt.Sprintf("Unique Card Types:     %d\n", len(currentReport.Statistics.CardsByType)))
	content.WriteString("\n")

	// Risk Assessment
	if totalCards > 0 {
		content.WriteString("RISK ASSESSMENT\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")
		content.WriteString(fmt.Sprintf("🔴 High Risk Files:    %d (5+ cards)\n", currentReport.Statistics.HighRiskFiles))
		content.WriteString(fmt.Sprintf("🟡 Medium Risk Files:  %d (2-4 cards)\n", currentReport.Statistics.MediumRiskFiles))
		content.WriteString(fmt.Sprintf("🟢 Low Risk Files:     %d (1 card)\n", currentReport.Statistics.LowRiskFiles))
		content.WriteString("\n")
	}

	// Card Type Distribution
	if len(currentReport.Statistics.CardsByType) > 0 {
		content.WriteString("CARD TYPE DISTRIBUTION\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		for cardType, count := range currentReport.Statistics.CardsByType {
			percentage := float64(count) / float64(totalCards) * 100
			// Create simple bar chart
			bars := int(percentage / 5) // Each bar = 5%
			barChart := strings.Repeat("█", bars)
			content.WriteString(fmt.Sprintf("%-15s %3d (%5.1f%%) %s\n",
				cardType, count, percentage, barChart))
		}
		content.WriteString("\n")
	}

	// File Type Distribution
	if len(currentReport.Statistics.FilesByType) > 0 {
		content.WriteString("FILE TYPE DISTRIBUTION\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		for fileType, count := range currentReport.Statistics.FilesByType {
			content.WriteString(fmt.Sprintf("%-10s %d files\n", fileType, count))
		}
		content.WriteString("\n")
	}

	// Top Files
	if len(currentReport.Statistics.TopFiles) > 0 {
		content.WriteString("TOP FILES BY CARD COUNT\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		for i, fs := range currentReport.Statistics.TopFiles {
			if i >= 5 { // Show top 5
				break
			}

			// Risk indicator
			risk := "🟢"
			if fs.CardCount >= 5 {
				risk = "🔴"
			} else if fs.CardCount >= 2 {
				risk = "🟡"
			}

			content.WriteString(fmt.Sprintf("%s #%d: %s\n", risk, i+1, filepath.Base(fs.FilePath)))
			content.WriteString(fmt.Sprintf("    Cards: %d", fs.CardCount))

			// Show card type breakdown
			if len(fs.CardTypes) > 0 {
				content.WriteString(" (")
				first := true
				for cardType, count := range fs.CardTypes {
					if !first {
						content.WriteString(", ")
					}
					content.WriteString(fmt.Sprintf("%s: %d", cardType, count))
					first = false
				}
				content.WriteString(")")
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Detailed Findings (Grouped by File)
	if len(currentReport.GroupedFindings) > 0 {
		content.WriteString("DETAILED FINDINGS\n")
		content.WriteString(strings.Repeat("═", 60) + "\n\n")

		// Sort files by path for consistent output
		filePaths := make([]string, 0, len(currentReport.GroupedFindings))
		for filePath := range currentReport.GroupedFindings {
			filePaths = append(filePaths, filePath)
		}
		// Simple bubble sort
		for i := 0; i < len(filePaths); i++ {
			for j := i + 1; j < len(filePaths); j++ {
				if filePaths[j] < filePaths[i] {
					filePaths[i], filePaths[j] = filePaths[j], filePaths[i]
				}
			}
		}

		// Print findings grouped by file
		for fileNum, filePath := range filePaths {
			findings := currentReport.GroupedFindings[filePath]

			// File header
			content.WriteString(fmt.Sprintf("[File %d] %s\n", fileNum+1, filePath))
			content.WriteString(fmt.Sprintf("         %d card(s) found\n", len(findings)))
			content.WriteString(strings.Repeat("─", 60) + "\n")

			// List findings in this file
			for i, finding := range findings {
				isLast := i == len(findings)-1
				prefix := "├─"
				if isLast {
					prefix = "└─"
				}

				content.WriteString(fmt.Sprintf("%s Line %4d: %-12s %s\n",
					prefix,
					finding.LineNumber,
					finding.CardType,
					finding.MaskedCard))
			}
			content.WriteString("\n")
		}
	} else {
		content.WriteString("\nNo credit card numbers found. ✓\n\n")
	}

	// Footer
	content.WriteString(strings.Repeat("═", 60) + "\n")
	content.WriteString("Report generated by BasicPanScanner v1.1.0\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// ============================================================================
// FILE SCANNING
// ============================================================================

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

// ============================================================================
// DIRECTORY SCANNING
// ============================================================================

// scanDirectoryWithOptionsConcurrent scans directory with goroutines
func scanDirectoryWithOptionsConcurrent(dirPath string, outputFile string, extensions []string, excludeDirs []string, maxFileSize int64, workers int) error {
	fmt.Printf("\nScanning directory: %s\n", dirPath)
	fmt.Printf("Workers: %d (concurrent scanning enabled)\n", workers)

	// Show max file size setting
	if maxFileSize > 0 {
		fmt.Printf("Max file size: %s\n", formatBytes(maxFileSize))
	} else {
		fmt.Printf("Max file size: unlimited\n")
	}

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

		// Skip large files BEFORE adding to scan list
		if maxFileSize > 0 && info.Size() > maxFileSize {
			mu.Lock()
			skippedFiles++
			// Debug logging for skipped files (show first 5)
			if skippedFiles <= 5 {
				fmt.Printf("⊘ Skipping (too large): %s (%s)\n",
					filepath.Base(path),
					formatBytes(info.Size()))
			}
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

	// Show skipped file summary
	if skippedFiles > 5 {
		fmt.Printf("⊘ ... and %d more files skipped (too large)\n", skippedFiles-5)
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
		fmt.Printf("  Skipped (size): %d\n", skippedFiles)
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

	// Show max file size
	if maxFileSize > 0 {
		fmt.Printf("Max file size: %s\n", formatBytes(maxFileSize))
	} else {
		fmt.Printf("Max file size: unlimited\n")
	}

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
			// Debug logging
			if skippedFiles <= 5 {
				fmt.Printf("⊘ Skipping (too large): %s (%s)\n",
					filepath.Base(path),
					formatBytes(info.Size()))
			}
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

	// Show skipped summary
	if skippedFiles > 5 {
		fmt.Printf("\n⊘ ... and %d more files skipped (too large)\n", skippedFiles-5)
	}

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
		fmt.Printf("  Skipped (size): %d\n", skippedFiles)
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

// ============================================================================
// UI FUNCTIONS
// ============================================================================

func showHelp() {
	fmt.Println(`
BasicPanScanner v1.1.0 - PCI Compliance Scanner
Usage: ./scanner -path <directory> [options]

Required:
    -path <directory>      Directory to scan

Options:
    -output <file>         Save results (.json, .csv, .html, .txt, .xml)
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

    # Full scan with XML output
    ./scanner -path /data -workers 4 -output report.xml

    # Full scan with JSON output
    ./scanner -path /data -workers 4 -output report.json

Performance:
    Default workers: CPU cores / 2 (safe for production)
    Max workers: CPU cores (automatically limited)
    More workers = faster scanning (2-4x speed improvement)

Supported Card Issuers (11):
    Visa, Mastercard, Amex, Discover, Diners Club, JCB,
    UnionPay, Maestro, RuPay, Troy, Mir

Configuration:
    Edit config.json to change default settings.
    CLI flags always override config values.

Export Formats:
    .json  - JSON format (machine-readable)
    .csv   - CSV format (spreadsheet import)
    .txt   - Plain text format (human-readable)
    .html  - HTML format (browser viewing)
    .xml   - XML format (enterprise systems)
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

// ============================================================================
// MAIN
// ============================================================================

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

	// Display banner
	displayBanner()

	// Initialize card detection patterns (MUST be called before scanning)
	initCardPatterns()

	// Load config file (use defaults if it fails)
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Warning: Could not load config.json: %v\n", err)
		fmt.Println("Using default settings\n")
		config = &Config{
			Extensions:  []string{".txt", ".log", ".csv"},
			ExcludeDirs: []string{".git", "node_modules"},
			MaxFileSize: "50MB",
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
