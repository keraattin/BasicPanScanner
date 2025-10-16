// Package report handles report generation and export
// This package takes scan results and generates reports in various formats
package report

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"../scanner"
)

// Report represents a complete scan report
// This is the main structure that holds all scan information and findings
type Report struct {
	// Metadata
	Version   string    // Scanner version (e.g., "3.0.0")
	ScanDate  time.Time // When the scan started
	Directory string    // Directory that was scanned

	// Configuration
	ScanMode   string   // "whitelist" or "blacklist"
	Extensions []string // Extensions that were scanned/excluded

	// Results
	TotalFiles    int // Total files found
	ScannedFiles  int // Files actually scanned
	SkippedBySize int // Files skipped due to size
	SkippedByExt  int // Files skipped by extension filter
	CardsFound    int // Total cards found

	// Timing
	Duration time.Duration // Total scan duration
	ScanRate float64       // Files per second

	// Findings
	Findings      []scanner.Finding            // All findings (flat list)
	GroupedByFile map[string][]scanner.Finding // Findings grouped by file

	// Statistics
	Statistics Statistics // Computed statistics
}

// Statistics holds computed statistics about the scan
// These are calculated after scanning completes
type Statistics struct {
	// Card distribution by type
	CardsByType map[string]int // {"Visa": 15, "Mastercard": 8}

	// File distribution by extension
	FilesByType map[string]int // {".txt": 50, ".log": 30}

	// Files with cards
	FilesWithCards int // Number of files containing at least one card

	// Top files by card count
	TopFiles []FileStats // Top 10 files with most cards

	// Risk assessment
	HighRiskFiles   int // Files with 5+ cards
	MediumRiskFiles int // Files with 2-4 cards
	LowRiskFiles    int // Files with 1 card
}

// FileStats holds statistics for a single file
// Used for "top files" reporting
type FileStats struct {
	FilePath  string         // File path
	CardCount int            // Number of cards in this file
	CardTypes map[string]int // Card type distribution in this file
}

// NewReport creates a new report from scan results
//
// Parameters:
//   - version: Scanner version
//   - directory: Directory that was scanned
//   - scanMode: "whitelist" or "blacklist"
//   - extensions: Extensions list
//   - result: Scan results from scanner package
//
// Returns:
//   - *Report: Initialized report with statistics
//
// Example:
//
//	report := report.NewReport(
//	    "3.0.0",
//	    "/var/log",
//	    "blacklist",
//	    []string{".exe", ".dll"},
//	    scanResult,
//	)
func NewReport(version, directory, scanMode string, extensions []string, result *scanner.ScanResult) *Report {
	rep := &Report{
		Version:       version,
		ScanDate:      time.Now(),
		Directory:     directory,
		ScanMode:      scanMode,
		Extensions:    extensions,
		TotalFiles:    result.TotalFiles,
		ScannedFiles:  result.ScannedFiles,
		SkippedBySize: result.SkippedBySize,
		SkippedByExt:  result.SkippedByExt,
		CardsFound:    result.CardsFound,
		Duration:      result.Duration,
		ScanRate:      result.ScanRate,
		Findings:      result.Findings,
		GroupedByFile: result.GroupedByFile,
	}

	// Calculate statistics
	rep.calculateStatistics()

	return rep
}

// calculateStatistics computes all statistics from the findings
// This analyzes the findings to generate useful metrics
func (r *Report) calculateStatistics() {
	stats := Statistics{
		CardsByType:    make(map[string]int),
		FilesByType:    make(map[string]int),
		FilesWithCards: len(r.GroupedByFile),
	}

	// ============================================================
	// Count cards by type and files by extension
	// ============================================================

	for filePath, findings := range r.GroupedByFile {
		// Get file extension
		ext := strings.ToLower(filepath.Ext(filePath))
		stats.FilesByType[ext]++

		// Count cards by type in this file
		cardTypesInFile := make(map[string]int)

		for _, finding := range findings {
			// Global count by type
			stats.CardsByType[finding.CardType]++

			// Per-file count
			cardTypesInFile[finding.CardType]++
		}

		// Determine risk level based on card count
		cardCount := len(findings)
		if cardCount >= 5 {
			stats.HighRiskFiles++
		} else if cardCount >= 2 {
			stats.MediumRiskFiles++
		} else {
			stats.LowRiskFiles++
		}
	}

	// ============================================================
	// Build "top files" list
	// ============================================================

	// Create list of file stats
	var fileStatsList []FileStats

	for filePath, findings := range r.GroupedByFile {
		// Count card types in this file
		cardTypes := make(map[string]int)
		for _, finding := range findings {
			cardTypes[finding.CardType]++
		}

		fileStatsList = append(fileStatsList, FileStats{
			FilePath:  filePath,
			CardCount: len(findings),
			CardTypes: cardTypes,
		})
	}

	// Sort by card count (descending) using simple bubble sort
	// For small lists (<100 files), this is fine
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

	r.Statistics = stats
}

