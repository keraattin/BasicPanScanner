// Package scanner - Basic PDF Text Extraction (Standard Library Only)
// File: internal/scanner/pdf_reader.go
//
// ⚠️ CRITICAL WARNINGS - READ THIS FIRST! ⚠️
//
// This implementation uses ONLY Go standard library for PDF text extraction.
// This approach has SEVERE LIMITATIONS:
//
// WHAT IT CAN DO:
//
//	✅ Extract readable text from simple, uncompressed PDFs
//	✅ Find credit card numbers in extracted text
//	✅ Work without external dependencies
//
// WHAT IT CANNOT DO (Important!):
//
//	❌ Parse compressed PDFs (most modern PDFs are compressed)
//	❌ Handle encoded text streams
//	❌ Extract text in correct order
//	❌ Handle scanned/image PDFs
//	❌ Process encrypted PDFs
//	❌ Handle complex fonts
//	❌ Parse PDF structure properly
//
// SUCCESS RATE: ~10-20% of real-world PDFs
//
// WHY SO LOW?
//   - 80-90% of PDFs use compression (FlateDecode, etc.)
//   - Modern PDF creators always compress content
//   - We can only read RAW uncompressed text
//
// RECOMMENDATION:
//
//	This is a "better than nothing" approach for learning.
//	For production use, consider external PDF library.
//
// THIS APPROACH IS PROVIDED FOR:
//  1. Educational purposes (understand PDF complexity)
//  2. Demonstration of limitations
//  3. Backup for simple PDFs
//
// USERS SHOULD BE WARNED when this method is used!
package scanner

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// ============================================================
// PDF DETECTION
// ============================================================

// isPDFFile checks if a file is a PDF based on extension and header
//
// This function performs two checks:
//  1. File extension check (.pdf)
//  2. File header check (%PDF-)
//
// Parameters:
//   - filePath: Full path to the file
//
// Returns:
//   - bool: true if file is a PDF file
//   - error: Error if file cannot be read
//
// Example:
//
//	if isPDF, _ := isPDFFile("document.pdf"); isPDF {
//	    // File is a PDF
//	}
func isPDFFile(filePath string) (bool, error) {
	// Check extension first (fast check)
	if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return false, nil
	}

	// Verify PDF header (more reliable)
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 5 bytes
	// All PDF files start with: %PDF-
	header := make([]byte, 5)
	n, err := file.Read(header)
	if err != nil || n < 5 {
		return false, nil
	}

	// Check if header matches PDF signature
	return string(header) == "%PDF-", nil
}

// ============================================================
// BASIC PDF TEXT EXTRACTION
// ============================================================

// readBasicPDF attempts to extract text from a PDF using only standard library
//
// ⚠️ LIMITATIONS WARNING:
//
//	This function can ONLY extract text from simple, uncompressed PDFs.
//	It will FAIL or return incomplete text for:
//	- Compressed PDFs (80-90% of modern PDFs)
//	- Scanned PDFs (images)
//	- Encrypted PDFs
//	- PDFs with complex encodings
//
// HOW IT WORKS:
//  1. Read entire PDF file into memory
//  2. Try to decompress Flate-encoded streams
//  3. Search for text operators (BT...ET blocks)
//  4. Extract text from () and <> strings
//  5. Clean and return extracted text
//
// This is a "best effort" approach and should not be relied upon
// for critical applications!
//
// Parameters:
//   - filePath: Full path to the PDF file
//
// Returns:
//   - string: Extracted text (may be incomplete or empty)
//   - error: Error if file cannot be read
//   - bool: Success flag (false if extraction was poor quality)
//
// Example:
//
//	text, err, success := readBasicPDF("simple.pdf")
//	if err != nil {
//	    log.Printf("Failed to read PDF: %v", err)
//	}
//	if !success {
//	    log.Printf("Warning: PDF text extraction was incomplete")
//	}
func readBasicPDF(filePath string) (string, error, bool) {
	// ============================================================
	// STEP 1: Read entire file
	// ============================================================

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF file: %w", err), false
	}

	// ============================================================
	// STEP 2: Check if it's a valid PDF
	// ============================================================

	if len(data) < 5 || string(data[:5]) != "%PDF-" {
		return "", fmt.Errorf("not a valid PDF file (missing header)"), false
	}

	// ============================================================
	// STEP 3: Try to decompress any Flate-encoded streams
	// ============================================================
	// Many PDFs use FlateDecode compression
	// We'll try to decompress these streams using zlib (standard library)

	decompressedData := tryDecompressFlateStreams(data)

	// ============================================================
	// STEP 4: Extract text using basic pattern matching
	// ============================================================

	// Approach 1: Look for text between BT (Begin Text) and ET (End Text)
	textFromOperators := extractTextFromOperators(decompressedData)

	// Approach 2: Search for readable strings in raw data
	textFromRaw := extractReadableStrings(decompressedData)

	// ============================================================
	// STEP 5: Combine and clean extracted text
	// ============================================================

	combinedText := textFromOperators + "\n" + textFromRaw
	cleanedText := cleanExtractedText(combinedText)

	// ============================================================
	// STEP 6: Check quality of extraction
	// ============================================================

	// If we got very little text, extraction probably failed
	// This is common with compressed/encrypted PDFs
	success := len(cleanedText) > 50 // At least 50 characters

	if !success {
		// Return what we got, but warn that it's incomplete
		return cleanedText, nil, false
	}

	return cleanedText, nil, true
}

