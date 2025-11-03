// Package report - Beautiful Native PDF exporter (NO EXTERNAL LIBRARIES!)
// Creates professional, colorful PDFs using only GO standard library
package report

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ============================================================
// BEAUTIFUL PDF EXPORTER - ZERO DEPENDENCIES!
// ============================================================
// This creates professional PDFs similar to HTML reports
// Features:
//   ✅ Colors (blue header, colored sections)
//   ✅ Multiple pages with automatic page breaks
//   ✅ Multiple fonts (Helvetica, Bold, Italic)
//   ✅ Boxes and borders
//   ✅ Professional layout
//   ✅ Beautiful typography
//   ✅ ALL using only GO standard library!

type PDFExporter struct {
	currentY      float64  // Current Y position on page
	pageHeight    float64  // Page height in points
	marginTop     float64  // Top margin
	marginBottom  float64  // Bottom margin
	pageNumber    int      // Current page number
	objectNumber  int      // Current PDF object number
	objects       []string // PDF objects
	objectOffsets []int    // Object byte offsets
}

// ============================================================
// MAIN EXPORT FUNCTION
// ============================================================
func (e *PDFExporter) Export(report *Report, filename string) error {
	// Initialize PDF state
	e.pageHeight = 792 // US Letter height
	e.marginTop = 50
	e.marginBottom = 50
	e.currentY = e.pageHeight - e.marginTop
	e.pageNumber = 1
	e.objectNumber = 0
	e.objects = []string{}
	e.objectOffsets = []int{}

	var pdf strings.Builder

	// ============================================================
	// PDF HEADER
	// ============================================================
	pdf.WriteString("%PDF-1.4\n")
	pdf.WriteString("%âãÏÓ\n\n")

	// ============================================================
	// CREATE ALL CONTENT PAGES
	// ============================================================
	contentPages := e.buildAllPages(report)

	// ============================================================
	// CATALOG OBJECT
	// ============================================================
	e.addObject(&pdf, fmt.Sprintf("<< /Type /Catalog /Pages 2 0 R >>"))

	// ============================================================
	// PAGES OBJECT
	// ============================================================
	// Build list of page references
	pageRefs := ""
	for i := 0; i < len(contentPages); i++ {
		if i > 0 {
			pageRefs += " "
		}
		pageRefs += fmt.Sprintf("%d 0 R", 3+i*3) // Page objects are 3, 6, 9, ...
	}

	e.addObject(&pdf, fmt.Sprintf(
		"<< /Type /Pages /Kids [%s] /Count %d /MediaBox [0 0 612 792] >>",
		pageRefs, len(contentPages)))

	// ============================================================
	// CREATE PAGE OBJECTS
	// ============================================================
	for i, content := range contentPages {
		pageNum := 3 + i*3
		contentNum := pageNum + 1
		fontsNum := pageNum + 2

		// Page object
		e.addObject(&pdf, fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /Contents %d 0 R /Resources << /Font %d 0 R >> >>",
			contentNum, fontsNum))

		// Content stream
		e.addObject(&pdf, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream",
			len(content), content))

		// Font resources (multiple fonts for beautiful typography)
		e.addObject(&pdf, `<< 
			/F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
			/F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>
			/F3 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>
			/F4 << /Type /Font /Subtype /Type1 /BaseFont /Times-Roman >>
		>>`)
	}

	// ============================================================
	// CROSS-REFERENCE TABLE
	// ============================================================
	xrefStart := pdf.Len()
	pdf.WriteString("xref\n")
	pdf.WriteString(fmt.Sprintf("0 %d\n", len(e.objectOffsets)+1))
	pdf.WriteString("0000000000 65535 f \n")
	for _, offset := range e.objectOffsets {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// ============================================================
	// TRAILER
	// ============================================================
	pdf.WriteString("\ntrailer\n")
	pdf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(e.objectOffsets)+1))
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	pdf.WriteString("%%EOF\n")

	// Write to file
	return os.WriteFile(filename, []byte(pdf.String()), 0644)
}

// ============================================================
// ADD PDF OBJECT HELPER
// ============================================================
func (e *PDFExporter) addObject(pdf *strings.Builder, content string) {
	e.objectNumber++
	e.objectOffsets = append(e.objectOffsets, pdf.Len())
	pdf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n\n", e.objectNumber, content))
}

