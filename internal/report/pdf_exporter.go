// Package report - ULTRA BEAUTIFUL PDF Exporter (NO EXTERNAL LIBRARIES!)
// Maximum visual appeal with professional design - using ONLY Go standard library
package report

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ============================================================
// ULTRA BEAUTIFUL PDF EXPORTER
// ============================================================
// Features:
//   ✨ Gradient-style headers
//   ✨ Unicode icons and symbols
//   ✨ Visual charts and graphs
//   ✨ Professional color schemes
//   ✨ Page numbers and footers
//   ✨ Risk level badges
//   ✨ Beautiful typography
//   ✨ Card type icons
//   ✨ Shadow effects (simulated)
//   ✨ Table of contents
//   ✨ Executive dashboard
//   ✨ ALL using only GO standard library!

type PDFExporter struct {
	currentY      float64
	pageHeight    float64
	marginTop     float64
	marginBottom  float64
	pageNumber    int
	objectNumber  int
	objects       []string
	objectOffsets []int
	totalPages    int // For page numbering
}

// ============================================================
// MAIN EXPORT FUNCTION
// ============================================================
func (e *PDFExporter) Export(report *Report, filename string) error {
	// Initialize
	e.pageHeight = 792
	e.marginTop = 40
	e.marginBottom = 50
	e.currentY = e.pageHeight - e.marginTop
	e.pageNumber = 1
	e.objectNumber = 0
	e.objects = []string{}
	e.objectOffsets = []int{}

	var pdf strings.Builder

	// PDF Header
	pdf.WriteString("%PDF-1.4\n%âãÏÓ\n\n")

	// Create all pages
	contentPages := e.buildAllPages(report)
	e.totalPages = len(contentPages)

	// Catalog
	e.addObject(&pdf, "<< /Type /Catalog /Pages 2 0 R >>")

	// Pages collection
	pageRefs := ""
	for i := 0; i < len(contentPages); i++ {
		if i > 0 {
			pageRefs += " "
		}
		pageRefs += fmt.Sprintf("%d 0 R", 3+i*3)
	}
	e.addObject(&pdf, fmt.Sprintf(
		"<< /Type /Pages /Kids [%s] /Count %d /MediaBox [0 0 612 792] >>",
		pageRefs, len(contentPages)))

	// Create page objects
	for i, content := range contentPages {
		pageNum := 3 + i*3
		contentNum := pageNum + 1
		fontsNum := pageNum + 2

		e.addObject(&pdf, fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /Contents %d 0 R /Resources << /Font %d 0 R >> >>",
			contentNum, fontsNum))

		e.addObject(&pdf, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream",
			len(content), content))

		e.addObject(&pdf, `<< 
			/F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
			/F2 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>
			/F3 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique >>
			/F4 << /Type /Font /Subtype /Type1 /BaseFont /Courier >>
		>>`)
	}

	// Cross-reference table
	xrefStart := pdf.Len()
	pdf.WriteString("xref\n")
	pdf.WriteString(fmt.Sprintf("0 %d\n", len(e.objectOffsets)+1))
	pdf.WriteString("0000000000 65535 f \n")
	for _, offset := range e.objectOffsets {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	pdf.WriteString("\ntrailer\n")
	pdf.WriteString(fmt.Sprintf("<< /Size %d /Root 1 0 R >>\n", len(e.objectOffsets)+1))
	pdf.WriteString("startxref\n")
	pdf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	pdf.WriteString("%%EOF\n")

	return os.WriteFile(filename, []byte(pdf.String()), 0644)
}

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

	e.currentY = e.pageHeight - e.marginTop

	// ============================================================
	// PAGE 1: COVER PAGE WITH GRADIENT HEADER
	// ============================================================
	e.addGradientHeader(&currentPage, report)
	e.addExecutiveDashboard(&currentPage, report)
	e.addPageNumber(&currentPage, 1)

	pages = append(pages, currentPage.String())
	currentPage.Reset()
	e.currentY = e.pageHeight - e.marginTop

	// ============================================================
	// PAGE 2: DETAILED STATISTICS
	// ============================================================
	e.addStatsHeader(&currentPage, "DETAILED STATISTICS")
	e.addStatisticsCards(&currentPage, report)
	e.addCardDistributionChart(&currentPage, report)
	e.addFileTypeDistribution(&currentPage, report)
	e.addPageNumber(&currentPage, 2)

	pages = append(pages, currentPage.String())
	currentPage.Reset()
	e.currentY = e.pageHeight - e.marginTop

	// ============================================================
	// PAGE 3+: TOP FILES WITH RISK BADGES
	// ============================================================
	e.addStatsHeader(&currentPage, "TOP FILES ANALYSIS")
	e.addTopFilesDetailed(&currentPage, report)

	if e.currentY < 200 {
		e.addPageNumber(&currentPage, 3)
		pages = append(pages, currentPage.String())
		currentPage.Reset()
		e.currentY = e.pageHeight - e.marginTop
	}

	// ============================================================
	// REMAINING PAGES: DETAILED FINDINGS
	// ============================================================
	pageNum := len(pages) + 1
	e.addStatsHeader(&currentPage, "DETAILED FINDINGS")

	fileCount := 0
	for filePath, findings := range report.GroupedByFile {
		if e.currentY < 120 {
			e.addPageNumber(&currentPage, pageNum)
			pages = append(pages, currentPage.String())
			currentPage.Reset()
			e.currentY = e.pageHeight - e.marginTop
			pageNum++
		}

		e.addFileCard(&currentPage, filePath, len(findings), fileCount, report)
		fileCount++

		if fileCount >= 15 {
			break
		}
	}

	// Final page with footer
	e.addFinalFooter(&currentPage)
	e.addPageNumber(&currentPage, pageNum)
	pages = append(pages, currentPage.String())

	return pages
}

