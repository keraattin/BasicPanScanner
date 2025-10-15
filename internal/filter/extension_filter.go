// Package filter handles file filtering based on extensions
// This package determines which files should be scanned based on
// whitelist/blacklist mode and extension lists
package filter

import (
	"path/filepath"
	"strings"
)

// ExtensionFilter provides file filtering based on extensions
// It supports two modes:
//   - Whitelist: Only scan files with specified extensions
//   - Blacklist: Scan all files EXCEPT those with specified extensions
type ExtensionFilter struct {
	// Mode determines filtering behavior ("whitelist" or "blacklist")
	mode string

	// Whitelist contains extensions to scan (when mode = "whitelist")
	whitelist map[string]bool

	// Blacklist contains extensions to skip (when mode = "blacklist")
	blacklist map[string]bool

	// Statistics
	totalChecked  int // Total files checked
	totalAccepted int // Files accepted for scanning
	totalRejected int // Files rejected
}

// NewExtensionFilter creates a new extension filter
//
// Parameters:
//   - mode: "whitelist" or "blacklist"
//   - whitelist: List of extensions for whitelist mode (e.g., [".txt", ".log"])
//   - blacklist: List of extensions for blacklist mode (e.g., [".exe", ".dll"])
//
// Returns:
//   - *ExtensionFilter: Configured filter ready to use
//
// Example:
//
//	filter := NewExtensionFilter("whitelist", []string{".txt", ".log"}, nil)
//	if filter.ShouldScan("document.txt") {
//	     Scan this file
//	}
func NewExtensionFilter(mode string, whitelist, blacklist []string) *ExtensionFilter {
	filter := &ExtensionFilter{
		mode:      mode,
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
	}

	// Convert whitelist slice to map for O(1) lookup
	// Map is faster than iterating through slice
	for _, ext := range whitelist {
		// Normalize: lowercase and ensure dot prefix
		cleanExt := normalizeExtension(ext)
		if cleanExt != "" {
			filter.whitelist[cleanExt] = true
		}
	}

	// Convert blacklist slice to map
	for _, ext := range blacklist {
		cleanExt := normalizeExtension(ext)
		if cleanExt != "" {
			filter.blacklist[cleanExt] = true
		}
	}

	return filter
}

// ShouldScan determines if a file should be scanned based on its extension
// This is the main function used throughout the application
//
// Parameters:
//   - filePath: Full path to the file
//
// Returns:
//   - bool: true if file should be scanned, false if it should be skipped
//
// Logic:
//   - Whitelist mode: Returns true ONLY if extension is in whitelist
//   - Blacklist mode: Returns true UNLESS extension is in blacklist
//   - Files without extensions: Scanned in blacklist mode, skipped in whitelist mode
//
// Example:
//
//	filter := NewExtensionFilter("whitelist", []string{".txt", ".log"}, nil)
//
//	filter.ShouldScan("file.txt")   // true  (in whitelist)
//	filter.ShouldScan("file.log")   // true  (in whitelist)
//	filter.ShouldScan("file.exe")   // false (not in whitelist)
//	filter.ShouldScan("README")     // false (no extension, whitelist mode)
func (ef *ExtensionFilter) ShouldScan(filePath string) bool {
	// Update statistics
	ef.totalChecked++

	// Extract file extension
	// filepath.Ext returns extension with dot (e.g., ".txt")
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext) // Normalize to lowercase

	// Handle files without extensions
	if ext == "" {
		// In blacklist mode: scan files without extensions
		// (they might be text files like README, Makefile, etc.)
		// In whitelist mode: skip files without extensions
		// (user explicitly listed what to scan)
		result := ef.mode == "blacklist"

		if result {
			ef.totalAccepted++
		} else {
			ef.totalRejected++
		}

		return result
	}

	// WHITELIST MODE: Only scan if extension is in whitelist
	if ef.mode == "whitelist" {
		// Check if extension exists in whitelist map
		// Map lookup is O(1) - very fast
		if ef.whitelist[ext] {
			ef.totalAccepted++
			return true
		}
		ef.totalRejected++
		return false
	}

	// BLACKLIST MODE: Scan everything EXCEPT blacklisted extensions
	if ef.mode == "blacklist" {
		// Check if extension exists in blacklist map
		if ef.blacklist[ext] {
			ef.totalRejected++
			return false // Skip this file
		}
		ef.totalAccepted++
		return true // Scan this file
	}

	// Default: don't scan (shouldn't reach here if mode is valid)
	ef.totalRejected++
	return false
}

