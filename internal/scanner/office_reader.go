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
// We support modern Office formats (2007+) and OpenDocument formats
// because they all use ZIP+XML structure.
// Old formats (.doc, .xls, .ppt) use binary format and would require
// complex parsing that's beyond standard library capabilities.
//
// SUPPORTED FORMATS:
//
//	Microsoft Office 2007+ (ZIP+XML):
//	  - DOCX, DOCM, DOTX, DOTM (Word family)
//	  - XLSX, XLSM, XLTX, XLTM (Excel family)
//	  - PPTX, PPTM, POTX, POTM (PowerPoint family)
//	OpenDocument Format (ZIP+XML):
//	  - ODT (Text documents)
//	  - ODS (Spreadsheets)
//	  - ODP (Presentations)
//
// Parameters:
//   - filePath: Full path to the file
//
// Returns:
//   - bool: true if file is a supported office document
//
// Example:
//
//	isOfficeDocument("report.docx")  // true - Word document
//	isOfficeDocument("data.xlsm")    // true - Excel with macros
//	isOfficeDocument("doc.odt")      // true - OpenDocument text
//	isOfficeDocument("old.doc")      // false - old format, not supported
func isOfficeDocument(filePath string) bool {
	// Extract file extension (e.g., ".docx")
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check if extension is a supported modern Office format
	// These are ZIP files with XML inside - we can parse them!
	switch ext {
	// Microsoft Word family (all use same structure)
	case ".docx": // Word 2007+
		return true
	case ".docm": // Word with macros
		return true
	case ".dotx": // Word template
		return true
	case ".dotm": // Word template with macros
		return true

	// Microsoft Excel family (all use same structure)
	case ".xlsx": // Excel 2007+
		return true
	case ".xlsm": // Excel with macros
		return true
	case ".xltx": // Excel template
		return true
	case ".xltm": // Excel template with macros
		return true

	// Microsoft PowerPoint family (all use same structure)
	case ".pptx": // PowerPoint 2007+
		return true
	case ".pptm": // PowerPoint with macros
		return true
	case ".potx": // PowerPoint template
		return true
	case ".potm": // PowerPoint template with macros
		return true

	// OpenDocument Format (LibreOffice/OpenOffice)
	case ".odt": // Text document (like Word)
		return true
	case ".ods": // Spreadsheet (like Excel)
		return true
	case ".odp": // Presentation (like PowerPoint)
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
//  2. Call appropriate parser (readDOCX, readXLSX, readPPTX, or OpenDocument parsers)
//  3. Return extracted text as string
//
// SUPPORTED FORMATS:
//   - Microsoft Office 2007+ (14 formats)
//   - OpenDocument Format (3 formats)
//   - Total: 17 office document formats!
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
//	Extract text from various formats
//	text, err := readOfficeDocument("report.docx")    // Word
//	text, err := readOfficeDocument("data.xlsm")      // Excel with macros
//	text, err := readOfficeDocument("slides.pptx")    // PowerPoint
//	text, err := readOfficeDocument("document.odt")   // OpenDocument
func readOfficeDocument(filePath string) (string, error) {
	// Get file extension to determine document type
	ext := strings.ToLower(filepath.Ext(filePath))

	// Route to appropriate parser based on file type
	switch ext {
	// ============================================================
	// MICROSOFT WORD FAMILY
	// ============================================================
	// All Word formats use the same XML structure (word/document.xml)
	// The difference is just additional files for macros/templates
	case ".docx", ".docm", ".dotx", ".dotm":
		return readDOCX(filePath)

	// ============================================================
	// MICROSOFT EXCEL FAMILY
	// ============================================================
	// All Excel formats use the same XML structure
	// (xl/sharedStrings.xml and xl/worksheets/*.xml)
	case ".xlsx", ".xlsm", ".xltx", ".xltm":
		return readXLSX(filePath)

	// ============================================================
	// MICROSOFT POWERPOINT FAMILY
	// ============================================================
	// All PowerPoint formats use the same XML structure
	// (ppt/slides/*.xml)
	case ".pptx", ".pptm", ".potx", ".potm":
		return readPPTX(filePath)

	// ============================================================
	// OPENDOCUMENT TEXT (LibreOffice/OpenOffice)
	// ============================================================
	// ODT uses content.xml for main content
	case ".odt":
		return readODT(filePath)

	// ============================================================
	// OPENDOCUMENT SPREADSHEET (LibreOffice/OpenOffice)
	// ============================================================
	// ODS uses content.xml with different XML structure
	case ".ods":
		return readODS(filePath)

	// ============================================================
	// OPENDOCUMENT PRESENTATION (LibreOffice/OpenOffice)
	// ============================================================
	// ODP uses content.xml with slides
	case ".odp":
		return readODP(filePath)

	default:
		// Unsupported file type
		return "", fmt.Errorf("unsupported office document format: %s", ext)
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

// ============================================================
// OPENDOCUMENT FORMAT READERS (LibreOffice/OpenOffice)
// ============================================================

// readODT extracts text from OpenDocument Text (.odt) file
//
// ODT FORMAT STRUCTURE:
//   - An .odt file is a ZIP archive (like DOCX)
//   - Inside: content.xml contains the main document text
//   - The XML has <text:p> tags for paragraphs
//   - And <text:span> tags for text runs
//
// WHAT WE EXTRACT:
//
//	✅ All paragraphs
//	✅ All text content
//	✅ Table content
//
// Parameters:
//   - filePath: Full path to the .odt file
//
// Returns:
//   - string: All text content from the document
//   - error: Error if file can't be opened or parsed
//
// Example:
//
//	text, err := readODT("/documents/report.odt")
func readODT(filePath string) (string, error) {
	// Open ODT file as ZIP archive
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open ODT as ZIP: %w", err)
	}
	defer zipReader.Close()

	// Find and read content.xml
	var contentXML string

	for _, file := range zipReader.File {
		// Look for the main content XML file
		// In OpenDocument, it's simply "content.xml" at the root
		if file.Name == "content.xml" {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open content.xml: %w", err)
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read content.xml: %w", err)
			}

			contentXML = string(xmlContent)
			break
		}
	}

	// Check if we found the content XML
	if contentXML == "" {
		return "", fmt.Errorf("content.xml not found in ODT file")
	}

	// Extract text from OpenDocument XML
	text := extractTextFromOpenDocumentXML(contentXML)

	return text, nil
}

// readODS extracts text from OpenDocument Spreadsheet (.ods) file
//
// ODS FORMAT STRUCTURE:
//   - An .ods file is a ZIP archive (like XLSX)
//   - Inside: content.xml contains all sheets and cells
//   - The XML has <table:table-cell> tags for cells
//
// WHAT WE EXTRACT:
//
//	✅ Text from all cells
//	✅ Numbers from all cells
//	✅ Content from all sheets
//
// Parameters:
//   - filePath: Full path to the .ods file
//
// Returns:
//   - string: All text content from all sheets
//   - error: Error if file can't be opened or parsed
//
// Example:
//
//	text, err := readODS("/reports/data.ods")
func readODS(filePath string) (string, error) {
	// Open ODS file as ZIP archive
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open ODS as ZIP: %w", err)
	}
	defer zipReader.Close()

	// Find and read content.xml
	var contentXML string

	for _, file := range zipReader.File {
		if file.Name == "content.xml" {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open content.xml: %w", err)
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read content.xml: %w", err)
			}

			contentXML = string(xmlContent)
			break
		}
	}

	if contentXML == "" {
		return "", fmt.Errorf("content.xml not found in ODS file")
	}

	// Extract text from OpenDocument XML
	text := extractTextFromOpenDocumentXML(contentXML)

	return text, nil
}