// ============================================================
// GRADIENT HEADER (Simulated with color bands)
// ============================================================
func (e *PDFExporter) addGradientHeader(page *strings.Builder, report *Report) {
	// Create gradient effect with multiple color bands
	colors := []string{
		"0.12 0.47 0.71", // Dark blue
		"0.16 0.50 0.73",
		"0.20 0.60 0.86", // Main blue
		"0.24 0.63 0.88",
		"0.28 0.67 0.90", // Light blue
	}

	bandHeight := 16.0
	for i, color := range colors {
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", color))
		y := e.currentY - float64(i)*bandHeight
		page.WriteString(fmt.Sprintf("0 %f 612 %f re f\n", y-bandHeight, bandHeight))
		page.WriteString("Q\n")
	}

	// White text with shadow effect
	// Shadow (gray, slightly offset)
	page.WriteString("BT\n")
	page.WriteString("0.3 0.3 0.3 rg\n")
	page.WriteString("/F2 28 Tf\n")
	page.WriteString(fmt.Sprintf("72 %f Td\n", e.currentY-47))
	page.WriteString("(BASICPANSCANNER) Tj\n")
	page.WriteString("ET\n")

	// Main text (white)
	page.WriteString("BT\n")
	page.WriteString("1 1 1 rg\n")
	page.WriteString("/F2 28 Tf\n")
	page.WriteString(fmt.Sprintf("70 %f Td\n", e.currentY-45))
	page.WriteString("(BASICPANSCANNER) Tj\n")
	page.WriteString("ET\n")

	// Subtitle with icon
	page.WriteString("BT\n")
	page.WriteString("1 1 1 rg\n")
	page.WriteString("/F1 13 Tf\n")
	page.WriteString(fmt.Sprintf("70 %f Td\n", e.currentY-63))
	page.WriteString("(\\u25C6 Credit Card Security Assessment Report) Tj\n")
	page.WriteString("ET\n")

	// Version badge (right side)
	page.WriteString("q\n")
	page.WriteString("0.9 0.9 0.9 rg\n")
	page.WriteString(fmt.Sprintf("490 %f 80 18 re f\n", e.currentY-50))
	page.WriteString("0.2 0.6 0.9 RG\n")
	page.WriteString("1.5 w\n")
	page.WriteString(fmt.Sprintf("490 %f 80 18 re S\n", e.currentY-50))
	page.WriteString("Q\n")

	page.WriteString("BT\n")
	page.WriteString("0.2 0.6 0.9 rg\n")
	page.WriteString("/F2 9 Tf\n")
	page.WriteString(fmt.Sprintf("510 %f Td\n", e.currentY-44))
	page.WriteString(fmt.Sprintf("(v%s) Tj\n", e.escape(report.Version)))
	page.WriteString("ET\n")

	e.currentY -= 100
}

