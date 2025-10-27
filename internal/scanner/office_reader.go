// Package scanner - Office Document Reader (Pure GO - Standard Library Only)
// File: internal/scanner/office_reader.go
//
// This file handles reading office documents using ONLY GO STANDARD LIBRARY
// NO external libraries, NO external servers, NO API calls - completely self-contained!
//
// HOW OFFICE DOCUMENTS WORK:
//   - DOCX/XLSX/PPTX are actually ZIP files
//   - Inside the ZIP, there are XML files with the content
//   - We use archive/zip to extract files
//   - We use encoding/xml to parse XML
//   - Both are GO standard library!
//
// SUPPORTED FORMATS:
//
//	✅ DOCX (Microsoft Word 2007+)
//	✅ XLSX (Microsoft Excel 2007+)
//	✅ PPTX (Microsoft PowerPoint 2007+) - Coming soon
//
// WHY THIS APPROACH?
//
//	✅ NO external dependencies
//	✅ NO external servers/processes
//	✅ NO API calls
//	✅ 100% self-contained
//	✅ Uses only standard library
//	✅ Complete control over code
//	✅ No security risks from external code
//
// LIMITATIONS:
//
//	❌ Old formats (.doc, .xls, .ppt) not supported - they use binary format
//	❌ PDF not supported - requires complex PDF parsing
//	⚠️  Only extracts plain text (no formatting, images, etc.)
//
// HOW IT WORKS:
//  1. Open file as ZIP archive (archive/zip)
//  2. Find the XML file with content inside the ZIP
//  3. Parse XML to extract text (encoding/xml)
//  4. Return plain text for credit card scanning
package scanner

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// ============================================================
// OFFICE DOCUMENT TYPE DETECTION
// ============================================================

// isOfficeDocument checks if a file is an office document we can parse
//
// We only support modern Office formats (2007+) because they are ZIP+XML.
// Old formats (.doc, .xls, .ppt) use binary format and would require
// complex parsing that's beyond standard library capabilities.
//
// Parameters:
//   - filePath: Full path to the file
//
// Returns:
//   - bool: true if file is a supported office document
//
// Example:
//
//	isOfficeDocument("report.docx")  // true - supported!
//	isOfficeDocument("data.xlsx")    // true - supported!
//	isOfficeDocument("old.doc")      // false - old format, not supported
//	isOfficeDocument("file.pdf")     // false - not supported
func isOfficeDocument(filePath string) bool {
	// Extract file extension (e.g., ".docx")
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check if extension is a supported modern Office format
	// These are ZIP files with XML inside - we can parse them!
	switch ext {
	case ".docx": // Word 2007+
		return true
	case ".xlsx": // Excel 2007+
		return true
	case ".pptx": // PowerPoint 2007+ (basic support)
		return true
	default:
		return false
	}
}

// ============================================================
// MAIN OFFICE DOCUMENT READER FUNCTION
// ============================================================

// readOfficeDocument extracts text from office documents
//
// This is the MAIN FUNCTION that handles all office document types.
// It automatically detects the file type and uses the appropriate parser.
//
// PROCESS:
//  1. Detect file type from extension
//  2. Call appropriate parser (readDOCX, readXLSX, or readPPTX)
//  3. Return extracted text as string
//
// Parameters:
//   - filePath: Full path to the office document
//
// Returns:
//   - string: Extracted text content from the document
//   - error: Error if file can't be read or format is unsupported
//
// Example:
//
//	// Extract text from Word document
//	text, err := readOfficeDocument("report.docx")
//	if err != nil {
//	    log.Printf("Failed to extract text: %v", err)
//	    return
//	}
//	// Now 'text' contains all text from the document
//	// We can scan it for credit card numbers
func readOfficeDocument(filePath string) (string, error) {
	// Get file extension to determine document type
	ext := strings.ToLower(filepath.Ext(filePath))

	// Route to appropriate parser based on file type
	switch ext {
	case ".docx":
		// Handle Microsoft Word documents
		return readDOCX(filePath)

	case ".xlsx":
		// Handle Microsoft Excel spreadsheets
		return readXLSX(filePath)

	case ".pptx":
		// Handle Microsoft PowerPoint presentations
		return readPPTX(filePath)

	default:
		// Unsupported file type
		return "", fmt.Errorf("unsupported office document format: %s (only .docx, .xlsx, .pptx supported)", ext)
	}
}

