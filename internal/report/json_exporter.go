// Package report - JSON exporter
// Exports reports in JSON format for machine-readable output
package report

import (
	"encoding/json"
	"os"
)

// JSONExporter exports reports in JSON format
// JSON is ideal for:
//   - Machine-readable output
//   - API integration
//   - Further processing with other tools
//   - Web applications
type JSONExporter struct{}

// Export implements the Exporter interface for JSON format
// Creates a well-formatted JSON file with all scan information
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .json)
//
// Returns:
//   - error: Error if file can't be written or JSON encoding fails
//
// Example output structure:
//
//	{
//	  "version": "3.0.0",
//	  "scan_info": {
//	    "scan_date": "2025-01-15T10:30:00Z",
//	    "directory": "/var/log",
//	    "duration": "1m23s"
//	  },
//	  "summary": {
//	    "total_cards": 12,
//	    "files_with_cards": 3
//	  },
//	  "statistics": {...},
//	  "findings": {...}
//	}
func (e *JSONExporter) Export(report *Report, filename string) error {
	// Create a clean JSON structure
	// This provides a well-organized format that's easy to parse
	type jsonReport struct {
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
		Findings map[string][]struct {
			LineNumber int    `json:"line_number"`
			CardType   string `json:"card_type"`
			MaskedCard string `json:"masked_card"`
			Timestamp  string `json:"timestamp"`
		} `json:"findings"`
	}

	// Build the JSON structure
	jr := jsonReport{}
	jr.Version = report.Version
	jr.ScanInfo.ScanDate = report.ScanDate.Format("2006-01-02T15:04:05Z07:00")
	jr.ScanInfo.Directory = report.Directory
	jr.ScanInfo.Duration = report.Duration.String()
	jr.ScanInfo.TotalFiles = report.TotalFiles
	jr.ScanInfo.ScannedFiles = report.ScannedFiles

	jr.Summary.TotalCards = report.CardsFound
	jr.Summary.FilesWithCards = report.Statistics.FilesWithCards
	jr.Summary.HighRiskFiles = report.Statistics.HighRiskFiles
	jr.Summary.MediumRiskFiles = report.Statistics.MediumRiskFiles
	jr.Summary.LowRiskFiles = report.Statistics.LowRiskFiles

	jr.Statistics.CardsByType = report.Statistics.CardsByType
	jr.Statistics.FilesByType = report.Statistics.FilesByType
	jr.Statistics.TopFiles = report.Statistics.TopFiles

	// Convert findings
	jr.Findings = make(map[string][]struct {
		LineNumber int    `json:"line_number"`
		CardType   string `json:"card_type"`
		MaskedCard string `json:"masked_card"`
		Timestamp  string `json:"timestamp"`
	})

	for filePath, findings := range report.GroupedByFile {
		var fileFindings []struct {
			LineNumber int    `json:"line_number"`
			CardType   string `json:"card_type"`
			MaskedCard string `json:"masked_card"`
			Timestamp  string `json:"timestamp"`
		}

		for _, f := range findings {
			fileFindings = append(fileFindings, struct {
				LineNumber int    `json:"line_number"`
				CardType   string `json:"card_type"`
				MaskedCard string `json:"masked_card"`
				Timestamp  string `json:"timestamp"`
			}{
				LineNumber: f.LineNumber,
				CardType:   f.CardType,
				MaskedCard: f.MaskedCard,
				Timestamp:  f.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		jr.Findings[filePath] = fileFindings
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filename, data, 0644)
}
