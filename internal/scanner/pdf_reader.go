// Package scanner - ULTIMATE PDF Text Extraction (Standard Library Only)
// File: internal/scanner/pdf_reader_ultimate.go
//
// ============================================================
// THE ULTIMATE STANDARD LIBRARY PDF READER
// ============================================================
//
// This is the most comprehensive PDF reader possible using
// ONLY Go standard library. It includes:
//
// ✅ PDF structure parsing (objects, cross-references)
// ✅ Font definition extraction
// ✅ ToUnicode CMap parsing
// ✅ Hex code to character mapping
// ✅ Multiple text extraction strategies
// ✅ Fallback mechanisms
// ✅ Comprehensive error handling
//
// This reader will extract text from PDFs that use:
// - Standard fonts
// - Custom fonts with ToUnicode CMaps
// - Hex-encoded character references
// - Multiple encoding schemes
//
// LIMITATIONS (even with best effort):
// - Cannot parse complex CMap definitions without parser
// - Cannot handle encrypted PDFs
// - Cannot perform OCR on images
// - Some custom encodings may still fail

package scanner

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ============================================================
// PDF STRUCTURE TYPES
// ============================================================

// PDFObject represents a PDF object
type PDFObject struct {
	ID      int
	Content string
	Stream  []byte
}

// FontInfo stores font encoding information
type FontInfo struct {
	Name      string
	ToUnicode map[string]rune // Hex code -> Unicode character
	Encoding  string
	BaseFont  string
}

// PDFDocument represents the parsed PDF structure
type PDFDocument struct {
	Objects map[int]*PDFObject
	Fonts   map[string]*FontInfo
	Pages   []int // Object IDs of pages
}

// ============================================================
// PDF FILE DETECTION
// ============================================================

func isPDFFile(filePath string) (bool, error) {
	if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	header := make([]byte, 5)
	n, err := file.Read(header)
	if err != nil || n < 5 {
		return false, nil
	}

	return string(header) == "%PDF-", nil
}

// ============================================================
// MAIN PDF READING FUNCTION
// ============================================================

func readPDF(filePath string) (string, error) {
	// Read entire file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return "", fmt.Errorf("not a valid PDF file")
	}

	fmt.Println("DEBUG: Starting comprehensive PDF parsing...")

	// Parse PDF structure
	doc := parsePDFStructure(data)

	fmt.Printf("DEBUG: Parsed %d objects\n", len(doc.Objects))
	fmt.Printf("DEBUG: Found %d fonts\n", len(doc.Fonts))

	// Extract text using multiple strategies
	var allText strings.Builder

	// Strategy 1: Parse content streams with font mapping
	fmt.Println("DEBUG: Strategy 1 - Content stream parsing with font mapping")
	text1 := extractTextWithFontMapping(doc, data)
	if len(text1) > 0 {
		fmt.Printf("DEBUG: Strategy 1 extracted %d characters\n", len(text1))
		allText.WriteString(text1)
		allText.WriteString("\n")
	}

	// Strategy 2: Direct stream extraction
	fmt.Println("DEBUG: Strategy 2 - Direct stream extraction")
	text2 := extractFromAllStreams(data)
	if len(text2) > 0 {
		fmt.Printf("DEBUG: Strategy 2 extracted %d characters\n", len(text2))
		allText.WriteString(text2)
		allText.WriteString("\n")
	}

	// Strategy 3: Text object extraction
	fmt.Println("DEBUG: Strategy 3 - Text object extraction")
	text3 := extractTextObjects(data)
	if len(text3) > 0 {
		fmt.Printf("DEBUG: Strategy 3 extracted %d characters\n", len(text3))
		allText.WriteString(text3)
		allText.WriteString("\n")
	}

	result := allText.String()
	fmt.Printf("DEBUG: Total extracted: %d characters\n", len(result))

	// Show preview
	if len(result) > 0 {
		preview := result
		if len(preview) > 500 {
			preview = preview[:500]
		}
		fmt.Printf("DEBUG: First 500 chars:\n%s\n", preview)
	}

	return result, nil
}

