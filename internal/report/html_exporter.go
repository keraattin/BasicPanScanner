// Package report - HTML exporter
// Exports reports in interactive HTML format
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
// Creates an interactive HTML report with accordion UI and statistics
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .html)
//
// Returns:
//   - error: Error if file can't be written
func (e *HTMLExporter) Export(report *Report, filename string) error {
	var html strings.Builder

	// Helper function to get card brand icon (using Unicode/CSS)
	getCardIcon := func(cardType string) string {
		icons := map[string]string{
			"Visa":       "💳", // We'll style these with CSS
			"MasterCard": "💳",
			"Amex":       "💳",
			"Discover":   "💳",
			"Diners":     "💳",
			"JCB":        "💳",
			"UnionPay":   "💳",
			"Maestro":    "💳",
			"RuPay":      "💳",
			"Troy":       "💳",
			"Mir":        "💳",
		}
		if icon, ok := icons[cardType]; ok {
			return icon
		}
		return "💳"
	}

	// Helper function to get card brand color
	getCardColor := func(cardType string) string {
		colors := map[string]string{
			"Visa":       "#1A1F71", // Visa blue
			"MasterCard": "#EB001B", // Mastercard red
			"Amex":       "#006FCF", // Amex blue
			"Discover":   "#FF6000", // Discover orange
			"Diners":     "#0079BE", // Diners blue
			"JCB":        "#0E4C96", // JCB blue
			"UnionPay":   "#E21836", // UnionPay red
			"Maestro":    "#0099DF", // Maestro blue
			"RuPay":      "#097CBE", // RuPay blue
			"Troy":       "#00ADEF", // Troy blue
			"Mir":        "#4DB45E", // Mir green
		}
		if color, ok := colors[cardType]; ok {
			return color
		}
		return "#667eea" // Default purple
	}

	// ============================================================
	// HTML HEAD AND STYLES
	// ============================================================

	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PAN Scanner Report - BasicPanScanner v` + report.Version + `</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/3.9.1/chart.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { font-size: 36px; margin-bottom: 10px; text-shadow: 2px 2px 4px rgba(0,0,0,0.2); }
        .header .version { font-size: 14px; opacity: 0.9; }
        .content { padding: 40px; }
        
        /* Executive Summary */
        .executive-summary {
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
            padding: 35px;
            border-radius: 12px;
            margin-bottom: 30px;
            box-shadow: 0 8px 20px rgba(0,0,0,0.15);
        }
        .executive-summary h2 {
            font-size: 28px;
            margin-bottom: 20px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .executive-summary .summary-text {
            font-size: 16px;
            line-height: 1.8;
            margin-bottom: 15px;
        }
        .executive-summary .highlight {
            font-weight: bold;
            font-size: 20px;
            background: rgba(255,255,255,0.2);
            padding: 2px 8px;
            border-radius: 4px;
        }
        .executive-summary .recommendation {
            margin-top: 20px;
            padding-top: 20px;
            border-top: 2px solid rgba(255,255,255,0.3);
            font-size: 15px;
        }
        .executive-summary .recommendation strong {
            display: block;
            margin-bottom: 10px;
            font-size: 18px;
        }
        
        /* Summary Cards */
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }
        .summary-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.2);
        }
        .summary-label {
            font-size: 13px;
            opacity: 0.9;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        .summary-value {
            font-size: 36px;
            font-weight: bold;
        }
        
        /* Statistics Section */
        .stats-section {
            margin: 40px 0;
        }
        .stats-section h2 {
            color: #2c3e50;
            margin-bottom: 25px;
            padding-bottom: 12px;
            border-bottom: 3px solid #667eea;
            font-size: 26px;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 30px;
            margin-bottom: 30px;
        }
        .stats-card {
            background: #f8f9fa;
            padding: 30px;
            border-radius: 12px;
            border-left: 4px solid #667eea;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .stats-card h3 {
            color: #2c3e50;
            margin-bottom: 20px;
            font-size: 20px;
        }
        .chart-container {
            position: relative;
            height: 300px;
        }
        
        /* Card Brand Badges */
        .card-badge {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 600;
            color: white;
            margin: 5px;
        }
        
        /* Risk Badges */
        .risk-item {
            margin: 15px 0;
            display: flex;
            align-items: center;
            gap: 15px;
            padding: 12px;
            background: white;
            border-radius: 8px;
        }
        .badge {
            display: inline-block;
            padding: 8px 20px;
            border-radius: 20px;
            font-size: 13px;
            font-weight: bold;
            min-width: 110px;
            text-align: center;
        }
        .badge-high { 
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
        }
        .badge-medium { 
            background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
            color: #8b4513;
        }
        .badge-low { 
            background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
            color: #2d7a6e;
        }
        
        /* Accordion */
        .findings-section { margin-top: 40px; }
        .accordion-item {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 10px;
            margin: 15px 0;
            overflow: hidden;
            transition: all 0.3s ease;
        }
        .accordion-item:hover {
            box-shadow: 0 6px 15px rgba(0,0,0,0.1);
        }
        .accordion-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px 30px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
            user-select: none;
            transition: background 0.3s ease;
        }
        .accordion-header:hover {
            background: linear-gradient(135deg, #5568d3 0%, #6a4291 100%);
        }
        .accordion-header .file-info {
            flex: 1;
        }
        .accordion-header .file-name {
            font-weight: 600;
            font-size: 16px;
            margin-bottom: 5px;
        }
        .accordion-header .file-path {
            font-size: 12px;
            opacity: 0.85;
            font-family: 'Courier New', monospace;
        }
        .accordion-header .card-count {
            margin: 0 20px;
            font-size: 15px;
        }
        .accordion-header .toggle-icon {
            font-size: 24px;
            transition: transform 0.3s ease;
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
            max-height: 3000px;
        }
        .accordion-content { padding: 25px 30px; }
        
        /* Finding Item */
        .finding-item {
            background: white;
            padding: 18px;
            margin: 12px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            display: grid;
            grid-template-columns: 100px 150px 1fr;
            gap: 20px;
            align-items: center;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .finding-item:hover {
            transform: translateX(5px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .finding-line {
            color: #7f8c8d;
            font-family: 'Courier New', monospace;
            font-weight: bold;
            font-size: 14px;
        }
        .finding-type {
            font-weight: 600;
            color: white;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 13px;
            text-align: center;
        }
        .finding-card {
            font-family: 'Courier New', monospace;
            color: #e74c3c;
            font-weight: bold;
            font-size: 16px;
        }
        
        /* No Findings */
        .no-findings {
            text-align: center;
            padding: 80px 20px;
            background: linear-gradient(135deg, #eafaf1 0%, #d5f4e6 100%);
            border-radius: 12px;
            margin: 30px 0;
        }
        .no-findings .icon {
            font-size: 72px;
            margin-bottom: 20px;
        }
        .no-findings h3 {
            color: #27ae60;
            font-size: 24px;
            margin-bottom: 10px;
        }
        .no-findings p {
            color: #2d7a6e;
            font-size: 16px;
        }
        
        /* Footer */
        footer {
            background: #2c3e50;
            color: white;
            padding: 30px;
            text-align: center;
            margin-top: 40px;
        }
        footer p { margin: 5px 0; opacity: 0.9; }
        
        /* Responsive */
        @media (max-width: 768px) {
            .content { padding: 20px; }
            .summary-grid { grid-template-columns: 1fr; }
            .stats-grid { grid-template-columns: 1fr; }
            .finding-item { grid-template-columns: 1fr; gap: 10px; }
            .accordion-header { flex-direction: column; gap: 10px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 BasicPanScanner Security Report</h1>
            <div class="version">Version ` + report.Version + ` | PCI DSS Compliance Scanner</div>
        </div>
        
        <div class="content">`)

	// ============================================================
	// EXECUTIVE SUMMARY
	// ============================================================

	riskLevel, riskColor := report.GetRiskLevel()

	// Build executive summary text
	summaryText := fmt.Sprintf(
		"On %s, a comprehensive security scan was conducted on directory <span class='highlight'>%s</span>, "+
			"covering <span class='highlight'>%d files</span> across various file types. "+
			"The scan completed in <span class='highlight'>%s</span>, analyzing files for credit card data "+
			"in compliance with PCI DSS requirements.",
		report.ScanDate.Format("January 2, 2006"),
		report.Directory,
		report.ScannedFiles,
		report.GetFormattedDuration(),
	)

	var findingsText, recommendationText string
	if report.CardsFound > 0 {
		findingsText = fmt.Sprintf(
			"The scan <strong>identified <span class='highlight' style='font-size: 24px;'>%d credit card number(s)</span></strong> "+
				"across <span class='highlight'>%d file(s)</span>. ",
			report.CardsFound,
			report.Statistics.FilesWithCards,
		)

		// Add card type breakdown
		if len(report.Statistics.CardsByType) > 0 {
			var cardTypes []string
			for cardType, count := range report.Statistics.CardsByType {
				cardTypes = append(cardTypes, fmt.Sprintf("%s (%d)", cardType, count))
			}
			findingsText += "Card types detected include: " + strings.Join(cardTypes, ", ") + ". "
		}

		// Risk level assessment
		findingsText += fmt.Sprintf(
			"Overall risk assessment: <strong style='color: %s; font-size: 20px;'>%s</strong>",
			riskColor, riskLevel,
		)

		// Recommendations based on findings
		if report.Statistics.HighRiskFiles > 0 {
			recommendationText = fmt.Sprintf(
				"<strong>⚠️ Critical Action Required</strong> "+
					"%d file(s) contain 5 or more credit card numbers, indicating a high risk of data exposure. "+
					"<strong>Immediate remediation is recommended:</strong> "+
					"<ul style='margin: 10px 0 0 20px;'>"+
					"<li>Review and remove or encrypt all exposed credit card data</li>"+
					"<li>Implement access controls to restrict file permissions</li>"+
					"<li>Conduct a thorough investigation of how card data entered these files</li>"+
					"<li>Update data handling procedures to prevent future exposure</li>"+
					"<li>Consider conducting regular automated scans for ongoing compliance</li>"+
					"</ul>",
				report.Statistics.HighRiskFiles,
			)
		} else if report.Statistics.MediumRiskFiles > 0 {
			recommendationText = fmt.Sprintf(
				"<strong>⚠️ Action Required</strong> "+
					"%d file(s) contain multiple credit card numbers. "+
					"<strong>Recommended actions:</strong> "+
					"<ul style='margin: 10px 0 0 20px;'>"+
					"<li>Review flagged files and remove sensitive data</li>"+
					"<li>Implement data masking or encryption where appropriate</li>"+
					"<li>Verify compliance with PCI DSS data storage requirements</li>"+
					"<li>Establish secure data handling protocols</li>"+
					"</ul>",
				report.Statistics.MediumRiskFiles,
			)
		} else {
			recommendationText =
				"<strong>✓ Low Risk Detected</strong> " +
					"While only single instances of card data were found, " +
					"<strong>action is still recommended:</strong> " +
					"<ul style='margin: 10px 0 0 20px;'>" +
					"<li>Review and remove identified card data</li>" +
					"<li>Verify the data is not required for business purposes</li>" +
					"<li>Implement preventive measures to avoid future exposure</li>" +
					"</ul>"
		}
	} else {
		findingsText = "<strong>✓ No credit card numbers were detected</strong> in any of the scanned files. "
		recommendationText =
			"<strong>✓ Compliant</strong> " +
				"The scanned directory appears to be compliant with PCI DSS requirements regarding cardholder data storage. " +
				"<strong>Recommendations for ongoing compliance:</strong> " +
				"<ul style='margin: 10px 0 0 20px;'>" +
				"<li>Continue regular security scans to maintain compliance</li>" +
				"<li>Ensure staff training on proper cardholder data handling</li>" +
				"<li>Review and update data security policies periodically</li>" +
				"</ul>"
	}

	html.WriteString(fmt.Sprintf(`
            <div class="executive-summary">
                <h2>📋 Executive Summary</h2>
                <div class="summary-text">%s</div>
                <div class="summary-text">%s</div>
                <div class="recommendation">%s</div>
            </div>`, summaryText, findingsText, recommendationText))

	// ============================================================
	// SUMMARY CARDS
	// ============================================================

	html.WriteString(`
            <div class="summary-grid">
                <div class="summary-card">
                    <div class="summary-label">Scan Duration</div>
                    <div class="summary-value">` + report.GetFormattedDuration() + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Files Scanned</div>
                    <div class="summary-value">` + fmt.Sprintf("%d", report.ScannedFiles) + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Cards Found</div>
                    <div class="summary-value">` + fmt.Sprintf("%d", report.CardsFound) + `</div>
                </div>
                <div class="summary-card">
                    <div class="summary-label">Risk Level</div>
                    <div class="summary-value">` + riskLevel + `</div>
                </div>
            </div>`)

	// ============================================================
	// STATISTICS WITH CHARTS
	// ============================================================

	if report.CardsFound > 0 {
		html.WriteString(`
            <div class="stats-section">
                <h2>📊 Detailed Statistics</h2>
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
                                borderWidth: 2,
                                borderColor: '#fff'
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                legend: {
                                    position: 'right',
                                    labels: { padding: 15, font: { size: 13 } }
                                },
                                tooltip: {
                                    callbacks: {
                                        label: function(context) {
                                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                            const percentage = ((context.parsed / total) * 100).toFixed(1);
                                            return context.label + ': ' + context.parsed + ' (' + percentage + '%)';
                                        }
                                    }
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
                            <span class="badge badge-high">HIGH RISK</span>
                            <span>` + fmt.Sprintf("%d files with 5+ cards", report.Statistics.HighRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-medium">MEDIUM RISK</span>
                            <span>` + fmt.Sprintf("%d files with 2-4 cards", report.Statistics.MediumRiskFiles) + `</span>
                        </div>
                        <div class="risk-item">
                            <span class="badge badge-low">LOW RISK</span>
                            <span>` + fmt.Sprintf("%d files with 1 card", report.Statistics.LowRiskFiles) + `</span>
                        </div>
                    </div>`)

		html.WriteString(`
                </div>
            </div>`)
	}

	// ============================================================
	// DETAILED FINDINGS (Accordion)
	// ============================================================

	if len(report.GroupedByFile) > 0 {
		html.WriteString(`
            <div class="stats-section">
                <h2>🔎 Detailed Findings</h2>`)

		// Sort file paths
		var filePaths []string
		for filePath := range report.GroupedByFile {
			filePaths = append(filePaths, filePath)
		}
		sort.Strings(filePaths)

		// Create accordion items
		for _, filePath := range filePaths {
			findings := report.GroupedByFile[filePath]

			html.WriteString(fmt.Sprintf(`
                <div class="accordion-item">
                    <div class="accordion-header" onclick="toggleAccordion(this)">
                        <span>%s</span>
                        <span>%d cards</span>
                    </div>
                    <div class="accordion-body">
                        <div class="accordion-content">`, filepath.Base(filePath), len(findings)))

			for _, finding := range findings {
				html.WriteString(fmt.Sprintf(`
                            <div class="finding-item">
                                <div class="finding-line">Line %d</div>
                                <div class="finding-type">%s %s</div>
                                <div class="finding-card">%s</div>
                            </div>`, finding.LineNumber, getCardIcon(finding.CardType), finding.CardType, finding.MaskedCard))
			}

			html.WriteString(`
                        </div>
                    </div>
                </div>`)
		}

		html.WriteString(`
            </div>`)
	} else {
		html.WriteString(`
            <div style="text-align: center; padding: 60px 20px; color: #27ae60; font-size: 20px;">
                <div style="font-size: 64px; margin-bottom: 20px;">✅</div>
                <div>No credit card numbers found</div>
                <p style="font-size: 16px; margin-top: 10px;">Files are compliant with PCI DSS</p>
            </div>`)
	}

	// ============================================================
	// FOOTER
	// ============================================================

	html.WriteString(`
        </div>
        
        <footer>
            <p><strong>BasicPanScanner v` + report.Version + `</strong> | PCI Compliance Tool</p>
            <p>Report generated on ` + report.ScanDate.Format("January 2, 2006 at 15:04:05") + `</p>
        </footer>
    </div>
    
    <script>
        function toggleAccordion(header) {
            const item = header.parentElement;
            const wasActive = item.classList.contains('active');
            
            document.querySelectorAll('.accordion-item').forEach(accordionItem => {
                accordionItem.classList.remove('active');
            });
            
            if (!wasActive) {
                item.classList.add('active');
            }
        }
    </script>
</body>
</html>`)

	// Write to file
	return os.WriteFile(filename, []byte(html.String()), 0644)
}

// Helper functions for JSON array conversion

// toJSONArray converts a string slice to JSON array format
func toJSONArray(items []string) string {
	var parts []string
	for _, item := range items {
		// Escape quotes in item
		escaped := strings.ReplaceAll(item, `"`, `\"`)
		parts = append(parts, `"`+escaped+`"`)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// toJSONIntArray converts an int slice to JSON array format
func toJSONIntArray(items []int) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%d", item))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
