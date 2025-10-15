// Package detector - Card pattern definitions
// This file contains regex patterns for detecting different card types
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
// Returns:
//   - error: Error if any regex pattern fails to compile
//
// Example:
//
//	func main() {
//	    if err := detector.InitPatterns(); err != nil {
//	        log.Fatal(err)
//	    }
//	    // Now ready to detect cards
//	}
func InitPatterns() error {
	// Define all patterns
	// Each regex handles:
	//   - Card number format (with or without spaces/dashes)
	//   - Correct length for that card type
	//   - Issuer-specific prefixes

	patterns := []struct {
		name    string
		pattern string
	}{
		// ============================================================
		// VISA
		// - Starts with: 4
		// - Length: 13, 16, or 19 digits
		// - Format: 4xxx xxxx xxxx xxxx (standard 16-digit)
		//           4xxx xxxx xxxx x    (13-digit - older cards)
		//           4xxx xxxx xxxx xxxx xxx (19-digit - newer)
		// ============================================================
		{
			"Visa",
			`\b4\d{3}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}(?:[\s\-]?\d{3})?(?:[\s\-]?\d{3})?\b`,
		},

		// ============================================================
		// MASTERCARD
		// - Starts with: 51-55 OR 2221-2720 (new range)
		// - Length: 16 digits only
		// - Format: 5xxx xxxx xxxx xxxx or 22xx xxxx xxxx xxxx
		// ============================================================
		{
			"MasterCard",
			`\b(?:5[1-5]|222[1-9]|22[3-9]\d|2[3-6]\d{2}|27[01]\d|2720)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// AMERICAN EXPRESS (Amex)
		// - Starts with: 34 or 37
		// - Length: 15 digits only
		// - Format: 3xxx xxxxxx xxxxx (different grouping than others)
		// ============================================================
		{
			"Amex",
			`\b3[47]\d{2}[\s\-]?\d{6}[\s\-]?\d{5}\b`,
		},

		// ============================================================
		// DISCOVER
		// - Starts with: 6011, 622126-622925, 644-649, or 65
		// - Length: 16 digits only
		// - Format: 6xxx xxxx xxxx xxxx
		// ============================================================
		{
			"Discover",
			`\b(?:6011|65\d{2}|64[4-9]\d|622(?:1[2-9]\d|[2-8]\d{2}|9[01]\d|92[0-5]))\d{0,2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// DINERS CLUB
		// - Starts with: 36, 38, or 300-305
		// - Length: 14 digits only
		// - Format: 3xxx xxxxxx xxxx
		// ============================================================
		{
			"Diners",
			`\b3(?:0[0-5]|[68]\d)\d{1}[\s\-]?\d{6}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// JCB
		// - Starts with: 3528-3589
		// - Length: 16 digits only
		// - Format: 35xx xxxx xxxx xxxx
		// ============================================================
		{
			"JCB",
			`\b35(?:2[89]|[3-8]\d)\d{1}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// CHINA UNIONPAY
		// - Starts with: 62
		// - Length: 16-19 digits
		// - Format: 62xx xxxx xxxx xxxx (and variants)
		// ============================================================
		{
			"UnionPay",
			`\b62\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}(?:[\s\-]?\d{3})?(?:[\s\-]?\d{3})?\b`,
		},

		// ============================================================
		// MAESTRO
		// - Starts with: 50, 56-69
		// - Length: 12-19 digits (very flexible)
		// - Format: Variable (most flexible card type)
		// ============================================================
		{
			"Maestro",
			`\b(?:5[06789]|6\d)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{1,7}\b`,
		},

		// ============================================================
		// RUPAY (India)
		// - Starts with: 60, 6521, or 6522
		// - Length: 16 digits
		// - Format: 60xx xxxx xxxx xxxx
		// ============================================================
		{
			"RuPay",
			`\b(?:60|6521|6522)\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// TROY (Turkey)
		// - Starts with: 9792
		// - Length: 16 digits
		// - Format: 9792 xxxx xxxx xxxx
		// ============================================================
		{
			"Troy",
			`\b9792[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},

		// ============================================================
		// MIR (Russia)
		// - Starts with: 2200-2204
		// - Length: 16 digits
		// - Format: 220x xxxx xxxx xxxx
		// ============================================================
		{
			"Mir",
			`\b220[0-4][\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}\b`,
		},
	}

	// Compile all patterns
	cardPatterns = make([]CardPattern, 0, len(patterns))

	for _, p := range patterns {
		// Compile the regex pattern
		regex, err := regexp.Compile(p.pattern)
		if err != nil {
			// If compilation fails, return error with pattern name
			return fmt.Errorf("failed to compile pattern for %s: %w", p.name, err)
		}

		// Add compiled pattern to list
		cardPatterns = append(cardPatterns, CardPattern{
			Name:    p.name,
			Pattern: regex,
		})
	}

	// Success! All patterns compiled
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
func GetPatternCount() int {
	return len(cardPatterns)
}
