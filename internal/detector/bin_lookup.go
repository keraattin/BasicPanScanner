// Package detector - BIN Database Lookup System
// File: internal/detector/bin_lookup.go
//
// This file performs card detection based on the BIN (Bank Identification Number) database
// It allows high-accuracy card type identification using the BIN database
//
// ADVANTAGES:
//
//	✅ 6-digit BIN check (instead of 2-4 digits) = fewer false positives
//	✅ Overlap management (resolves conflicts using a priority system)
//	✅ Regional card support (Elo, Verve, Troy, Mir, RuPay, etc.)
//	✅ Offline operation (no dependency on an API)
//	✅ Fast querying (binary search O(log n))
//
// USAGE:
//
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("bindata/bin_ranges.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	issuer, ok := db.LookupBIN("453201")
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Card type: Visa
//	}
package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// ============================================================
// CONSTANTS - DEFAULT VALUES
// ============================================================

// DefaultBINDatabasePath default location for the BIN database file
// The application first looks in the current directory to find this file
const DefaultBINDatabasePath = "internal/detector/bindata/bin_ranges.json"

// ============================================================
// DATA STRUCTURES - DATA STRUCTURES
// ============================================================

// BINRange represents the BIN range for a card issuer
//
// Example:
//
//	For Visa cards:
//	  BINRange{Start: 400000, End: 499999}
//	This indicates that all BINs between 400000 and 499999 are Visa cards
//
// Note: BINs are stored as integers because range comparison
// is faster than string comparison
type BINRange struct {
	// Start the starting value of the range (inclusive)
	// Example: 400000 (Visa's first BIN)
	Start int `json:"start"`

	// End the ending value of the range (inclusive)
	// Example: 499999 (Visa's last BIN)
	End int `json:"end"`

	// Issuer the card issuer for this range
	// Example: "Visa", "MasterCard", "Amex"
	Issuer string `json:"-"` // Not read from JSON, assigned later
}

// BINIssuer holds all information about a card issuer
//
// This structure contains all the properties and BIN ranges of a card network
// (Visa, Mastercard, etc.)
//
// Example:
//
//	For Visa:
//	  BINIssuer{
//	    Issuer: "Visa",
//	    DisplayName: "Visa",
//	    Ranges: []BINRange{{Start: 400000, End: 499999}},
//	    Lengths: []int{16, 19},
//	    Priority: 40,
//	    Active: true,
//	  }
type BINIssuer struct {
	// Issuer short name (for internal use)
	// Example: "Visa", "MasterCard", "Amex"
	Issuer string `json:"issuer"`

	// DisplayName the name to display to the user
	// Example: "Visa", "American Express", "China UnionPay"
	DisplayName string `json:"display_name"`

	// Ranges all BIN ranges for this issuer
	// An issuer can have multiple ranges
	// Example: Discover has 4 different ranges
	Ranges []struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"ranges"`

	// Lengths card lengths supported by this issuer (number of digits)
	// Example: Visa [16, 19], Amex [15], UnionPay [16, 17, 18, 19]
	Lengths []int `json:"lengths"`

	// Priority priority order for overlapping ranges
	// Higher value = checked first
	// Example: Discover (70) is checked before UnionPay (60)
	//        Thus, the range 622126-622925 is detected as Discover
	Priority int `json:"priority"`

	// Region the region where this card is used
	// Example: "Global", "Brazil", "India", "Turkey"
	Region string `json:"region"`

	// Active whether this issuer is active or not
	// If false, it is ignored during lookup
	// Used to disable old/obsolete cards
	Active bool `json:"active"`

	// Notes additional notes (optional)
	// For informational purposes for developers
	Notes string `json:"notes"`
}

// BINDatabase is the main database structure managing BIN ranges
//
// This structure holds all card issuers and BIN ranges
// and is optimized for fast searching
//
// CONTENT:
//   - All card issuers (Visa, Mastercard, etc.)
//   - Sorted BIN ranges (for binary search)
//   - Metadata (version, update date, etc.)
//
// PERFORMANCE:
//   - Lookup: O(log n) - uses binary search
//   - Memory: ~50-100KB (for all BIN ranges)
type BINDatabase struct {
	// Version database version
	// Example: "3.0.0"
	Version string

	// LastUpdated last update date
	// Example: "2025-01-15"
	LastUpdated string

	// Issuers list of all card issuers
	// Sorted by priority (from high to low)
	Issuers []BINIssuer

	// sortedRanges sorted list of all BIN ranges
	// Used for binary search
	// NOTE: This list is created during Load()
	sortedRanges []BINRange
}

// ============================================================
// JSON LOADING - LOADING DATABASE
// ============================================================

// BINDatabaseLoader loads the BIN database from a JSON file
//
// This struct is only used for the loading process
// After loading, the BINDatabase is used
//
// USAGE:
//
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("path/to/bin_ranges.json")
type BINDatabaseLoader struct {
	// Currently empty, can later be used for cache or config
}

// NewBINDatabaseLoader creates a new BIN database loader
//
// Returns:
//   - *BINDatabaseLoader: A new loader instance
//
// Example:
//
//	loader := NewBINDatabaseLoader()
func NewBINDatabaseLoader() *BINDatabaseLoader {
	return &BINDatabaseLoader{}
}

// jsonStructure represents the structure of the JSON file
// This is only used for JSON parsing
type jsonStructure struct {
	Info struct {
		Version     string `json:"version"`
		LastUpdated string `json:"last_updated"`
	} `json:"_info"`
	BINRanges []BINIssuer `json:"bin_ranges"`
}

// Load loads the BIN database from a JSON file
//
// This function:
//  1. Reads the JSON file
//  2. Parses it
//  3. Sorts the BIN ranges (for binary search)
//  4. Validates it
//  5. Returns a BINDatabase
//
// Parameters:
//   - filePath: Path to the JSON file
//     Example: "internal/detector/bindata/bin_ranges.json"
//
// Returns:
//   - *BINDatabase: The loaded and ready database
//   - error: Error if the file cannot be read or parsed
//
// Error cases:
//   - "file not found" if the file cannot be found
//   - "invalid JSON" if the JSON is invalid
//   - warning if BIN ranges overlap (not an error)
//
// Example:
//
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("bindata/bin_ranges.json")
//	if err != nil {
//	    log.Fatalf("Failed to load BIN database: %v", err)
//	}
//	fmt.Printf("Loaded BIN database v%s\n", db.Version)
func (l *BINDatabaseLoader) Load(filePath string) (*BINDatabase, error) {
	// ============================================================
	// STEP 1: Read the file
	// ============================================================

	// Read the JSON file from disk
	// os.ReadFile loads the entire file into memory
	// Since the BIN database is small (~50KB), this is acceptable
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read BIN database file '%s': %w", filePath, err)
	}

	// ============================================================
	// STEP 2: Parse the JSON
	// ============================================================

	var jsonData jsonStructure
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BIN database JSON: %w", err)
	}

	// ============================================================
	// STEP 3: Create the BINDatabase structure
	// ============================================================

	db := &BINDatabase{
		Version:     jsonData.Info.Version,
		LastUpdated: jsonData.Info.LastUpdated,
		Issuers:     jsonData.BINRanges,
	}

	// ============================================================
	// STEP 4: Sort issuers by priority
	// ============================================================
	// Issuers with higher priority should be checked first
	// This ensures the correct issuer is detected in overlap situations
	//
	// Example: Discover (priority 70) is checked before UnionPay (priority 60)
	//        So, the range 622126-622925 is detected as Discover

	sort.Slice(db.Issuers, func(i, j int) bool {
		// Önceliği yüksek olan öne gelsin (descending order)
		return db.Issuers[i].Priority > db.Issuers[j].Priority
	})

	// ============================================================
	// STEP 5: Prepare BIN ranges for binary search
	// ============================================================
	// Collect all BIN ranges into a single sorted list
	// This allows us to perform a O(log n) binary search

	err = db.buildSortedRanges()
	if err != nil {
		return nil, fmt.Errorf("failed to build sorted BIN ranges: %w", err)
	}

	return db, nil
}

// buildSortedRanges collects all BIN ranges into a single sorted list
//
// This function:
//  1. Converts each issuer's ranges to integers
//  2. Collects all ranges into a single list
//  3. Sorts by the start value (for binary search)
//  4. Checks for overlaps (warns but continues)
//
// Process:
//   - Converts string BINs to integers ("400000" -> 400000)
//   - Adds issuer information to each range
//   - Processes in priority order (from high to low)
//   - Sorts and validates the ranges
//
// This method is private and is only called by Load()
func (db *BINDatabase) buildSortedRanges() error {
	var allRanges []BINRange

	// ============================================================
	// Process each issuer's ranges
	// ============================================================
	// Issuers are already sorted by priority (done in Load)

	for _, issuer := range db.Issuers {
		// Skip inactive issuers
		// Example: Old/retired cards
		if !issuer.Active {
			continue
		}

		// Process each range of this issuer
		for _, r := range issuer.Ranges {
			// Convert string BIN to integer
			// Example: "400000" -> 400000
			start, err := strconv.Atoi(r.Start)
			if err != nil {
				return fmt.Errorf("invalid BIN range start '%s' for issuer '%s': %w",
					r.Start, issuer.Issuer, err)
			}

			end, err := strconv.Atoi(r.End)
			if err != nil {
				return fmt.Errorf("invalid BIN range end '%s' for issuer '%s': %w",
					r.End, issuer.Issuer, err)
			}

			// Logical check: Start <= End should be true
			if start > end {
				return fmt.Errorf("invalid BIN range for issuer '%s': start (%d) > end (%d)",
					issuer.Issuer, start, end)
			}

			// dd range to the list
			allRanges = append(allRanges, BINRange{
				Start:  start,
				End:    end,
				Issuer: issuer.Issuer,
			})
		}
	}

	// ============================================================
	// Sort all ranges by the start value
	// ============================================================
	// This sorting is necessary for binary search
	// We use sort.Slice for in-place sorting

	sort.Slice(allRanges, func(i, j int) bool {
		// Smaller start values come first (ascending order)
		return allRanges[i].Start < allRanges[j].Start
	})

	db.sortedRanges = allRanges

	return nil
}

// ============================================================
// BIN LOOKUP - DETERMINE CARD TYPE
// ============================================================

// LookupBIN determines the card type from a BIN number
//
// This function uses the BIN database to identify the card type
// Resolves overlap situations using a priority system
//
// PERFORMANCE:
//   - Checks the first 6 digits (most accurate method)
//   - Priority-based lookup: O(n), but n is small (~15 issuers)
//   - Binary search for each issuer: O(log m), m = number of ranges
//   - Total: O(n * log m), but n and m are small, practically very fast
//
// FALSE POSITIVE PREVENTION:
//   - 6-digit BIN check (instead of 2-4 digits)
//   - Length validation (does the card length match one of the supported lengths?)
//   - Priority system (chooses the correct issuer in overlap cases)
//
// Parameters:
//
//   - bin: 6-digit BIN number (string)
//     Example: "453201" (Visa), "622127" (Discover), "600001" (RuPay)
//
//   - cardLength: Total number of digits of the card
//     Example: 16, 15, 19
//     This parameter is optional; if 0 is provided, length validation is not performed
//
// Returns:
//   - string: Issuer name ("Visa", "MasterCard", "Amex", etc.)
//   - bool: true = found, false = not found
//
// Example Usage:
//
//	// Simple lookup (only with BIN)
//	issuer, ok := db.LookupBIN("453201", 0)
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Card type: Visa
//	}
//
//	// With length check (more accurate)
//	issuer, ok := db.LookupBIN("453201", 16)
//	if ok {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Card type: Visa
//	}
//
//	// RuPay vs Visa overlap example
//	issuer1, _ := db.LookupBIN("600001", 16)
//	fmt.Println(issuer1) // Output: RuPay (priority 65, checked before Visa)
//
//	// Discover vs UnionPay overlap example
//	issuer2, _ := db.LookupBIN("622127", 16)
//	fmt.Println(issuer2) // Output: Discover (priority 70, checked before UnionPay)
func (db *BINDatabase) LookupBIN(bin string, cardLength int) (string, bool) {
	// ============================================================
	// STEP 1: Input validation
	// ============================================================

	// BIN must be at least 6 digits
	// Shorter BINs are not reliable for identification
	if len(bin) < 6 {
		return "", false
	}

	// Get the first 6 digits
	// According to BIN standards, the first 6 digits determine the issuer
	bin6 := bin[:6]

	// Convert BIN to integer (for range check)
	// Example: "453201" -> 453201
	binInt, err := strconv.Atoi(bin6)
	if err != nil {
		// BIN sayısal değilse invalid
		return "", false
	}

	// ============================================================
	// STEP 2: Priority-based lookup
	// ============================================================
	// Issuers are already sorted by priority
	// Start with the highest-priority issuer, continue until a match is found
	//
	// This approach automatically resolves overlap situations:
	// Example: 622127 is in both Discover and UnionPay ranges
	//        But Discover (priority 70) is checked first
	//        Therefore, it is identified as Discover ✓

	for _, issuer := range db.Issuers {
		// Skip inactive issuers
		if !issuer.Active {
			continue
		}

		// Check the ranges of this issuer
		// An issuer may have multiple ranges
		for _, r := range issuer.Ranges {
			// Convert the range to integer
			start, _ := strconv.Atoi(r.Start)
			end, _ := strconv.Atoi(r.End)

			// Is the BIN within this range?
			// Check if start <= binInt <= end
			if binInt >= start && binInt <= end {
				// ============================================================
				// STEP 3: Length validation (optional)
				// ============================================================
				// If cardLength is provided (not 0)
				// Check if the card length is valid for this issuer
				//
				// This extra check reduces false positives further
				// Example: A 15-digit card cannot be Visa (Visa supports 16 or 19 digits)

				if cardLength > 0 {
					// Does this issuer support this length?
					lengthSupported := false
					for _, supportedLen := range issuer.Lengths {
						if cardLength == supportedLen {
							lengthSupported = true
							break
						}
					}

					// If the length is not supported, skip to the next issuer
					if !lengthSupported {
						continue
					}
				}

				// ============================================================
				// SUCCESS! BIN and length matched
				// ============================================================
				// Due to the priority system, the first matching issuer is the correct one
				return issuer.Issuer, true
			}
		}
	}

	// ============================================================
	// No issuer matched
	// ============================================================
	// This BIN is unknown or unsupported
	return "", false
}

// GetIssuerInfo returns detailed information about an issuer
//
// This function is useful for debugging and reporting
// It provides all details of an issuer (ranges, lengths, priority, etc.)
//
// Parameters:
//   - issuerName: Issuer name ("Visa", "MasterCard", etc.)
//
// Returns:
//   - *BINIssuer: Issuer information
//   - bool: true = found, false = not found
//
// Example:
//
//	info, ok := db.GetIssuerInfo("Visa")
//	if ok {
//	    fmt.Printf("Display Name: %s\n", info.DisplayName)
//	    fmt.Printf("Region: %s\n", info.Region)
//	    fmt.Printf("Supported Lengths: %v\n", info.Lengths)
//	}
func (db *BINDatabase) GetIssuerInfo(issuerName string) (*BINIssuer, bool) {
	for i := range db.Issuers {
		if db.Issuers[i].Issuer == issuerName {
			return &db.Issuers[i], true
		}
	}
	return nil, false
}

// GetVersion returns the version of the database
//
// Returns:
//   - string: Version (e.g., "3.0.0")
//
// Example:
//
//	version := db.GetVersion()
//	fmt.Printf("BIN Database v%s\n", version)
func (db *BINDatabase) GetVersion() string {
	return db.Version
}

// GetLastUpdated returns the last update date
//
// Returns:
//   - string: Date (e.g., "2025-01-15")
//
// Example:
//
//	updated := db.GetLastUpdated()
//	fmt.Printf("Last updated: %s\n", updated)
func (db *BINDatabase) GetLastUpdated() string {
	return db.LastUpdated
}

// GetIssuerCount returns the number of active issuers
//
// Returns:
//   - int: Number of active issuers
//
// Example:
//
//	count := db.GetIssuerCount()
//	fmt.Printf("Supporting %d card issuers\n", count)
func (db *BINDatabase) GetIssuerCount() int {
	count := 0
	for _, issuer := range db.Issuers {
		if issuer.Active {
			count++
		}
	}
	return count
}

// ============================================================
// GLOBAL INSTANCE - SINGLETON PATTERN
// ============================================================
// Global BIN database instance
// Loaded once at the start of the application, then shared across the app
//
// This singleton pattern is important for memory and performance:
//   ✅ Loaded only once from disk
//   ✅ Everyone uses the same instance (thread-safe)
//   ✅ Memory efficient (no multiple copies)

var globalBINDB *BINDatabase

// InitGlobalBINDatabase initializes the global BIN database instance
//
// This function should be called ONCE at the start of the application
// Typically used in main() or init()
//
// Thread Safety:
//   - This function is NOT thread-safe
//   - Should be called only once at the application start
//   - Subsequent accesses should be done through GetGlobalBINDatabase()
//
// Parameters:
//   - filePath: Path to the BIN database JSON file
//     If empty, DefaultBINDatabasePath is used
//
// Returns:
//   - error: Error if loading fails, nil if successful
//
// Example Usage:
//
//	// In main.go
//	func main() {
//	    // Initialize the BIN database
//	    err := InitGlobalBINDatabase("")
//	    if err != nil {
//	        log.Fatalf("Failed to initialize BIN database: %v", err)
//	    }
//
//	    // Now accessible everywhere
//	    // ...
//	}
func InitGlobalBINDatabase(filePath string) error {
	// Use default if path is not provided
	if filePath == "" {
		filePath = DefaultBINDatabasePath
	}

	// DLoad the database
	loader := NewBINDatabaseLoader()
	db, err := loader.Load(filePath)
	if err != nil {
		return fmt.Errorf("failed to load global BIN database: %w", err)
	}

	// Assign to global instance
	globalBINDB = db

	return nil
}

// GetGlobalBINDatabase returns the global BIN database instance
//
// This function can be called from anywhere in the application
// Should be used after InitGlobalBINDatabase() is called
//
// Thread Safety:
//   - This function is thread-safe (only performs reading)
//   - The BINDatabase struct is immutable (unchangeable)
//   - Hence, it can be safely used from multiple goroutines
//
// Returns:
//   - *BINDatabase: Global database instance
//   - error: Error if the database is not yet initialized
//
// Example Usage:
//
//	// From any package
//	func detectCardType(cardNumber string) (string, error) {
//	    db, err := GetGlobalBINDatabase()
//	    if err != nil {
//	        return "", err
//	    }
//
//	    bin := cardNumber[:6]
//	    issuer, ok := db.LookupBIN(bin, len(cardNumber))
//	    if !ok {
//	        return "", fmt.Errorf("unknown card type")
//	    }
//
//	    return issuer, nil
//	}
func GetGlobalBINDatabase() (*BINDatabase, error) {
	if globalBINDB == nil {
		return nil, fmt.Errorf("BIN database not initialized, call InitGlobalBINDatabase() first")
	}
	return globalBINDB, nil
}
