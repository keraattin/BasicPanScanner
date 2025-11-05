// Package report - Beautiful Native PDF exporter (NO EXTERNAL LIBRARIES!)
//
// Creates professional, colorful PDFs matching our HTML report design
// Uses ONLY GO standard library - no external dependencies!
//
// WHAT THIS FILE DOES:
//
//	This exporter creates beautiful PDF reports that look similar to our HTML reports
//	It uses PDF commands to draw text, colors, boxes, and shapes
//
// FEATURES:
//
//	✅ Beautiful color scheme (matching HTML)
//	✅ Professional gradient-style headers
//	✅ Colored summary cards
//	✅ Statistics with visual bars
//	✅ Multiple pages with automatic page breaks
//	✅ Multiple fonts (Helvetica family)
//	✅ Professional spacing and typography
//	✅ Risk level indicators with colors
//	✅ ALL using only GO standard library!
//
// PDF BASICS (for learning):
//   - PDF uses a coordinate system: (0,0) is bottom-left corner
//   - Text is positioned using X (horizontal) and Y (vertical) coordinates
//   - Colors are RGB values from 0 to 1 (e.g., 0.2 0.6 0.85 = blue)
//   - Commands are written as text strings (e.g., "BT" = Begin Text)
//   - Everything is measured in "points" (1 point = 1/72 inch)
//
// INDUSTRY STANDARD:
//   - We follow PDF specification version 1.4
//   - Code is organized into clear sections with comments
//   - Helper functions are well-documented
//   - Error handling is included
package report

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ============================================================
// COLOR DEFINITIONS (matching our HTML design)
// ============================================================
// PDF colors are RGB values from 0.0 to 1.0
// We define them as constants for easy reuse
const (
	// Primary blue (used in headers and accents)
	colorBluePrimary = "0.204 0.596 0.859" // #3498db

	// Darker blue (for header background gradient effect)
	colorBlueDark = "0.161 0.502 0.725" // #2980b9

	// Risk level colors
	colorRedHigh   = "0.906 0.298 0.235" // #e74c3c - High risk
	colorYellowMed = "0.945 0.769 0.059" // #f1c40f - Medium risk
	colorGreenLow  = "0.153 0.682 0.376" // #27ae60 - Low risk

	// Text colors
	colorBlack     = "0 0 0"          // Black text
	colorWhite     = "1 1 1"          // White text
	colorGray      = "0.5 0.5 0.5"    // Gray text
	colorGrayLight = "0.95 0.95 0.95" // Light gray background

	// Card background (light blue tint)
	colorCardBg = "0.941 0.965 0.984" // Very light blue
)

// ============================================================
// PDF EXPORTER STRUCTURE
// ============================================================
// This struct holds all the state we need while creating the PDF
type PDFExporter struct {
	// Page positioning
	currentY     float64 // Current Y position on page (starts at top, moves down)
	pageHeight   float64 // Total page height in points (US Letter = 792)
	pageWidth    float64 // Total page width in points (US Letter = 612)
	marginTop    float64 // Top margin
	marginBottom float64 // Bottom margin
	marginLeft   float64 // Left margin
	marginRight  float64 // Right margin

	// PDF structure
	pageNumber    int      // Current page number (for multi-page PDFs)
	objectNumber  int      // Current PDF object number (each element is an object)
	objects       []string // List of all PDF objects (pages, fonts, content)
	objectOffsets []int    // Byte offset of each object (for cross-reference table)
}

