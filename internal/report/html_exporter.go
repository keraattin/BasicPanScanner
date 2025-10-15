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

	// Helper function to get card icon
	getCardIcon := func(cardType string) string {
		return "💳" // Simple credit card emoji for all types
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
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
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
        .header h1 { font-size: 32px; margin-bottom: 10px; }
        .header .version { font-size: 14px; opacity: 0.9; }
        .content { padding: 40px; }
        
        /* Summary Cards */
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 25px;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .summary-label {
            font-size: 13px;
            opacity: 0.9;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .summary-value {
            font-size: 32px;
            font-weight: bold;
            margin-top: 10px;
        }
        
        /* Statistics */
        .stats-section { margin: 30px 0; }
        .stats-section h2 {
            color: #2c3e50;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid #667eea;
        }
        
        /* Accordion */
        .accordion-item {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 8px;
            margin: 12px 0;
            overflow: hidden;
        }
        .accordion-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 18px 25px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .accordion-header:hover {
            background: linear-gradient(135deg, #5568d3 0%, #6a4291 100%);
        }
        .accordion-body {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.4s ease;
            background: #f8f9fa;
        }
        .accordion-item.active .accordion-body {
            max-height: 2000px;
        }
        .accordion-content { padding: 20px 25px; }
        
        /* Finding Item */
        .finding-item {
            background: white;
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            display: grid;
            grid-template-columns: 80px 120px 1fr;
            gap: 20px;
            align-items: center;
        }
        .finding-line {
            color: #7f8c8d;
            font-family: 'Courier New', monospace;
            font-weight: bold;
        }
        .finding-type {
            font-weight: 600;
            color: #2c3e50;
        }
        .finding-card {
            font-family: 'Courier New', monospace;
            color: #e74c3c;
            font-weight: bold;
        }
        
        footer {
            background: #2c3e50;
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        @media (max-width: 768px) {
            .content { padding: 20px; }
            .summary-grid { grid-template-columns: 1fr; }
            .finding-item { grid-template-columns: 1fr; gap: 10px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 BasicPanScanner Security Report</h1>
            <div class="version">Version ` + report.Version + ` | PCI Compliance Scanner</div>
        </div>
        
        <div class="content">`)

	// ============================================================
	// SUMMARY CARDS
	// ============================================================

	riskLevel, riskColor := report.GetRiskLevel()
	_ = riskColor // We'll use this for styling

	html.WriteString(`
            <div class="summary-grid">
                <div class="summary-card">
                    <div class="summary-label">Scan Duration</div>
                    <div class="summary-value">` + report.Duration.Round(1).String() + `</div>
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
