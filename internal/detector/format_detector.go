// Package detector handles credit card detection and validation
// File: internal/detector/format_detector.go
//
// This file implements PHASE 1 of the new pipeline detection algorithm:
// Finding card-like patterns in text using simple, fast regex patterns
package detector

import (
	"regexp"
	"strings"
)

// ============================================================
// DATA STRUCTURES
// ============================================================

// CardLikePattern represents a potential card number found in text
// This is just a "candidate" - not yet validated as a real card
//
// The pattern might be:
//   - A real credit card number
//   - A phone number that looks like a card
//   - A random sequence of digits
//   - An ID number or tracking code
//
// Later stages will filter out false positives
type CardLikePattern struct {
	// OriginalText is the text as found in the file
	// Examples: "4532-0151-1283-0366", "4532 0151 1283 0366"
	OriginalText string

	// Normalized is the same text with only digits (no separators)
	// Example: "4532015112830366"
	Normalized string

	// StartIndex is the position in the original text where this pattern starts
	// Used to calculate line numbers later
	StartIndex int

	// EndIndex is the position where this pattern ends
	EndIndex int
}

// ============================================================
// REGEX PATTERNS FOR CARD-LIKE SEQUENCES
// ============================================================
//
// These patterns are VERY GENERAL and FAST
// They find anything that LOOKS like a card number
// We'll validate later whether it's actually a valid card
//
// Pattern design principles:
//   1. Simple and fast (no complex lookaheads/lookbehinds)
//   2. Support common separators: space, dash, underscore
//   3. Match word boundaries (\b) to avoid partial matches
//   4. Cover all valid card lengths (14-19 digits)

