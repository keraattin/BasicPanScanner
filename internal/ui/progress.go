// Package ui - Progress display
// This file handles progress indicators during scanning
package ui

import (
	"fmt"
	"strings"
	"time"
)

// ProgressTracker tracks and displays scan progress
// This provides real-time feedback to the user
type ProgressTracker struct {
	startTime    time.Time
	lastUpdate   time.Time
	totalFiles   int
	scannedFiles int
	cardsFound   int
}

// NewProgressTracker creates a new progress tracker
//
// Returns:
//   - *ProgressTracker: New progress tracker
//
// Example:
//
//	tracker := ui.NewProgressTracker()
//	tracker.Start()
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		lastUpdate: time.Now(),
	}
}

// Start marks the beginning of scanning
// This records the start time for duration calculations
func (pt *ProgressTracker) Start() {
	pt.startTime = time.Now()
}

// Update updates the progress display
// This is called after each file is scanned
//
// Parameters:
//   - scannedFiles: Number of files scanned so far
//   - totalFiles: Total number of files to scan
//   - cardsFound: Total cards found so far
//
// Example:
//
//	tracker.Update(50, 100, 5)  // Shows: [Scanned: 50/100 | Cards: 5]
func (pt *ProgressTracker) Update(scannedFiles, totalFiles, cardsFound int) {
	pt.scannedFiles = scannedFiles
	pt.totalFiles = totalFiles
	pt.cardsFound = cardsFound

	// Only update display every 100ms to avoid flickering
	if time.Since(pt.lastUpdate) < 100*time.Millisecond {
		return
	}
	pt.lastUpdate = time.Now()

	// Clear previous line and write new progress
	fmt.Printf("\r[Scanned: %d/%d | Cards: %d]", scannedFiles, totalFiles, cardsFound)
}

// Finish completes the progress tracking and shows final summary
// This clears the progress line and shows completion message
func (pt *ProgressTracker) Finish() {
	// Clear the progress line
	fmt.Print("\r" + strings.Repeat(" ", 60) + "\r")
}

// ShowScanInfo displays scan configuration before starting
// This helps users verify what will be scanned
//
// Parameters:
//   - directory: Directory being scanned
//   - mode: Scan mode ("whitelist" or "blacklist")
//   - extensions: Number of extensions in active list
//   - workers: Number of worker goroutines
//   - maxSize: Maximum file size (formatted string like "50.00 MB")
//
// Example:
//
//	ui.ShowScanInfo("/var/log", "blacklist", 80, 2, "50.00 MB")
func ShowScanInfo(directory, mode string, extensions, workers int, maxSize string) {
	fmt.Printf("\nScanning directory: %s\n", directory)
	fmt.Printf("Scan mode: %s\n", mode)

	if mode == "whitelist" {
		fmt.Printf("Whitelist: %d extensions (scanning only these)\n", extensions)
	} else {
		fmt.Printf("Blacklist: %d extensions (scanning everything except these)\n", extensions)
	}

	fmt.Printf("Workers: %d", workers)
	if workers > 1 {
		fmt.Printf(" (concurrent scanning enabled)")
	} else {
		fmt.Printf(" (single-threaded mode)")
	}
	fmt.Println()

	if maxSize != "" {
		fmt.Printf("Max file size: %s\n", maxSize)
	} else {
		fmt.Printf("Max file size: unlimited\n")
	}

	fmt.Println(strings.Repeat("=", 60))
}

// ShowSummary displays the final scan summary
// This shows the complete results after scanning finishes
//
// Parameters:
//   - duration: Total scan duration
//   - totalFiles: Total files found
//   - scannedFiles: Files actually scanned
//   - skippedBySize: Files skipped due to size
//   - skippedByExt: Files skipped by extension filter
//   - cardsFound: Total cards found
//   - scanRate: Files per second
//
// Example:
//
//	ui.ShowSummary(time.Minute, 1000, 800, 20, 180, 15, 13.3)
func ShowSummary(duration time.Duration, totalFiles, scannedFiles, skippedBySize, skippedByExt, cardsFound int, scanRate float64) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("✓ Scan complete!\n")
	fmt.Printf("  Time: %s\n", formatDuration(duration)) // Use formatted duration
	fmt.Printf("  Total files: %d\n", totalFiles)
	fmt.Printf("  Scanned: %d\n", scannedFiles)

	if skippedBySize > 0 {
		fmt.Printf("  Skipped (size): %d\n", skippedBySize)
	}

	if skippedByExt > 0 {
		fmt.Printf("  Skipped (extension): %d\n", skippedByExt)
	}

	fmt.Printf("  Cards found: %d\n", cardsFound)

	if scanRate > 0 {
		fmt.Printf("  Scan rate: %.1f files/second\n", scanRate)
	}
}

// formatDuration formats a duration in a human-readable way
// This is a local helper function for UI display
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Formatted duration (e.g., "1h 23m", "2m 15s", "3.5s", "250ms")
func formatDuration(d time.Duration) string {
	// Handle very short durations
	if d < time.Millisecond {
		// Less than 1ms, show microseconds
		us := float64(d.Microseconds())
		if us < 10 {
			return fmt.Sprintf("%.1fµs", us)
		}
		return fmt.Sprintf("%dµs", d.Microseconds())
	}

	if d < time.Second {
		// Less than 1 second, show milliseconds
		ms := float64(d.Milliseconds())
		if ms < 10 {
			return fmt.Sprintf("%.1fms", d.Seconds()*1000)
		}
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		// Less than 1 minute, show seconds
		s := d.Seconds()
		if s < 10 {
			return fmt.Sprintf("%.1fs", s)
		}
		return fmt.Sprintf("%ds", int(s))
	}

	if d < time.Hour {
		// Less than 1 hour, show minutes and seconds
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	// 1 hour or more, show hours, minutes, and seconds
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if minutes == 0 && seconds == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	if seconds == 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

// ShowFileFound displays a message when cards are found in a file
// This provides immediate feedback during scanning
//
// Parameters:
//   - filename: Name of the file (just the basename, not full path)
//   - count: Number of cards found in this file
//
// Example:
//
//	ui.ShowFileFound("app.log", 3)
//	 Output: ✓ Found 3 cards in: app.log
func ShowFileFound(filename string, count int) {
	fmt.Printf("✓ Found %d cards in: %s\n", count, filename)
}

// ShowExportSuccess displays a message when report is exported successfully
//
// Parameters:
//   - filename: Name of the exported file
//
// Example:
//
//	ui.ShowExportSuccess("report.html")
//	 Output:
//	   ✓ Saved: report.html
func ShowExportSuccess(filename string) {
	fmt.Printf("\n  ✓ Saved: %s\n", filename)
}
