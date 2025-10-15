// Package report - Exporter interface
// This file defines the interface for all report exporters
package report

// Exporter defines the interface that all export formats must implement
// This allows easy addition of new export formats
//
// Example implementation:
//   type MyExporter struct{}
//
//   func (e *MyExporter) Export(report *Report, filename string) error {
//       Implementation here
//       return nil
//   }
type Exporter interface {
	// Export writes the report to a file in the specific format
	//
	// Parameters:
	//   - report: The report to export
	//   - filename: Output filename
	//
	// Returns:
	//   - error: Error if export fails
	Export(report *Report, filename string) error
}

// Available exporters (implemented in separate files):
// - JSONExporter  - json_exporter.go
// - CSVExporter   - csv_exporter.go
// - TXTExporter   - txt_exporter.go
// - XMLExporter   - xml_exporter.go
// - HTMLExporter  - html_exporter.go