// ============================================================
// MAIN EXPORT FUNCTION
// ============================================================
// Export creates the complete PDF file
// This is the main entry point called from Report.Export()
//
// HOW IT WORKS:
//  1. Initialize PDF state (page size, margins, etc.)
//  2. Build all page content (header, summary, statistics, findings)
//  3. Create PDF structure (catalog, pages, fonts)
//  4. Write cross-reference table (index of all objects)
//  5. Write file to disk
//
// Parameters:
//   - report: The Report struct containing all scan data
//   - filename: Output filename (e.g., "scan_report.pdf")
//
// Returns:
//   - error: Error if file can't be written or PDF creation fails
//
// Example:
//
//	exporter := &PDFExporter{}
//	err := exporter.Export(report, "report.pdf")
func (e *PDFExporter) Export(report *Report, filename string) error {
	// ============================================================
	// STEP 1: Initialize PDF state
	// ============================================================
	// Set up page dimensions (US Letter size)
	e.pageHeight = 792 // 11 inches * 72 points/inch
	e.pageWidth = 612  // 8.5 inches * 72 points/inch

	// Set margins (in points)
	e.marginTop = 40
	e.marginBottom = 40
	e.marginLeft = 40
	e.marginRight = 40

	// Start at top of first page
	e.currentY = e.pageHeight - e.marginTop
	e.pageNumber = 1

	// Initialize PDF object tracking
	e.objectNumber = 0
	e.objects = []string{}
	e.objectOffsets = []int{}

	// ============================================================
	// STEP 2: Build all page content
	// ============================================================
	// This creates the actual visible content of the PDF
	// Returns a list of content strings (one per page)
	contentPages := e.buildAllPages(report)

	// ============================================================
	// STEP 3: Start building the PDF file
	// ============================================================
	var pdf strings.Builder

	// PDF Header - required by PDF specification
	// The comment below with special characters tells PDF readers this is binary
	pdf.WriteString("%PDF-1.4\n")
	pdf.WriteString("%âãÏÓ\n\n")

	// ============================================================
	// STEP 4: Create PDF structure objects
	// ============================================================

	// Object 1: Catalog (root of PDF document tree)
	// This tells the PDF reader where to find the pages
	e.addObject(&pdf, "<< /Type /Catalog /Pages 2 0 R >>")

	// Object 2: Pages collection
	// This lists all the pages in the document
	pageRefs := ""
	for i := 0; i < len(contentPages); i++ {
		if i > 0 {
			pageRefs += " "
		}
		// Page objects are numbered 3, 6, 9, 12, ... (every 3rd number)
		pageRefs += fmt.Sprintf("%d 0 R", 3+i*3)
	}
	e.addObject(&pdf, fmt.Sprintf(
		"<< /Type /Pages /Kids [%s] /Count %d /MediaBox [0 0 %.1f %.1f] >>",
		pageRefs, len(contentPages), e.pageWidth, e.pageHeight))

	// ============================================================
	// STEP 5: Create each page with content and fonts
	// ============================================================
	// For each page, we create 3 objects:
	//   1. Page object (references content and fonts)
	//   2. Content stream (the actual text and graphics)
	//   3. Font resources (fonts used on this page)
	for i, content := range contentPages {
		pageNum := 3 + i*3
		contentNum := pageNum + 1
		fontsNum := pageNum + 2

		// Page object - links to parent pages collection and its content
		e.addObject(&pdf, fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /Contents %d 0 R /Resources << /Font %d 0 R >> >>",
			contentNum, fontsNum))

		// Content stream - the actual page content (text, shapes, colors)
		e.addObject(&pdf, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream",
			len(content), content))

		// Font resources - we use Helvetica family fonts
		// F1 = Regular, F2 = Bold, F3 = Italic, F4 = Bold Italic
		e.addObject(&pdf, `<< 
			/F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
			/F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>
			/F3 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>
			/F4 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-BoldOblique >>
		>>`)
	}

	// ============================================================
	// STEP 6: Create cross-reference table (xref)
	// ============================================================
	// This table tells the PDF reader where each object is in the file
	// It's required for random access to PDF objects
	xrefStart := pdf.Len() // Remember where xref table starts
	pdf.WriteString("xref\n")
	pdf.WriteString(fmt.Sprintf("0 %d\n", len(e.objectOffsets)+1))
	pdf.WriteString("0000000000 65535 f \n") // Special entry for object 0
	for _, offset := range e.objectOffsets {
		// Each entry is exactly 20 bytes: 10 digits + space + 5 digits + space + n + space + newline
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// ============================================================
	// STEP 7: Write trailer and EOF marker
	// ============================================================
	// Trailer tells reader where to start reading and how many objects exist
	pdf.WriteString("\ntrailer\n")
	pdf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(e.objectOffsets)+1))
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	pdf.WriteString("%%EOF\n")

	// ============================================================
	// STEP 8: Write completed PDF to file
	// ============================================================
	return os.WriteFile(filename, []byte(pdf.String()), 0644)
}

