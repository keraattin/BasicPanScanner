// Package detector handles credit card detection and validation
// This file implements the Luhn algorithm for card number validation
package detector

import "strings"

// ValidateLuhn checks if a card number passes the Luhn algorithm
// The Luhn algorithm (also known as Luhn formula or modulus 10 algorithm)
// is a checksum formula used to validate credit card numbers
//
// How it works:
//   1. Start from the rightmost digit
//   2. Double every second digit (from right to left)
//   3. If doubled value > 9, subtract 9 (or sum the digits)
//   4. Sum all digits
//   5. Valid if sum is divisible by 10
//
// Parameters:
//   - cardNumber: String containing only digits (no spaces or dashes)
//
// Returns:
//   - bool: true if valid, false if invalid
//
// Example:
//   ValidateLuhn("4532015112830366") => true  (valid Visa)
//   ValidateLuhn("4532015112830367") => false (invalid - wrong checksum)
//
// Card number length requirements:
//   - Minimum: 13 digits (some Visa cards)
//   - Maximum: 19 digits (some UnionPay cards)
func ValidateLuhn(cardNumber string) bool {
	// Step 1: Remove any non-digit characters
	// This handles cases where spaces/dashes weren't already removed
	cleaned := cleanDigits(cardNumber)

	// Step 2: Validate length
	// Credit cards are typically 13-19 digits
	length := len(cleaned)
	if length < 13 || length > 19 {
		return false
	}

	// Step 3: Apply Luhn algorithm
	sum := 0
	isEven := false // Track if we're on an even position (from right)

	// Process digits from right to left
	// Example: 4532015112830366
	// We process: 6,6,3,0,8,2,1,1,5,1,0,2,3,5,4
	for i := length - 1; i >= 0; i-- {
		// Convert character to integer
		// '0' has ASCII value 48, so '0'-'0'=0, '1'-'0'=1, etc.
		digit := int(cleaned[i] - '0')

		// Step 4: Double every second digit (from right)
		if isEven {
			digit *= 2

			// Step 5: If doubled digit > 9, subtract 9
			// This is equivalent to summing the two digits
			// Example: 14 -> 1+4=5, and 14-9=5
			if digit > 9 {
				digit -= 9
			}
		}

		// Step 6: Add to running sum
		sum += digit

		// Toggle even/odd flag for next iteration
		isEven = !isEven
	}

	// Step 7: Valid if sum is divisible by 10
	// Example: sum=60, 60%10=0, so valid
	return sum%10 == 0
}

// cleanDigits extracts only digit characters from a string
// This removes spaces, dashes, and any other non-digit characters
//
// Parameters:
//   - text: Input string
//
// Returns:
//   - string: String containing only digits
//
// Example:
//   cleanDigits("4532-0151-1283-0366") => "4532015112830366"
//   cleanDigits("4532 0151 1283 0366") => "4532015112830366"
//   cleanDigits("Card: 4532015112830366") => "4532015112830366"
func cleanDigits(text string) string {
	// Pre-allocate builder with capacity for efficiency
	// Most card numbers are 16 digits, so this avoids reallocation
	var builder strings.Builder
	builder.Grow(19) // Max card length

	// Iterate through each character
	for i := 0; i < len(text); i++ {
		char := text[i]

		// Check if character is a digit (ASCII 48-57 are '0'-'9')
		if char >= '0' && char <= '9' {
			builder.WriteByte(char)
		}
	}

	return builder.String()
}

// MaskCardNumber returns a PCI-compliant masked version of a card number
// PCI DSS compliance requires showing only:
//   - First 6 digits (BIN - Bank Identification Number)
//   - Last 4 digits
//   - Everything else replaced with asterisks
//
// Parameters:
//   - cardNumber: Full card number (digits only)
//
// Returns:
//   - string: Masked card number
//
// Examples:
//   MaskCardNumber("4532015112830366") => "453201******0366"
//   MaskCardNumber("378282246310005")  => "378282*****0005" (Amex - 15 digits)
//
// Why mask?
//   - PCI DSS Requirement 3.3: Display maximum first 6 and last 4 digits
//   - Prevents unauthorized access to full card numbers
//   - Safe for logs, reports, and displays
func MaskCardNumber(cardNumber string) string {
	length := len(cardNumber)

	// If card number is too short, return as-is
	// This shouldn't happen with valid cards, but prevents panic
	if length <= 10 {
		return cardNumber
	}

	// Build masked version using strings.Builder for efficiency
	var masked strings.Builder
	masked.Grow(length) // Pre-allocate exact size needed

	// Add first 6 digits (BIN)
	// BIN identifies the issuing bank
	masked.WriteString(cardNumber[0:6])

	// Add asterisks for middle digits
	// Total length - 10 (6 from start + 4 from end)
	middleDigits := length - 10
	for i := 0; i < middleDigits; i++ {
		masked.WriteByte('*')
	}

	// Add last 4 digits
	// Used for transaction verification
	masked.WriteString(cardNumber[length-4:])

	return masked.String()
}