// ============================================================
// BUILD ALL PAGES
// ============================================================
func (e *PDFExporter) buildAllPages(report *Report) []string {
	var pages []string
	var currentPage strings.Builder

	// ============================================================
	// PAGE 1: HEADER + EXECUTIVE SUMMARY
	// ============================================================
	e.currentY = e.pageHeight - e.marginTop

	// Draw blue header
	e.addBlueHeader(&currentPage, report)

	// Executive Summary Section
	e.addExecutiveSummary(&currentPage, report)

	// Start Statistics
	e.addSectionHeader(&currentPage, "STATISTICS")
	e.addStatisticsOverview(&currentPage, report)

	// Check if we need new page
	if e.currentY < 200 {
		pages = append(pages, currentPage.String())
		currentPage.Reset()
		e.currentY = e.pageHeight - e.marginTop
	}

	// Card Distribution
	e.addCardDistribution(&currentPage, report)

	// ============================================================
	// PAGE 2: TOP FILES
	// ============================================================
	if e.currentY < 250 {
		pages = append(pages, currentPage.String())
		currentPage.Reset()
		e.currentY = e.pageHeight - e.marginTop
	}

	e.addSectionHeader(&currentPage, "TOP FILES BY CARD COUNT")
	e.addTopFiles(&currentPage, report)

	// ============================================================
	// REMAINING PAGES: DETAILED FINDINGS
	// ============================================================
	if e.currentY < 200 {
		pages = append(pages, currentPage.String())
		currentPage.Reset()
		e.currentY = e.pageHeight - e.marginTop
	}

	e.addSectionHeader(&currentPage, "DETAILED FINDINGS")

	// Add findings with automatic page breaks
	fileCount := 0
	for filePath, findings := range report.GroupedByFile {
		// Check if we need new page
		if e.currentY < 150 {
			pages = append(pages, currentPage.String())
			currentPage.Reset()
			e.currentY = e.pageHeight - e.marginTop
		}

		e.addFileFindings(&currentPage, filePath, len(findings), fileCount)
		fileCount++

		// Limit to first 20 files to keep PDF reasonable
		if fileCount >= 20 {
			break
		}
	}

	// Add footer to last page
	e.addFooter(&currentPage)

	// Add the last page
	pages = append(pages, currentPage.String())

	return pages
}

// ============================================================
// DRAW BLUE HEADER (Like HTML report!)
// ============================================================
func (e *PDFExporter) addBlueHeader(page *strings.Builder, report *Report) {
	// Blue color: RGB(52, 152, 219) -> PDF: 0.204 0.596 0.859
	page.WriteString("q\n") // Save graphics state

	// Draw blue rectangle
	page.WriteString("0.204 0.596 0.859 rg\n")                         // Set fill color (blue)
	page.WriteString(fmt.Sprintf("0 %f 612 70 re f\n", e.currentY-70)) // Rectangle

	// White text
	page.WriteString("1 1 1 rg\n") // White color
	page.WriteString("BT\n")
	page.WriteString("/F2 24 Tf\n") // Bold, 24pt
	page.WriteString(fmt.Sprintf("100 %f Td\n", e.currentY-35))
	page.WriteString("(BASICPANSCANNER) Tj\n")
	page.WriteString("ET\n")

	// Subtitle
	page.WriteString("BT\n")
	page.WriteString("/F1 12 Tf\n") // Regular, 12pt
	page.WriteString(fmt.Sprintf("130 %f Td\n", e.currentY-52))
	page.WriteString("(Security Report - Credit Card Discovery) Tj\n")
	page.WriteString("ET\n")

	// Version (small)
	page.WriteString("BT\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString(fmt.Sprintf("220 %f Td\n", e.currentY-65))
	page.WriteString(fmt.Sprintf("(Version %s) Tj\n", e.escape(report.Version)))
	page.WriteString("ET\n")

	page.WriteString("Q\n") // Restore graphics state

	e.currentY -= 90 // Move down past header
}

// ============================================================
// EXECUTIVE SUMMARY BOX
// ============================================================
func (e *PDFExporter) addExecutiveSummary(page *strings.Builder, report *Report) {
	// Light blue background box
	page.WriteString("q\n")
	page.WriteString("0.85 0.92 0.98 rg\n") // Light blue
	page.WriteString(fmt.Sprintf("40 %f 532 100 re f\n", e.currentY-100))

	// Border
	page.WriteString("0.204 0.596 0.859 RG\n") // Blue border
	page.WriteString("2 w\n")                  // 2pt width
	page.WriteString(fmt.Sprintf("40 %f 532 100 re S\n", e.currentY-100))
	page.WriteString("Q\n")

	// Title
	page.WriteString("BT\n")
	page.WriteString("0.204 0.596 0.859 rg\n") // Blue text
	page.WriteString("/F2 14 Tf\n")            // Bold
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY-25))
	page.WriteString("(EXECUTIVE SUMMARY) Tj\n")
	page.WriteString("ET\n")

	// Content in black
	page.WriteString("0 0 0 rg\n") // Black
	y := e.currentY - 45

	info := []struct {
		label string
		value string
	}{
		{"Scan Date:", report.ScanDate.Format("January 2, 2006 at 3:04 PM")},
		{"Directory:", e.truncate(report.Directory, 60)},
		{"Duration:", e.formatDuration(report.Duration)},
		{"Files Scanned:", fmt.Sprintf("%d / %d (%.1f%%)", report.ScannedFiles, report.TotalFiles,
			float64(report.ScannedFiles)/float64(report.TotalFiles)*100)},
	}

	for _, item := range info {
		page.WriteString("BT\n")
		page.WriteString("/F2 10 Tf\n") // Bold label
		page.WriteString(fmt.Sprintf("50 %f Td\n", y))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(item.label)))
		page.WriteString("ET\n")

		page.WriteString("BT\n")
		page.WriteString("/F1 10 Tf\n") // Regular value
		page.WriteString(fmt.Sprintf("140 %f Td\n", y))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(item.value)))
		page.WriteString("ET\n")

		y -= 18
	}

	e.currentY -= 120
}

