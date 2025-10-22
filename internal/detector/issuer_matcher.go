// Package detector handles credit card detection and validation
// File: internal/detector/issuer_matcher.go
//
// # VERSION 3.0 - BIN DATABASE BASED SYSTEM
//
// This file implements PHASE 2: Determining the card issuer
//
// OLD METHOD (v2.0):
//
//	❌ Hard-coded prefix check (if first == '4' then Visa)
//	❌ Manual range checking (if first2 >= "51" && first2 <= "55")
//	❌ Code change needed for every new card
//	❌ Manual prioritization in overlap cases
//	❌ 400+ lines of switch-case logic
//
// NEW METHOD (v3.0):
//
//	✅ BIN database-based (loaded from JSON)
//	✅ 6-digit BIN check (more accurate)
//	✅ Priority system (automatically resolves overlaps)
//	✅ Easy update (just change the JSON)
//	✅ Regional card support (Elo, Verve, Troy, Mir, RuPay)
//	✅ ~50 lines of clean code
//
// PERFORMANCE:
//   - Old: O(1) but many comparisons
//   - New: O(n * log m) but n is small (~15 issuers), m is smaller (~5 ranges/issuer)
//   - In practice: Same speed, but much more maintainable
//
// FALSE POSITIVE REDUCTION:
//
//	✅ 6-digit BIN check (instead of 2-4 digits)
//	✅ Length validation (16-digit vs 15-digit)
//	✅ Priority-based matching (Discover first, UnionPay later)
//	✅ Region-specific cards (RuPay first, Visa later)
//
// OVERLAP SOLUTIONS:
//   - RuPay (60xxxx) vs Visa (4xxxxx): RuPay priority 65, checked first ✓
//   - Discover (622126-622925) vs UnionPay (62xxxx): Discover priority 70, checked first ✓
//   - Mir (2200-2204) vs Mastercard (2221-2720): Mir priority 75, checked first ✓
//
// USAGE:
//
//	// Initialize once at app startup (in main.go)
//	err := InitGlobalBINDatabase("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use anywhere (in any package)
//	issuer, ok := MatchIssuer("4532015112830366")
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
package detector

// ============================================================
// MAIN MATCHER FUNCTION - MAIN DETECTION FUNCTION
// ============================================================

// MatchIssuer determines the issuer from a normalized card number
//
// VERSION 3.0 IMPLEMENTATION:
// This function now uses a BIN database (instead of hard-coded checks)
//
// HOW IT WORKS:
//  1. Get the global BIN database (GetGlobalBINDatabase)
//  2. Extract the first 6 digits (BIN)
//  3. Calculate the card length
//  4. Search the database for the BIN (priority-based lookup)
//  5. If a match is found, return the issuer
//
// ADVANTAGES:
//
//	✅ Much cleaner code (~20 lines vs 400+ lines)
//	✅ Automatic overlap resolution (priority system)
//	✅ 6-digit BIN check (fewer false positives)
//	✅ Length validation (extra security)
//	✅ Easy update (change JSON, no code change)
//
// EXAMPLES OF FALSE POSITIVES (Old system):
//
//	❌ "600001..." → Visa (incorrect, should be RuPay)
//	❌ "622127..." → UnionPay (incorrect, should be Discover)
//	❌ "220100..." → Mastercard (incorrect, should be Mir)
//
// CORRECT DETECTIONS (New system):
//
//	✅ "600001..." → RuPay (priority 65, before Visa)
//	✅ "622127..." → Discover (priority 70, before UnionPay)
//	✅ "220100..." → Mir (priority 75, before Mastercard)
//
// Parameters:
//   - normalized: A card number consisting of digits only (13-19 digits)
//     Example: "4532015112830366" (Visa 16-digit)
//     "378282246310005" (Amex 15-digit)
//     "6000010000000000" (RuPay 16-digit)
//
// Returns:
//   - string: Issuer name ("Visa", "MasterCard", "Amex", etc.)
//   - bool: true = found, false = not found
//
// Example usage:
//
//	// Simple usage
//	issuer, ok := MatchIssuer("4532015112830366")
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
//
//	// RuPay detection example (incorrectly detected as Visa in the old system)
//	issuer, ok := MatchIssuer("6000010000000000")
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: RuPay ✓
//	}
//
//	// Discover vs UnionPay overlap example
//	issuer, ok := MatchIssuer("6221270000000000")
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Discover ✓
//	}
//
// Note: This function does NOT perform Luhn validation
//
//	Luhn validation is done in a separate step (PHASE 3)
func MatchIssuer(normalized string) (string, bool) {
	// ============================================================
	// STEP 1: Input validation
	// ============================================================

	// Minimum length check
	// Credit cards must be at least 13 digits long (old Visa format)
	// Modern cards are generally between 14-19 digits
	if len(normalized) < 13 || len(normalized) > 19 {
		return "", false
	}

	// At least 6 digits required for BIN
	// According to BIN standards, the first 6 digits determine the issuer
	if len(normalized) < 6 {
		return "", false
	}

	// ============================================================
	// STEP 2: Get the global BIN database
	// ============================================================

	// Get the global BIN database instance
	// This database is loaded once at the application start (in main.go)
	// Here, we are just referencing it
	db, err := GetGlobalBINDatabase()
	if err != nil {
		// Database is not initialized
		// This should not normally happen, but this is defensive programming
		// Fallback: Return to the old system (hard-coded checks)
		return matchIssuerFallback(normalized)
	}

	// ============================================================
	// STEP 3: Search the BIN database
	// ============================================================

	// Get the first 6 digits (BIN)
	// Example: "4532015112830366" -> "453201"
	bin := normalized[:6]

	// Get the card length
	// This is needed for length validation
	// Example: Amex is only 15-digit, Visa is 16 or 19-digit
	cardLength := len(normalized)

	// Search the BIN database
	// The LookupBIN function:
	//   1. Compares the BIN to all issuer ranges
	//   2. Checks based on priority order (high to low)
	//   3. Performs length validation
	//   4. Returns the first matching issuer
	issuer, found := db.LookupBIN(bin, cardLength)

	// ============================================================
	// STEP 4: Return the result
	// ============================================================

	if found {
		// Success! BIN and length match
		// The priority system resolves overlaps
		// and the correct issuer is detected
		return issuer, true
	}

	// BIN not found in the database
	// This BIN is unknown or unsupported
	return "", false
}

