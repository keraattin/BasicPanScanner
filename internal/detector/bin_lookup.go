// Package detector - BIN Database Lookup System
// File: internal/detector/bin_lookup.go
//
// This file implements the BIN (Bank Identification Number) database system
// for accurate credit card type detection.
//
// ADVANTAGES OF BIN DATABASE APPROACH:
//
//	✅ 6-digit BIN matching (vs 2-4 digit prefix matching)
//	✅ Priority-based overlap resolution
//	✅ Regional card support (RuPay, Mir, Troy, etc.)
//	✅ Offline operation (no API dependencies)
//	✅ Easy updates (modify JSON, no code changes)
//	✅ Length validation (additional accuracy)
//
// PERFORMANCE:
//   - Lookup time: O(n) where n = number of issuers (~15)
//   - Memory usage: ~50-100KB for complete database
//   - Initialization: One-time at application startup
//
// USAGE EXAMPLE:
//
//	// Initialize once at startup
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("bindata/bin_ranges.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Query the database
//	issuer, found := db.LookupBIN("453201", 16)
//	if found {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

// ============================================================
// CONSTANTS
// ============================================================

// DefaultBINDatabasePath is the default location for the BIN database file
// Relative to the project root directory
const DefaultBINDatabasePath = "internal/detector/bindata/bin_ranges.json"

// ============================================================
// DATA STRUCTURES
// ============================================================

// BINRange represents a range of BIN numbers for a card issuer
//
// BINs are stored as strings (not integers) in JSON for leading zero support
// Example: "000000" to "099999" for certain test cards
//
// Example for Visa:
//
//	BINRange{
//	    Start: "400000",
//	    End:   "499999",
//	}
type BINRange struct {
	// Start is the beginning of the BIN range (inclusive)
	// Format: 6-digit string
	// Example: "400000" (Visa's starting BIN)
	Start string `json:"start"`

	// End is the end of the BIN range (inclusive)
	// Format: 6-digit string
	// Example: "499999" (Visa's ending BIN)
	End string `json:"end"`
}

// BINIssuer contains all information about a card issuer
//
// This structure holds:
//   - Issuer identification (name, display name)
//   - BIN ranges (multiple ranges per issuer)
//   - Card length specifications
//   - Priority for overlap resolution
//   - Regional information
//   - Active status
type BINIssuer struct {
	// Issuer is the canonical issuer name
	// Example: "Visa", "MasterCard", "Amex"
	// This is used internally and in reports
	Issuer string `json:"issuer"`

	// DisplayName is the user-friendly name
	// Example: "American Express", "UnionPay"
	DisplayName string `json:"display_name"`

	// Ranges contains all BIN ranges for this issuer
	// An issuer may have multiple non-contiguous ranges
	// Example: Maestro has ranges in 50, 56, 57, 58, 6 prefixes
	Ranges []BINRange `json:"ranges"`

	// Lengths specifies valid card number lengths
	// Example: [15] for Amex (15-digit only)
	//          [16, 19] for Visa (16 or 19 digits)
	Lengths []int `json:"lengths"`

	// Priority determines checking order for overlapping ranges
	// Higher priority = checked first
	// Range: 0-100, where:
	//   90-100: Highest priority (regional cards with conflicts)
	//   70-89:  High priority (major international cards)
	//   40-69:  Medium priority (common cards)
	//   1-39:   Low priority (less common cards)
	Priority int `json:"priority"`

	// Region indicates primary geographic region
	// Example: "Global", "India", "Russia", "Brazil"
	Region string `json:"region"`

	// Active indicates if this issuer should be used for detection
	// Set to false for deprecated or testing entries
	Active bool `json:"active"`

	// Notes contains additional information
	// Used for documentation and special cases
	Notes string `json:"notes"`
}

// BINDatabase is the main structure for BIN lookups
//
// This structure contains:
//   - Metadata (version, last update date)
//   - Issuer list (sorted by priority)
//   - Sorted ranges (for binary search)
//
// The database is loaded once at startup and kept in memory
type BINDatabase struct {
	// Version is the database version
	// Format: Semantic versioning (e.g., "3.0.0")
	Version string

	// LastUpdated is the last database update date
	// Format: ISO 8601 (YYYY-MM-DD)
	LastUpdated string

	// Issuers contains all card issuers
	// Sorted by priority (highest first) for overlap resolution
	Issuers []BINIssuer

	// sortedRanges contains all BIN ranges in sorted order
	// Used for binary search optimization
	// Created during Load() operation
	sortedRanges []BINRange
}

// ============================================================
// DATABASE LOADER
// ============================================================

// BINDatabaseLoader handles loading the BIN database from JSON
//
// This is a simple struct that could be extended in the future
// for caching or configuration options
type BINDatabaseLoader struct {
	// Currently empty - placeholder for future enhancements
	// Possible additions:
	//   - Cache settings
	//   - Validation options
	//   - Custom error handlers
}

// NewBINDatabaseLoader creates a new BIN database loader
//
// Returns:
//   - *BINDatabaseLoader: A new loader instance
//
// Example:
//
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("path/to/bin_ranges.json")
func NewBINDatabaseLoader() *BINDatabaseLoader {
	return &BINDatabaseLoader{}
}

// jsonStructure represents the JSON file structure
//
// This is only used during JSON parsing
// The actual database uses BINDatabase structure
type jsonStructure struct {
	// Info contains metadata
	Info struct {
		Version     string `json:"version"`
		LastUpdated string `json:"last_updated"`
	} `json:"_info"`

	// BINRanges contains all issuer definitions
	BINRanges []BINIssuer `json:"bin_ranges"`
}

// Load reads and parses the BIN database from a JSON file
//
// This function performs the following steps:
//  1. Read JSON file from disk
//  2. Parse JSON structure
//  3. Sort issuers by priority (high to low)
//  4. Build sorted range list for binary search
//  5. Validate database integrity
//
// Parameters:
//   - filePath: Path to the JSON file
//     Example: "internal/detector/bindata/bin_ranges.json"
//
// Returns:
//   - *BINDatabase: Loaded and ready database
//   - error: Error if file not found, invalid JSON, or validation fails
//
// Error cases:
//   - File not found or unreadable
//   - Invalid JSON syntax
//   - Missing required fields
//   - Invalid BIN range format
//
// Example:
//
//	loader := NewBINDatabaseLoader()
//	db, err := loader.Load("bindata/bin_ranges.json")
//	if err != nil {
//	    log.Fatalf("Failed to load BIN database: %v", err)
//	}
//	fmt.Printf("Loaded BIN database v%s with %d issuers\n",
//	    db.Version, len(db.Issuers))
func (l *BINDatabaseLoader) Load(filePath string) (*BINDatabase, error) {
	// ============================================================
	// STEP 1: Read file from disk
	// ============================================================

	// Read entire file into memory
	// BIN database is small (~50KB), so this is efficient
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read BIN database file '%s': %w", filePath, err)
	}

	// Validate file is not empty
	if len(data) == 0 {
		return nil, fmt.Errorf("BIN database file '%s' is empty", filePath)
	}

	// ============================================================
	// STEP 2: Parse JSON
	// ============================================================

	var jsonData jsonStructure
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BIN database JSON: %w", err)
	}

	// Validate metadata exists
	if jsonData.Info.Version == "" {
		return nil, fmt.Errorf("BIN database missing version information")
	}

	// ============================================================
	// STEP 3: Create database structure
	// ============================================================

	db := &BINDatabase{
		Version:     jsonData.Info.Version,
		LastUpdated: jsonData.Info.LastUpdated,
		Issuers:     jsonData.BINRanges,
	}

	// Validate we have at least some issuers
	if len(db.Issuers) == 0 {
		return nil, fmt.Errorf("BIN database contains no issuers")
	}

	// ============================================================
	// STEP 4: Sort issuers by priority
	// ============================================================
	// Higher priority issuers are checked first
	// This ensures correct detection in overlap cases
	//
	// Example: Discover (priority 70) is checked before UnionPay (priority 60)
	// Result: BIN 622127 is correctly identified as Discover

	sort.Slice(db.Issuers, func(i, j int) bool {
		// Sort in descending order (high priority first)
		return db.Issuers[i].Priority > db.Issuers[j].Priority
	})

	// ============================================================
	// STEP 5: Build sorted range list
	// ============================================================
	// Collect all ranges for potential optimization

	err = db.buildSortedRanges()
	if err != nil {
		return nil, fmt.Errorf("failed to build sorted BIN ranges: %w", err)
	}

	return db, nil
}