// normalizeExtension ensures extension is lowercase and has dot prefix
//
// Parameters:
//   - ext: Extension string (e.g., "txt", ".txt", "TXT", ".TXT")
//
// Returns:
//   - string: Normalized extension (e.g., ".txt")
//
// Examples:
//
//	normalizeExtension("txt")   => ".txt"
//	normalizeExtension(".txt")  => ".txt"
//	normalizeExtension("TXT")   => ".txt"
//	normalizeExtension(".TXT")  => ".txt"
//	normalizeExtension("")      => ""
func normalizeExtension(ext string) string {
	// Remove whitespace
	ext = strings.TrimSpace(ext)

	// Empty string stays empty
	if ext == "" {
		return ""
	}

	// Convert to lowercase for consistent comparison
	ext = strings.ToLower(ext)

	// Add dot prefix if missing
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return ext
}

// GetStats returns filtering statistics
// Useful for reporting and monitoring
//
// Returns:
//   - totalChecked: Total files evaluated
//   - totalAccepted: Files accepted for scanning
//   - totalRejected: Files rejected (skipped)
//
// Example:
//
//	checked, accepted, rejected := filter.GetStats()
//	fmt.Printf("Checked: %d, Accepted: %d, Rejected: %d\n",
//	           checked, accepted, rejected)
func (ef *ExtensionFilter) GetStats() (totalChecked, totalAccepted, totalRejected int) {
	return ef.totalChecked, ef.totalAccepted, ef.totalRejected
}

// ResetStats resets all statistics counters
// Useful when starting a new scan
func (ef *ExtensionFilter) ResetStats() {
	ef.totalChecked = 0
	ef.totalAccepted = 0
	ef.totalRejected = 0
}

// GetMode returns the current filtering mode
func (ef *ExtensionFilter) GetMode() string {
	return ef.mode
}

// GetWhitelistCount returns the number of whitelisted extensions
func (ef *ExtensionFilter) GetWhitelistCount() int {
	return len(ef.whitelist)
}

// GetBlacklistCount returns the number of blacklisted extensions
func (ef *ExtensionFilter) GetBlacklistCount() int {
	return len(ef.blacklist)
}

// IsWhitelisted checks if an extension is in the whitelist
// Useful for debugging and reporting
func (ef *ExtensionFilter) IsWhitelisted(ext string) bool {
	cleanExt := normalizeExtension(ext)
	return ef.whitelist[cleanExt]
}

// IsBlacklisted checks if an extension is in the blacklist
// Useful for debugging and reporting
func (ef *ExtensionFilter) IsBlacklisted(ext string) bool {
	cleanExt := normalizeExtension(ext)
	return ef.blacklist[cleanExt]
}

// DirectoryFilter handles directory exclusion logic
// This is separate from extension filtering
type DirectoryFilter struct {
	excludeDirs map[string]bool // Directories to skip
}

// NewDirectoryFilter creates a new directory filter
//
// Parameters:
//   - excludeDirs: List of directory names to exclude (e.g., [".git", "node_modules"])
//
// Returns:
//   - *DirectoryFilter: Configured directory filter
//
// Example:
//
//	dirFilter := NewDirectoryFilter([]string{".git", "node_modules"})
//	if dirFilter.ShouldSkip("/path/to/.git") {
//	     Skip this directory
//	}
func NewDirectoryFilter(excludeDirs []string) *DirectoryFilter {
	filter := &DirectoryFilter{
		excludeDirs: make(map[string]bool),
	}

	// Convert slice to map for fast lookup
	for _, dir := range excludeDirs {
		cleanDir := strings.TrimSpace(dir)
		if cleanDir != "" {
			filter.excludeDirs[cleanDir] = true
		}
	}

	return filter
}

// ShouldSkip determines if a directory should be skipped
//
// Parameters:
//   - dirPath: Full path to the directory
//
// Returns:
//   - bool: true if directory should be skipped, false if it should be scanned
//
// Example:
//
//	dirFilter := NewDirectoryFilter([]string{".git", "node_modules"})
//
//	dirFilter.ShouldSkip("/project/.git")           // true
//	dirFilter.ShouldSkip("/project/node_modules")   // true
//	dirFilter.ShouldSkip("/project/src")            // false
func (df *DirectoryFilter) ShouldSkip(dirPath string) bool {
	// Extract just the directory name (last part of path)
	// Example: "/home/user/project/.git" -> ".git"
	dirName := filepath.Base(dirPath)

	// Check if this directory name is in the exclusion list
	return df.excludeDirs[dirName]
}

// GetExcludeCount returns the number of excluded directories
func (df *DirectoryFilter) GetExcludeCount() int {
	return len(df.excludeDirs)
}

// IsExcluded checks if a directory name is excluded
func (df *DirectoryFilter) IsExcluded(dirName string) bool {
	return df.excludeDirs[dirName]
}
