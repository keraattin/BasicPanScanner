// Package config handles configuration loading and management
// This package is responsible for reading config.json and providing
// configuration data to other parts of the application
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration settings for the scanner
// This struct uses JSON tags to map JSON fields to Go struct fields
type Config struct {
	// ScanMode determines how extensions are interpreted
	// Valid values: "whitelist" or "blacklist"
	ScanMode string `json:"scan_mode"`

	// WhitelistExtensions contains file extensions to scan (when mode = "whitelist")
	// Example: ["txt", "log", "csv"]
	WhitelistExtensions []string `json:"whitelist_extensions"`

	// BlacklistExtensions contains file extensions to skip (when mode = "blacklist")
	// Example: ["exe", "dll", "jpg", "png"]
	BlacklistExtensions []string `json:"blacklist_extensions"`

	// ExcludeDirs contains directory names to skip during scanning
	// Example: [".git", "node_modules", "vendor"]
	ExcludeDirs []string `json:"exclude_dirs"`

	// MaxFileSize is the maximum file size to scan (e.g., "50MB")
	// Files larger than this will be skipped
	MaxFileSize string `json:"max_file_size"`
}

// Load reads and parses the configuration file
// It returns a Config struct or an error if loading/parsing fails
//
// Parameters:
//   - filename: Path to the configuration file (usually "config.json")
//
// Returns:
//   - *Config: Parsed configuration
//   - error: Error if file doesn't exist, is invalid, or fails validation
//
// Example:
//
//	cfg, err := config.Load("config.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load(filename string) (*Config, error) {
	// Read the entire file into memory
	data, err := os.ReadFile(filename)
	if err != nil {
		// Return descriptive error if file can't be read
		return nil, fmt.Errorf("could not read config file '%s': %w", filename, err)
	}

	// Check if file is empty (common mistake)
	if len(data) == 0 {
		return nil, fmt.Errorf("config file '%s' is empty", filename)
	}

	// Parse JSON into Config struct
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		// Return error if JSON is invalid
		return nil, fmt.Errorf("could not parse config (invalid JSON): %w", err)
	}

	// Validate the configuration
	// This ensures the config has valid values before using it
	err = Validate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetMaxFileSizeBytes converts the MaxFileSize string to bytes
// This function handles size suffixes like "MB", "GB", etc.
//
// Returns:
//   - int64: Size in bytes (0 means no limit)
//   - error: Error if format is invalid
//
// Example:
//
//	bytes, err := cfg.GetMaxFileSizeBytes()
//	"50MB" returns 52428800
func (c *Config) GetMaxFileSizeBytes() (int64, error) {
	return ParseFileSize(c.MaxFileSize)
}

// ParseFileSize converts human-readable size to bytes
// Supports: B, KB, MB, GB suffixes
//
// Parameters:
//   - sizeStr: Size string (e.g., "100MB", "1GB", "500KB")
//
// Returns:
//   - int64: Size in bytes (0 means no limit/empty string)
//   - error: Error if format is invalid
//
// Examples:
//
//	ParseFileSize("100MB")  => 104857600, nil
//	ParseFileSize("1GB")    => 1073741824, nil
//	ParseFileSize("")       => 0, nil (no limit)
//	ParseFileSize("100XB")  => 0, error
func ParseFileSize(sizeStr string) (int64, error) {
	// Normalize to uppercase and remove whitespace
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Empty string means no limit
	if sizeStr == "" {
		return 0, nil
	}

	// Define size suffixes with their multipliers
	// IMPORTANT: Check longest suffixes first to avoid partial matches
	// Example: "MB" should match before "B"
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"GB", 1024 * 1024 * 1024}, // Gigabyte
		{"MB", 1024 * 1024},        // Megabyte
		{"KB", 1024},               // Kilobyte
		{"B", 1},                   // Byte
	}

	// Try each suffix
	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			// Remove suffix to get the number part
			numStr := strings.TrimSuffix(sizeStr, s.suffix)
			numStr = strings.TrimSpace(numStr)

			// Parse the number
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size format '%s': cannot parse number", sizeStr)
			}

			// Validate number is not negative
			if num < 0 {
				return 0, fmt.Errorf("invalid size '%s': size cannot be negative", sizeStr)
			}

			// Zero means no limit
			if num == 0 {
				return 0, nil
			}

			// Calculate result
			result := num * s.multiplier

			// Check for overflow (result became negative due to overflow)
			if result < 0 {
				return 0, fmt.Errorf("invalid size '%s': value too large", sizeStr)
			}

			return result, nil
		}
	}

	// No valid suffix found
	return 0, fmt.Errorf("invalid size format '%s': must end with B, KB, MB, or GB", sizeStr)
}

// FormatBytes converts bytes to human-readable format
// This is useful for displaying file sizes to users
//
// Parameters:
//   - bytes: Size in bytes
//
// Returns:
//   - string: Human-readable size (e.g., "50.00 MB")
//
// Examples:
//
//	FormatBytes(1024)      => "1.00 KB"
//	FormatBytes(52428800)  => "50.00 MB"
//	FormatBytes(500)       => "500 B"
func FormatBytes(bytes int64) string {
	// Base unit for calculations (1 KB = 1024 bytes)
	const unit = 1024

	// If less than 1 KB, show in bytes
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	// Calculate which unit to use (KB, MB, GB, TB)
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Size names (corresponding to exponential levels)
	sizes := []string{"KB", "MB", "GB", "TB"}

	// Format with 2 decimal places
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), sizes[exp])
}

// GetActiveExtensions returns the extension list based on scan mode
// This helper function simplifies getting the correct extension list
//
// Returns:
//   - []string: Whitelist if mode is "whitelist", otherwise blacklist
//
// Example:
//
//	exts := cfg.GetActiveExtensions()
//	// If scan_mode = "whitelist", returns whitelist_extensions
//	// If scan_mode = "blacklist", returns blacklist_extensions
func (c *Config) GetActiveExtensions() []string {
	if c.ScanMode == "whitelist" {
		return c.WhitelistExtensions
	}
	return c.BlacklistExtensions
}

// NormalizeExtensions ensures all extensions have a dot prefix and are lowercase
// This prevents issues with extension matching
//
// Example:
//
//	Before: ["txt", "LOG", ".csv"]
//	After:  [".txt", ".log", ".csv"]
func (c *Config) NormalizeExtensions() {
	// Normalize whitelist
	for i := range c.WhitelistExtensions {
		ext := strings.TrimSpace(c.WhitelistExtensions[i])
		ext = strings.ToLower(ext)
		if ext != "" && !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		c.WhitelistExtensions[i] = ext
	}

	// Normalize blacklist
	for i := range c.BlacklistExtensions {
		ext := strings.TrimSpace(c.BlacklistExtensions[i])
		ext = strings.ToLower(ext)
		if ext != "" && !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		c.BlacklistExtensions[i] = ext
	}
}