// ============================================================
// HELPER: DECOMPRESS FLATE STREAMS
// ============================================================

// tryDecompressFlateStreams attempts to decompress FlateDecode streams
//
// PDF streams can be compressed using various methods.
// This function handles FlateDecode (zlib compression) which is
// the most common and supported by Go standard library.
//
// It searches for stream markers and tries to decompress data
// between "stream" and "endstream" keywords.
//
// Parameters:
//   - data: Raw PDF file data
//
// Returns:
//   - []byte: Data with decompressed streams (if any)
func tryDecompressFlateStreams(data []byte) []byte {
	// This is a simplified approach
	// Real PDF parsing would use cross-reference tables

	result := make([]byte, 0, len(data)*2)

	// Convert to string for easier searching
	content := string(data)

	// Find all stream objects
	// Pattern: "stream\n...binary data...endstream"
	streamRegex := regexp.MustCompile(`stream\r?\n([\s\S]*?)endstream`)
	matches := streamRegex.FindAllStringIndex(content, -1)

	lastEnd := 0

	for _, match := range matches {
		// Add content before stream
		result = append(result, data[lastEnd:match[0]]...)

		// Extract stream data (between "stream\n" and "endstream")
		streamStart := match[0] + 7 // Length of "stream\n"
		streamEnd := match[1] - 9   // Before "endstream"

		streamData := data[streamStart:streamEnd]

		// Try to decompress with zlib
		decompressed, err := tryZlibDecompress(streamData)
		if err == nil {
			// Successfully decompressed!
			result = append(result, decompressed...)
		} else {
			// Keep original data
			result = append(result, streamData...)
		}

		lastEnd = match[1]
	}

	// Add remaining data
	result = append(result, data[lastEnd:]...)

	return result
}

