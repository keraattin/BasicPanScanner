// Package report - Plain text exporter
// Exports reports in human-readable plain text format
package report

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// TXTExporter exports reports in plain text format
// Plain text is ideal for:
//   - Human reading
//   - Email reports
//   - Terminal viewing
//   - Simple archiving
type TXTExporter struct{}

// Export implements the Exporter interface for plain text format
// Creates a nicely formatted text file with ASCII art and tables
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .txt)
//
// Returns:
//   - error: Error if file can't be written
func (e *TXTExporter) Export(report *Report, filename string) error {
	var content strings.Builder

	// ============================================================
	// HEADER
	// ============================================================

	content.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	content.WriteString("║          BASICPANSCANNER SECURITY REPORT                   ║\n")
	content.WriteString("╚════════════════════════════════════════════════════════════╝\n\n")

	// ============================================================
	// SCAN INFORMATION
	// ============================================================

	content.WriteString("SCAN INFORMATION\n")
	content.WriteString(strings.Repeat("─", 60) + "\n")
	content.WriteString(fmt.Sprintf("Date:           %s\n", report.ScanDate.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("Directory:      %s\n", report.Directory))
	content.WriteString(fmt.Sprintf("Duration:       %s\n", report.GetFormattedDuration())) // Use formatted duration
	content.WriteString(fmt.Sprintf("Files Scanned:  %d / %d (%.1f%%)\n",
		report.ScannedFiles,
		report.TotalFiles,
		float64(report.ScannedFiles)/float64(report.TotalFiles)*100))
	content.WriteString("\n")

	// ============================================================
	// SUMMARY STATISTICS
	// ============================================================

	content.WriteString("SUMMARY\n")
	content.WriteString(strings.Repeat("─", 60) + "\n")
	content.WriteString(fmt.Sprintf("Total Cards Found:     %d\n", report.CardsFound))
	content.WriteString(fmt.Sprintf("Files with Cards:      %d\n", report.Statistics.FilesWithCards))
	content.WriteString(fmt.Sprintf("Unique Card Types:     %d\n", len(report.Statistics.CardsByType)))
	content.WriteString("\n")

	// ============================================================
	// RISK ASSESSMENT
	// ============================================================

	if report.CardsFound > 0 {
		content.WriteString("RISK ASSESSMENT\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")
		content.WriteString(fmt.Sprintf("🔴 High Risk Files:    %d (5+ cards)\n", report.Statistics.HighRiskFiles))
		content.WriteString(fmt.Sprintf("🟡 Medium Risk Files:  %d (2-4 cards)\n", report.Statistics.MediumRiskFiles))
		content.WriteString(fmt.Sprintf("🟢 Low Risk Files:     %d (1 card)\n", report.Statistics.LowRiskFiles))
		content.WriteString("\n")
	}

	// ============================================================
	// CARD TYPE DISTRIBUTION
	// ============================================================

	if len(report.Statistics.CardsByType) > 0 {
		content.WriteString("CARD TYPE DISTRIBUTION\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		// Sort by count (descending)
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

		// Display with simple bar charts
		for _, cc := range counts {
			percentage := float64(cc.count) / float64(report.CardsFound) * 100
			bars := int(percentage / 5) // Each bar = 5%
			barChart := strings.Repeat("█", bars)
			content.WriteString(fmt.Sprintf("%-15s %3d (%5.1f%%) %s\n",
				cc.name, cc.count, percentage, barChart))
		}
		content.WriteString("\n")
	}

	// ============================================================
	// FILE TYPE DISTRIBUTION
	// ============================================================

	if len(report.Statistics.FilesByType) > 0 {
		content.WriteString("FILE TYPE DISTRIBUTION\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		// Sort by count (descending)
		type fileCount struct {
			ext   string
			count int
		}
		var counts []fileCount
		for fileType, count := range report.Statistics.FilesByType {
			counts = append(counts, fileCount{fileType, count})
		}
		sort.Slice(counts, func(i, j int) bool {
			return counts[i].count > counts[j].count
		})

		// Display top 10
		for i, fc := range counts {
			if i >= 10 {
				break
			}
			content.WriteString(fmt.Sprintf("%-10s %d files\n", fc.ext, fc.count))
		}
		content.WriteString("\n")
	}

	// ============================================================
	// TOP FILES
	// ============================================================

	if len(report.Statistics.TopFiles) > 0 {
		content.WriteString("TOP FILES BY CARD COUNT\n")
		content.WriteString(strings.Repeat("─", 60) + "\n")

		for i, fs := range report.Statistics.TopFiles {
			if i >= 5 {
				break
			}

			// Risk indicator
			risk := "🟢"
			if fs.CardCount >= 5 {
				risk = "🔴"
			} else if fs.CardCount >= 2 {
				risk = "🟡"
			}

			content.WriteString(fmt.Sprintf("%s #%d: %s\n", risk, i+1, fs.FilePath))
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

	// ============================================================
	// DETAILED FINDINGS (Grouped by File)
	// ============================================================

	if len(report.GroupedByFile) > 0 {
		content.WriteString("DETAILED FINDINGS\n")
		content.WriteString(strings.Repeat("═", 60) + "\n\n")

		// Sort files by path
		var filePaths []string
		for filePath := range report.GroupedByFile {
			filePaths = append(filePaths, filePath)
		}
		sort.Strings(filePaths)

		// Print findings grouped by file
		for fileNum, filePath := range filePaths {
			findings := report.GroupedByFile[filePath]

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

	// ============================================================
	// FOOTER
	// ============================================================

	content.WriteString(strings.Repeat("═", 60) + "\n")
	content.WriteString(fmt.Sprintf("Report generated by BasicPanScanner v%s\n", report.Version))
	content.WriteString(fmt.Sprintf("Generated: %s\n", report.ScanDate.Format("2006-01-02 15:04:05")))

	// Write to file
	return os.WriteFile(filename, []byte(content.String()), 0644)
}