// ============================================================
// ADD PDF OBJECT HELPER
// ============================================================
// addObject adds a new object to the PDF and tracks its position
//
// HOW IT WORKS:
//   - Increments object number
//   - Records current byte position (for xref table)
//   - Writes object in standard PDF format
//
// Parameters:
//   - pdf: The string builder containing the PDF
//   - content: The object's content (dictionary, stream, etc.)
//
// PDF object format:
//
//	N 0 obj
//	<content>
//	endobj
func (e *PDFExporter) addObject(pdf *strings.Builder, content string) {
	e.objectNumber++
	// Record byte offset of this object for the xref table
	e.objectOffsets = append(e.objectOffsets, pdf.Len())
	// Write object in standard format
	pdf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n\n", e.objectNumber, content))
}

// ============================================================
// BUILD ALL PAGES
// ============================================================
// buildAllPages creates the content for all pages
// Returns a slice of strings (one string per page)
//
// HOW IT WORKS:
//  1. Create first page with header and executive summary
//  2. Add statistics section
//  3. Add file findings (may span multiple pages)
//  4. Automatically creates new pages when content doesn't fit
//
// Parameters:
//   - report: The Report struct containing all scan data
//
// Returns:
//   - []string: Slice of page contents (one per page)
func (e *PDFExporter) buildAllPages(report *Report) []string {
	var pages []string
	var currentPage strings.Builder

	// ============================================================
	// PAGE 1: HEADER + EXECUTIVE SUMMARY
	// ============================================================
	e.currentY = e.pageHeight - e.marginTop

	// Draw beautiful blue header (similar to HTML)
	e.addModernHeader(&currentPage, report)

	// Add executive summary section
	e.addExecutiveSummary(&currentPage, report)

	// Add statistics section
	e.addStatistics(&currentPage, report)

	// ============================================================
	// DETAILED FINDINGS
	// ============================================================
	if report.CardsFound > 0 {
		e.checkPageBreak(&currentPage, &pages, 100)
		e.addSectionTitle(&currentPage, "DETAILED FINDINGS")

		// Add each file's findings
		// We need to iterate over the GroupedByFile map
		fileIndex := 0
		for filePath, findings := range report.GroupedByFile {
			// Check if we need a new page
			e.checkPageBreak(&currentPage, &pages, 80)

			// Add this file's findings
			fileIndex++
			e.addFileFindings(&currentPage, filePath, len(findings), fileIndex)
		}
	}

	// Add footer to last page
	e.addFooter(&currentPage)

	// Add the final page
	pages = append(pages, currentPage.String())

	return pages
}

// ============================================================
// CHECK PAGE BREAK
// ============================================================
// checkPageBreak checks if we need to start a new page
// If currentY is too low, it saves the current page and starts a new one
//
// Parameters:
//   - currentPage: The string builder for current page content
//   - pages: Slice of all completed pages
//   - spaceNeeded: How much vertical space we need (in points)
func (e *PDFExporter) checkPageBreak(currentPage *strings.Builder, pages *[]string, spaceNeeded float64) {
	if e.currentY < (e.marginBottom + spaceNeeded) {
		// Save current page
		*pages = append(*pages, currentPage.String())

		// Start new page
		currentPage.Reset()
		e.currentY = e.pageHeight - e.marginTop
		e.pageNumber++
	}
}

// ============================================================
// MODERN HEADER (matching HTML design)
// ============================================================
// addModernHeader creates the beautiful blue header at the top
// Similar to the gradient header in the HTML report
//
// WHAT IT DRAWS:
//   - Blue gradient-style background
//   - Title "BasicPanScanner Security Report"
//   - Subtitle "PCI DSS Compliance Scanner"
//   - Version number
//
// Parameters:
//   - page: The string builder for page content
//   - report: The Report struct (for version info)
func (e *PDFExporter) addModernHeader(page *strings.Builder, report *Report) {
	// Draw blue background box (simulating gradient)
	// We use a slightly darker blue for depth
	page.WriteString("q\n")                   // Save graphics state
	page.WriteString(colorBlueDark + " rg\n") // Set fill color to dark blue
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		0.0,              // x position (full width)
		e.pageHeight-120, // y position
		e.pageWidth,      // width (full page)
		120.0))           // height
	page.WriteString("Q\n") // Restore graphics state

	// Add lighter blue accent at bottom of header
	page.WriteString("q\n")
	page.WriteString(colorBluePrimary + " rg\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		0.0,
		e.pageHeight-124,
		e.pageWidth,
		4.0)) // 4pt accent line
	page.WriteString("Q\n")

	// Draw title text (white, bold, large)
	page.WriteString("BT\n")               // Begin text
	page.WriteString(colorWhite + " rg\n") // White color
	page.WriteString("/F2 22 Tf\n")        // Bold font, 22pt
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n",
		e.marginLeft,
		e.pageHeight-65))
	page.WriteString("(BasicPanScanner Security Report) Tj\n")
	page.WriteString("ET\n") // End text

	// Draw subtitle (white, regular, medium)
	page.WriteString("BT\n")
	page.WriteString("0.9 0.9 0.9 rg\n") // Light white/gray
	page.WriteString("/F1 12 Tf\n")      // Regular font, 12pt
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n",
		e.marginLeft,
		e.pageHeight-85))
	page.WriteString("(PCI DSS Compliance Scanner) Tj\n")
	page.WriteString("ET\n")

	// Draw version (white, small)
	page.WriteString("BT\n")
	page.WriteString("0.8 0.8 0.8 rg\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n",
		e.marginLeft,
		e.pageHeight-100))
	page.WriteString(fmt.Sprintf("(Version %s) Tj\n", e.escape(report.Version)))
	page.WriteString("ET\n")

	// Move Y position down past the header
	e.currentY = e.pageHeight - 145
}

