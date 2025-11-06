// Package detector handles credit card detection and validation
// File: internal/detector/issuer_matcher.go
//
// This file implements PHASE 2 of card detection: Issuer Matching
//
// EVOLUTION OF ISSUER DETECTION:
//
// v1.0 (Old Method - Deprecated):
//
//	❌ Hard-coded prefix checks (if first == '4' then Visa)
//	❌ Manual range checking (if first2 >= "51" && first2 <= "55")
//	❌ Switch-case statements (400+ lines of code)
//	❌ Code changes required for every new card type
//	❌ Manual priority management in overlaps
//
// v3.0.0 (Current Method - BIN Database Only):
//
//	✅ JSON-based BIN database (REQUIRED)
//	✅ 6-digit BIN matching (more accurate)
//	✅ Automatic priority-based overlap resolution
//	✅ Easy updates (modify JSON, no code changes)
//	✅ Regional card support (RuPay, Mir, Troy, Elo, etc.)
//	✅ Clean, maintainable code (~50 lines vs 400+)
//	✅ No fallback - database is mandatory
//
// ADVANTAGES:
//   - Accuracy: 6-digit BIN vs 2-4 digit prefix matching
//   - Maintainability: Update JSON instead of code
//   - Scalability: Add new issuers without code changes
//   - Correctness: Priority system prevents false positives
//   - Simplicity: Single source of truth (BIN database)
//
// OVERLAP RESOLUTION EXAMPLES:
//
//	✅ RuPay (60xxxx) vs Visa (4xxxxx): RuPay priority 65 > Visa priority 55
//	✅ Discover (622126-622925) vs UnionPay (62xxxx): Discover priority 70 > UnionPay priority 60
//	✅ Mir (2200-2204) vs Mastercard (2221-2720): Mir priority 75 > Mastercard priority 60
//
// CRITICAL REQUIREMENT:
//
//	The BIN database MUST be initialized before any card detection.
//	If the database is not initialized, the application will panic.
//	This is intentional - the application cannot function without it.
//
// USAGE:
//
//	// Initialize once at application startup (in main.go)
//	err := InitGlobalBINDatabase("")
//	if err != nil {
//	    log.Fatal("Failed to initialize BIN database:", err)
//	}
//
//	// Use anywhere in the application
//	issuer, found := MatchIssuer("4532015112830366")
//	if found {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
//
// NOTE: All global database management functions (InitGlobalBINDatabase,
// GetGlobalBINDatabase, etc.) are defined in bin_lookup.go
package detector

import (
	"fmt"
)

// ============================================================
// MAIN ISSUER MATCHING FUNCTION
// ============================================================

// MatchIssuer determines the card issuer from a normalized card number
//
// This is the PRIMARY FUNCTION for issuer detection in the entire application.
//
// REQUIREMENTS:
//   - BIN database MUST be initialized before calling this function
//   - If database is not initialized, this function will panic
//   - This is intentional: the application cannot work without the database
//
// ALGORITHM:
//  1. Validate input (length check)
//  2. Get global BIN database (panic if not initialized)
//  3. Extract first 6 digits (BIN)
//  4. Query database with BIN and card length
//  5. Return issuer name if found
//
// Parameters:
//   - normalized: Card number with digits only (no spaces, dashes, etc.)
//     Length: 13-19 digits
//     Example: "4532015112830366" (Visa 16-digit)
//     "378282246310005" (Amex 15-digit)
//     "6000010000000000" (RuPay 16-digit)
//
// Returns:
//   - string: Issuer name ("Visa", "MasterCard", "Amex", "Discover", etc.)
//   - bool: true if issuer identified, false if unknown
//
// Example usage:
//
//	// Basic usage
//	issuer, found := MatchIssuer("4532015112830366")
//	if found {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	} else {
//	    fmt.Println("Unknown card type")
//	}
//
//	// RuPay detection (would be Visa in old system)
//	issuer, _ := MatchIssuer("6000010000000000")
//	fmt.Println(issuer) // Output: RuPay ✓
//
//	// Discover vs UnionPay overlap
//	issuer, _ := MatchIssuer("6221270000000000")
//	fmt.Println(issuer) // Output: Discover ✓ (not UnionPay)
//
// Important notes:
//   - This function does NOT perform Luhn validation
//   - Luhn validation is done separately in PHASE 3 (pipeline_detector.go)
//   - This separation allows early filtering of non-card numbers
//   - BIN database must be initialized or this function will panic
func MatchIssuer(normalized string) (string, bool) {
	// ============================================================
	// STEP 1: Input validation
	// ============================================================

	// Validate card number length
	// Valid credit cards are between 13-19 digits
	// 13 digits: Old Visa format (mostly deprecated)
	// 14-16 digits: Most common (Visa, Mastercard, Discover, etc.)
	// 19 digits: Some newer Visa cards
	cardLength := len(normalized)
	if cardLength < 13 || cardLength > 19 {
		return "", false
	}

	// Validate we have at least 6 digits for BIN extraction
	// BIN (Bank Identification Number) is the first 6 digits
	if cardLength < 6 {
		return "", false
	}

	// ============================================================
	// STEP 2: Get global BIN database
	// ============================================================

	// Retrieve the global database instance
	// This MUST succeed - if it doesn't, the application is misconfigured
	// GetGlobalBINDatabase is defined in bin_lookup.go
	db, err := GetGlobalBINDatabase()
	if err != nil {
		// CRITICAL ERROR: BIN database not initialized
		// This should NEVER happen if main.go properly initializes the database
		// We panic here because the application cannot function without the database
		panic(fmt.Sprintf("CRITICAL: BIN database not initialized - call InitGlobalBINDatabase() in main(): %v", err))
	}

	// ============================================================
	// STEP 3: Extract BIN (first 6 digits)
	// ============================================================

	// Extract the Bank Identification Number
	// This is the standard 6-digit issuer identifier
	// Example: "4532015112830366" → "453201"
	bin := normalized[:6]

	// ============================================================
	// STEP 4: Query BIN database
	// ============================================================

	// Perform lookup with both BIN and card length
	// The database will:
	//   1. Check BIN against all issuer ranges (sorted by priority)
	//   2. Validate card length matches issuer specifications
	//   3. Return first match (priority ensures correctness)
	issuer, found := db.LookupBIN(bin, cardLength)

	// ============================================================
	// STEP 5: Return result
	// ============================================================

	if found {
		// Successfully identified issuer
		return issuer, true
	}

	// BIN not found in database
	// This could mean:
	//   - New/unknown issuer
	//   - Test card number
	//   - Invalid card number
	//   - Regional card not in our database
	return "", false
}