// buildSortedRanges collects all BIN ranges into a sorted list
//
// This function:
//  1. Collects all ranges from all issuers
//  2. Sorts them by start BIN
//  3. Stores in database for potential binary search
//
// Currently not used for lookup (linear search is sufficient)
// but kept for future optimization possibilities
//
// Returns:
//   - error: Error if BIN range format is invalid
func (db *BINDatabase) buildSortedRanges() error {
	// Collect all ranges
	var allRanges []BINRange

	for _, issuer := range db.Issuers {
		// Skip inactive issuers
		if !issuer.Active {
			continue
		}

		// Add all ranges for this issuer
		for _, r := range issuer.Ranges {
			// Validate range format
			if len(r.Start) != 6 || len(r.End) != 6 {
				return fmt.Errorf("invalid BIN range format for %s: start=%s, end=%s",
					issuer.Issuer, r.Start, r.End)
			}

			allRanges = append(allRanges, r)
		}
	}

	// Sort ranges by start BIN
	sort.Slice(allRanges, func(i, j int) bool {
		return allRanges[i].Start < allRanges[j].Start
	})

	db.sortedRanges = allRanges
	return nil
}

// ============================================================
// DATABASE QUERY METHODS
// ============================================================

// LookupBIN searches for a card issuer by BIN number
//
// This is the main query function used for card detection.
//
// Algorithm:
//  1. Validate BIN format (must be 6 digits)
//  2. Iterate through issuers (sorted by priority)
//  3. For each issuer, check all ranges
//  4. If BIN matches, optionally validate card length
//  5. Return first match (priority system ensures correctness)
//
// Priority-based matching automatically resolves overlaps:
//   - RuPay (60xxxx, priority 65) vs Visa (4xxxxx, priority 55)
//     → RuPay checked first, correctly identified
//   - Discover (622126-622925, priority 70) vs UnionPay (62xxxx, priority 60)
//     → Discover checked first, correctly identified
//   - Mir (2200-2204, priority 75) vs Mastercard (2221-2720, priority 60)
//     → Mir checked first, correctly identified
//
// Parameters:
//
//   - bin: 6-digit BIN number (string format)
//     Example: "453201" (Visa), "622127" (Discover), "600001" (RuPay)
//
//   - cardLength: Total digits in card number
//     Example: 16, 15, 19
//     Pass 0 to skip length validation
//
// Returns:
//   - string: Issuer name ("Visa", "MasterCard", "Amex", etc.)
//   - bool: true if issuer found, false otherwise
//
// Example usage:
//
//	// Simple lookup (BIN only)
//	issuer, found := db.LookupBIN("453201", 0)
//	if found {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
//
//	// With length validation (more accurate)
//	issuer, found := db.LookupBIN("453201", 16)
//	if found {
//	    fmt.Printf("Card type: %s\n", issuer) // Output: Visa
//	}
//
//	// Overlap example: RuPay vs Visa
//	issuer, _ := db.LookupBIN("600001", 16)
//	fmt.Println(issuer) // Output: RuPay (correct, not Visa)
//
//	// Overlap example: Discover vs UnionPay
//	issuer, _ := db.LookupBIN("622127", 16)
//	fmt.Println(issuer) // Output: Discover (correct, not UnionPay)
func (db *BINDatabase) LookupBIN(bin string, cardLength int) (string, bool) {
	// ============================================================
	// STEP 1: Input validation
	// ============================================================

	// BIN must be at least 6 digits
	// Shorter BINs are not reliable for identification
	if len(bin) < 6 {
		return "", false
	}

	// Extract first 6 digits
	// This is the standard BIN length according to ISO/IEC 7812
	bin6 := bin[:6]

	// Convert BIN to integer for range comparison
	// Example: "453201" → 453201
	binInt, err := strconv.Atoi(bin6)
	if err != nil {
		// BIN is not numeric - invalid
		return "", false
	}

	// ============================================================
	// STEP 2: Priority-based lookup
	// ============================================================
	// Issuers are already sorted by priority (high to low)
	// We check each issuer until we find a match
	//
	// This approach automatically resolves overlaps:
	//   - Higher priority issuer is checked first
	//   - First match wins
	//   - Result: Correct issuer even in overlap cases

	for _, issuer := range db.Issuers {
		// Skip inactive issuers
		if !issuer.Active {
			continue
		}

		// Check all ranges for this issuer
		// An issuer may have multiple non-contiguous ranges
		for _, r := range issuer.Ranges {
			// Convert range boundaries to integers
			start, err1 := strconv.Atoi(r.Start)
			end, err2 := strconv.Atoi(r.End)

			// Skip invalid ranges (should not happen with valid JSON)
			if err1 != nil || err2 != nil {
				continue
			}

			// Check if BIN is within this range
			// Range is inclusive: start <= BIN <= end
			if binInt >= start && binInt <= end {
				// ============================================================
				// STEP 3: Optional length validation
				// ============================================================
				// If cardLength is provided (not 0), verify it matches
				// This provides additional accuracy
				//
				// Example: 15-digit card cannot be Visa (Visa is 16 or 19)

				if cardLength > 0 {
					// Check if this issuer supports this length
					lengthSupported := false
					for _, supportedLen := range issuer.Lengths {
						if cardLength == supportedLen {
							lengthSupported = true
							break
						}
					}

					// Length not supported - skip to next issuer
					if !lengthSupported {
						continue
					}
				}

				// ============================================================
				// SUCCESS: BIN and length matched
				// ============================================================
				// Due to priority system, first match is correct
				return issuer.Issuer, true
			}
		}
	}

	// ============================================================
	// No match found
	// ============================================================
	// This BIN is either:
	//   - Unknown issuer
	//   - Test/invalid BIN
	//   - Not in our database
	return "", false
}