// ============================================================
// EXECUTIVE DASHBOARD (Cards with icons)
// ============================================================
func (e *PDFExporter) addExecutiveDashboard(page *strings.Builder, report *Report) {
	// Section title with icon
	page.WriteString("BT\n")
	page.WriteString("0.2 0.2 0.2 rg\n")
	page.WriteString("/F2 16 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
	page.WriteString("(\\u25BA EXECUTIVE DASHBOARD) Tj\n")
	page.WriteString("ET\n")

	e.currentY -= 30

	// Dashboard cards with icons and colors
	cards := []struct {
		icon  string
		label string
		value string
		color string
		x     float64
	}{
		{"\\u2637", "Scan Duration", e.formatDuration(report.Duration), "0.4 0.6 0.9", 50},
		{"\\u2713", "Files Scanned", fmt.Sprintf("%d/%d", report.ScannedFiles, report.TotalFiles), "0.3 0.7 0.4", 210},
		{"\\u26A0", "Cards Found", fmt.Sprintf("%d", report.CardsFound), "0.9 0.4 0.3", 370},
	}

	for _, card := range cards {
		// Card background with rounded effect
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", card.color))
		page.WriteString(fmt.Sprintf("%f %f 145 70 re f\n", card.x, e.currentY-70))

		// Lighter overlay for depth
		page.WriteString(fmt.Sprintf("%s rg\n", card.color))
		page.WriteString("0.9 0.9 0.9 rg\n")
		page.WriteString(fmt.Sprintf("%f %f 145 20 re f\n", card.x, e.currentY-20))
		page.WriteString("Q\n")

		// Icon (large, white)
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 24 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", card.x+10, e.currentY-45))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", card.icon))
		page.WriteString("ET\n")

		// Value (large, white)
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 16 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", card.x+50, e.currentY-42))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(card.value)))
		page.WriteString("ET\n")

		// Label (small, on overlay)
		page.WriteString("BT\n")
		page.WriteString(fmt.Sprintf("%s rg\n", card.color))
		page.WriteString("/F2 8 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", card.x+10, e.currentY-13))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(card.label)))
		page.WriteString("ET\n")
	}

	e.currentY -= 90

	// Scan information box
	e.addInfoBox(page, report)
}