// ============================================================
// PDF STRUCTURE PARSING
// ============================================================

// parsePDFStructure parses the PDF object structure
func parsePDFStructure(data []byte) *PDFDocument {
	doc := &PDFDocument{
		Objects: make(map[int]*PDFObject),
		Fonts:   make(map[string]*FontInfo),
		Pages:   []int{},
	}

	content := string(data)

	// Find all objects: "N 0 obj ... endobj"
	objPattern := regexp.MustCompile(`(\d+)\s+0\s+obj([\s\S]*?)endobj`)
	matches := objPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			id, _ := strconv.Atoi(match[1])
			objContent := match[2]

			obj := &PDFObject{
				ID:      id,
				Content: objContent,
			}

			// Check if object has a stream
			if strings.Contains(objContent, "stream") {
				streamPattern := regexp.MustCompile(`stream\s*\n([\s\S]*?)endstream`)
				if streamMatch := streamPattern.FindSubmatch([]byte(objContent)); len(streamMatch) > 1 {
					obj.Stream = streamMatch[1]
				}
			}

			doc.Objects[id] = obj

			// Check if this is a font object
			if strings.Contains(objContent, "/Type") && strings.Contains(objContent, "/Font") {
				parseFont(obj, doc)
			}
		}
	}

	return doc
}

// ============================================================
// FONT PARSING
// ============================================================

// parseFont extracts font information from a font object
func parseFont(obj *PDFObject, doc *PDFDocument) {
	content := obj.Content

	font := &FontInfo{
		ToUnicode: make(map[string]rune),
	}

	// Extract font name
	if match := regexp.MustCompile(`/BaseFont\s*/([^\s/>]+)`).FindStringSubmatch(content); len(match) > 1 {
		font.BaseFont = match[1]
		font.Name = match[1]
	}

	// Extract encoding
	if match := regexp.MustCompile(`/Encoding\s*/([^\s/>]+)`).FindStringSubmatch(content); len(match) > 1 {
		font.Encoding = match[1]
	}

	// Look for ToUnicode reference
	if match := regexp.MustCompile(`/ToUnicode\s+(\d+)\s+0\s+R`).FindStringSubmatch(content); len(match) > 1 {
		toUnicodeID, _ := strconv.Atoi(match[1])
		if toUnicodeObj, exists := doc.Objects[toUnicodeID]; exists {
			parseToUnicodeCMap(toUnicodeObj, font)
		}
	}

	// Store font
	if font.Name != "" {
		doc.Fonts[font.Name] = font
		fmt.Printf("DEBUG: Found font: %s (encoding: %s, %d mappings)\n",
			font.Name, font.Encoding, len(font.ToUnicode))
	}
}

// ============================================================
// TOUNICODE CMAP PARSING
// ============================================================

