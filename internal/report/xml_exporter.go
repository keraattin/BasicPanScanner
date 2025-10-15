// Package report - XML exporter
// Exports reports in XML format for enterprise systems
package report

import (
	"encoding/xml"
	"os"
	"sort"
)

// XMLExporter exports reports in XML format
// XML is ideal for:
//   - Enterprise system integration
//   - Legacy system compatibility
//   - Structured data exchange
//   - Compliance documentation
type XMLExporter struct{}

// Export implements the Exporter interface for XML format
// Creates a well-structured XML document
//
// Parameters:
//   - report: The report to export
//   - filename: Output filename (should end with .xml)
//
// Returns:
//   - error: Error if file can't be written or XML encoding fails
func (e *XMLExporter) Export(report *Report, filename string) error {
	// XML-friendly structures
	// (Go's XML encoder doesn't work well with maps, so we convert to slices)

	type XMLCardType struct {
		Name  string `xml:"name,attr"`
		Count int    `xml:"count,attr"`
	}

	type XMLFileType struct {
		Extension string `xml:"extension,attr"`
		Count     int    `xml:"count,attr"`
	}

	type XMLCardTypeInFile struct {
		Name  string `xml:"name,attr"`
		Count int    `xml:"count,attr"`
	}

	type XMLFileStats struct {
		FilePath  string              `xml:"FilePath"`
		CardCount int                 `xml:"CardCount"`
		CardTypes []XMLCardTypeInFile `xml:"CardTypes>CardType"`
	}

	type XMLStatistics struct {
		FilesWithCards  int            `xml:"FilesWithCards"`
		HighRiskFiles   int            `xml:"HighRiskFiles"`
		MediumRiskFiles int            `xml:"MediumRiskFiles"`
		LowRiskFiles    int            `xml:"LowRiskFiles"`
		CardsByType     []XMLCardType  `xml:"CardsByType>CardType"`
		FilesByType     []XMLFileType  `xml:"FilesByType>FileType"`
		TopFiles        []XMLFileStats `xml:"TopFiles>File"`
	}

	type XMLFinding struct {
		LineNumber int    `xml:"LineNumber"`
		CardType   string `xml:"CardType"`
		MaskedCard string `xml:"MaskedCard"`
		Timestamp  string `xml:"Timestamp"`
	}

	type XMLFileGroup struct {
		FilePath string       `xml:"path,attr"`
		Count    int          `xml:"count,attr"`
		Findings []XMLFinding `xml:"Finding"`
	}

	type XMLReport struct {
		XMLName      xml.Name `xml:"ScanReport"`
		Version      string   `xml:"version,attr"`
		ScanDate     string   `xml:"ScanInfo>ScanDate"`
		Directory    string   `xml:"ScanInfo>Directory"`
		Duration     string   `xml:"ScanInfo>Duration"`
		TotalFiles   int      `xml:"ScanInfo>TotalFiles"`
		ScannedFiles int      `xml:"ScanInfo>ScannedFiles"`
		TotalCards   int      `xml:"Summary>TotalCards"`
		Statistics   XMLStatistics
		FileGroups   []XMLFileGroup `xml:"Findings>FileGroup"`
	}

	// ============================================================
	// Convert map data to slices for XML
	// ============================================================

	// Convert CardsByType
	var cardTypes []XMLCardType
	for cardType, count := range report.Statistics.CardsByType {
		cardTypes = append(cardTypes, XMLCardType{
			Name:  cardType,
			Count: count,
		})
	}
	// Sort by count (descending)
	sort.Slice(cardTypes, func(i, j int) bool {
		return cardTypes[i].Count > cardTypes[j].Count
	})

	// Convert FilesByType
	var fileTypes []XMLFileType
	for fileExt, count := range report.Statistics.FilesByType {
		fileTypes = append(fileTypes, XMLFileType{
			Extension: fileExt,
			Count:     count,
		})
	}
	sort.Slice(fileTypes, func(i, j int) bool {
		return fileTypes[i].Count > fileTypes[j].Count
	})

	// Convert TopFiles
	var topFiles []XMLFileStats
	for _, fs := range report.Statistics.TopFiles {
		// Convert card types map in this file
		var cardTypesInFile []XMLCardTypeInFile
		for cardType, count := range fs.CardTypes {
			cardTypesInFile = append(cardTypesInFile, XMLCardTypeInFile{
				Name:  cardType,
				Count: count,
			})
		}

		topFiles = append(topFiles, XMLFileStats{
			FilePath:  fs.FilePath,
			CardCount: fs.CardCount,
			CardTypes: cardTypesInFile,
		})
	}

	// Convert findings
	var fileGroups []XMLFileGroup

	// Sort file paths for consistent output
	var filePaths []string
	for filePath := range report.GroupedByFile {
		filePaths = append(filePaths, filePath)
	}
	sort.Strings(filePaths)

	for _, filePath := range filePaths {
		findings := report.GroupedByFile[filePath]

		var xmlFindings []XMLFinding
		for _, f := range findings {
			xmlFindings = append(xmlFindings, XMLFinding{
				LineNumber: f.LineNumber,
				CardType:   f.CardType,
				MaskedCard: f.MaskedCard,
				Timestamp:  f.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		fileGroups = append(fileGroups, XMLFileGroup{
			FilePath: filePath,
			Count:    len(findings),
			Findings: xmlFindings,
		})
	}

	// ============================================================
	// Build XML report structure
	// ============================================================

	xmlReport := XMLReport{
		Version:      report.Version,
		ScanDate:     report.ScanDate.Format("2006-01-02T15:04:05Z07:00"),
		Directory:    report.Directory,
		Duration:     report.Duration.String(),
		TotalFiles:   report.TotalFiles,
		ScannedFiles: report.ScannedFiles,
		TotalCards:   report.CardsFound,
		Statistics: XMLStatistics{
			CardsByType:     cardTypes,
			FilesByType:     fileTypes,
			TopFiles:        topFiles,
			FilesWithCards:  report.Statistics.FilesWithCards,
			HighRiskFiles:   report.Statistics.HighRiskFiles,
			MediumRiskFiles: report.Statistics.MediumRiskFiles,
			LowRiskFiles:    report.Statistics.LowRiskFiles,
		},
		FileGroups: fileGroups,
	}

	// ============================================================
	// Write XML file
	// ============================================================

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write XML header
	_, err = file.WriteString(xml.Header)
	if err != nil {
		return err
	}

	// Create encoder with indentation
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	// Encode
	err = encoder.Encode(xmlReport)
	if err != nil {
		return err
	}

	// Final newline
	_, err = file.WriteString("\n")
	return err
}
