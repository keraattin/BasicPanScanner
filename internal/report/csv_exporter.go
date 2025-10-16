// Package report - CSV exporter
// Exports reports in CSV format for spreadsheet applications
package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
)

// CSVExporter exports reports in CSV format
// CSV is ideal for:
//   - Excel and spreadsheet applications
//   - Data analysis tools
//   - Easy viewing and filtering
//   - Database imports
type CSVExporter struct{}

// Export implements the Exporter interface for CSV format
// Creates a CSV file with summary and detailed findings
//
// CSV Structure:
//   - Header section with scan information
//   - Statistics section with card distribution
//   - Top files section
//   - Detailed findings section (grouped by file)
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .csv)
//
// Returns:
//   - error: Error if file can't be written
func (e *CSVExporter) Export(report *Report, filename string) error {
	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ============================================================
	// SECTION 1: Header and Summary
	// ============================================================

	writer.Write([]string{"BasicPanScanner Report - Version " + report.Version})
	writer.Write([]string{""}) // Empty line

	writer.Write([]string{"SCAN INFORMATION"})
	writer.Write([]string{"Scan Date", report.ScanDate.Format("2006-01-02 15:04:05")})
	writer.Write([]string{"Directory", report.Directory})
	writer.Write([]string{"Duration", report.GetFormattedDuration()}) // Use formatted duration
	writer.Write([]string{"Total Files", fmt.Sprintf("%d", report.TotalFiles)})
	writer.Write([]string{"Scanned Files", fmt.Sprintf("%d", report.ScannedFiles)})
	writer.Write([]string{""})

	writer.Write([]string{"SUMMARY"})
	writer.Write([]string{"Total Cards Found", fmt.Sprintf("%d", report.CardsFound)})
	writer.Write([]string{"Files with Cards", fmt.Sprintf("%d", report.Statistics.FilesWithCards)})
	writer.Write([]string{"High Risk Files (5+ cards)", fmt.Sprintf("%d", report.Statistics.HighRiskFiles)})
	writer.Write([]string{"Medium Risk Files (2-4 cards)", fmt.Sprintf("%d", report.Statistics.MediumRiskFiles)})
	writer.Write([]string{"Low Risk Files (1 card)", fmt.Sprintf("%d", report.Statistics.LowRiskFiles)})
	writer.Write([]string{""})

	// ============================================================
	// SECTION 2: Card Type Distribution
	// ============================================================

	if len(report.Statistics.CardsByType) > 0 {
		writer.Write([]string{"CARD TYPE DISTRIBUTION"})
		writer.Write([]string{"Card Type", "Count", "Percentage"})

		// Sort card types by count (descending)
		type cardCount struct {
			name  string
			count int
		}
		var counts []cardCount
		for cardType, count := range report.Statistics.CardsByType {
			counts = append(counts, cardCount{cardType, count})
		}
		sort.Slice(counts, func(i, j int) bool {
			return counts[i].count > counts[j].count
		})

		// Write card type rows
		for _, cc := range counts {
			percentage := float64(cc.count) / float64(report.CardsFound) * 100
			writer.Write([]string{
				cc.name,
				fmt.Sprintf("%d", cc.count),
				fmt.Sprintf("%.1f%%", percentage),
			})
		}
		writer.Write([]string{""})
	}

	// ============================================================
	// SECTION 3: Top Files
	// ============================================================

	if len(report.Statistics.TopFiles) > 0 {
		writer.Write([]string{"TOP FILES BY CARD COUNT"})
		writer.Write([]string{"Rank", "File Path", "Card Count", "Risk Level"})

		for i, fs := range report.Statistics.TopFiles {
			if i >= 10 {
				break
			}

			// Determine risk level
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

	// ============================================================
	// SECTION 4: Detailed Findings (Grouped by File)
	// ============================================================

	writer.Write([]string{"DETAILED FINDINGS"})
	writer.Write([]string{""})

	// Sort file paths for consistent output
	var filePaths []string
	for filePath := range report.GroupedByFile {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	// Write findings for each file
	for _, filePath := range filePaths {
		findings := report.GroupedByFile[filePath]

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