// ============================================================
// DOCX READER (MICROSOFT WORD)
// ============================================================

// readDOCX extracts text from a Microsoft Word (.docx) file
//
// DOCX FORMAT STRUCTURE:
//   - A .docx file is a ZIP archive
//   - Inside: word/document.xml contains the main document text
//   - The XML has <w:t> tags that contain the actual text
//
// WHAT WE EXTRACT:
//
//	✅ All paragraphs (<w:p>)
//	✅ All text runs (<w:t>)
//	✅ Table content
//	✅ Headers and footers (if in word/document.xml)
//
// WHAT WE DON'T EXTRACT:
//
//	❌ Images (not needed for credit card scanning)
//	❌ Formatting (bold, italic, colors - we only need text)
//	❌ Comments and tracked changes
//	❌ Embedded objects
//
// Parameters:
//   - filePath: Full path to the .docx file
//
// Returns:
//   - string: All text content from the document
//   - error: Error if file can't be opened or parsed
//
// Example:
//
//	text, err := readDOCX("/documents/contract.docx")
//	if err != nil {
//	    log.Printf("Failed to read DOCX: %v", err)
//	    return
//	}
//	fmt.Printf("Extracted %d characters\n", len(text))
func readDOCX(filePath string) (string, error) {
	// ============================================================
	// STEP 1: Open DOCX file as ZIP archive
	// ============================================================
	// A .docx file is actually a ZIP file
	// We use archive/zip from standard library
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}
	defer zipReader.Close()

	// ============================================================
	// STEP 2: Find and read word/document.xml
	// ============================================================
	// The main document content is in word/document.xml
	// We need to find this file in the ZIP archive
	var documentXML string

	for _, file := range zipReader.File {
		// Look for the main document XML file
		// Path is: word/document.xml
		if file.Name == "word/document.xml" {
			// Open the file inside the ZIP
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			// Read the entire XML content
			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read document.xml: %w", err)
			}

			documentXML = string(xmlContent)
			break
		}
	}

	// Check if we found the document XML
	if documentXML == "" {
		return "", fmt.Errorf("document.xml not found in DOCX file (corrupted file?)")
	}

	// ============================================================
	// STEP 3: Parse XML to extract text
	// ============================================================
	// The XML contains text in <w:t> tags
	// Example: <w:t>This is some text</w:t>
	//
	// We'll parse the XML and extract all text from <w:t> tags
	text := extractTextFromWordXML(documentXML)

	return text, nil
}

