// Package report - HTML exporter
// Exports reports in interactive HTML format with professional styling
package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HTMLExporter exports reports in HTML format
// HTML is ideal for:
//   - Browser viewing
//   - Interactive reports
//   - Presentation to management
//   - Sharing via web
type HTMLExporter struct{}

// Export implements the Exporter interface for HTML format
// Creates an interactive HTML report with professional styling and real card icons
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .html)
//
// Returns:
//   - error: Error if file can't be written
func (e *HTMLExporter) Export(report *Report, filename string) error {
	var html strings.Builder

	// ============================================================
	// HELPER FUNCTIONS
	// ============================================================

	// getCardIcon returns the image URL for each card brand
	// Using placeholder URLs - replace with actual icon URLs or local files
	getCardIcon := func(cardType string) string {
		// Map card types to their icon URLs
		// You can replace these URLs with:
		// 1. Local file paths (e.g., "./icons/visa.png")
		// 2. Your own hosted icons
		// 3. CDN URLs from icon libraries
		icons := map[string]string{
			"Visa":               "https://img.icons8.com/color/48/visa.png",
			"MasterCard":         "https://img.icons8.com/color/48/mastercard.png",
			"Amex":               "https://img.icons8.com/color/48/amex.png",
			"Discover":           "https://img.icons8.com/color/48/discover.png",
			"Diners":             "https://img.icons8.com/color/48/diners-club.png",
			"JCB":                "https://img.icons8.com/color/48/jcb.png",
			"UnionPay":           "https://img.icons8.com/color/48/unionpay.png",
			"Maestro":            "https://img.icons8.com/color/48/maestro.png",
			"RuPay":              "https://img.icons8.com/color/48/rupay.png",
			"Troy":               "https://www.freelogovectors.net/wp-content/uploads/2024/01/troy-odeme-logo-freelogovectors.net_.png",
			"Mir":                "https://upload.wikimedia.org/wikipedia/commons/b/b9/Mir-logo.SVG.svg",
			"Visa Electron":      "https://e7.pngegg.com/pngimages/598/1008/png-clipart-visa-electron-credit-card-debit-card-payment-mastercard-purple-text.png",
			"Elo":                "https://w7.pngwing.com/pngs/755/199/png-transparent-elo-payment-method-card-icon.png",
			"Hipercard":          "https://cdn.freebiesupply.com/logos/large/2x/hipercard-logo-svg-vector.svg",
			"Aura":               "https://images.seeklogo.com/logo-png/1/2/aura-logo-png_seeklogo-13605.png",
			"Argencard":          "https://static.openfintech.io/payment_methods/argencard/logo.png",
			"Naranja":            "https://upload.wikimedia.org/wikipedia/commons/f/f5/Logo_Naranja.png",
			"Cabal":              "https://www.novalnet.com/wp-content/uploads/2020/08/cabal_credito.png",
			"Banamex":            "https://images.seeklogo.com/logo-png/1/2/banamex-logo-png_seeklogo-15891.png",
			"BCCard":             "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRuDKVOgMVkXgd9lHg9BT3glkDWWfpyId0QrA&s",
			"Uzcard":             "https://kdb.uz/storage/cards/October2021/hNE9Tjbf0qf181qpgGah.jpg",
			"Humo":               "https://humocard.uz/upload/medialibrary/208/8x0p9hi3h9jww0flwdm92dayhn0flulj/humo-logo-more.png",
			"Verve":              "https://w7.pngwing.com/pngs/339/296/png-transparent-verve-payment-method-icon.png",
			"Dankort":            "https://images.ctfassets.net/gxv815mh8y8i/2sNXfrUCyCkEzMFTAZsxDa/b6b33ad387b650009185a809e57beef5/Dankort_logo.png",
			"Forbrugsforeningen": "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTfjlPoGCHt30UT-ASkFAtWVLK0AxIvQmAlFg&s",
			"InterPayment":       "https://portal.interpayments.com/assets/logo.png",
			"InstaPayment":       "https://cdn-icons-png.flaticon.com/512/2695/2695971.png",
			"NPS Pridnestrovie":  "https://cdn-icons-png.flaticon.com/512/2695/2695971.png",
			"UATP":               "https://uatp.com/wp-content/uploads/2022/11/UATP_CMYK_RegisteredMark.png",
			"EFTPOS":             "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcT8j2xQl15Fo4REBTVnjVVuKmzIRxbRw0DIng&s",
			"EBT":                "https://cdn-icons-png.flaticon.com/512/2695/2695971.png",
			"UkrCard":            "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQQJUHv9ljvG3-wQy9TRGS8sHkv1OllNAeGiQ&s",
			"BelCart":            "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcR_iAOAbIwwFS4fOhz525o8ZHTyyYzYop8nSA&s",
			"NSMEP":              "https://cdn-icons-png.flaticon.com/512/2695/2695971.png",
			"LankaPay":           "https://lankapay.net/public/website/assets/images/logo.png",
			"PayPak":             "https://crystalpng.com/wp-content/uploads/2025/10/paypak-logo.png",
		}

		if iconURL, ok := icons[cardType]; ok {
			return fmt.Sprintf(`<img src="%s" alt="%s" style="width: 32px; height: 20px; object-fit: contain;">`, iconURL, cardType)
		}

		// Default generic card icon
		return `<img src="https://img.icons8.com/color/48/bank-cards.png" alt="Card" style="width: 32px; height: 20px; object-fit: contain;">`
	}

	// getCardColor returns professional colors for each card brand
	// Colors are based on:
	// - Official brand guidelines (for major brands)
	// - National colors (for government-issued cards)
	// - Industry standards (for specialized cards)
	// - Regional color preferences (for regional cards)
	getCardColor := func(cardType string) string {
		colors := map[string]string{
			"Visa":               "#1A1F71",
			"MasterCard":         "#EB001B",
			"Amex":               "#006FCF",
			"Discover":           "#FF6000",
			"Diners":             "#0079BE",
			"JCB":                "#0E4C96",
			"UnionPay":           "#E21836",
			"Maestro":            "#0099DF",
			"RuPay":              "#097CBE",
			"Troy":               "#00ADEF",
			"Mir":                "#4DB45E",
			"Visa Electron":      "#1A1F71",
			"Elo":                "#FFCB05",
			"Hipercard":          "#D32F2F",
			"Aura":               "#FF6B35",
			"Argencard":          "#00A859",
			"Naranja":            "#FF6F00",
			"Cabal":              "#003DA5",
			"Banamex":            "#ED1C24",
			"BCCard":             "#0033A0",
			"Uzcard":             "#0066CC",
			"Humo":               "#00A3E0",
			"Verve":              "#00425F",
			"PayPak":             "#006747",
			"LankaPay":           "#FF7900",
			"Meeza":              "#C8102E",
			"ArCa":               "#0033A0",
			"Dankort":            "#ED1B2E",
			"Forbrugsforeningen": "#8B4513",
			"InterPayment":       "#2E3192",
			"InstaPayment":       "#009FE3",
			"NPS Pridnestrovie":  "#6B8E23",
			"UkrCard":            "#005BBB",
			"BelCart":            "#C8102E",
			"NSMEP":              "#DA291C",
			"UATP":               "#003087",
			"EFTPOS":             "#00A651",
			"EBT":                "#4A90E2",
		}

		if color, ok := colors[cardType]; ok {
			return color
		}

		// Default color for unknown card types
		// Using Bootstrap's secondary gray
		return "#6C757D"
	}

	// ============================================================
	// HTML HEAD AND PROFESSIONAL STYLES
	// ============================================================

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PAN Scanner Report - BasicPanScanner v` + report.Version + `</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/3.9.1/chart.min.js"></script>
    <style>
        /* ============================================================
           GLOBAL STYLES
           ============================================================ */
        * { 
            margin: 0; 
            padding: 0; 
            box-sizing: border-box; 
        }
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
            background: #f5f7fa;
            color: #2c3e50;
            line-height: 1.6;
            padding: 20px;
        }
        
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 2px 20px rgba(0,0,0,0.08);
            overflow: hidden;
        }

        /* ============================================================
           HEADER SECTION
           ============================================================ */
        .header {
            background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
            color: white;
            padding: 50px 40px;
            text-align: center;
            border-bottom: 4px solid #3498db;
        }
        
        .header h1 { 
            font-size: 32px; 
            margin-bottom: 8px; 
            font-weight: 600;
            letter-spacing: -0.5px;
        }
        
        .header .subtitle { 
            font-size: 16px; 
            opacity: 0.9;
            font-weight: 400;
        }
        
        .header .version { 
            font-size: 13px; 
            opacity: 0.75;
            margin-top: 10px;
            font-weight: 300;
        }

        /* ============================================================
           CONTENT AREA
           ============================================================ */
        .content { 
            padding: 40px; 
        }

        /* ============================================================
           EXECUTIVE SUMMARY - Professional Design
           ============================================================ */
        .executive-summary {
            background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
            color: white;
            padding: 40px;
            border-radius: 10px;
            margin-bottom: 35px;
            box-shadow: 0 4px 15px rgba(52, 152, 219, 0.2);
        }
        
        .executive-summary h2 {
            font-size: 24px;
            margin-bottom: 25px;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 12px;
            border-bottom: 2px solid rgba(255,255,255,0.3);
            padding-bottom: 15px;
        }
        
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 25px;
        }
        
        .summary-item {
            background: rgba(255,255,255,0.15);
            padding: 20px;
            border-radius: 8px;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255,255,255,0.2);
        }
        
        .summary-item .label {
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 1px;
            opacity: 0.9;
            margin-bottom: 8px;
            font-weight: 500;
        }
        
        .summary-item .value {
            font-size: 28px;
            font-weight: 700;
            line-height: 1;
        }
        
        .summary-item .subtext {
            font-size: 13px;
            opacity: 0.85;
            margin-top: 5px;
        }

        /* ============================================================
           STATISTICS CARDS - Professional Layout
           ============================================================ */
        .stats-section {
            margin: 40px 0;
        }
        
        .stats-section h2 {
            color: #2c3e50;
            margin-bottom: 25px;
            padding-bottom: 12px;
            border-bottom: 2px solid #e8eaf0;
            font-size: 22px;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(450px, 1fr));
            gap: 25px;
            margin-bottom: 30px;
        }
        
        .stats-card {
            background: #ffffff;
            padding: 30px;
            border-radius: 10px;
            border: 1px solid #e8eaf0;
            box-shadow: 0 2px 8px rgba(0,0,0,0.04);
            transition: all 0.3s ease;
        }
        
        .stats-card:hover {
            box-shadow: 0 4px 16px rgba(0,0,0,0.08);
            transform: translateY(-2px);
        }
        
        .stats-card h3 {
            color: #2c3e50;
            margin-bottom: 20px;
            font-size: 18px;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        
        .chart-container {
            position: relative;
            height: 300px;
        }

        /* ============================================================
           RISK ASSESSMENT - Professional Badges
           ============================================================ */
        .risk-item {
            margin: 12px 0;
            display: flex;
            align-items: center;
            gap: 15px;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 8px;
            border-left: 4px solid transparent;
            transition: all 0.2s ease;
        }
        
        .risk-item:hover {
            background: #f0f2f5;
            transform: translateX(5px);
        }
        
        .badge {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 8px 16px;
            border-radius: 6px;
            font-size: 12px;
            font-weight: 600;
            min-width: 110px;
            text-align: center;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .badge-high { 
            background: #e74c3c;
            color: white;
        }
        
        .risk-item:has(.badge-high) {
            border-left-color: #e74c3c;
        }
        
        .badge-medium { 
            background: #f39c12;
            color: white;
        }
        
        .risk-item:has(.badge-medium) {
            border-left-color: #f39c12;
        }
        
        .badge-low { 
            background: #27ae60;
            color: white;
        }
        
        .risk-item:has(.badge-low) {
            border-left-color: #27ae60;
        }

        /* ============================================================
           ACCORDION - Professional File Findings
           ============================================================ */
        .findings-section { 
            margin-top: 40px; 
        }
        
        .accordion-item {
            background: white;
            border: 1px solid #e8eaf0;
            border-radius: 10px;
            margin: 15px 0;
            overflow: hidden;
            transition: all 0.3s ease;
        }
        
        .accordion-item:hover {
            box-shadow: 0 4px 12px rgba(0,0,0,0.08);
        }
        
        .accordion-header {
            background: linear-gradient(135deg, #34495e 0%, #2c3e50 100%);
            color: white;
            padding: 20px 30px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
            user-select: none;
            transition: all 0.3s ease;
        }
        
        .accordion-header:hover {
            background: linear-gradient(135deg, #2c3e50 0%, #1a252f 100%);
        }
        
        .accordion-header .file-info {
            flex: 1;
            min-width: 0; /* Allow text truncation */
        }
        
        .accordion-header .file-name {
            font-weight: 600;
            font-size: 16px;
            margin-bottom: 6px;
        }
        
        .accordion-header .file-path {
            font-size: 12px;
            opacity: 0.85;
            font-family: 'Courier New', 'Consolas', monospace;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
        
        .accordion-header .card-count {
            background: rgba(255,255,255,0.2);
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 600;
            margin: 0 15px;
            min-width: 60px;
            text-align: center;
        }
        
        .accordion-header .toggle-icon {
            font-size: 24px;
            transition: transform 0.3s ease;
            opacity: 0.9;
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
            max-height: 5000px; /* Increased for full paths */
        }
        
        .accordion-content { 
            padding: 25px 30px; 
        }

        /* ============================================================
           FINDING ITEMS - Professional Card Display
           ============================================================ */
        .finding-item {
            background: white;
            padding: 18px 20px;
            margin: 12px 0;
            border-radius: 8px;
            border-left: 4px solid #3498db;
            display: grid;
            grid-template-columns: 80px auto 1fr;
            gap: 20px;
            align-items: center;
            transition: all 0.2s ease;
            box-shadow: 0 1px 3px rgba(0,0,0,0.06);
        }
        
        .finding-item:hover {
            transform: translateX(5px);
            box-shadow: 0 3px 8px rgba(0,0,0,0.12);
        }
        
        .finding-line {
            color: #7f8c8d;
            font-family: 'Courier New', 'Consolas', monospace;
            font-weight: 600;
            font-size: 13px;
        }
        
        .finding-type {
            display: flex;
            align-items: center;
            gap: 10px;
            font-weight: 600;
            font-size: 14px;
        }
        
        .finding-card {
            font-family: 'Courier New', 'Consolas', monospace;
            color: #e74c3c;
            font-weight: 700;
            font-size: 15px;
            letter-spacing: 0.5px;
        }

        /* ============================================================
           NO FINDINGS STATE
           ============================================================ */
        .no-findings {
            text-align: center;
            padding: 80px 20px;
            background: linear-gradient(135deg, #eafaf1 0%, #d5f4e6 100%);
            border-radius: 12px;
            margin: 30px 0;
            border: 2px dashed #27ae60;
        }
        
        .no-findings .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        
        .no-findings h3 {
            color: #27ae60;
            font-size: 24px;
            margin-bottom: 10px;
            font-weight: 600;
        }
        
        .no-findings p {
            color: #2d7a6e;
            font-size: 16px;
            font-weight: 400;
        }

        /* ============================================================
           FOOTER
           ============================================================ */
        footer {
            background: #2c3e50;
            color: white;
            padding: 30px;
            text-align: center;
            border-top: 4px solid #3498db;
        }
        
        footer p { 
            margin: 5px 0; 
            opacity: 0.9;
            font-size: 14px;
        }
        
        footer .timestamp {
            font-size: 12px;
            opacity: 0.7;
            margin-top: 10px;
        }

        /* ============================================================
           RESPONSIVE DESIGN
           ============================================================ */
        @media (max-width: 768px) {
            .content { padding: 20px; }
            .header { padding: 30px 20px; }
            .summary-grid { grid-template-columns: 1fr; }
            .stats-grid { grid-template-columns: 1fr; }
            .finding-item { 
                grid-template-columns: 1fr; 
                gap: 10px; 
            }
            .accordion-header { 
                flex-direction: column; 
                gap: 10px; 
                align-items: flex-start;
            }
            .accordion-header .card-count {
                margin: 0;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- ============================================================
             HEADER
             ============================================================ -->
        <div class="header">
            <h1>🔐 BasicPanScanner Security Report</h1>
            <div class="subtitle">PCI DSS Compliance Scanner</div>
            <div class="version">Version ` + report.Version + `</div>
        </div>
        
        <div class="content">`)

	// ============================================================
	// EXECUTIVE SUMMARY - Professional Format
	// ============================================================

	riskLevel, riskColor := report.GetRiskLevel()

	// Risk indicator emoji
	riskEmoji := "🟢"
	if riskLevel == "High" {
		riskEmoji = "🔴"
	} else if riskLevel == "Medium" {
		riskEmoji = "🟡"
	}

	html.WriteString(`
            <div class="executive-summary">
                <h2>📊 Executive Summary</h2>
                <div class="summary-grid">
                    <div class="summary-item">
                        <div class="label">Scan Date</div>
                        <div class="value">` + report.ScanDate.Format("Jan 2, 2006") + `</div>
                        <div class="subtext">` + report.ScanDate.Format("15:04:05 MST") + `</div>
                    </div>
                    <div class="summary-item">
                        <div class="label">Duration</div>
                        <div class="value">` + report.GetFormattedDuration() + `</div>
                        <div class="subtext">` + fmt.Sprintf("%.1f files/sec", report.ScanRate) + `</div>
                    </div>
                    <div class="summary-item">
                        <div class="label">Files Scanned</div>
                        <div class="value">` + fmt.Sprintf("%d", report.ScannedFiles) + `</div>
                        <div class="subtext">of ` + fmt.Sprintf("%d", report.TotalFiles) + ` total files</div>
                    </div>
                    <div class="summary-item">
                        <div class="label">Cards Found</div>
                        <div class="value">` + fmt.Sprintf("%d", report.CardsFound) + `</div>
                        <div class="subtext">in ` + fmt.Sprintf("%d", report.Statistics.FilesWithCards) + ` files</div>
                    </div>
                    <div class="summary-item" style="background: ` + riskColor + `20; border: 2px solid ` + riskColor + `;">
                        <div class="label">Risk Level</div>
                        <div class="value">` + riskEmoji + ` ` + riskLevel + `</div>
                        <div class="subtext">Overall assessment</div>
                    </div>
                </div>
            </div>`)

	// ============================================================
	// STATISTICS WITH CHARTS
	// ============================================================

	if report.CardsFound > 0 {
		html.WriteString(`
            <div class="stats-section">
                <h2>📈 Detailed Statistics</h2>
                <div class="stats-grid">`)

		// Card Type Distribution Chart
		if len(report.Statistics.CardsByType) > 0 {
			// Prepare data for chart
			var cardLabels []string
			var cardCounts []int
			var cardColors []string

			// Sort by count for better visualization
			type cardStat struct {
				name  string
				count int
			}
			var cardStats []cardStat
			for cardType, count := range report.Statistics.CardsByType {
				cardStats = append(cardStats, cardStat{cardType, count})
			}
			sort.Slice(cardStats, func(i, j int) bool {
				return cardStats[i].count > cardStats[j].count
			})

			for _, cs := range cardStats {
				cardLabels = append(cardLabels, cs.name)
				cardCounts = append(cardCounts, cs.count)
				cardColors = append(cardColors, getCardColor(cs.name))
			}

			html.WriteString(`
                    <div class="stats-card">
                        <h3>💳 Card Type Distribution</h3>
                        <div class="chart-container">
                            <canvas id="cardTypeChart"></canvas>
                        </div>
                    </div>

                    <script>
                    new Chart(document.getElementById('cardTypeChart'), {
                        type: 'doughnut',
                        data: {
                            labels: ` + toJSONArray(cardLabels) + `,
                            datasets: [{
                                data: ` + toJSONIntArray(cardCounts) + `,
                                backgroundColor: ` + toJSONArray(cardColors) + `,
                                borderWidth: 3,
                                borderColor: '#fff'
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    position: 'right',
                                    labels: { 
                                        padding: 15, 
                                        font: { size: 13, weight: '600' },
                                        color: '#2c3e50'
                                    }
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                            const percentage = ((context.parsed / total) * 100).toFixed(1);
                                            return context.label + ': ' + context.parsed + ' (' + percentage + '%)';
                                        }
                                    },
                                    backgroundColor: 'rgba(44, 62, 80, 0.9)',
                                    padding: 12,
                                    bodyFont: { size: 13 },
                                    titleFont: { size: 14, weight: 'bold' }
                                }
                            }
                        }
                    });
                    </script>`)
		}

		// Risk Assessment Card
		html.WriteString(`
                    <div class="stats-card">
                        <h3>⚠️ Risk Assessment</h3>
                        <div class="risk-item">
                            <span class="badge badge-high">High Risk</span>
                            <span style="flex: 1;">` + fmt.Sprintf("%d files with 5+ cards", report.Statistics.HighRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-medium">Medium Risk</span>
                            <span style="flex: 1;">` + fmt.Sprintf("%d files with 2-4 cards", report.Statistics.MediumRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-low">Low Risk</span>
                            <span style="flex: 1;">` + fmt.Sprintf("%d files with 1 card", report.Statistics.LowRiskFiles) + `</span>
                        </div>
                    </div>`)

		html.WriteString(`
                </div>
            </div>`)
	}

	// ============================================================
	// DETAILED FINDINGS (Accordion with Full Paths)
	// ============================================================

	if len(report.GroupedByFile) > 0 {
		html.WriteString(`
            <div class="stats-section">
                <h2>🔍 Detailed Findings</h2>`)

		// Sort file paths
		var filePaths []string
		for filePath := range report.GroupedByFile {
			filePaths = append(filePaths, filePath)
		}
		sort.Strings(filePaths)

		// Create accordion items with full paths
		for _, filePath := range filePaths {
			findings := report.GroupedByFile[filePath]
			fileName := filepath.Base(filePath)

			html.WriteString(fmt.Sprintf(`
                <div class="accordion-item">
                    <div class="accordion-header" onclick="toggleAccordion(this)">
                        <div class="file-info">
                            <div class="file-name">📄 %s</div>
                            <div class="file-path" title="%s">%s</div>
                        </div>
                        <span class="card-count">%d cards</span>
                        <span class="toggle-icon">▼</span>
                    </div>
                    <div class="accordion-body">
                        <div class="accordion-content">`,
				fileName,
				filePath,       // Full path in title (shows on hover)
				filePath,       // Full path displayed in header
				len(findings))) // THIS WAS MISSING - Card count!

			// Display each finding in this file
			for _, finding := range findings {
				cardIcon := getCardIcon(finding.CardType)

				html.WriteString(fmt.Sprintf(`
                            <div class="finding-item">
                                <div class="finding-line">Line %d</div>
                                <div class="finding-type">
                                    %s
                                    <span>%s</span>
                                </div>
                                <div class="finding-card">%s</div>
                            </div>`,
					finding.LineNumber,
					cardIcon,
					finding.CardType,
					finding.MaskedCard))
			}

			html.WriteString(`
                        </div>
                    </div>
                </div>`)
		}

		html.WriteString(`
            </div>`)
	} else {
		// No findings - show success message
		html.WriteString(`
            <div class="no-findings">
                <div class="icon">✅</div>
                <h3>No Credit Card Numbers Found</h3>
                <p>All scanned files are compliant with PCI DSS requirements</p>
            </div>`)
	}

	// ============================================================
	// FOOTER
	// ============================================================

	html.WriteString(`
        </div>
        
        <footer>
            <p><strong>BasicPanScanner v` + report.Version + `</strong></p>
            <p>PCI DSS Compliance Tool | Secure Card Detection</p>
            <p class="timestamp">Report generated on ` + report.ScanDate.Format("Monday, January 2, 2006 at 15:04:05 MST") + `</p>
        </footer>
    </div>
    
    <script>
        // ============================================================
        // ACCORDION TOGGLE FUNCTIONALITY
        // ============================================================
        // This function handles the expand/collapse behavior of findings
        //
        // Parameters:
        //   - header: The clicked accordion header element
        //
        // Behavior:
        //   - Collapses all other accordion items
        //   - Toggles the clicked item (expand if collapsed, collapse if expanded)
        //   - Smooth animation via CSS transition
        function toggleAccordion(header) {
            // Get the parent accordion item
            const item = header.parentElement;
            
            // Check if this item is currently active
            const wasActive = item.classList.contains('active');
            
            // Close all accordion items first
            // This ensures only one item is open at a time
            document.querySelectorAll('.accordion-item').forEach(accordionItem => {
                accordionItem.classList.remove('active');
            });
            
            // If the clicked item wasn't active, activate it
            // This creates a toggle effect
            if (!wasActive) {
                item.classList.add('active');
            }
        }
    </script>
</body>
</html>`)

	// ============================================================
	// WRITE TO FILE
	// ============================================================

	// Write the complete HTML to the file
	// Using 0644 permissions: owner can read/write, group and others can read
	return os.WriteFile(filename, []byte(html.String()), 0644)
}

// ============================================================
// HELPER FUNCTIONS FOR JSON CONVERSION
// ============================================================

// toJSONArray converts a string slice to JSON array format
// This is used for Chart.js data labels
//
// Parameters:
//   - items: Slice of strings to convert
//
// Returns:
//   - string: JSON array format (e.g., '["Visa","Mastercard","Amex"]')
//
// Example:
//
//	input:  []string{"Visa", "Mastercard", "Amex"}
//	output: '["Visa","Mastercard","Amex"]'
//
// Note: This function properly escapes quotes in item values
func toJSONArray(items []string) string {
	var parts []string
	for _, item := range items {
		// Escape any quotes in the item value
		// This prevents JSON syntax errors
		escaped := strings.ReplaceAll(item, `"`, `\"`)
		parts = append(parts, `"`+escaped+`"`)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// toJSONIntArray converts an int slice to JSON array format
// This is used for Chart.js data values
//
// Parameters:
//   - items: Slice of integers to convert
//
// Returns:
//   - string: JSON array format (e.g., '[15,8,3]')
//
// Example:
//
//	input:  []int{15, 8, 3}
//	output: '[15,8,3]'
//
// Note: Integers don't need quotes in JSON
func toJSONIntArray(items []int) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%d", item))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