// ============================================================
// EXECUTIVE SUMMARY (beautiful cards layout)
// ============================================================
// addExecutiveSummary creates the summary section with colored cards
// Similar to the blue summary section in HTML report
//
// WHAT IT DRAWS:
//   - Blue background box
//   - Summary cards with key metrics
//   - Risk level indicator with appropriate color
//
// Parameters:
//   - page: The string builder for page content
//   - report: The Report struct with scan data
func (e *PDFExporter) addExecutiveSummary(page *strings.Builder, report *Report) {
	// Draw light blue background box
	boxHeight := 180.0
	page.WriteString("q\n")
	page.WriteString(colorCardBg + " rg\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		e.marginLeft,
		e.currentY-boxHeight,
		e.pageWidth-e.marginLeft-e.marginRight,
		boxHeight))
	page.WriteString("Q\n")

	// Section title
	page.WriteString("BT\n")
	page.WriteString(colorBluePrimary + " rg\n")
	page.WriteString("/F2 16 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-25))
	page.WriteString("(EXECUTIVE SUMMARY) Tj\n")
	page.WriteString("ET\n")

	// Get risk level and color
	riskLevel, riskColor := report.GetRiskLevel()
	riskColorRGB := e.getRiskColorRGB(riskLevel)

	// Define summary items in a grid layout
	summaryItems := []struct {
		label string
		value string
		x     float64 // X position for this card
	}{
		{
			label: "SCAN DATE",
			value: report.ScanDate.Format("Jan 2, 2006"),
			x:     e.marginLeft + 10,
		},
		{
			label: "DURATION",
			value: e.formatDuration(report.Duration),
			x:     e.marginLeft + 150,
		},
		{
			label: "FILES SCANNED",
			value: fmt.Sprintf("%d", report.ScannedFiles),
			x:     e.marginLeft + 290,
		},
		{
			label: "CARDS FOUND",
			value: fmt.Sprintf("%d", report.CardsFound),
			x:     e.marginLeft + 430,
		},
	}

	// Draw each summary card
	cardY := e.currentY - 60
	for _, item := range summaryItems {
		// Draw label (small, uppercase, gray)
		page.WriteString("BT\n")
		page.WriteString("0.4 0.4 0.4 rg\n")
		page.WriteString("/F2 8 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", item.x, cardY))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", item.label))
		page.WriteString("ET\n")

		// Draw value (large, bold, blue)
		page.WriteString("BT\n")
		page.WriteString(colorBluePrimary + " rg\n")
		page.WriteString("/F2 18 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", item.x, cardY-22))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(item.value)))
		page.WriteString("ET\n")
	}

	// Draw risk level card (special, with colored box)
	riskX := e.marginLeft + 10
	riskY := e.currentY - 130

	// Draw colored box background for risk
	page.WriteString("q\n")
	page.WriteString(riskColorRGB + " rg\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		riskX-5,
		riskY-30,
		140.0,
		50.0))
	page.WriteString("Q\n")

	// Risk label
	page.WriteString("BT\n")
	page.WriteString(colorWhite + " rg\n")
	page.WriteString("/F2 9 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", riskX, riskY))
	page.WriteString("(RISK LEVEL) Tj\n")
	page.WriteString("ET\n")

	// Risk value
	page.WriteString("BT\n")
	page.WriteString(colorWhite + " rg\n")
	page.WriteString("/F2 20 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", riskX, riskY-22))
	page.WriteString(fmt.Sprintf("(%s) Tj\n", riskLevel))
	page.WriteString("ET\n")

	// Additional details (smaller text below the cards)
	page.WriteString("BT\n")
	page.WriteString("0.3 0.3 0.3 rg\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+160, riskY-15))
	page.WriteString(fmt.Sprintf("(Directory: %s) Tj\n", e.escape(e.truncate(report.Directory, 55))))
	page.WriteString("ET\n")

	// Move Y position down
	e.currentY -= boxHeight + 20

	// Unused variable to avoid linter error
	_ = riskColor
}