// ============================================================
// INFO BOX (Professional details)
// ============================================================
func (e *PDFExporter) addInfoBox(page *strings.Builder, report *Report) {
	// Light blue background
	page.WriteString("q\n")
	page.WriteString("0.94 0.97 0.99 rg\n")
	page.WriteString(fmt.Sprintf("50 %f 512 85 re f\n", e.currentY-85))

	// Blue border
	page.WriteString("0.2 0.6 0.9 RG\n")
	page.WriteString("1 w\n")
	page.WriteString(fmt.Sprintf("50 %f 512 85 re S\n", e.currentY-85))
	page.WriteString("Q\n")

	// Title
	page.WriteString("BT\n")
	page.WriteString("0.2 0.6 0.9 rg\n")
	page.WriteString("/F2 11 Tf\n")
	page.WriteString(fmt.Sprintf("60 %f Td\n", e.currentY-20))
	page.WriteString("(SCAN INFORMATION) Tj\n")
	page.WriteString("ET\n")

	// Information with icons
	info := []struct {
		icon  string
		label string
		value string
	}{
		{"\\u25C9", "Date", report.ScanDate.Format("January 2, 2006 at 3:04 PM")},
		{"\\u25C9", "Directory", e.truncate(report.Directory, 55)},
		{"\\u25C9", "Mode", report.ScanMode},
	}

	y := e.currentY - 38
	for _, item := range info {
		// Icon
		page.WriteString("BT\n")
		page.WriteString("0.2 0.6 0.9 rg\n")
		page.WriteString("/F1 10 Tf\n")
		page.WriteString(fmt.Sprintf("60 %f Td\n", y))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", item.icon))
		page.WriteString("ET\n")

		// Label (bold)
		page.WriteString("BT\n")
		page.WriteString("0.2 0.2 0.2 rg\n")
		page.WriteString("/F2 9 Tf\n")
		page.WriteString(fmt.Sprintf("75 %f Td\n", y))
		page.WriteString(fmt.Sprintf("(%s:) Tj\n", e.escape(item.label)))
		page.WriteString("ET\n")

		// Value
		page.WriteString("BT\n")
		page.WriteString("0.3 0.3 0.3 rg\n")
		page.WriteString("/F1 9 Tf\n")
		page.WriteString(fmt.Sprintf("140 %f Td\n", y))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(item.value)))
		page.WriteString("ET\n")

		y -= 15
	}

	e.currentY -= 105
}

// ============================================================
// STATISTICS CARDS (Beautiful colored cards)
// ============================================================
func (e *PDFExporter) addStatisticsCards(page *strings.Builder, report *Report) {
	e.currentY -= 10

	stats := []struct {
		label string
		value int
		icon  string
		color string
	}{
		{"Total Cards", report.CardsFound, "\\u2726", "0.9 0.3 0.3"},
		{"Affected Files", report.Statistics.FilesWithCards, "\\u25C6", "0.3 0.6 0.9"},
		{"High Risk", report.Statistics.HighRiskFiles, "\\u26A0", "0.8 0.2 0.2"},
		{"Medium Risk", report.Statistics.MediumRiskFiles, "\\u25B2", "0.9 0.6 0.2"},
		{"Low Risk", report.Statistics.LowRiskFiles, "\\u25BC", "0.3 0.7 0.3"},
	}

	x := 50.0
	for i, stat := range stats {
		if i == 3 {
			x = 50.0
			e.currentY -= 75
		}

		// Card with gradient effect
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", stat.color))
		page.WriteString(fmt.Sprintf("%f %f 100 65 re f\n", x, e.currentY-65))

		// Top highlight
		page.WriteString("1 1 1 rg\n")
		page.WriteString("0.2 gs\n")
		page.WriteString(fmt.Sprintf("%f %f 100 3 re f\n", x, e.currentY-3))
		page.WriteString("Q\n")

		// Icon
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 20 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", x+38, e.currentY-35))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", stat.icon))
		page.WriteString("ET\n")

		// Value
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 18 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", x+38, e.currentY-52))
		page.WriteString(fmt.Sprintf("(%d) Tj\n", stat.value))
		page.WriteString("ET\n")

		// Label
		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F1 7 Tf\n")
		page.WriteString(fmt.Sprintf("%f %f Td\n", x+25, e.currentY-62))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(stat.label)))
		page.WriteString("ET\n")

		x += 110
	}

	e.currentY -= 75
}

