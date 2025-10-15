// Package config - Validator handles configuration validation
// This file contains all validation logic for configuration settings
package config

import (
	"fmt"
	"os"
	"strings"
)

// Validate checks if the configuration is valid and usable
// It performs multiple checks and returns descriptive errors
//
// Validation checks:
//   - Scan mode is valid ("whitelist" or "blacklist")
//   - Appropriate extension lists are populated
//   - No duplicate extensions
//   - No conflicts between whitelist and blacklist
//   - Valid max file size format
//
// Returns:
//   - error: Descriptive error if validation fails, nil if valid
//
// Note: Some checks produce warnings (printed) but don't fail validation
func Validate(cfg *Config) error {
	warningCount := 0

	// ============================================================
	// SCAN MODE VALIDATION
	// ============================================================

	// List of valid scan modes
	validModes := map[string]bool{
		"whitelist": true,
		"blacklist": true,
	}

	// Check if scan_mode is set
	if cfg.ScanMode == "" {
		// Not critical - we can default to blacklist
		fmt.Println("⚠ Warning: scan_mode not set, defaulting to 'blacklist'")
		cfg.ScanMode = "blacklist"
		warningCount++
	} else if !validModes[cfg.ScanMode] {
		// This IS critical - invalid mode means we can't operate
		return fmt.Errorf("config error: scan_mode must be 'whitelist' or 'blacklist', got '%s'", cfg.ScanMode)
	}

	// ============================================================
	// EXTENSION LIST VALIDATION
	// ============================================================

	// Whitelist mode MUST have extensions defined
	if cfg.ScanMode == "whitelist" && len(cfg.WhitelistExtensions) == 0 {
		return fmt.Errorf("config error: whitelist mode requires whitelist_extensions to be set")
	}

	// Blacklist mode with empty blacklist is allowed but should warn
	if cfg.ScanMode == "blacklist" && len(cfg.BlacklistExtensions) == 0 {
		fmt.Println("⚠ Warning: blacklist mode with empty blacklist means ALL files will be scanned")
		fmt.Println("  Tip: Add common binary/media extensions to blacklist for better performance")
		warningCount++
	}

	// ============================================================
	// DUPLICATE DETECTION IN WHITELIST
	// ============================================================

	if len(cfg.WhitelistExtensions) > 0 {
		duplicates := findDuplicates(cfg.WhitelistExtensions)
		if len(duplicates) > 0 {
			fmt.Printf("⚠ Warning: duplicate extensions in whitelist: %v\n", duplicates)
			fmt.Println("  Tip: Edit config.json to remove duplicates")
			warningCount++
		}
	}

	// ============================================================
	// DUPLICATE DETECTION IN BLACKLIST
	// ============================================================

	if len(cfg.BlacklistExtensions) > 0 {
		duplicates := findDuplicates(cfg.BlacklistExtensions)
		if len(duplicates) > 0 {
			fmt.Printf("⚠ Warning: duplicate extensions in blacklist: %v\n", duplicates)
			fmt.Println("  Tip: Edit config.json to remove duplicates")
			warningCount++
		}
	}

	// ============================================================
	// CONFLICT DETECTION (Extension in BOTH lists)
	// ============================================================

	if len(cfg.WhitelistExtensions) > 0 && len(cfg.BlacklistExtensions) > 0 {
		conflicts := findConflicts(cfg.WhitelistExtensions, cfg.BlacklistExtensions)
		if len(conflicts) > 0 {
			fmt.Printf("⚠ Warning: extensions in BOTH whitelist and blacklist: %v\n", conflicts)
			fmt.Println("  Note: Only the active mode's list will be used")
			warningCount++
		}
	}

	// ============================================================
	// MAX FILE SIZE VALIDATION
	// ============================================================

	if cfg.MaxFileSize != "" {
		_, err := ParseFileSize(cfg.MaxFileSize)
		if err != nil {
			// This IS critical - invalid size format
			return fmt.Errorf("config error: invalid max_file_size '%s': %v", cfg.MaxFileSize, err)
		}
	}

	// ============================================================
	// EXCLUDE DIRECTORIES VALIDATION
	// ============================================================

	if len(cfg.ExcludeDirs) == 0 {
		// Not critical, but user should know
		fmt.Println("⚠ Warning: exclude_dirs is empty - will scan all directories")
		fmt.Println("  Tip: Add common directories like .git, node_modules for better performance")
		warningCount++
	}

	// Check for duplicate exclude directories
	duplicateDirs := findDuplicates(cfg.ExcludeDirs)
	if len(duplicateDirs) > 0 {
		fmt.Printf("⚠ Warning: duplicate exclude_dirs found: %v\n", duplicateDirs)
		fmt.Println("  Tip: Edit config.json to remove duplicates")
		warningCount++
	}

	// ============================================================
	// FINAL SUMMARY
	// ============================================================

	// If there were warnings, show summary
	if warningCount > 0 {
		fmt.Printf("\n✓ Config loaded with %d warning(s)\n\n", warningCount)
	}

	return nil
}