// readODP extracts text from OpenDocument Presentation (.odp) file
//
// ODP FORMAT STRUCTURE:
//   - An .odp file is a ZIP archive (like PPTX)
//   - Inside: content.xml contains all slides
//   - The XML has <draw:page> tags for slides
//
// WHAT WE EXTRACT:
//
//	✅ Text from all slides
//	✅ Text from text boxes
//	✅ Slide titles
//
// Parameters:
//   - filePath: Full path to the .odp file
//
// Returns:
//   - string: All text content from all slides
//   - error: Error if file can't be opened or parsed
//
// Example:
//
//	text, err := readODP("/presentations/slides.odp")
func readODP(filePath string) (string, error) {
	// Open ODP file as ZIP archive
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open ODP as ZIP: %w", err)
	}
	defer zipReader.Close()

	// Find and read content.xml
	var contentXML string

	for _, file := range zipReader.File {
		if file.Name == "content.xml" {
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open content.xml: %w", err)
			}
			defer rc.Close()

			xmlContent, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read content.xml: %w", err)
			}

			contentXML = string(xmlContent)
			break
		}
	}

	if contentXML == "" {
		return "", fmt.Errorf("content.xml not found in ODP file")
	}

	// Extract text from OpenDocument XML
	text := extractTextFromOpenDocumentXML(contentXML)

	return text, nil
}

// extractTextFromOpenDocumentXML extracts text from OpenDocument XML content
//
// OpenDocument XML structure (simplified):
//
//	<office:document-content>
//	  <office:body>
//	    <office:text>           <!-- For ODT -->
//	      <text:p>Hello</text:p>  <!-- Paragraph -->
//	      <text:p>World</text:p>
//	    </office:text>
//	    <office:spreadsheet>    <!-- For ODS -->
//	      <table:table-cell>
//	        <text:p>Data</text:p>
//	      </table:table-cell>
//	    </office:spreadsheet>
//	    <office:presentation>   <!-- For ODP -->
//	      <draw:page>
//	        <text:p>Slide text</text:p>
//	      </draw:page>
//	    </office:presentation>
//	  </office:body>
//	</office:document-content>
//
// We extract all text from <text:p> tags (paragraphs)
// and from actual text content in various tags.
//
// Parameters:
//   - xmlContent: The XML content from content.xml
//
// Returns:
//   - string: Extracted plain text
func extractTextFromOpenDocumentXML(xmlContent string) string {
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
			// OpenDocument uses various text tags
			// text:p = paragraph
			// text:h = heading
			// text:span = text span
			// We look for any tag with "text" in the namespace
			if strings.Contains(t.Name.Space, "text") || t.Name.Local == "p" || t.Name.Local == "h" || t.Name.Local == "span" {
				inTextTag = true
			}

		case xml.CharData:
			// If we're in a text tag, this is text content
			if inTextTag {
				text := strings.TrimSpace(string(t))
				if text != "" {
					result.WriteString(text)
					result.WriteString(" ")
				}
			}

		case xml.EndElement:
			// End of text tag
			if strings.Contains(t.Name.Space, "text") || t.Name.Local == "p" || t.Name.Local == "h" {
				inTextTag = false
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}