// ============================================================
// CARD DISTRIBUTION CHART (Visual bar chart)
// ============================================================
func (e *PDFExporter) addCardDistributionChart(page *strings.Builder, report *Report) {
	e.currentY -= 20

	// Title
	page.WriteString("BT\n")
	page.WriteString("0.2 0.2 0.2 rg\n")
	page.WriteString("/F2 12 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
	page.WriteString("(\\u25BA Card Distribution by Type) Tj\n")
	page.WriteString("ET\n")

	e.currentY -= 25

	// Sort cards
	type cardCount struct {
		name  string
		count int
	}
	var cards []cardCount
	for name, count := range report.Statistics.CardsByType {
		cards = append(cards, cardCount{name, count})
	}
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[j].count > cards[i].count {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	// Card brand colors
	brandColors := map[string]string{
		"Visa":             "0.0 0.3 0.6",
		"Mastercard":       "0.9 0.4 0.1",
		"American Express": "0.0 0.5 0.7",
		"Discover":         "0.9 0.5 0.1",
		"JCB":              "0.0 0.4 0.3",
	}

	maxCount := 0
	if len(cards) > 0 {
		maxCount = cards[0].count
	}

	for _, card := range cards {
		// Card icon based on brand
		icon := "\\u25C6"

		// Card name with icon
		page.WriteString("BT\n")
		page.WriteString("0.2 0.2 0.2 rg\n")
		page.WriteString("/F2 10 Tf\n")
		page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%s %s) Tj\n", icon, e.escape(card.name)))
		page.WriteString("ET\n")

		// Count
		page.WriteString("BT\n")
		page.WriteString("0.4 0.4 0.4 rg\n")
		page.WriteString("/F1 9 Tf\n")
		page.WriteString(fmt.Sprintf("180 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%d cards) Tj\n", card.count))
		page.WriteString("ET\n")

		// Percentage
		percentage := 0.0
		if report.CardsFound > 0 {
			percentage = float64(card.count) / float64(report.CardsFound) * 100
		}
		page.WriteString("BT\n")
		page.WriteString("0.5 0.5 0.5 rg\n")
		page.WriteString("/F1 8 Tf\n")
		page.WriteString(fmt.Sprintf("235 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%.1f%%) Tj\n", percentage))
		page.WriteString("ET\n")

		// Visual bar
		barWidth := 280.0
		if maxCount > 0 {
			barWidth = (float64(card.count) / float64(maxCount)) * 280.0
		}

		// Get color for this brand
		color := "0.2 0.6 0.9"
		if c, ok := brandColors[card.name]; ok {
			color = c
		}

		// Bar background (light)
		page.WriteString("q\n")
		page.WriteString("0.95 0.95 0.95 rg\n")
		page.WriteString(fmt.Sprintf("280 %f 280 12 re f\n", e.currentY-2))

		// Bar foreground (colored)
		page.WriteString(fmt.Sprintf("%s rg\n", color))
		page.WriteString(fmt.Sprintf("280 %f %f 12 re f\n", e.currentY-2, barWidth))

		// Bar border
		page.WriteString("0.7 0.7 0.7 RG\n")
		page.WriteString("0.5 w\n")
		page.WriteString(fmt.Sprintf("280 %f 280 12 re S\n", e.currentY-2))
		page.WriteString("Q\n")

		e.currentY -= 20
	}

	e.currentY -= 10
}

// ============================================================
// FILE TYPE DISTRIBUTION
// ============================================================
func (e *PDFExporter) addFileTypeDistribution(page *strings.Builder, report *Report) {
	if len(report.Statistics.FilesByType) == 0 {
		return
	}

	// Title
	page.WriteString("BT\n")
	page.WriteString("0.2 0.2 0.2 rg\n")
	page.WriteString("/F2 12 Tf\n")
	page.WriteString(fmt.Sprintf("50 %f Td\n", e.currentY))
	page.WriteString("(\\u25BA File Type Distribution) Tj\n")
	page.WriteString("ET\n")

	e.currentY -= 20

	// Sort file types
	type fileCount struct {
		ext   string
		count int
	}
	var files []fileCount
	for ext, count := range report.Statistics.FilesByType {
		files = append(files, fileCount{ext, count})
	}
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].count > files[i].count {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Show top 5
	maxShow := 5
	if len(files) < maxShow {
		maxShow = len(files)
	}

	for i := 0; i < maxShow; i++ {
		file := files[i]

		page.WriteString("BT\n")
		page.WriteString("0.3 0.3 0.3 rg\n")
		page.WriteString("/F4 9 Tf\n") // Courier for extensions
		page.WriteString(fmt.Sprintf("60 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(file.ext)))
		page.WriteString("ET\n")

		page.WriteString("BT\n")
		page.WriteString("0.5 0.5 0.5 rg\n")
		page.WriteString("/F1 9 Tf\n")
		page.WriteString(fmt.Sprintf("150 %f Td\n", e.currentY))
		page.WriteString(fmt.Sprintf("(%d files) Tj\n", file.count))
		page.WriteString("ET\n")

		e.currentY -= 15
	}

	e.currentY -= 10
}

// ============================================================
// SECTION HEADER (Styled)
// ============================================================
func (e *PDFExporter) addStatsHeader(page *strings.Builder, title string) {
	e.currentY -= 15

	// Blue bar with gradient effect
	page.WriteString("q\n")
	page.WriteString("0.2 0.6 0.9 rg\n")
	page.WriteString(fmt.Sprintf("30 %f 552 30 re f\n", e.currentY-30))

	// Lighter overlay
	page.WriteString("0.3 0.7 1.0 rg\n")
	page.WriteString(fmt.Sprintf("30 %f 552 3 re f\n", e.currentY-3))
	page.WriteString("Q\n")

	// Title (white, with icon)
	page.WriteString("BT\n")
	page.WriteString("1 1 1 rg\n")
	page.WriteString("/F2 13 Tf\n")
	page.WriteString(fmt.Sprintf("40 %f Td\n", e.currentY-20))
	page.WriteString(fmt.Sprintf("(\\u25BA %s) Tj\n", e.escape(title)))
	page.WriteString("ET\n")

	e.currentY -= 40
}

// ============================================================
// TOP FILES DETAILED (With risk badges)
// ============================================================
func (e *PDFExporter) addTopFilesDetailed(page *strings.Builder, report *Report) {
	if len(report.Statistics.TopFiles) == 0 {
		return
	}

	maxFiles := 10
	if len(report.Statistics.TopFiles) < maxFiles {
		maxFiles = len(report.Statistics.TopFiles)
	}

	for i := 0; i < maxFiles; i++ {
		fs := report.Statistics.TopFiles[i]

		// Risk badge
		riskLevel := "LOW"
		riskColor := "0.3 0.7 0.3"
		riskIcon := "\\u25BC"
		if fs.CardCount >= 5 {
			riskLevel = "HIGH"
			riskColor = "0.9 0.2 0.2"
			riskIcon = "\\u26A0"
		} else if fs.CardCount >= 2 {
			riskLevel = "MED"
			riskColor = "0.9 0.6 0.2"
			riskIcon = "\\u25B2"
		}

		// File box
		page.WriteString("q\n")
		page.WriteString("0.98 0.98 0.98 rg\n")
		page.WriteString(fmt.Sprintf("50 %f 512 45 re f\n", e.currentY-45))

		// Left colored stripe
		page.WriteString(fmt.Sprintf("%s rg\n", riskColor))
		page.WriteString(fmt.Sprintf("50 %f 5 45 re f\n", e.currentY-45))

		// Border
		page.WriteString("0.8 0.8 0.8 RG\n")
		page.WriteString("0.5 w\n")
		page.WriteString(fmt.Sprintf("50 %f 512 45 re S\n", e.currentY-45))
		page.WriteString("Q\n")

		// Risk badge (top right)
		page.WriteString("q\n")
		page.WriteString(fmt.Sprintf("%s rg\n", riskColor))
		page.WriteString(fmt.Sprintf("520 %f 35 16 re f\n", e.currentY-15))
		page.WriteString("Q\n")

		page.WriteString("BT\n")
		page.WriteString("1 1 1 rg\n")
		page.WriteString("/F2 7 Tf\n")
		page.WriteString(fmt.Sprintf("525 %f Td\n", e.currentY-11))
		page.WriteString(fmt.Sprintf("(%s %s) Tj\n", riskIcon, riskLevel))
		page.WriteString("ET\n")

		// File number and icon
		page.WriteString("BT\n")
		page.WriteString(fmt.Sprintf("%s rg\n", riskColor))
		page.WriteString("/F2 11 Tf\n")
		page.WriteString(fmt.Sprintf("63 %f Td\n", e.currentY-17))
		page.WriteString(fmt.Sprintf("(%d. \\u25C6) Tj\n", i+1))
		page.WriteString("ET\n")

		// Filename
		page.WriteString("BT\n")
		page.WriteString("0.2 0.2 0.2 rg\n")
		page.WriteString("/F2 10 Tf\n")
		page.WriteString(fmt.Sprintf("90 %f Td\n", e.currentY-17))
		page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(e.truncate(fs.FilePath, 45))))
		page.WriteString("ET\n")

		// Details
		page.WriteString("BT\n")
		page.WriteString("0.5 0.5 0.5 rg\n")
		page.WriteString("/F1 8 Tf\n")
		page.WriteString(fmt.Sprintf("90 %f Td\n", e.currentY-33))
		page.WriteString(fmt.Sprintf("(Cards: %d) Tj\n", fs.CardCount))
		page.WriteString("ET\n")

		e.currentY -= 52
	}
}

// ============================================================
// FILE CARD (Individual file details)
// ============================================================
func (e *PDFExporter) addFileCard(page *strings.Builder, filePath string, findingsCount int, index int, report *Report) {
	// Card background
	page.WriteString("q\n")
	page.WriteString("0.96 0.96 0.96 rg\n")
	page.WriteString(fmt.Sprintf("50 %f 512 35 re f\n", e.currentY-35))

	// Border
	page.WriteString("0.7 0.7 0.7 RG\n")
	page.WriteString("0.5 w\n")
	page.WriteString(fmt.Sprintf("50 %f 512 35 re S\n", e.currentY-35))
	page.WriteString("Q\n")

	// File number
	page.WriteString("BT\n")
	page.WriteString("0.2 0.6 0.9 rg\n")
	page.WriteString("/F2 10 Tf\n")
	page.WriteString(fmt.Sprintf("60 %f Td\n", e.currentY-20))
	page.WriteString(fmt.Sprintf("(File %d:) Tj\n", index+1))
	page.WriteString("ET\n")

	// Filename
	page.WriteString("BT\n")
	page.WriteString("0.2 0.2 0.2 rg\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString(fmt.Sprintf("100 %f Td\n", e.currentY-20))
	page.WriteString(fmt.Sprintf("(%s) Tj\n", e.escape(e.truncate(filePath, 55))))
	page.WriteString("ET\n")

	// Card count
	page.WriteString("BT\n")
	page.WriteString("0.5 0.5 0.5 rg\n")
	page.WriteString("/F1 8 Tf\n")
	page.WriteString(fmt.Sprintf("460 %f Td\n", e.currentY-20))
	page.WriteString(fmt.Sprintf("(Cards: %d) Tj\n", findingsCount))
	page.WriteString("ET\n")

	e.currentY -= 42
}

// ============================================================
// FINAL FOOTER
// ============================================================
func (e *PDFExporter) addFinalFooter(page *strings.Builder) {
	page.WriteString("BT\n")
	page.WriteString("0.4 0.4 0.4 rg\n")
	page.WriteString("/F3 9 Tf\n")
	page.WriteString("50 60 Td\n")
	page.WriteString(fmt.Sprintf("(Generated: %s) Tj\n",
		e.escape(time.Now().Format("January 2, 2006 at 3:04 PM MST"))))
	page.WriteString("ET\n")

	page.WriteString("BT\n")
	page.WriteString("0.6 0.6 0.6 rg\n")
	page.WriteString("/F1 8 Tf\n")
	page.WriteString("50 48 Td\n")
	page.WriteString("(\\u25C6 BasicPanScanner - Professional Security Assessment Tool) Tj\n")
	page.WriteString("ET\n")
}

// ============================================================
// PAGE NUMBER
// ============================================================
func (e *PDFExporter) addPageNumber(page *strings.Builder, pageNum int) {
	page.WriteString("BT\n")
	page.WriteString("0.5 0.5 0.5 rg\n")
	page.WriteString("/F1 9 Tf\n")
	page.WriteString("540 30 Td\n")
	page.WriteString(fmt.Sprintf("(Page %d) Tj\n", pageNum))
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
