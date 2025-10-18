// Package detector - Enhanced Card Pattern Definitions
// This file contains improved regex patterns for detecting different card types
// with reduced false positives and better accuracy
package detector

import (
	"fmt"
	"regexp"
)

// CardPattern represents a single card issuer's detection pattern
// Each pattern combines:
//   - Name: Card issuer (e.g., "Visa", "Mastercard")
//   - Pattern: Regex that matches the card format
//   - MinLength: Minimum valid length for this card type
//   - MaxLength: Maximum valid length for this card type
//   - ValidLengths: Specific valid lengths (more precise than min/max)
type CardPattern struct {
	Name         string         // Card issuer name
	Pattern      *regexp.Regexp // Compiled regex pattern
	MinLength    int            // Minimum card length
	MaxLength    int            // Maximum card length
	ValidLengths []int          // Valid specific lengths
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
//	     Now ready to detect cards
//	}
func InitPatterns() error {
	// Define all patterns with enhanced boundary detection
	// We use word boundaries and negative lookahead/lookbehind to reduce false positives

	patterns := []struct {
		name         string
		pattern      string
		minLength    int
		maxLength    int
		validLengths []int
	}{
		// ============================================================
		// VISA
		// - Starts with: 4
		// - Length: 13, 16, or 19 digits (most common: 16)
		// - BIN ranges: 4xxxxx
		// - Enhanced: Added boundary checks to avoid matching longer numbers
		// ============================================================
		{
			name: "Visa",
			// Pattern breakdown:
			// (?:^|[^0-9]) - Start of string OR non-digit (negative lookbehind alternative)
			// (4) - Must start with 4
			// (?:[0-9][\s\-]?){12} - Exactly 12 more digits for 13-digit cards
			// (?:...) groups for 16 and 19 digit variants
			// (?:[^0-9]|$) - End with non-digit or end of string
			pattern:      `(?:^|[^\d])(4\d{3}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d(?:\d{3})?(?:[\s\-]?\d{3})?)(?:[^\d]|$)`,
			minLength:    13,
			maxLength:    19,
			validLengths: []int{13, 16, 19},
		},

		// ============================================================
		// MASTERCARD
		// - Starts with: 51-55 OR 2221-2720 (newer BIN range)
		// - Length: 16 digits only
		// - Enhanced: More precise BIN range matching
		// ============================================================
		{
			name: "MasterCard",
			// Pattern breakdown:
			// (?:^|[^\d]) - Boundary check
			// (?:5[1-5]\d{2}|222[1-9]\d{1}|22[3-9]\d{2}|2[3-6]\d{3}|27[01]\d{2}|2720)
			//   - 51-55 followed by 2 digits (traditional range)
			//   - 2221-2720 (new range, broken down for regex efficiency)
			// [\s\-]? - Optional separator
			// \d{4} groups - 4-digit groups
			pattern:      `(?:^|[^\d])((?:5[1-5]\d{2}|222[1-9]\d{1}|22[3-9]\d{2}|2[3-6]\d{3}|27[01]\d{2}|2720)[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},

		// ============================================================
		// AMERICAN EXPRESS (Amex)
		// - Starts with: 34 or 37
		// - Length: 15 digits only
		// - Format: 3xxx xxxxxx xxxxx (4-6-5 grouping)
		// - Enhanced: Strict 15-digit enforcement
		// ============================================================
		{
			name: "Amex",
			// Pattern ensures exactly 15 digits with Amex-specific grouping
			pattern:      `(?:^|[^\d])(3[47]\d{2}[\s\-]?\d{6}[\s\-]?\d{5})(?:[^\d]|$)`,
			minLength:    15,
			maxLength:    15,
			validLengths: []int{15},
		},

		// ============================================================
		// DISCOVER
		// - Starts with: 6011, 622126-622925, 644-649, or 65
		// - Length: 16 digits (some sources say 16-19, but 16 is standard)
		// - Enhanced: Comprehensive BIN coverage
		// ============================================================
		{
			name: "Discover",
			// Complex pattern covering all Discover BIN ranges
			pattern:      `(?:^|[^\d])((?:6011|65\d{2}|64[4-9]\d|622(?:1[2-9]\d|[2-8]\d{2}|9[01]\d|92[0-5]))\d{0,2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},

		// ============================================================
		// DINERS CLUB
		// - Starts with: 36, 38, or 300-305
		// - Length: 14 digits (US/Canada), 14-19 internationally
		// - Enhanced: Focus on 14-digit standard
		// ============================================================
		{
			name: "Diners",
			// Pattern for 14-digit Diners Club cards
			pattern:      `(?:^|[^\d])(3(?:0[0-5]|[68]\d)\d[\s\-]?\d{6}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    14,
			maxLength:    14,
			validLengths: []int{14},
		},

		// ============================================================
		// JCB
		// - Starts with: 3528-3589
		// - Length: 16 digits (sometimes 15-19, but 16 is standard)
		// - Enhanced: Precise range matching
		// ============================================================
		{
			name: "JCB",
			// JCB cards starting with 35xx range
			pattern:      `(?:^|[^\d])(35(?:2[89]|[3-8]\d)\d[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},

		// ============================================================
		// CHINA UNIONPAY
		// - Starts with: 62
		// - Length: 16-19 digits
		// - Enhanced: Flexible length handling
		// ============================================================
		{
			name: "UnionPay",
			// UnionPay with variable length support
			pattern:      `(?:^|[^\d])(62\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4}(?:[\s\-]?\d{1,3})?)(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    19,
			validLengths: []int{16, 17, 18, 19},
		},

		// ============================================================
		// MAESTRO
		// - Starts with: 50, 56-69, 6390, 67
		// - Length: 12-19 digits (very flexible)
		// - Enhanced: Support for all length variants
		// ============================================================
		{
			name: "Maestro",
			// Maestro with highly variable length
			pattern:      `(?:^|[^\d])((?:5[06-9]|6[0-9])\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{2,7})(?:[^\d]|$)`,
			minLength:    12,
			maxLength:    19,
			validLengths: []int{12, 13, 14, 15, 16, 17, 18, 19},
		},

		// ============================================================
		// RUPAY (India)
		// - Starts with: 60, 6521, 6522, 6531-6536
		// - Length: 16 digits
		// - Enhanced: Added more BIN ranges
		// ============================================================
		{
			name: "RuPay",
			// RuPay with comprehensive BIN coverage
			pattern:      `(?:^|[^\d])((?:60|652[12]|653[1-6])\d{2}[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},

		// ============================================================
		// TROY (Turkey)
		// - Starts with: 9792
		// - Length: 16 digits
		// - Enhanced: Precise pattern
		// ============================================================
		{
			name:         "Troy",
			pattern:      `(?:^|[^\d])(9792[\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},

		// ============================================================
		// MIR (Russia)
		// - Starts with: 2200-2204
		// - Length: 16 digits
		// - Enhanced: Precise range
		// ============================================================
		{
			name:         "Mir",
			pattern:      `(?:^|[^\d])(220[0-4][\s\-]?\d{4}[\s\-]?\d{4}[\s\-]?\d{4})(?:[^\d]|$)`,
			minLength:    16,
			maxLength:    16,
			validLengths: []int{16},
		},
	}

	// ============================================================
	// Compile all patterns and validate
	// ============================================================

	cardPatterns = make([]CardPattern, 0, len(patterns))

	for _, p := range patterns {
		// Compile the regex pattern with case-insensitive flag
		// This helps detect cards even if they're in mixed contexts
		regex, err := regexp.Compile(p.pattern)
		if err != nil {
			// If compilation fails, return error with pattern name
			return fmt.Errorf("failed to compile pattern for %s: %w", p.name, err)
		}

		// Validate pattern configuration
		if p.minLength <= 0 || p.maxLength <= 0 {
			return fmt.Errorf("invalid length configuration for %s", p.name)
		}

		if p.minLength > p.maxLength {
			return fmt.Errorf("min length > max length for %s", p.name)
		}

		// Add compiled pattern to list
		cardPatterns = append(cardPatterns, CardPattern{
			Name:         p.name,
			Pattern:      regex,
			MinLength:    p.minLength,
			MaxLength:    p.maxLength,
			ValidLengths: p.validLengths,
		})
	}

	// Success! All patterns compiled
	fmt.Printf("✓ Loaded %d card issuer patterns (enhanced)\n", len(cardPatterns))
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
//	    fmt.Printf("Pattern: %s (lengths: %v)\n", pattern.Name, pattern.ValidLengths)
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

// ValidateCardLength checks if a card number has a valid length for its type
// This is an additional validation step beyond Luhn
//
// Parameters:
//   - cardNumber: Card number (digits only)
//   - cardType: Card issuer name
//
// Returns:
//   - bool: true if length is valid for this card type
//
// Example:
//
//	ValidateCardLength("4532015112830366", "Visa") => true (16 digits)
//	ValidateCardLength("453201511283036", "Visa") => false (15 digits)
func ValidateCardLength(cardNumber string, cardType string) bool {
	// Find the pattern for this card type
	for _, pattern := range cardPatterns {
		if pattern.Name == cardType {
			length := len(cardNumber)

			// Check against valid lengths if specified
			if len(pattern.ValidLengths) > 0 {
				for _, validLen := range pattern.ValidLengths {
					if length == validLen {
						return true
					}
				}
				return false
			}

			// Otherwise check min/max range
			return length >= pattern.MinLength && length <= pattern.MaxLength
		}
	}

	// Pattern not found
	return false
}