// ============================================================
// FALLBACK SYSTEM - BACKUP SYSTEM
// ============================================================

// matchIssuerFallback is used when the BIN database cannot be loaded
//
// This function is only called when the BIN database fails to load
// Under normal circumstances, it should never be called
//
// It implements a simple fallback:
//   - Only major card types (Visa, MC, Amex, Discover)
//   - Simple prefix checking (1-2 digits instead of 6)
//   - Overlap cases are not resolved
//
// This function is a minimal safety net
// The real detection should use the BIN database
//
// Parameters:
//   - normalized: The normalized card number
//
// Returns:
//   - string: Issuer name (only major types)
//   - bool: true = found, false = not found
func matchIssuerFallback(normalized string) (string, bool) {
	// This function only works if the BIN database fails
	// It performs minimal detection, cannot resolve overlaps

	length := len(normalized)

	// Amex (15-digit only)
	if length == 15 {
		first2 := normalized[0:2]
		if first2 == "34" || first2 == "37" {
			return "Amex", true
		}
	}

	// Diners (14-digit)
	if length == 14 {
		first2 := normalized[0:2]
		first3 := normalized[0:3]
		if first2 == "36" || first2 == "38" || first2 == "39" {
			return "Diners", true
		}
		if first3 >= "300" && first3 <= "305" {
			return "Diners", true
		}
	}

	// 16-digit cards
	if length == 16 {
		first := normalized[0]
		first2 := normalized[0:2]
		first4 := normalized[0:4]

		// Mastercard
		if first2 >= "51" && first2 <= "55" {
			return "MasterCard", true
		}
		if first4 >= "2221" && first4 <= "2720" {
			return "MasterCard", true
		}

		// Discover
		if first4 == "6011" || first2 == "65" {
			return "Discover", true
		}

		// Visa (check last, because of wide range)
		if first == '4' {
			return "Visa", true
		}
	}

	return "", false
}

// ============================================================
// BACKWARD COMPATIBILITY - BACKWARD COMPATIBILITY
// ============================================================
// The following functions are kept to support the old API
// The new code should not use these functions
// They exist only to ensure that existing code continues to work

// match16DigitCards - DEPRECATED: Use MatchIssuer instead
// This function is no longer used, only for backward compatibility
func match16DigitCards(normalized string) (string, bool) {
	if len(normalized) != 16 {
		return "", false
	}
	return MatchIssuer(normalized)
}

// match15DigitCards - DEPRECATED: Use MatchIssuer instead
// This function is no longer used, only for backward compatibility
func match15DigitCards(normalized string) (string, bool) {
	if len(normalized) != 15 {
		return "", false
	}
	return MatchIssuer(normalized)
}

// match14DigitCards - DEPRECATED: Use MatchIssuer instead
// This function is no longer used, only for backward compatibility
func match14DigitCards(normalized string) (string, bool) {
	if len(normalized) != 14 {
		return "", false
	}
	return MatchIssuer(normalized)
}

// matchVariableLengthCards - DEPRECATED: Use MatchIssuer instead
// This function is no longer used, only for backward compatibility
func matchVariableLengthCards(normalized string) (string, bool) {
	length := len(normalized)
	if length < 17 || length > 19 {
		return "", false
	}
	return MatchIssuer(normalized)
}

// ============================================================
// MIGRATION NOTES - MIGRATION NOTES
// ============================================================
//
// OLD CODE (v2.0):
//   issuer, ok := MatchIssuer("4532015112830366")
//
// NEW CODE (v3.0):
//   issuer, ok := MatchIssuer("4532015112830366")
//
// The API is the same! Only the implementation has changed.
// Existing code will continue to work without any changes.
//
// ============================================================
// INITIALIZATION REQUIRED - INITIALIZATION REQUIRED
// ============================================================
//
// In main.go (at application startup):
//
//   func main() {
//       // Load the BIN database
//       err := detector.InitGlobalBINDatabase("")
//       if err != nil {
//           log.Fatalf("Failed to load BIN database: %v", err)
//       }
//
//       // Now MatchIssuer can be used anywhere
//       // ...
//   }
//
// This initialization should only be done ONCE.
// After that, all packages can use MatchIssuer.