// ============================================================
// STATISTICS SECTION
// ============================================================
// addStatistics creates the statistics section with visual bars
//
// WHAT IT DRAWS:
//   - Section title
//   - Card type distribution with bars
//   - Top files with cards
//
// Parameters:
//   - page: The string builder for page content
//   - report: The Report struct with statistics
func (e *PDFExporter) addStatistics(page *strings.Builder, report *Report) {
	e.addSectionTitle(page, "STATISTICS")

	// Check if there are any cards found
	if report.CardsFound == 0 {
		page.WriteString("BT\n")
		page.WriteString(colorGray + " rg\n")
		page.WriteString("/F3 11 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-10))
		page.WriteString("(No payment cards found in scan.) Tj\n")
		page.WriteString("ET\n")
		e.currentY -= 40
		return
	}

	// Card type distribution
	page.WriteString("BT\n")
	page.WriteString(colorBlack + " rg\n")
	page.WriteString("/F2 12 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-10))
	page.WriteString("(Card Type Distribution) Tj\n")
	page.WriteString("ET\n")

	y := e.currentY - 35
	for cardType, count := range report.Statistics.CardsByType {
		// Card type name
		page.WriteString("BT\n")
		page.WriteString(colorBlack + " rg\n")
		page.WriteString("/F1 10 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+20, y))
		page.WriteString(fmt.Sprintf("(%s:) Tj\n", e.escape(cardType)))
		page.WriteString("ET\n")

		// Count
		page.WriteString("BT\n")
		page.WriteString(colorBluePrimary + " rg\n")
		page.WriteString("/F2 10 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+120, y))
		page.WriteString(fmt.Sprintf("(%d cards) Tj\n", count))
		page.WriteString("ET\n")

		// Visual bar (proportional to count)
		barWidth := float64(count) / float64(report.CardsFound) * 300
		if barWidth < 5 {
			barWidth = 5
		}
		page.WriteString("q\n")
		page.WriteString(colorBluePrimary + " rg\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
			e.marginLeft+200,
			y-3,
			barWidth,
			10.0))
		page.WriteString("Q\n")

		y -= 25
	}

	e.currentY = y - 20

	// Top files
	if len(report.Statistics.TopFiles) > 0 {
		page.WriteString("BT\n")
		page.WriteString(colorBlack + " rg\n")
		page.WriteString("/F2 12 Tf\n")
		page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY))
		page.WriteString("(Top Files with Cards) Tj\n")
		page.WriteString("ET\n")

		y = e.currentY - 25
		for i, fs := range report.Statistics.TopFiles {
			if i >= 5 { // Show top 5
				break
			}

			// File number and name
			page.WriteString("BT\n")
			page.WriteString(colorBlack + " rg\n")
			page.WriteString("/F1 9 Tf\n")
			page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+20, y))
			page.WriteString(fmt.Sprintf("(%d. %s) Tj\n", i+1, e.escape(e.truncate(fs.FilePath, 60))))
			page.WriteString("ET\n")

			// Card count
			page.WriteString("BT\n")
			page.WriteString(colorBluePrimary + " rg\n")
			page.WriteString("/F2 9 Tf\n")
			page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.pageWidth-e.marginRight-60, y))
			page.WriteString(fmt.Sprintf("(%d cards) Tj\n", fs.CardCount))
			page.WriteString("ET\n")

			y -= 20
		}
		e.currentY = y - 20
	}
}

