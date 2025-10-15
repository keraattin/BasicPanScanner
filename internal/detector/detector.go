// Package detector - Main card detection logic
// This file provides the high-level card detection functionality
package detector

// FindCardsInText scans a text string for credit card numbers
// It combines regex pattern matching with Luhn validation
//
// Process:
//   1. Try each card type's regex pattern
//   2. For each match, extract only digits
//   3. Validate with Luhn algorithm
//   4. Return valid cards with their types
//
// Parameters:
//   - text: Text to scan (e.g., a line from a file)
//
// Returns:
//   - map[string]string: Map of card number -> card type
//                        Example: {"4532015112830366": "Visa"}
//
// Example:
//   line := "Payment with card 4532-0151-1283-0366 processed"
//   cards := FindCardsInText(line)
//   // Returns: map[string]string{"4532015112830366": "Visa"}
func FindCardsInText(text string) map[string]string {
	// Map to store found cards
	// Key: card number (digits only)
	// Value: card type (e.g., "Visa")
	foundCards := make(map[string]string)

	// Try each card pattern
	for _, pattern := range cardPatterns {
		// Find all matches for this pattern in the text
		// Example: For Visa pattern, might match "4532-0151-1283-0366"
		matches := pattern.Pattern.FindAllString(text, -1)

		// Process each match
		for _, match := range matches {
			// Extract only the digits (remove spaces/dashes)
			// "4532-0151-1283-0366" -> "4532015112830366"
			cardNumber := cleanDigits(match)

			// Skip if we already found this card number
			// This prevents duplicate entries
			if _, exists := foundCards[cardNumber]; exists {
				continue
			}

			// Validate with Luhn algorithm
			// This eliminates false positives (numbers that look like cards but aren't)
			if ValidateLuhn(cardNumber) {
				// Valid card! Store it with its type
				foundCards[cardNumber] = pattern.Name
			}
		}
	}

	return foundCards
}

// CardFinding represents a single credit card finding
// This struct contains all information about a found card
type CardFinding struct {
	CardNumber string // Full card number (digits only)
	CardType   string // Card issuer (e.g., "Visa")
	Masked     string // PCI-compliant masked version
}

// FindAndMaskCards finds cards in text and returns CardFinding structs
// This is a convenience function that combines finding and masking
//
// Parameters:
//   - text: Text to scan
//
// Returns:
//   - []CardFinding: List of found cards with masking
//
// Example:
//   findings := FindAndMaskCards("Card: 4532015112830366")
//   for _, finding := range findings {
//       fmt.Printf("Found %s: %s\n", finding.CardType, finding.Masked)
//   }
//   Output: Found Visa: 453201******0366
func FindAndMaskCards(text string) []CardFinding {
	// Find all cards in the text
	cards := FindCardsInText(text)

	// Convert map to slice of CardFinding structs
	findings := make([]CardFinding, 0, len(cards))

	for cardNumber, cardType := range cards {
		findings = append(findings, CardFinding{
			CardNumber: cardNumber,
			CardType:   cardType,
			Masked:     MaskCardNumber(cardNumber),
		})
	}

	return findings
}

// DetectionStats holds statistics about card detection
// Useful for reporting and monitoring
type DetectionStats struct {
	TotalScanned  int            // Total texts scanned
	TotalFound    int            // Total cards found
	ByType        map[string]int // Cards found by type
	FalsePositive int            // Pattern matches that failed Luhn
}

// NewDetectionStats creates a new DetectionStats instance
func NewDetectionStats() *DetectionStats {
	return &DetectionStats{
		ByType: make(map[string]int),
	}
}

// Update updates statistics with new findings
func (ds *DetectionStats) Update(cards map[string]string) {
	ds.TotalScanned++

	for _, cardType := range cards {
		ds.TotalFound++
		ds.ByType[cardType]++
	}
}