// Export exports the report to the specified file format
// Format is determined by file extension
//
// Supported formats:
//   - .json - JSON format
//   - .csv  - CSV format
//   - .txt  - Plain text format
//   - .xml  - XML format
//   - .html - HTML format
//
// Parameters:
//   - filename: Output filename with extension
//
// Returns:
//   - error: Error if export fails
//
// Example:
//
//	err := report.Export("scan_results.html")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *Report) Export(filename string) error {
	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(filename))

	var exporter Exporter

	switch ext {
	case ".json":
		exporter = &JSONExporter{}
	case ".csv":
		exporter = &CSVExporter{}
	case ".txt":
		exporter = &TXTExporter{}
	case ".xml":
		exporter = &XMLExporter{}
	case ".html":
		exporter = &HTMLExporter{}
	default:
		return fmt.Errorf("unsupported format: %s (use .json, .csv, .txt, .xml, or .html)", ext)
	}

	// Use the exporter to write the report
	return exporter.Export(r, filename)
}

// GetRiskLevel returns the overall risk level of the scan
// Based on the number of high-risk files
//
// Returns:
//   - string: "High", "Medium", or "Low"
//   - string: Color code for display ("#e74c3c" for high, etc.)
func (r *Report) GetRiskLevel() (level string, color string) {
	if r.Statistics.HighRiskFiles > 0 {
		return "High", "#e74c3c" // Red
	}
	if r.Statistics.MediumRiskFiles > 0 {
		return "Medium", "#f39c12" // Orange
	}
	return "Low", "#27ae60" // Green
}

// GetFormattedDuration returns a human-readable duration string
// This converts Go's verbose duration format to something cleaner
//
// Returns:
//   - string: Formatted duration (e.g., "1h 23m 45s", "2m 15s", "3.5s", "250ms")
//
// Examples:
//
//	1h23m45.123456789s -> "1h 23m 45s"
//	2m15.5s            -> "2m 16s"
//	3.567s             -> "3.6s"
//	123.456ms          -> "123ms"
//	1.234ms            -> "1.2ms"
func (r *Report) GetFormattedDuration() string {
	return FormatDuration(r.Duration)
}

// FormatDuration formats a duration in a human-readable way
// This helper function can be used throughout the application
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Formatted duration
//
// Formatting rules:
//   - Hours + minutes + seconds: "1h 23m 45s"
//   - Minutes + seconds: "2m 15s"
//   - Seconds only (>= 10s): "45s"
//   - Seconds with decimal (1-10s): "3.5s"
//   - Milliseconds (>= 10ms): "123ms"
//   - Milliseconds with decimal (< 10ms): "1.2ms"
//   - Microseconds: "500µs"
//
// Examples:
//
//	5000000000000 ns (1h23m20s) -> "1h 23m 20s"
//	125000000000 ns (2m5s)      -> "2m 5s"
//	45000000000 ns (45s)        -> "45s"
//	3567000000 ns (3.567s)      -> "3.6s"
//	123456000 ns (123.456ms)    -> "123ms"
//	1234000 ns (1.234ms)        -> "1.2ms"
//	567000 ns (567µs)           -> "567µs"
func FormatDuration(d time.Duration) string {
	// Handle very short durations
	if d < time.Millisecond {
		// Less than 1ms, show microseconds
		us := float64(d.Microseconds())
		if us < 10 {
			return fmt.Sprintf("%.1fµs", us)
		}
		return fmt.Sprintf("%dµs", d.Microseconds())
	}

	if d < time.Second {
		// Less than 1 second, show milliseconds
		ms := float64(d.Milliseconds())
		if ms < 10 {
			return fmt.Sprintf("%.1fms", d.Seconds()*1000)
		}
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		// Less than 1 minute, show seconds
		s := d.Seconds()
		if s < 10 {
			return fmt.Sprintf("%.1fs", s)
		}
		return fmt.Sprintf("%ds", int(s))
	}

	if d < time.Hour {
		// Less than 1 hour, show minutes and seconds
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	// 1 hour or more, show hours, minutes, and seconds
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if minutes == 0 && seconds == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if seconds == 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}