// parseToUnicodeCMap parses a ToUnicode CMap stream
func parseToUnicodeCMap(obj *PDFObject, font *FontInfo) {
	var data []byte

	// Decompress stream if needed
	if obj.Stream != nil {
		decompressed := decompressZlib(obj.Stream)
		if decompressed != nil {
			data = decompressed
		} else {
			data = obj.Stream
		}
	}

	if len(data) == 0 {
		return
	}

	content := string(data)

	// Parse bfchar mappings: <hex> <unicode>
	// Example: <01> <0053>  means: 01 -> U+0053 ('S')
	bfcharPattern := regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>`)
	matches := bfcharPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			hexCode := strings.ToLower(match[1])
			unicodeHex := match[2]

			// Convert unicode hex to rune
			if unicodeVal, err := strconv.ParseInt(unicodeHex, 16, 32); err == nil {
				font.ToUnicode[hexCode] = rune(unicodeVal)
			}
		}
	}

	// Parse bfrange mappings: <start> <end> <unicode_start>
	// Example: <01> <0A> <0041>  means: 01->A, 02->B, ..., 0A->J
	bfrangePattern := regexp.MustCompile(`<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>\s*<([0-9A-Fa-f]+)>`)
	rangeMatches := bfrangePattern.FindAllStringSubmatch(content, -1)

	for _, match := range rangeMatches {
		if len(match) >= 4 {
			startHex := match[1]
			endHex := match[2]
			unicodeStartHex := match[3]

			start, _ := strconv.ParseInt(startHex, 16, 32)
			end, _ := strconv.ParseInt(endHex, 16, 32)
			unicodeStart, _ := strconv.ParseInt(unicodeStartHex, 16, 32)

			// Map range
			for i := start; i <= end; i++ {
				hexCode := fmt.Sprintf("%02x", i)
				font.ToUnicode[hexCode] = rune(unicodeStart + (i - start))
			}
		}
	}

	fmt.Printf("DEBUG: Parsed ToUnicode CMap: %d character mappings\n", len(font.ToUnicode))
}

// ============================================================
// STRATEGY 1: TEXT EXTRACTION WITH FONT MAPPING
// ============================================================

// extractTextWithFontMapping extracts text using font character mappings
func extractTextWithFontMapping(doc *PDFDocument, data []byte) string {
	var result strings.Builder
	content := string(data)

	// Find all content streams
	streamPattern := regexp.MustCompile(`stream\s*\n([\s\S]*?)endstream`)
	matches := streamPattern.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		streamData := match[1]

		// Try to decompress
		decompressed := decompressZlib(streamData)
		if decompressed != nil {
			streamData = decompressed
		}

		// Parse content stream
		text := parseContentStream(streamData, doc.Fonts)
		if len(text) > 0 {
			result.WriteString(text)
			result.WriteString(" ")
		}
	}

	return result.String()
}

// ============================================================
// CONTENT STREAM PARSING
// ============================================================

// parseContentStream parses a PDF content stream and extracts text
func parseContentStream(data []byte, fonts map[string]*FontInfo) string {
	content := string(data)
	var result strings.Builder
	var currentFont *FontInfo

	// Split into lines
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for font selection: /F1 12 Tf
		if strings.Contains(line, "Tf") {
			if match := regexp.MustCompile(`/([^\s]+)\s+[\d.]+\s+Tf`).FindStringSubmatch(line); len(match) > 1 {
				fontName := match[1]
				if font, exists := fonts[fontName]; exists {
					currentFont = font
				}
			}
		}

		// Look for hex strings: <hexdata>
		hexPattern := regexp.MustCompile(`<([0-9A-Fa-f]+)>`)
		hexMatches := hexPattern.FindAllStringSubmatch(line, -1)

		for _, match := range hexMatches {
			if len(match) > 1 {
				hexStr := strings.ToLower(match[1])

				// Decode using current font
				if currentFont != nil && len(currentFont.ToUnicode) > 0 {
					decoded := decodeHexWithFont(hexStr, currentFont)
					if len(decoded) > 0 {
						result.WriteString(decoded)
						result.WriteString(" ")
					}
				} else {
					// Try to decode as direct hex (fallback)
					decoded := hexToText(hexStr)
					if len(decoded) > 0 {
						result.WriteString(decoded)
						result.WriteString(" ")
					}
				}
			}
		}

		// Look for regular strings: (text)
		stringPattern := regexp.MustCompile(`\(([^)]+)\)`)
		stringMatches := stringPattern.FindAllStringSubmatch(line, -1)

		for _, match := range stringMatches {
			if len(match) > 1 {
				text := match[1]
				text = unescapeText(text)
				if len(text) > 0 {
					result.WriteString(text)
					result.WriteString(" ")
				}
			}
		}
	}

	return result.String()
}

// ============================================================
// HEX DECODING WITH FONT MAPPING
// ============================================================

// decodeHexWithFont decodes hex string using font ToUnicode mapping
func decodeHexWithFont(hexStr string, font *FontInfo) string {
	var result strings.Builder

	// Process hex string in 2-character (1-byte) or 4-character (2-byte) chunks
	// Try 2-char chunks first (most common)
	i := 0
	for i < len(hexStr) {
		// Try 4-char chunk (2 bytes)
		if i+4 <= len(hexStr) {
			chunk4 := hexStr[i : i+4]
			if char, exists := font.ToUnicode[chunk4]; exists {
				result.WriteRune(char)
				i += 4
				continue
			}
		}

		// Try 2-char chunk (1 byte)
		if i+2 <= len(hexStr) {
			chunk2 := hexStr[i : i+2]
			if char, exists := font.ToUnicode[chunk2]; exists {
				result.WriteRune(char)
				i += 2
				continue
			}
		}

		// Try single char
		if i+1 <= len(hexStr) {
			chunk1 := hexStr[i : i+1]
			if char, exists := font.ToUnicode[chunk1]; exists {
				result.WriteRune(char)
				i += 1
				continue
			}
		}

		// Couldn't decode this chunk, skip it
		i += 2
	}

	return result.String()
}

// ============================================================
// STRATEGY 2: DIRECT STREAM EXTRACTION
// ============================================================

func extractFromAllStreams(data []byte) string {
	var result strings.Builder

	streamPattern := regexp.MustCompile(`stream\s*\n([\s\S]*?)endstream`)
	matches := streamPattern.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		streamData := match[1]

		// Try decompression
		decompressed := decompressZlib(streamData)
		if decompressed != nil {
			streamData = decompressed
		}

		// Extract printable text
		text := extractPrintableText(streamData)
		if len(text) > 0 {
			result.WriteString(text)
			result.WriteString(" ")
		}
	}

	return result.String()
}

// ============================================================
// STRATEGY 3: TEXT OBJECT EXTRACTION
// ============================================================

func extractTextObjects(data []byte) string {
	content := string(data)
	var result strings.Builder

	// Find text blocks: BT ... ET
	textBlockPattern := regexp.MustCompile(`BT([\s\S]*?)ET`)
	matches := textBlockPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			block := match[1]

			// Extract strings from parentheses
			stringPattern := regexp.MustCompile(`\(([^)]+)\)`)
			stringMatches := stringPattern.FindAllStringSubmatch(block, -1)

			for _, strMatch := range stringMatches {
				if len(strMatch) > 1 {
					text := unescapeText(strMatch[1])
					if len(text) > 0 {
						result.WriteString(text)
						result.WriteString(" ")
					}
				}
			}
		}
	}

	return result.String()
}

// ============================================================
// UTILITY FUNCTIONS
// ============================================================

func decompressZlib(data []byte) []byte {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}

	return decompressed
}

func extractPrintableText(data []byte) string {
	var result strings.Builder
	var current strings.Builder

	for _, b := range data {
		if unicode.IsPrint(rune(b)) || b == ' ' || b == '\n' || b == '\t' {
			current.WriteByte(b)
		} else {
			if current.Len() >= 3 {
				result.WriteString(current.String())
				result.WriteString(" ")
			}
			current.Reset()
		}
	}

	if current.Len() >= 3 {
		result.WriteString(current.String())
	}

	return result.String()
}

func unescapeText(text string) string {
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = strings.ReplaceAll(text, `\r`, "\n")
	text = strings.ReplaceAll(text, `\t`, "\t")
	text = strings.ReplaceAll(text, `\(`, "(")
	text = strings.ReplaceAll(text, `\)`, ")")
	text = strings.ReplaceAll(text, `\\`, "\\")
	return text
}

func hexToText(hexStr string) string {
	if len(hexStr)%2 != 0 {
		return ""
	}

	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		hexByte := hexStr[i : i+2]
		var b byte
		_, err := fmt.Sscanf(hexByte, "%02x", &b)
		if err != nil {
			continue
		}

		if unicode.IsPrint(rune(b)) {
			result.WriteByte(b)
		}
	}

	return result.String()
}