// tryZlibDecompress attempts to decompress data using zlib
//
// Parameters:
//   - data: Compressed data
//
// Returns:
//   - []byte: Decompressed data
//   - error: Error if decompression fails
func tryZlibDecompress(data []byte) ([]byte, error) {
	// Create zlib reader
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Read decompressed data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// ============================================================
// HELPER: EXTRACT TEXT FROM OPERATORS
// ============================================================

// extractTextFromOperators extracts text from PDF text operators
//
// PDF text is enclosed in BT (Begin Text) and ET (End Text) operators.
// Between these, text strings appear in parentheses () or angle brackets <>.
//
// Example PDF text syntax:
//
//	BT
//	/F1 12 Tf
//	100 700 Td
//	(Hello World) Tj
//	ET
//
// We extract: "Hello World"
//
// Parameters:
//   - data: PDF data (potentially decompressed)
//
// Returns:
//   - string: Extracted text
func extractTextFromOperators(data []byte) string {
	var result strings.Builder

	// Convert to string
	content := string(data)

	// Find all text blocks (BT...ET)
	// This is a simplified regex - real PDF parsing is more complex
	textBlockRegex := regexp.MustCompile(`BT[\s\S]*?ET`)
	textBlocks := textBlockRegex.FindAllString(content, -1)

	for _, block := range textBlocks {
		// Extract strings in parentheses: (text)
		// This captures text shown with Tj, TJ, and similar operators
		stringRegex := regexp.MustCompile(`\(((?:[^()]|\\\))*)\)`)
		matches := stringRegex.FindAllStringSubmatch(block, -1)

		for _, match := range matches {
			if len(match) > 1 {
				text := match[1]
				// Unescape escaped characters
				text = strings.ReplaceAll(text, `\)`, `)`)
				text = strings.ReplaceAll(text, `\(`, `(`)
				text = strings.ReplaceAll(text, `\\`, `\`)
				result.WriteString(text)
				result.WriteString(" ")
			}
		}

		// Also extract hex strings: <text>
		// These are used for special characters
		hexRegex := regexp.MustCompile(`<([0-9A-Fa-f]+)>`)
		hexMatches := hexRegex.FindAllStringSubmatch(block, -1)

		for _, match := range hexMatches {
			if len(match) > 1 {
				// Convert hex to ASCII (simplified)
				hexStr := match[1]
				result.WriteString(hexToText(hexStr))
				result.WriteString(" ")
			}
		}

		result.WriteString("\n")
	}

	return result.String()
}

// ============================================================
// HELPER: EXTRACT READABLE STRINGS
// ============================================================

// extractReadableStrings searches for readable ASCII strings in raw data
//
// This is a fallback method that looks for sequences of printable
// characters in the raw PDF data. It's very crude but can find
// text that other methods miss.
//
// Parameters:
//   - data: Raw PDF data
//
// Returns:
//   - string: Extracted readable strings
func extractReadableStrings(data []byte) string {
	var result strings.Builder

	// Look for sequences of printable ASCII characters
	// Minimum length: 4 characters (to avoid noise)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanBytes)

	currentString := make([]byte, 0, 100)

	for scanner.Scan() {
		b := scanner.Bytes()[0]

		// Check if character is printable ASCII
		if b >= 32 && b <= 126 {
			currentString = append(currentString, b)
		} else {
			// End of string
			if len(currentString) >= 4 {
				result.Write(currentString)
				result.WriteString(" ")
			}
			currentString = currentString[:0]
		}
	}

	// Add last string if any
	if len(currentString) >= 4 {
		result.Write(currentString)
	}

	return result.String()
}

// ============================================================
// HELPER: CLEAN EXTRACTED TEXT
// ============================================================

// cleanExtractedText cleans and normalizes extracted text
//
// This function:
//   - Removes duplicate whitespace
//   - Removes control characters
//   - Normalizes line breaks
//   - Removes PDF artifacts
//
// Parameters:
//   - text: Raw extracted text
//
// Returns:
//   - string: Cleaned text
func cleanExtractedText(text string) string {
	// Remove control characters except newlines and tabs
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == '\r' {
			return r
		}
		if r < 32 || r == 127 {
			return -1 // Remove character
		}
		return r
	}, text)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(` +`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")

	// Replace multiple newlines with double newline
	newlineRegex := regexp.MustCompile(`\n\n+`)
	cleaned = newlineRegex.ReplaceAllString(cleaned, "\n\n")

	// Remove PDF artifacts (common patterns)
	cleaned = strings.ReplaceAll(cleaned, "/Type", "")
	cleaned = strings.ReplaceAll(cleaned, "/Length", "")
	cleaned = strings.ReplaceAll(cleaned, "endobj", "")

	return strings.TrimSpace(cleaned)
}

// ============================================================
// HELPER: HEX TO TEXT CONVERSION
// ============================================================

// hexToText converts hex string to ASCII text
//
// PDF sometimes encodes text in hexadecimal format.
// Example: <48656C6C6F> = "Hello"
//
// Parameters:
//   - hexStr: Hexadecimal string (without < >)
//
// Returns:
//   - string: Converted text
func hexToText(hexStr string) string {
	// Make sure length is even
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	result := make([]byte, 0, len(hexStr)/2)

	for i := 0; i < len(hexStr); i += 2 {
		hexByte := hexStr[i : i+2]

		// Convert hex to byte
		var b byte
		fmt.Sscanf(hexByte, "%02x", &b)

		// Only add printable characters
		if b >= 32 && b <= 126 {
			result = append(result, b)
		}
	}

	return string(result)
}