// extractTextFromWordXML extracts text from Word XML content
//
// Word XML structure (simplified):
//
//	<w:document>
//	  <w:body>
//	    <w:p>              <!-- Paragraph -->
//	      <w:r>            <!-- Run (text with same formatting) -->
//	        <w:t>Hello</w:t>  <!-- Text content -->
//	      </w:r>
//	    </w:p>
//	  </w:body>
//	</w:document>
//
// We extract all <w:t> tags and concatenate them with spaces.
//
// Parameters:
//   - xmlContent: The XML content from document.xml
//
// Returns:
//   - string: Extracted plain text
func extractTextFromWordXML(xmlContent string) string {
	// We'll use a simple approach: extract all text between <w:t> and </w:t>
	// This is simpler than full XML parsing and works for our use case

	var result strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))

	// Flag to track if we're inside a <w:t> tag
	inTextTag := false

	for {
		// Read next XML token
		token, err := decoder.Token()
		if err == io.EOF {
			break // End of XML
		}
		if err != nil {
			// If parsing fails, return what we have so far
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check if this is a <w:t> tag (text tag)
			if t.Name.Local == "t" {
				inTextTag = true
			}

		case xml.CharData:
			// If we're inside a <w:t> tag, this is text content
			if inTextTag {
				text := string(t)
				result.WriteString(text)
				result.WriteString(" ") // Add space between text runs
			}

		case xml.EndElement:
			// Check if this is the end of <w:t> tag
			if t.Name.Local == "t" {
				inTextTag = false
			}
			// Add newline for paragraph end
			if t.Name.Local == "p" {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

// ============================================================
// XLSX READER (MICROSOFT EXCEL)
// ============================================================

// readXLSX extracts text from a Microsoft Excel (.xlsx) file
//
// XLSX FORMAT STRUCTURE:
//   - A .xlsx file is a ZIP archive
//   - Inside: xl/sharedStrings.xml contains all text strings
//   - Inside: xl/worksheets/sheet*.xml contains cell references and numbers
//
// WHAT WE EXTRACT:
//
//	✅ All text from sharedStrings (cell text values)
//	✅ Numbers from worksheet cells
//	✅ Content from all sheets
//
// WHAT WE DON'T EXTRACT:
//
//	❌ Formulas (only formula results)
//	❌ Charts and images
//	❌ Cell formatting (colors, borders, etc.)
//	❌ Hidden sheets or cells
//
// Parameters:
//   - filePath: Full path to the .xlsx file
//
// Returns:
//   - string: All text content from all sheets
//   - error: Error if file can't be opened or parsed
//
// Example:
//
//	text, err := readXLSX("/reports/financials.xlsx")
//	if err != nil {
//	    log.Printf("Failed to read XLSX: %v", err)
//	    return
//	}
//	fmt.Printf("Extracted %d characters\n", len(text))
func readXLSX(filePath string) (string, error) {
	// ============================================================
	// STEP 1: Open XLSX file as ZIP archive
	// ============================================================
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open XLSX as ZIP: %w", err)
	}
	defer zipReader.Close()

	// ============================================================
	// STEP 2: Read shared strings (text values)
	// ============================================================
	// Excel stores all text strings in xl/sharedStrings.xml
	// Cells reference these strings by index
	// For our purpose (credit card scanning), we just need all the text
	var result strings.Builder

	for _, file := range zipReader.File {
		// Look for shared strings XML
		if file.Name == "xl/sharedStrings.xml" {
			rc, err := file.Open()
			if err != nil {
				continue // Skip if can't open
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				continue // Skip if can't read
			}

			// Extract text from shared strings
			text := extractTextFromSharedStringsXML(string(xmlContent))
			result.WriteString(text)
			result.WriteString("\n")
		}

		// Also read worksheet files for numbers and additional data
		// Worksheet files are: xl/worksheets/sheet1.xml, sheet2.xml, etc.
		if strings.HasPrefix(file.Name, "xl/worksheets/sheet") && strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				continue // Skip if can't open
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				continue // Skip if can't read
			}

			// Extract values from worksheet
			text := extractTextFromWorksheetXML(string(xmlContent))
			result.WriteString(text)
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// extractTextFromSharedStringsXML extracts text from Excel shared strings XML
//
// Shared strings XML structure (simplified):
//
//	<sst>
//	  <si>              <!-- String item -->
//	    <t>Hello</t>    <!-- Text -->
//	  </si>
//	  <si>
//	    <t>World</t>
//	  </si>
//	</sst>
//
// Parameters:
//   - xmlContent: The XML content from sharedStrings.xml
//
// Returns:
//   - string: Extracted plain text
func extractTextFromSharedStringsXML(xmlContent string) string {
	var result strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	inTextTag := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check if this is a <t> tag (text tag in Excel)
			if t.Name.Local == "t" {
				inTextTag = true
			}

		case xml.CharData:
			// If we're inside a <t> tag, this is text content
			if inTextTag {
				text := string(t)
				result.WriteString(text)
				result.WriteString(" ")
			}

		case xml.EndElement:
			// Check if this is the end of <t> tag
			if t.Name.Local == "t" {
				inTextTag = false
			}
		}
	}

	return result.String()
}

// extractTextFromWorksheetXML extracts values from Excel worksheet XML
//
// Worksheet XML structure (simplified):
//
//	<worksheet>
//	  <sheetData>
//	    <row>
//	      <c>              <!-- Cell -->
//	        <v>123</v>     <!-- Value -->
//	      </c>
//	    </row>
//	  </sheetData>
//	</worksheet>
//
// Parameters:
//   - xmlContent: The XML content from sheet*.xml
//
// Returns:
//   - string: Extracted values
func extractTextFromWorksheetXML(xmlContent string) string {
	var result strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	inValueTag := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check if this is a <v> tag (value tag in Excel)
			if t.Name.Local == "v" {
				inValueTag = true
			}

		case xml.CharData:
			// If we're inside a <v> tag, this is a value
			if inValueTag {
				value := string(t)
				result.WriteString(value)
				result.WriteString(" ")
			}

		case xml.EndElement:
			// Check if this is the end of <v> tag
			if t.Name.Local == "v" {
				inValueTag = false
			}
		}
	}

	return result.String()
}