// ============================================================
// SECTION HEADER (Blue background)
// ============================================================
func (e *PDFExporter) addSectionHeader(page *strings.Builder, title string) {
	e.currentY -= 20

	// Blue bar
	page.WriteString("q\n")
	page.WriteString("0.204 0.596 0.859 rg\n")
	page.WriteString(fmt.Sprintf("40 %f 532 25 re f\n", e.currentY-25))
	page.WriteString("Q\n")

	// White text
	page.WriteString("BT\n")
	page.WriteString("1 1 1 rg\n")  // White
	page.WriteString("/F2 12 Tf\n") // Bold
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY-17))
	page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(title)))
	page.WriteString("ET\n")

	e.currentY -= 35
}

// ============================================================
// STATISTICS OVERVIEW (Colored boxes)
// ============================================================
func (e *PDFExporter) addStatisticsOverview(page *strings.Builder, report *Report) {
	stats := []struct {
		label string
		value string
		color string // RGB in PDF format
	}{
		{"Total Cards Found", fmt.Sprintf("%d", report.CardsFound), "0.9 0.3 0.3"},               // Red
		{"Files With Cards", fmt.Sprintf("%d", report.Statistics.FilesWithCards), "0.2 0.7 0.9"}, // Blue
		{"High-Risk Files", fmt.Sprintf("%d", report.Statistics.HighRiskFiles), "0.8 0.2 0.2"},   // Dark red
	}

	x := 50.0
	for _, stat := range stats {
		// Colored box
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", stat.color))
		page.WriteString(fmt.Sprintf("%f %f 160 50 re f\n", x, e.currentY-50))
		page.WriteString("Q\n")

		// White text - Value (large)
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 20 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", x+10, e.currentY-30))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", stat.value))
		page.WriteString("ET\n")

		// White text - Label (small)
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F1 9 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", x+10, e.currentY-45))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(stat.label)))
		page.WriteString("ET\n")

		x += 175
	}

	e.currentY -= 70
}

