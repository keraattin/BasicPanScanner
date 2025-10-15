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