// GetIssuerInfo retrieves detailed information about an issuer
//
// This function is useful for:
//   - Debugging
//   - Detailed reporting
//   - Understanding issuer properties
//
// Parameters:
//   - issuerName: Issuer name (e.g., "Visa", "MasterCard")
//
// Returns:
//   - *BINIssuer: Complete issuer information
//   - error: Error if issuer not found
//
// Example:
//
//	info, err := db.GetIssuerInfo("Visa")
//	if err == nil {
//	    fmt.Printf("Display name: %s\n", info.DisplayName)
//	    fmt.Printf("Priority: %d\n", info.Priority)
//	    fmt.Printf("Number of ranges: %d\n", len(info.Ranges))
//	}
func (db *BINDatabase) GetIssuerInfo(issuerName string) (*BINIssuer, error) {
	for _, issuer := range db.Issuers {
		if issuer.Issuer == issuerName {
			return &issuer, nil
		}
	}
	return nil, fmt.Errorf("issuer '%s' not found in database", issuerName)
}

// GetVersion returns the database version
func (db *BINDatabase) GetVersion() string {
	return db.Version
}

// GetLastUpdated returns the last update date
func (db *BINDatabase) GetLastUpdated() string {
	return db.LastUpdated
}

// GetIssuerCount returns the number of active issuers
func (db *BINDatabase) GetIssuerCount() int {
	count := 0
	for _, issuer := range db.Issuers {
		if issuer.Active {
			count++
		}
	}
	return count
}