// ============================================================
// CARD DISTRIBUTION (with visual bars)
// ============================================================
func (e *PDFExporter) addCardDistribution(page *strings.Builder, report *Report) {
	e.currentY -= 15

	// Title
	page.WriteString("BT\n")
	page.WriteString("0 0 0 rg\n")
	page.WriteString("/F2 11 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
	page.WriteString("(Card Distribution by Type) Tj\n")
	page.WriteString("ET\n")

	e.currentY -= 25

	// Sort cards by count
	type cardCount struct {
		name  string
		count int
	}
	var cards []cardCount
	for name, count := range report.Statistics.CardsByType {
		cards = append(cards, cardCount{name, count})
	}
	// Bubble sort
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[j].count > cards[i].count {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	// Draw each card type with colored bar
	maxCount := 0
	if len(cards) > 0 {
		maxCount = cards[0].count
	}

	for _, card := range cards {
		// Card name
		page.WriteString("BT\n")
		page.WriteString("0 0 0 rg\n")
		page.WriteString("/F2 10 Tf\n")
		page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%s:) Tj\n", e.escape(card.name)))
		page.WriteString("ET\n")

		// Count
		page.WriteString("BT\n")
		page.WriteString("/F1 10 Tf\n")
		page.WriteString(fmt.Sprintf("180 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%d cards) Tj\n", card.count))
		page.WriteString("ET\n")

		// Visual bar (blue)
		barWidth := 250.0
		if maxCount > 0 {
			barWidth = (float64(card.count) / float64(maxCount)) * 250.0
		}
		page.WriteString("q\n")
		page.WriteString("0.204 0.596 0.859 rg\n")
		page.WriteString(fmt.Sprintf("250 %f %f 12 re f\n", e.currentY-2, barWidth))
		page.WriteString("Q\n")

		e.currentY -= 20
	}
}

// ============================================================
// TOP FILES
// ============================================================
func (e *PDFExporter) addTopFiles(page *strings.Builder, report *Report) {
	if len(report.Statistics.TopFiles) == 0 {
		page.WriteString("BT\n")
		page.WriteString("0 0 0 rg\n")
		page.WriteString("/F3 10 Tf\n") // Italic
		page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
		page.WriteString("(No files with cards found.) Tj\n")
		page.WriteString("ET\n")
		e.currentY -= 30
		return
	}

	maxFiles := 10
	if len(report.Statistics.TopFiles) < maxFiles {
		maxFiles = len(report.Statistics.TopFiles)
	}

	for i := 0; i < maxFiles; i++ {
		fs := report.Statistics.TopFiles[i]

		// Risk indicator (colored circle)
		color := "0.2 0.8 0.2" // Green
		if fs.CardCount >= 5 {
			color = "0.9 0.2 0.2" // Red
		} else if fs.CardCount >= 2 {
			color = "0.9 0.7 0.2" // Yellow
		}

		// Draw circle
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", color))
		page.WriteString(fmt.Sprintf("50 %f 5 0 360 arc f\n", e.currentY-3))
		page.WriteString("Q\n")

		// File number and name
		page.WriteString("BT\n")
		page.WriteString("0 0 0 rg\n")
		page.WriteString("/F2 10 Tf\n")
		page.WriteString(fmt.Sprintf("65 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%d. %s) Tj\n", i+1, e.escape(e.truncate(fs.FilePath, 55))))
		page.WriteString("ET\n")

		// Card count (right aligned)
		page.WriteString("BT\n")
		page.WriteString("/F1 10 Tf\n")
		page.WriteString(fmt.Sprintf("520 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%d cards) Tj\n", fs.CardCount))
		page.WriteString("ET\n")

		e.currentY -= 18
	}

	e.currentY -= 10
}

// ============================================================
// FILE FINDINGS (Detailed)
// ============================================================
// Note: We don't need to know the Finding structure details
// We only need to count them, so we use interface{} for flexibility
func (e *PDFExporter) addFileFindings(page *strings.Builder, filePath string, findingsCount int, index int) {
	// Gray box background
	page.WriteString("q\n")
	page.WriteString("0.95 0.95 0.95 rg\n")
	page.WriteString(fmt.Sprintf("40 %f 532 30 re f\n", e.currentY-30))
	page.WriteString("Q\n")

	// File number and path
	page.WriteString("BT\n")
	page.WriteString("0 0 0 rg\n")
	page.WriteString("/F2 10 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY-18))
	page.WriteString(fmt.Sprintf("(File %d: %s) Tj\n", index+1, e.escape(e.truncate(filePath, 70))))
	page.WriteString("ET\n")

	e.currentY -= 40

	// Cards count
	page.WriteString("BT\n")
	page.WriteString("0.3 0.3 0.3 rg\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
	page.WriteString(fmt.Sprintf("(Cards found: %d) Tj\n", findingsCount))
	page.WriteString("ET\n")

	e.currentY -= 25
}

// ============================================================
// FOOTER
// ============================================================
func (e *PDFExporter) addFooter(page *strings.Builder) {
	page.WriteString("BT\n")
	page.WriteString("0.5 0.5 0.5 rg\n") // Gray
	page.WriteString("/F3 8 Tf\n")       // Italic, small
	page.WriteString(fmt.Sprintf("50 30 Td\n"))
	page.WriteString(fmt.Sprintf("(Generated: %s) Tj\n",
		e.escape(time.Now().Format("January 2, 2006 at 3:04 PM MST"))))
	page.WriteString("ET\n")
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func (e *PDFExporter) escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func (e *PDFExporter) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

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