// ============================================================
// PPTX READER (MICROSOFT POWERPOINT)
// ============================================================

// readPPTX extracts text from a Microsoft PowerPoint (.pptx) file
//
// PPTX FORMAT STRUCTURE:
//   - A .pptx file is a ZIP archive
//   - Inside: ppt/slides/slide*.xml contains slide content
//   - Each slide has <a:t> tags with text
//
// WHAT WE EXTRACT:
//
//	✅ All text from all slides
//	✅ Text from text boxes
//	✅ Slide titles
//
// WHAT WE DON'T EXTRACT:
//
//	❌ Speaker notes (usually not needed for card scanning)
//	❌ Images and charts
//	❌ Slide formatting
//
// Parameters:
//   - filePath: Full path to the .pptx file
//
// Returns:
//   - string: All text content from all slides
//   - error: Error if file can't be opened or parsed
func readPPTX(filePath string) (string, error) {
	// ============================================================
	// STEP 1: Open PPTX file as ZIP archive
	// ============================================================
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PPTX as ZIP: %w", err)
	}
	defer zipReader.Close()

	// ============================================================
	// STEP 2: Read all slide files
	// ============================================================
	var result strings.Builder

	for _, file := range zipReader.File {
		// Look for slide XML files
		// Slide files are: ppt/slides/slide1.xml, slide2.xml, etc.
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				continue // Skip if can't open
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				continue // Skip if can't read
			}

			// Extract text from slide
			text := extractTextFromSlideXML(string(xmlContent))
			result.WriteString(text)
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// extractTextFromSlideXML extracts text from PowerPoint slide XML
//
// Slide XML structure (simplified):
//
//	<p:sld>
//	  <p:cSld>
//	    <p:spTree>
//	      <p:sp>              <!-- Shape (text box) -->
//	        <a:t>Text</a:t>   <!-- Text -->
//	      </p:sp>
//	    </p:spTree>
//	  </p:cSld>
//	</p:sld>
//
// Parameters:
//   - xmlContent: The XML content from slide*.xml
//
// Returns:
//   - string: Extracted plain text
func extractTextFromSlideXML(xmlContent string) string {
	var result strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	inTextTag := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check if this is a <a:t> tag (text tag in PowerPoint)
			if t.Name.Local == "t" {
				inTextTag = true
			}

		case xml.CharData:
			// If we're inside a <a:t> tag, this is text content
			if inTextTag {
				text := string(t)
				result.WriteString(text)
				result.WriteString(" ")
			}

		case xml.EndElement:
			// Check if this is the end of <a:t> tag
			if t.Name.Local == "t" {
				inTextTag = false
			}
		}
	}

	return result.String()
}
