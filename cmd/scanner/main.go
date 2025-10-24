// BasicPanScanner v3.1.1 - PCI Compliance Scanner
// Main application entry point with BIN Database support
//
// INITIALIZATION ORDER:
//  1. Parse CLI flags
//  2. Show banner
//  3. Load configuration
//  4. Initialize BIN database (CRITICAL)
//  5. Create filters
//  6. Create scanner
//  7. Run scan
//  8. Generate report
//
// Author: BasicPanScanner Contributors
// License: MIT
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"../../internal/config"
	"../../internal/detector"
	"../../internal/filter"
	"../../internal/report"
	"../../internal/scanner"
	"../../internal/ui"
)

// Version is the application version
// Update this for each release
const Version = "3.0.0"

func main() {
	// ============================================================
	// STEP 1: Parse command line flags
	// ============================================================
	// Define and parse CLI arguments
	// All flags are optional except -path

	pathFlag := flag.String("path", "", "Directory or file to scan (required)")
	outputFlag := flag.String("output", "", "Output file path (.json, .csv, .html, .txt, .xml)")
	modeFlag := flag.String("mode", "", "Scan mode: 'whitelist' or 'blacklist' (overrides config.json)")
	extensionsFlag := flag.String("ext", "", "File extensions to process (comma-separated, e.g., txt,log,csv)")
	excludeFlag := flag.String("exclude", "", "Directories to exclude (comma-separated, e.g., .git,vendor)")
	workersFlag := flag.Int("workers", 0, "Number of concurrent workers (default: CPU cores / 2)")
	helpFlag := flag.Bool("help", false, "Show help information")

	flag.Parse()

	// Show help if requested or no arguments provided
	if *helpFlag || len(os.Args) == 1 {
		ui.ShowHelp()
		return
	}

	// Validate required path flag
	if *pathFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -path flag is required")
		fmt.Fprintln(os.Stderr, "Use -help for usage information")
		os.Exit(1)
	}

	// ============================================================
	// STEP 2: Display application banner
	// ============================================================

	ui.ShowBanner(Version)

	// ============================================================
	// STEP 3: Initialize BIN Database (CRITICAL)
	// ============================================================
	// This must be done BEFORE loading config and creating scanner
	// The scanner depends on the BIN database for card type detection

	fmt.Println("Initializing BIN database...")

	// Initialize the global BIN database
	// Empty string means use default path: internal/detector/bindata/bin_ranges.json
	err := detector.InitGlobalBINDatabase("")
	if err != nil {
		// CRITICAL ERROR: Cannot proceed without BIN database
		fmt.Fprintf(os.Stderr, "\n✗ CRITICAL ERROR: Failed to initialize BIN database\n")
		fmt.Fprintf(os.Stderr, "  Error: %v\n\n", err)
		fmt.Fprintln(os.Stderr, "  This error means the BIN database file could not be loaded.")
		fmt.Fprintln(os.Stderr, "  The scanner cannot detect card types without this database.")
		fmt.Fprintln(os.Stderr, "\n  Troubleshooting steps:")
		fmt.Fprintln(os.Stderr, "  1. Verify file exists: internal/detector/bindata/bin_ranges.json")
		fmt.Fprintln(os.Stderr, "  2. Check file permissions (must be readable)")
		fmt.Fprintln(os.Stderr, "  3. Validate JSON syntax using a JSON validator")
		fmt.Fprintln(os.Stderr, "  4. Ensure file is not corrupted")
		os.Exit(1)
	}

	// Database loaded successfully - show info
	db, _ := detector.GetGlobalBINDatabase()
	fmt.Printf("✓ BIN Database v%s loaded successfully\n", db.GetVersion())
	fmt.Printf("  Last updated: %s\n", db.GetLastUpdated())
	fmt.Printf("  Supporting %d card issuers\n", db.GetIssuerCount())
	fmt.Println()

	// ============================================================
	// STEP 4: Load configuration from config.json
	// ============================================================
	// Configuration is optional - use defaults if file missing

	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("⚠ Warning: Could not load config.json: %v\n", err)
		fmt.Println("  Using default configuration\n")

		// Fallback to default configuration
		cfg = &config.Config{
			ScanMode:            "blacklist",
			WhitelistExtensions: []string{},
			BlacklistExtensions: []string{
				".exe", ".dll", ".so", ".dylib", // Binaries
				".jpg", ".png", ".gif", ".mp4", // Media
				".zip", ".tar", ".gz", ".rar", // Archives
			},
			ExcludeDirs: []string{
				".git", "node_modules", "vendor", // Common dev dirs
			},
			MaxFileSize: "50MB",
		}
	}

	// ============================================================
	// STEP 5: Apply CLI overrides to configuration
	// ============================================================
	// CLI flags take precedence over config.json

	// Start with config file values
	scanMode := cfg.ScanMode
	whitelistExts := cfg.WhitelistExtensions
	blacklistExts := cfg.BlacklistExtensions
	excludeDirs := cfg.ExcludeDirs

	// Override scan mode if specified via CLI
	if *modeFlag != "" {
		if *modeFlag != "whitelist" && *modeFlag != "blacklist" {
			fmt.Fprintf(os.Stderr, "Error: invalid mode '%s', must be 'whitelist' or 'blacklist'\n", *modeFlag)
			os.Exit(1)
		}
		scanMode = *modeFlag
		fmt.Printf("✓ Scan mode overridden via CLI: %s\n", scanMode)
	}

	// Override extensions if specified via CLI
	if *extensionsFlag != "" {
		extensions := strings.Split(*extensionsFlag, ",")
		// Trim whitespace from each extension
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		// Apply to active scan mode
		if scanMode == "whitelist" {
			whitelistExts = extensions
			fmt.Printf("✓ Whitelist extensions overridden via CLI: %d extensions\n", len(extensions))
		} else {
			blacklistExts = extensions
			fmt.Printf("✓ Blacklist extensions overridden via CLI: %d extensions\n", len(extensions))
		}
	}

	// Override exclude directories if specified via CLI
	if *excludeFlag != "" {
		excludeDirs = strings.Split(*excludeFlag, ",")
		// Trim whitespace from each directory name
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
		fmt.Printf("✓ Exclude directories overridden via CLI: %d directories\n", len(excludeDirs))
	}

	// Normalize all extensions (add dots, convert to lowercase)
	cfg.NormalizeExtensions()

	// ============================================================
	// STEP 6: Determine optimal worker count
	// ============================================================
	// Balance between performance and resource usage

	numCPU := runtime.NumCPU()
	workers := *workersFlag

	// If not specified, use half of CPU cores (minimum 1)
	if workers == 0 {
		workers = numCPU / 2
		if workers < 1 {
			workers = 1
		}
	}

	// Validate worker count
	if workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be at least 1")
		os.Exit(1)
	}

	// Warn if worker count exceeds CPU cores
	if workers > numCPU {
		fmt.Printf("⚠ Warning: workers (%d) exceeds CPU cores (%d), limiting to %d\n",
			workers, numCPU, numCPU)
		workers = numCPU
	}

	// ============================================================
	// STEP 7: Validate target path
	// ============================================================
	// Ensure the path exists and is accessible

	err = config.ValidatePath(*pathFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 8: Create file and directory filters
	// ============================================================

	// Extension filter (whitelist or blacklist mode)
	extFilter := filter.NewExtensionFilter(scanMode, whitelistExts, blacklistExts)

	// Directory filter (always applied)
	dirFilter := filter.NewDirectoryFilter(excludeDirs)

	// ============================================================
	// STEP 9: Parse maximum file size
	// ============================================================

	maxFileSize, err := cfg.GetMaxFileSizeBytes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 10: Create scanner with configuration
	// ============================================================

	// Progress tracker for UI updates
	progressTracker := ui.NewProgressTracker()

	scannerConfig := &scanner.Config{
		ExtFilter:   extFilter,
		DirFilter:   dirFilter,
		MaxFileSize: maxFileSize,
		Workers:     workers,
		// Progress callback for real-time updates
		ProgressCallback: func(scanned, total, cards int) {
			progressTracker.Update(scanned, total, cards)
		},
	}

	s := scanner.NewScanner(scannerConfig)

	// ============================================================
	// STEP 11: Display scan configuration
	// ============================================================

	var extensionCount int
	if scanMode == "whitelist" {
		extensionCount = len(whitelistExts)
	} else {
		extensionCount = len(blacklistExts)
	}

	maxSizeStr := config.FormatBytes(maxFileSize)
	if maxFileSize == 0 {
		maxSizeStr = "unlimited"
	}

	ui.ShowScanInfo(*pathFlag, scanMode, extensionCount, workers, maxSizeStr)

	// ============================================================
	// STEP 12: Execute the scan
	// ============================================================

	var result *scanner.ScanResult

	fmt.Println("Starting scan...")
	result, err = s.ScanDirectory(*pathFlag)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ Scan failed: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 13: Display results
	// ============================================================

	ui.ShowSummary(
		result.Duration,
		result.TotalFiles,
		result.ScannedFiles,
		result.SkippedBySize,
		result.SkippedByExt,
		result.CardsFound,
		result.ScanRate,
	)

	// ============================================================
	// STEP 14: Export report if output file specified
	// ============================================================

	if *outputFlag != "" {
		// Determine which extensions list to show in report
		var reportExtensions []string
		if scanMode == "whitelist" {
			reportExtensions = whitelistExts
		} else {
			reportExtensions = blacklistExts
		}

		// Create report instance
		rep := report.NewReport(
			Version,
			*pathFlag,
			scanMode,
			reportExtensions,
			result,
		)

		// Generate and save report
		// Format is determined automatically from file extension
		fmt.Printf("\nGenerating report...\n")
		err = rep.Export(*outputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ Failed to generate report: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Report saved: %s\n", *outputFlag)
	}

	// ============================================================
	// STEP 15: Exit with appropriate code
	// ============================================================

	// Exit with error code if cards were found (for CI/CD integration)
	if result.CardsFound > 0 {
		os.Exit(2) // Exit code 2 indicates cards were found
	}

	// Success - no cards found
	os.Exit(0)
}
