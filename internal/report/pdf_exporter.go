// Package report - Native PDF exporter (NO EXTERNAL LIBRARIES!)
// Exports reports in PDF format using only GO standard library
package report

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ============================================================
// PDF EXPORTER STRUCT (Native - No External Dependencies!)
// ============================================================
// PDFExporter exports reports in PDF format
// This version uses ONLY GO's standard library
// No external dependencies needed!
//
// PDF is ideal for:
//   - Professional reports
//   - Archiving and compliance
//   - Print-ready documents
//   - Sharing with management
type PDFExporter struct{}

// ============================================================
// MAIN EXPORT FUNCTION
// ============================================================
// Export implements the Exporter interface for PDF format
// Creates a PDF using ONLY GO standard library!
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .pdf)
//
// Returns:
//   - error: Error if PDF can't be created or written
func (e *PDFExporter) Export(report *Report, filename string) error {
	// Build the PDF content step by step
	var pdf strings.Builder

	// Track the byte position of each object
	// This is needed for the cross-reference table
	objectPositions := make(map[int]int)

	// ============================================================
	// STEP 1: PDF HEADER
	// ============================================================
	// Every PDF must start with this header
	// It tells readers: "I'm a PDF version 1.4"
	pdf.WriteString("%PDF-1.4\n")
	pdf.WriteString("%âãÏÓ\n") // Binary comment (helps identify as binary)
	pdf.WriteString("\n")

	// ============================================================
	// STEP 2: CATALOG OBJECT (Object 1)
	// ============================================================
	// The Catalog is the root of the PDF
	// It points to the Pages object
	objectPositions[1] = pdf.Len()
	pdf.WriteString("1 0 obj\n")
	pdf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	pdf.WriteString("endobj\n\n")

	// ============================================================
	// STEP 3: PAGES OBJECT (Object 2)
	// ============================================================
	// The Pages object lists all pages in the document
	// We'll create multiple pages if needed
	objectPositions[2] = pdf.Len()
	pdf.WriteString("2 0 obj\n")
	pdf.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 ")
	pdf.WriteString("/MediaBox [0 0 612 792] >>\n") // US Letter size
	pdf.WriteString("endobj\n\n")

	// ============================================================
	// STEP 4: PAGE OBJECT (Object 3)
	// ============================================================
	// This defines a single page
	// It references the content (Object 4) and font (Object 5)
	objectPositions[3] = pdf.Len()
	pdf.WriteString("3 0 obj\n")
	pdf.WriteString("<< /Type /Page /Parent 2 0 R ")
	pdf.WriteString("/Contents 4 0 R ")
	pdf.WriteString("/Resources << /Font << /F1 5 0 R >> >> >>\n")
	pdf.WriteString("endobj\n\n")

	// ============================================================
	// STEP 5: CONTENT STREAM (Object 4)
	// ============================================================
	// This is where we write the actual text content
	// We'll build the content string first
	content := e.buildPDFContent(report)

	objectPositions[4] = pdf.Len()
	pdf.WriteString("4 0 obj\n")
	pdf.WriteString(fmt.Sprintf("<< /Length %d >>\n", len(content)))
	pdf.WriteString("stream\n")
	pdf.WriteString(content)
	pdf.WriteString("\nendstream\n")
	pdf.WriteString("endobj\n\n")

	// ============================================================
	// STEP 6: FONT OBJECT (Object 5)
	// ============================================================
	// Defines the font we'll use (Helvetica)
	// Helvetica is a standard PDF font (always available)
	objectPositions[5] = pdf.Len()
	pdf.WriteString("5 0 obj\n")
	pdf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	pdf.WriteString("endobj\n\n")

	// ============================================================
	// STEP 7: CROSS-REFERENCE TABLE
	// ============================================================
	// The xref table tells the PDF reader where each object is located
	// This is like an index for the PDF
	xrefStart := pdf.Len()
	pdf.WriteString("xref\n")
	pdf.WriteString("0 6\n")                 // 6 objects (0-5)
	pdf.WriteString("0000000000 65535 f \n") // Object 0 is always free

	// Write the position of each object
	for i := 1; i <= 5; i++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", objectPositions[i]))
	}

	// ============================================================
	// STEP 8: TRAILER
	// ============================================================
	// The trailer tells where to find the xref table and catalog
	pdf.WriteString("\ntrailer\n")
	pdf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	pdf.WriteString("%%EOF\n")

	// ============================================================
	// STEP 9: WRITE TO FILE
	// ============================================================
	// Write the PDF content to disk
	err := os.WriteFile(filename, []byte(pdf.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
}

// ============================================================
// BUILD PDF CONTENT
// ============================================================
// buildPDFContent creates the text content for the PDF
// This uses PDF's content stream syntax
//
// PDF Content Stream Commands:
//
//	BT    = Begin Text
//	ET    = End Text
//	Tf    = Set font and size
//	Td    = Move to position
//	Tj    = Show text
//	TL    = Set leading (line spacing)
//	T*    = Move to next line
//
// Parameters:
//   - report: The report to format
//
// Returns:
//   - string: PDF content stream
func (e *PDFExporter) buildPDFContent(report *Report) string {
	var content strings.Builder

	// Start text block
	content.WriteString("BT\n")

	// Set font: Helvetica Bold, 16 point
	content.WriteString("/F1 16 Tf\n")

	// Position: X=50, Y=750 (coordinates from bottom-left)
	// In PDF, (0,0) is bottom-left corner
	// US Letter is 612x792 points
	content.WriteString("50 750 Td\n")

	// Write title
	content.WriteString("(BASICPANSCANNER SECURITY REPORT) Tj\n")

	// Set line spacing (leading) to 20 points
	content.WriteString("20 TL\n")

	// Move to next line
	content.WriteString("T*\n")

	// Change to smaller font for body text
	content.WriteString("/F1 10 Tf\n")
	content.WriteString("15 TL\n") // Smaller line spacing

	// Add scan information
	content.WriteString("T*\n") // Blank line
	content.WriteString(fmt.Sprintf("(Scan Date: %s) Tj\n",
		e.escapeString(report.ScanDate.Format("2006-01-02 15:04:05"))))
	content.WriteString("T*\n")
	content.WriteString(fmt.Sprintf("(Directory: %s) Tj\n",
		e.escapeString(report.Directory)))
	content.WriteString("T*\n")
	content.WriteString(fmt.Sprintf("(Duration: %s) Tj\n",
		e.formatDuration(report.Duration)))
	content.WriteString("T*\n")
	content.WriteString(fmt.Sprintf("(Files Scanned: %d / %d) Tj\n",
		report.ScannedFiles, report.TotalFiles))

	// Add statistics
	content.WriteString("T*\nT*\n")    // Two blank lines
	content.WriteString("/F1 12 Tf\n") // Slightly larger for section header
	content.WriteString("(STATISTICS) Tj\n")
	content.WriteString("/F1 10 Tf\n") // Back to normal size
	content.WriteString("T*\n")

	// FIXED: Use correct field names
	content.WriteString(fmt.Sprintf("(Total Cards Found: %d) Tj\n",
		report.CardsFound))
	content.WriteString("T*\n")
	content.WriteString(fmt.Sprintf("(Affected Files: %d) Tj\n",
		report.Statistics.FilesWithCards))
	content.WriteString("T*\n")
	content.WriteString(fmt.Sprintf("(High-Risk Files: %d) Tj\n",
		report.Statistics.HighRiskFiles))

	// Add card distribution
	content.WriteString("T*\nT*\n")
	content.WriteString("(Card Distribution:) Tj\n")

	// Sort card types by count
	type cardCount struct {
		cardType string
		count    int
	}
	var cardCounts []cardCount
	for cardType, count := range report.Statistics.CardsByType {
		cardCounts = append(cardCounts, cardCount{cardType, count})
	}
	// Simple sort (bubble sort)
	for i := 0; i < len(cardCounts); i++ {
		for j := i + 1; j < len(cardCounts); j++ {
			if cardCounts[j].count > cardCounts[i].count {
				cardCounts[i], cardCounts[j] = cardCounts[j], cardCounts[i]
			}
		}
	}

	for _, cc := range cardCounts {
		content.WriteString("T*\n")
		content.WriteString(fmt.Sprintf("(  %s: %d cards) Tj\n",
			e.escapeString(cc.cardType), cc.count))
	}

	// Add top files section
	content.WriteString("T*\nT*\n")
	content.WriteString("/F1 12 Tf\n")
	content.WriteString("(TOP FILES) Tj\n")
	content.WriteString("/F1 10 Tf\n")
	content.WriteString("T*\n")

	// Limit to first 5 files (simple PDF - limited space)
	maxFiles := 5
	if len(report.Statistics.TopFiles) < maxFiles {
		maxFiles = len(report.Statistics.TopFiles)
	}

	for i := 0; i < maxFiles; i++ {
		fileStats := report.Statistics.TopFiles[i]
		content.WriteString("T*\n")
		// FIXED: Use FilePath instead of Filename
		content.WriteString(fmt.Sprintf("(%d. %s - %d cards) Tj\n",
			i+1,
			e.escapeString(e.truncateString(fileStats.FilePath, 50)),
			fileStats.CardCount))
	}

	// Add findings summary
	content.WriteString("T*\nT*\n")
	content.WriteString("/F1 12 Tf\n")
	content.WriteString("(DETAILED FINDINGS) Tj\n")
	content.WriteString("/F1 10 Tf\n")
	content.WriteString("T*\n")

	// FIXED: Use GroupedByFile instead of GroupedFindings
	if len(report.GroupedByFile) == 0 {
		content.WriteString("(No credit cards found.) Tj\n")
	} else {
		content.WriteString(fmt.Sprintf("(Found cards in %d files - see full report for details) Tj\n",
			len(report.GroupedByFile)))

		// Show first few files as examples (space limited in simple PDF)
		maxFindings := 3
		fileCount := 0

		// FIXED: Iterate over GroupedByFile map
		for filePath, findings := range report.GroupedByFile {
			if fileCount >= maxFindings {
				break
			}

			content.WriteString("T*\n")
			content.WriteString(fmt.Sprintf("(File: %s) Tj\n",
				e.escapeString(e.truncateString(filePath, 50))))
			content.WriteString("T*\n")
			content.WriteString(fmt.Sprintf("(  Cards: %d) Tj\n",
				len(findings)))

			fileCount++
		}
	}

	// Add footer
	content.WriteString("T*\nT*\nT*\n")
	content.WriteString("/F1 8 Tf\n") // Small font for footer
	content.WriteString(fmt.Sprintf("(Generated: %s) Tj\n",
		time.Now().Format("2006-01-02 15:04:05 MST")))

	// End text block
	content.WriteString("ET\n")

	return content.String()
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// escapeString escapes special characters for PDF
// PDF strings use () to delimit, so we need to escape those
//
// Parameters:
//   - s: String to escape
//
// Returns:
//   - string: Escaped string safe for PDF
func (e *PDFExporter) escapeString(s string) string {
	// Replace special characters that break PDF syntax
	s = strings.ReplaceAll(s, "\\", "\\\\") // Backslash must be first!
	s = strings.ReplaceAll(s, "(", "\\(")   // Left paren
	s = strings.ReplaceAll(s, ")", "\\)")   // Right paren
	s = strings.ReplaceAll(s, "\r", "\\r")  // Carriage return
	s = strings.ReplaceAll(s, "\n", "\\n")  // Newline

	return s
}

// truncateString limits string length
// Prevents text from running off the page
//
// Parameters:
//   - s: String to truncate
//   - maxLen: Maximum length
//
// Returns:
//   - string: Truncated string
func (e *PDFExporter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats duration nicely
// Converts time.Duration to readable string
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Formatted duration
func (e *PDFExporter) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