// GetAllIssuers returns a list of all active issuer names
//
// Returns:
//   - []string: List of issuer names sorted by priority
//
// Example:
//
//	issuers := db.GetAllIssuers()
//	fmt.Printf("Supported cards: %s\n", strings.Join(issuers, ", "))
func (db *BINDatabase) GetAllIssuers() []string {
	var issuers []string
	for _, issuer := range db.Issuers {
		if issuer.Active {
			issuers = append(issuers, issuer.Issuer)
		}
	}
	return issuers
}

// ============================================================
// GLOBAL DATABASE SINGLETON
// ============================================================

// globalBINDatabase is the singleton instance of the BIN database
// This is initialized once at application startup and shared across
// all goroutines safely using sync.Once
var (
	globalBINDatabase *BINDatabase
	dbOnce            sync.Once
	dbInitError       error
)

// InitGlobalBINDatabase initializes the global BIN database
//
// This function MUST be called once at application startup
// before any card detection operations.
//
// Thread-safety: Uses sync.Once to ensure single initialization
// even if called from multiple goroutines concurrently.
//
// Parameters:
//   - path: Path to the BIN database JSON file
//     Pass empty string "" to use default path
//     Default: "internal/detector/bindata/bin_ranges.json"
//
// Returns:
//   - error: Error if database fails to load
//
// Example usage in main.go:
//
//	func main() {
//	    // Initialize BIN database (use default path)
//	    err := detector.InitGlobalBINDatabase("")
//	    if err != nil {
//	        log.Fatalf("Fatal: Failed to initialize BIN database: %v", err)
//	    }
//
//	    // Now MatchIssuer() can be used anywhere in the app
//	    // ...
//	}
func InitGlobalBINDatabase(path string) error {
	// Use sync.Once to ensure initialization happens exactly once
	// Even if called from multiple goroutines concurrently
	dbOnce.Do(func() {
		// Use default path if not specified
		if path == "" {
			path = DefaultBINDatabasePath
		}

		// Create loader and load database
		loader := NewBINDatabaseLoader()
		db, err := loader.Load(path)
		if err != nil {
			dbInitError = fmt.Errorf("failed to load BIN database from '%s': %w", path, err)
			return
		}

		// Success - store database globally
		globalBINDatabase = db
		dbInitError = nil
	})

	return dbInitError
}

// GetGlobalBINDatabase returns the global BIN database instance
//
// This function is used internally by MatchIssuer() and can also
// be used by other parts of the application for direct database access.
//
// Returns:
//   - *BINDatabase: The global database instance
//   - error: Error if database is not initialized
//
// Example:
//
//	// Get database info
//	db, err := detector.GetGlobalBINDatabase()
//	if err == nil {
//	    fmt.Printf("Database version: %s\n", db.GetVersion())
//	    fmt.Printf("Number of issuers: %d\n", db.GetIssuerCount())
//	}
func GetGlobalBINDatabase() (*BINDatabase, error) {
	if globalBINDatabase == nil {
		return nil, fmt.Errorf("BIN database not initialized - call InitGlobalBINDatabase() first")
	}
	return globalBINDatabase, nil
}