// findDuplicates finds duplicate items in a string slice
// This helper function is used to detect duplicate extensions or directories
//
// Parameters:
//   - items: Slice of strings to check
//
// Returns:
//   - []string: List of duplicate items found
//
// Example:
//
//	findDuplicates([]string{"txt", "log", "txt", "csv"}) => ["txt"]
func findDuplicates(items []string) []string {
	// Map to track seen items
	seen := make(map[string]bool)

	// Map to track duplicates (using map to avoid duplicate duplicates!)
	duplicateMap := make(map[string]bool)

	// Check each item
	for _, item := range items {
		// Normalize: lowercase and trim whitespace
		cleanItem := strings.ToLower(strings.TrimSpace(item))

		// Skip empty items
		if cleanItem == "" {
			continue
		}

		// If we've seen this before, it's a duplicate
		if seen[cleanItem] {
			duplicateMap[cleanItem] = true
		} else {
			seen[cleanItem] = true
		}
	}

	// Convert map to slice for return
	var duplicates []string
	for dup := range duplicateMap {
		duplicates = append(duplicates, dup)
	}

	return duplicates
}

// findConflicts finds items that appear in both lists
// This is used to detect extensions that are in both whitelist and blacklist
//
// Parameters:
//   - list1: First list of items
//   - list2: Second list of items
//
// Returns:
//   - []string: Items that appear in both lists
//
// Example:
//
//	list1 := []string{"txt", "log", "csv"}
//	list2 := []string{"txt", "exe", "dll"}
//	findConflicts(list1, list2) => ["txt"]
func findConflicts(list1, list2 []string) []string {
	// Create map of items in first list
	inList1 := make(map[string]bool)
	for _, item := range list1 {
		cleanItem := strings.ToLower(strings.TrimSpace(item))
		if cleanItem != "" {
			inList1[cleanItem] = true
		}
	}

	// Check second list against first list
	conflictMap := make(map[string]bool)
	for _, item := range list2 {
		cleanItem := strings.ToLower(strings.TrimSpace(item))
		if cleanItem != "" && inList1[cleanItem] {
			conflictMap[cleanItem] = true
		}
	}

	// Convert map to slice
	var conflicts []string
	for conflict := range conflictMap {
		conflicts = append(conflicts, conflict)
	}

	return conflicts
}

// ValidatePath checks if a directory path exists and is accessible
// This is used to validate the scan path before starting
//
// Parameters:
//   - path: Directory path to validate
//
// Returns:
//   - error: Error if path doesn't exist or isn't a directory
//
// Example:
//
//	err := ValidatePath("/var/log")
//	if err != nil {
//	    log.Fatal(err)
//	}
func ValidatePath(path string) error {
	// Use os.Stat to check if path exists
	info, err := os.Stat(path)
	if err != nil {
		// Check if error is because path doesn't exist
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		// Other error (permission denied, etc.)
		return fmt.Errorf("error accessing path: %w", err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}