// ============================================================
// SECTION TITLE
// ============================================================
// addSectionTitle adds a section title with blue bar
//
// Parameters:
//   - page: The string builder for page content
//   - title: The section title text
func (e *PDFExporter) addSectionTitle(page *strings.Builder, title string) {
	e.currentY -= 10

	// Draw blue bar background
	page.WriteString("q\n")
	page.WriteString(colorBluePrimary + " rg\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		e.marginLeft,
		e.currentY-25,
		e.pageWidth-e.marginLeft-e.marginRight,
		25.0))
	page.WriteString("Q\n")

	// Draw white text
	page.WriteString("BT\n")
	page.WriteString(colorWhite + " rg\n")
	page.WriteString("/F2 14 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-17))
	page.WriteString(fmt.Sprintf("(%s) Tj\n", title))
	page.WriteString("ET\n")

	e.currentY -= 40
}

// ============================================================
// FILE FINDINGS
// ============================================================
// addFileFindings adds a file's findings section
//
// Parameters:
//   - page: The string builder for page content
//   - filePath: Path to the file
//   - findingsCount: Number of cards found
//   - index: File number (for display)
func (e *PDFExporter) addFileFindings(page *strings.Builder, filePath string, findingsCount int, index int) {
	// Draw light gray box background
	page.WriteString("q\n")
	page.WriteString(colorGrayLight + " rg\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f %.1f %.1f re f\n",
		e.marginLeft,
		e.currentY-35,
		e.pageWidth-e.marginLeft-e.marginRight,
		35.0))
	page.WriteString("Q\n")

	// File number and path
	page.WriteString("BT\n")
	page.WriteString(colorBlack + " rg\n")
	page.WriteString("/F2 10 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-15))
	page.WriteString(fmt.Sprintf("(File %d: %s) Tj\n", index, e.escape(e.truncate(filePath, 70))))
	page.WriteString("ET\n")

	// Card count
	page.WriteString("BT\n")
	page.WriteString(colorBluePrimary + " rg\n")
	page.WriteString("/F2 10 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft+10, e.currentY-30))
	page.WriteString(fmt.Sprintf("(%d cards found) Tj\n", findingsCount))
	page.WriteString("ET\n")

	e.currentY -= 50
}

// ============================================================
// FOOTER
// ============================================================
// addFooter adds a footer at the bottom of the page
//
// Parameters:
//   - page: The string builder for page content
func (e *PDFExporter) addFooter(page *strings.Builder) {
	// Draw timestamp at bottom
	page.WriteString("BT\n")
	page.WriteString(colorGray + " rg\n")
	page.WriteString("/F3 8 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.marginLeft, 30.0))
	page.WriteString(fmt.Sprintf("(Generated: %s) Tj\n",
		e.escape(time.Now().Format("January 2, 2006 at 3:04 PM MST"))))
	page.WriteString("ET\n")

	// Draw page number at bottom right
	page.WriteString("BT\n")
	page.WriteString(colorGray + " rg\n")
	page.WriteString("/F1 8 Tf\n")
	page.WriteString(fmt.Sprintf("%.1f %.1f Td\n", e.pageWidth-e.marginRight-40, 30.0))
	page.WriteString(fmt.Sprintf("(Page %d) Tj\n", e.pageNumber))
	page.WriteString("ET\n")
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// escape escapes special characters for PDF text strings
// PDF requires certain characters to be escaped
//
// Parameters:
//   - s: The string to escape
//
// Returns:
//   - string: Escaped string safe for PDF
func (e *PDFExporter) escape(s string) string {
	// Escape backslashes first (must be first!)
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// Escape parentheses (they're special in PDF)
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	// Replace newlines with spaces
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// truncate truncates a string to maximum length
// Adds "..." if truncated
//
// Parameters:
//   - s: The string to truncate
//   - maxLen: Maximum length
//
// Returns:
//   - string: Truncated string
func (e *PDFExporter) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a time.Duration in human-readable format
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Formatted duration (e.g., "1.5s", "2m 30s")
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

// getRiskColorRGB returns the RGB color string for a risk level
//
// Parameters:
//   - riskLevel: "High", "Medium", or "Low"
//
// Returns:
//   - string: RGB color value for PDF
func (e *PDFExporter) getRiskColorRGB(riskLevel string) string {
	switch riskLevel {
	case "High":
		return colorRedHigh
	case "Medium":
		return colorYellowMed
	case "Low":
		return colorGreenLow
	default:
		return colorGray
	}
}