var (
	// pattern16Digits matches most credit cards (16 digits)
	// Format: XXXX XXXX XXXX XXXX or XXXX-XXXX-XXXX-XXXX or XXXXXXXXXXXXXXXX
	//
	// This is the MOST COMMON pattern - used by:
	//   - Visa (standard)
	//   - Mastercard
	//   - Discover
	//   - JCB
	//   - Troy
	//   - Mir
	//   - UnionPay (credit cards)
	//
	// Performance: This pattern will match 80-90% of real cards
	pattern16Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}\b`,
	)

	// pattern15Digits matches American Express cards only
	// Format: XXXX XXXXXX XXXXX or XXXX-XXXXXX-XXXXX
	//
	// AmEx uses a unique 4-6-5 grouping
	// This is the ONLY card type that uses 15 digits in 2025
	pattern15Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{6}[\s\-_]?\d{5}\b`,
	)

	// pattern14Digits matches Diners Club traditional format
	// Format: XXXX XXXXXX XXXX or XXXX-XXXXXX-XXXX
	//
	// Note: Diners Club is transitioning to 16 digits
	// 14-digit format is legacy but still valid in 2025
	pattern14Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{6}[\s\-_]?\d{4}\b`,
	)

	// pattern19Digits matches extended format cards
	// Format: XXXX XXXX XXXX XXXX XXX
	//
	// Used by:
	//   - UnionPay (debit cards)
	//   - RuPay (some variants)
	//   - VPay (Europe-only Visa product)
	//
	// Less common but important for international support
	pattern19Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{3}\b`,
	)

	// pattern18Digits matches 18-digit cards
	// Format: XXXX XXXX XXXX XXXX XX
	//
	// Used by some UnionPay and RuPay variants
	pattern18Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{2}\b`,
	)

	// pattern17Digits matches 17-digit cards
	// Format: XXXX XXXX XXXX XXXX X
	//
	// Used by some UnionPay and RuPay variants
	pattern17Digits = regexp.MustCompile(
		`\b\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{4}[\s\-_]?\d{1}\b`,
	)
)

// ============================================================
// MAIN DETECTION FUNCTION
// ============================================================

// FindCardLikePatterns finds all potential card numbers in text
//
// This function performs PHASE 1 of the detection pipeline:
// It finds anything that LOOKS like a card number based on format
//
// Process:
//  1. Run simple regex patterns for each card length (14-19 digits)
//  2. Extract matches with their positions in the text
//  3. Normalize each match (remove separators, keep only digits)
//  4. Return all candidates for further validation
//
// Performance characteristics:
//   - Very fast: O(n) single pass through text
//   - Regex patterns are simple and optimized
//   - False positives are expected and acceptable
//   - Later stages will filter out non-cards
//
// Parameters:
//   - text: Text to scan (can be entire file or chunk)
//
// Returns:
//   - []CardLikePattern: All potential card numbers found
//
// Example:
//
//	text := "Card: 4532-0151-1283-0366 and phone: 555-1234-5678"
//	patterns := FindCardLikePatterns(text)
//	 Returns 2 patterns (both will be checked later)
//	 The phone number will be eliminated by Luhn validation
func FindCardLikePatterns(text string) []CardLikePattern {
	var patterns []CardLikePattern

	// ============================================================
	// Helper function to process matches from a regex
	// ============================================================
	// This closure avoids code duplication for each pattern
	//
	// Parameters:
	//   - matches: List of [startIndex, endIndex] pairs from regex
	//
	// Process:
	//   1. Extract original text using indices
	//   2. Normalize to digits only
	//   3. Create CardLikePattern struct
	//   4. Append to results
	processMatches := func(matches [][]int) {
		for _, match := range matches {
			// match[0] = start index in text
			// match[1] = end index in text
			original := text[match[0]:match[1]]

			// Remove all non-digit characters
			normalized := normalizeCardNumber(original)

			// Create pattern record
			patterns = append(patterns, CardLikePattern{
				OriginalText: original,
				Normalized:   normalized,
				StartIndex:   match[0],
				EndIndex:     match[1],
			})
		}
	}

	// ============================================================
	// Run all patterns in order of frequency
	// ============================================================
	// Most common pattern first = better performance on average
	//
	// Order: 16 > 15 > 19 > 18 > 17 > 14 digits
	//
	// Why this order?
	//   - 16 digits: ~85% of all cards (Visa, MC, Discover, JCB)
	//   - 15 digits: ~8% of cards (Amex only)
	//   - 14-19 digits: ~7% combined (international + Diners)

	// Pattern 1: 16 digits (MOST COMMON)
	processMatches(pattern16Digits.FindAllStringIndex(text, -1))

	// Pattern 2: 15 digits (Amex)
	processMatches(pattern15Digits.FindAllStringIndex(text, -1))

	// Pattern 3: 19 digits (UnionPay debit, RuPay, VPay)
	processMatches(pattern19Digits.FindAllStringIndex(text, -1))

	// Pattern 4: 18 digits (UnionPay, RuPay variants)
	processMatches(pattern18Digits.FindAllStringIndex(text, -1))

	// Pattern 5: 17 digits (UnionPay, RuPay variants)
	processMatches(pattern17Digits.FindAllStringIndex(text, -1))

	// Pattern 6: 14 digits (Diners Club legacy)
	processMatches(pattern14Digits.FindAllStringIndex(text, -1))

	return patterns
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// normalizeCardNumber removes all non-digit characters
//
// This converts various card number formats to a standard form:
//   - "4532-0151-1283-0366" → "4532015112830366"
//   - "4532 0151 1283 0366" → "4532015112830366"
//   - "4532_0151_1283_0366" → "4532015112830366"
//   - "4532015112830366"     → "4532015112830366"
//
// Performance: O(n) single pass, uses strings.Builder for efficiency
//
// Parameters:
//   - text: Card number with possible separators
//
// Returns:
//   - string: Only digits, no separators
func normalizeCardNumber(text string) string {
	// Use strings.Builder for efficient string construction
	// Pre-allocate capacity for maximum card length (19 digits)
	var builder strings.Builder
	builder.Grow(19)

	// Iterate through each character
	for i := 0; i < len(text); i++ {
		char := text[i]

		// Check if character is a digit (ASCII 48-57 are '0'-'9')
		if char >= '0' && char <= '9' {
			builder.WriteByte(char)
		}
		// All other characters (spaces, dashes, underscores) are ignored
	}

	return builder.String()
}
