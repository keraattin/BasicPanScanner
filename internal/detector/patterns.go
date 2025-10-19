// Package detector - Enhanced Card pattern definitions
// This file contains improved regex patterns for detecting different card types
// with better precision and reduced false positives
package detector

import (
	"fmt"
	"regexp"
)

// CardPattern represents a single card issuer's detection pattern
// Each pattern combines:
//   - Name: Card issuer (e.g., "Visa", "Mastercard")
//   - Pattern: Regex that matches the card format
type CardPattern struct {
	Name    string         // Card issuer name
	Pattern *regexp.Regexp // Compiled regex pattern
}

// Global variable holding all card patterns
// This is initialized once at startup for performance
var cardPatterns []CardPattern

// InitPatterns compiles and initializes all card detection patterns
// This function MUST be called before using any detection functions
// It should be called once during application startup
//
// Important Note About Go's RE2 Engine:
// Go uses the RE2 regex engine which does NOT support backreferences (\1, \2, etc.)
// This is a deliberate design choice for performance and security.
//
// Impact on Detection:
// - We cannot enforce consistent separators (e.g., all spaces OR all dashes)
// - This means "4532 0151-1283 0366" (mixed separators) will match
// - However, this is ACCEPTABLE because:
//  1. The Luhn algorithm still validates the card number
//  2. Mixed separators are rare in real-world data
//  3. Word boundaries (\b) prevent most false positives
//  4. The performance gain from RE2 is significant
//
// Returns:
//   - error: Error if any regex pattern fails to compile
//
// Example:
//
//	func main() {
//	    if err := detector.InitPatterns(); err != nil {
//	        log.Fatal(err)
//	    }
//	     Now ready to detect cards
//	}
func InitPatterns() error {
	// Define all patterns with enhanced precision
	// Each regex is designed to:
	// 1. Match valid card number formats (with or without separators)
	// 2. Use word boundaries to avoid matching within larger numbers
	// 3. Handle cards embedded in text
	// 4. Prevent matching invalid sequences
	//
	// NOTE: NO backreferences (\1, \2) are used because Go's RE2 doesn't support them

	patterns := []struct {
		name    string
		pattern string
	}{
		// ============================================================
		// VISA
		// - Starts with: 4
		// - Length: 13, 16, or 19 digits
		// - Format: 4xxx xxxx xxxx xxxx (standard 16-digit)
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 4           : Starts with 4 (Visa prefix)
		// - \d{3}       : 3 more digits (makes 4xxx)
		// - [\s\-]*     : Zero or more separators (handles multiple spaces)
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits (16 digits total so far)
		// - (?:[\s\-]*\d{3}){0,2} : Optional extra digits for 19-digit cards
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Visa",
			`\b4\d{3}[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}(?:[\s\-]*\d{3}){0,2}\b`,
		},

		// ============================================================
		// MASTERCARD
		// - Starts with: 51-55 (old range) OR 2221-2720 (new range)
		// - Length: 16 digits only
		// - Format: 5xxx xxxx xxxx xxxx or 22xx xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - (?:5[1-5]|2(?:22[1-9]|2[3-9]\d|[3-6]\d{2}|7[01]\d|720))
		//   This matches the valid prefixes
		// - \d{2}       : 2 more digits (completes first 4)
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"MasterCard",
			`\b(?:5[1-5]|2(?:22[1-9]|2[3-9]\d|[3-6]\d{2}|7[01]\d|720))\d{2}[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// AMERICAN EXPRESS (Amex)
		// - Starts with: 34 or 37
		// - Length: 15 digits only
		// - Format: 3xxx xxxxxx xxxxx (4-6-5 grouping)
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 3[47]       : Starts with 34 or 37
		// - \d{2}       : 2 more digits (makes 3xxx)
		// - [\s\-]*     : Zero or more separators
		// - \d{6}       : 6 digits (Amex uses 4-6-5 grouping)
		// - [\s\-]*     : Zero or more separators
		// - \d{5}       : 5 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Amex",
			`\b3[47]\d{2}[\s\-]*\d{6}[\s\-]*\d{5}\b`,
		},

		// ============================================================
		// DISCOVER
		// - Starts with: 6011, 622126-622925, 644-649, or 65
		// - Length: 16 digits only
		// - Format: 6xxx xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - (?:6011|65\d{2}|64[4-9]\d|622(?:1[2-9]\d|[2-8]\d{2}|9[01]\d|92[0-5]))
		//   This matches the valid prefixes
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Discover",
			`\b(?:6011|65\d{2}|64[4-9]\d|622(?:1[2-9]\d|[2-8]\d{2}|9[01]\d|92[0-5]))[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// DINERS CLUB
		// - Starts with: 36, 38, or 300-305
		// - Length: 14 digits only
		// - Format: 3xxx xxxxxx xxxx (4-6-4 grouping)
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 3(?:0[0-5]|[68]\d)
		//   This matches the valid prefixes
		// - \d          : 1 more digit (makes 3xxx)
		// - [\s\-]*     : Zero or more separators
		// - \d{6}       : 6 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Diners",
			`\b3(?:0[0-5]|[68]\d)\d[\s\-]*\d{6}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// JCB
		// - Starts with: 3528-3589
		// - Length: 16 digits only
		// - Format: 35xx xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 35(?:2[89]|[3-8]\d)
		//   This matches the valid prefixes
		// - \d          : 1 more digit (makes 35xx)
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"JCB",
			`\b35(?:2[89]|[3-8]\d)\d[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// CHINA UNIONPAY
		// - Starts with: 62
		// - Length: 16-19 digits
		// - Format: 62xx xxxx xxxx xxxx (and variants)
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 62\d{2}     : Starts with 62 + 2 more digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits (16 digits total so far)
		// - (?:[\s\-]*\d{3}){0,2} : Optional extra digits for 19-digit cards
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"UnionPay",
			`\b62\d{2}[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}(?:[\s\-]*\d{3}){0,2}\b`,
		},

		// ============================================================
		// MAESTRO
		// - Starts with: 50, 56-69
		// - Length: 12-19 digits (very flexible)
		// - Format: Variable (most flexible card type)
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - (?:5[06789]|6\d)
		//   This matches the valid prefixes
		// - \d{2}       : 2 more digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{1,7}     : 1-7 digits (flexible length)
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Maestro",
			`\b(?:5[06789]|6\d)\d{2}[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{1,7}\b`,
		},

		// ============================================================
		// RUPAY (India)
		// - Starts with: 60, 6521, or 6522
		// - Length: 16 digits
		// - Format: 60xx xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - (?:60|652[12])
		//   This matches the valid prefixes
		// - \d{2}       : 2 more digits (unless prefix is already 4 digits)
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"RuPay",
			`\b(?:60|652[12])\d{2}[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// TROY (Turkey)
		// - Starts with: 9792
		// - Length: 16 digits
		// - Format: 9792 xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 9792        : Exact prefix
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Troy",
			`\b9792[\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},

		// ============================================================
		// MIR (Russia)
		// - Starts with: 2200-2204
		// - Length: 16 digits
		// - Format: 220x xxxx xxxx xxxx
		//
		// Pattern explanation:
		// - \b          : Word boundary (start)
		// - 220[0-4]    : Starts with 2200-2204
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - [\s\-]*     : Zero or more separators
		// - \d{4}       : 4 digits
		// - \b          : Word boundary (end)
		// ============================================================
		{
			"Mir",
			`\b220[0-4][\s\-]*\d{4}[\s\-]*\d{4}[\s\-]*\d{4}\b`,
		},
	}

	// Compile all patterns
	// Pre-compiling patterns once at startup is much faster than
	// compiling them every time we need to detect cards
	cardPatterns = make([]CardPattern, 0, len(patterns))

	for _, p := range patterns {
		// Compile the regex pattern
		// This validates the pattern and creates an optimized matcher
		regex, err := regexp.Compile(p.pattern)
		if err != nil {
			// If compilation fails, return error with pattern name
			// This helps developers identify which pattern has a syntax error
			return fmt.Errorf("failed to compile pattern for %s: %w", p.name, err)
		}

		// Add compiled pattern to list
		cardPatterns = append(cardPatterns, CardPattern{
			Name:    p.name,
			Pattern: regex,
		})
	}

	// Success! All patterns compiled without errors
	fmt.Printf("✓ Loaded %d card issuer patterns\n", len(cardPatterns))
	return nil
}

// GetPatterns returns all card patterns
// This is useful for testing or iterating through patterns
//
// Returns:
//   - []CardPattern: List of all card patterns
//
// Example:
//
//	for _, pattern := range detector.GetPatterns() {
//	    fmt.Printf("Pattern: %s\n", pattern.Name)
//	}
func GetPatterns() []CardPattern {
	return cardPatterns
}

// GetPatternCount returns the number of loaded patterns
// Useful for verification and logging
//
// Returns:
//   - int: Number of patterns loaded
//
// Example:
//
//	count := detector.GetPatternCount()
//	fmt.Printf("Loaded %d patterns\n", count)
func GetPatternCount() int {
	return len(cardPatterns)
}
