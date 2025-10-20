// BasicPanScanner - PCI Compliance Scanner
// Main application entry point
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"../../internal/config"
	"../../internal/filter"
	"../../internal/report"
	"../../internal/scanner"
	"../../internal/ui"
)

// Application version
const Version = "3.0.0"

func main() {
	// ============================================================
	// STEP 1: Parse command line flags
	// ============================================================

	pathFlag := flag.String("path", "", "Directory to scan")
	outputFlag := flag.String("output", "", "Output file")
	modeFlag := flag.String("mode", "", "Scan mode: 'whitelist' or 'blacklist' (overrides config)")
	extensionsFlag := flag.String("ext", "", "Extensions (e.g., txt,log,csv)")
	excludeFlag := flag.String("exclude", "", "Exclude dirs (e.g., .git,vendor)")
	workersFlag := flag.Int("workers", 0, "Number of concurrent workers (default: CPU/2)")
	helpFlag := flag.Bool("help", false, "Show help")

	flag.Parse()

	// Show help if requested or no arguments
	if *helpFlag || len(os.Args) == 1 {
		ui.ShowHelp()
		return
	}

	// Path is required
	if *pathFlag == "" {
		fmt.Println("Error: -path is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// ============================================================
	// STEP 2: Display banner
	// ============================================================

	ui.ShowBanner(Version)

	// ============================================================
	// STEP 4: Load configuration
	// ============================================================

	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("Warning: Could not load config.json: %v\n", err)
		fmt.Println("Using default settings\n")

		// Use default config
		cfg = &config.Config{
			ScanMode:            "blacklist",
			WhitelistExtensions: []string{},
			BlacklistExtensions: []string{".exe", ".dll", ".so", ".jpg", ".png", ".mp4"},
			ExcludeDirs:         []string{".git", "node_modules"},
			MaxFileSize:         "50MB",
		}
	}

	// ============================================================
	// STEP 5: Apply CLI overrides
	// ============================================================

	// Start with config values
	scanMode := cfg.ScanMode
	whitelistExts := cfg.WhitelistExtensions
	blacklistExts := cfg.BlacklistExtensions
	excludeDirs := cfg.ExcludeDirs

	// Override mode if specified
	if *modeFlag != "" {
		validModes := map[string]bool{"whitelist": true, "blacklist": true}
		if !validModes[*modeFlag] {
			fmt.Printf("Error: invalid mode '%s', must be 'whitelist' or 'blacklist'\n", *modeFlag)
			os.Exit(1)
		}
		scanMode = *modeFlag
		fmt.Printf("✓ Mode override from CLI: %s\n", scanMode)
	}

	// Override extensions if specified
	if *extensionsFlag != "" {
		extensions := strings.Split(*extensionsFlag, ",")
		for i := range extensions {
			extensions[i] = strings.TrimSpace(extensions[i])
		}

		// Apply to the active mode
		if scanMode == "whitelist" {
			whitelistExts = extensions
			fmt.Printf("✓ Whitelist override from CLI: %d extensions\n", len(extensions))
		} else {
			blacklistExts = extensions
			fmt.Printf("✓ Blacklist override from CLI: %d extensions\n", len(extensions))
		}
	}

	// Override exclude dirs if specified
	if *excludeFlag != "" {
		excludeDirs = strings.Split(*excludeFlag, ",")
		for i := range excludeDirs {
			excludeDirs[i] = strings.TrimSpace(excludeDirs[i])
		}
	}

	// Normalize extensions (add dots, lowercase)
	cfg.NormalizeExtensions()

	// ============================================================
	// STEP 6: Determine number of workers
	// ============================================================

	numCPU := runtime.NumCPU()
	workers := *workersFlag

	if workers == 0 {
		// Default: Half of CPU cores (minimum 1)
		workers = numCPU / 2
		if workers < 1 {
			workers = 1
		}
	}

	// Validate worker count
	if workers < 1 {
		fmt.Println("Error: workers must be at least 1")
		os.Exit(1)
	}

	if workers > numCPU {
		fmt.Printf("Warning: workers (%d) exceeds CPU cores (%d), limiting to %d\n", workers, numCPU, numCPU)
		workers = numCPU
	}

	// ============================================================
	// STEP 7: Validate directory
	// ============================================================

	err = config.ValidatePath(*pathFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 8: Create filters
	// ============================================================

	// Extension filter (whitelist or blacklist)
	extFilter := filter.NewExtensionFilter(scanMode, whitelistExts, blacklistExts)

	// Directory filter (always applied)
	dirFilter := filter.NewDirectoryFilter(excludeDirs)

	// ============================================================
	// STEP 9: Parse max file size
	// ============================================================

	maxFileSize, err := cfg.GetMaxFileSizeBytes()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// ============================================================
	// STEP 10: Create scanner
	// ============================================================

	// Progress callback
	progressTracker := ui.NewProgressTracker()

	scannerConfig := &scanner.Config{
		ExtFilter:   extFilter,
		DirFilter:   dirFilter,
		MaxFileSize: maxFileSize,
		Workers:     workers,
		ProgressCallback: func(scanned, total, cards int) {
			progressTracker.Update(scanned, total, cards)
		},
	}

	s := scanner.NewScanner(scannerConfig)

	// ============================================================
	// STEP 11: Show scan information
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
	// STEP 12: Run the scan!
	// ============================================================

	progressTracker.Start()

	result, err := s.ScanDirectory(*pathFlag)
	if err != nil {
		fmt.Printf("\nScan failed: %v\n", err)
		os.Exit(1)
	}

	progressTracker.Finish()

	// ============================================================
	// STEP 13: Show summary
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
	// STEP 14: Generate and export report (if requested)
	// ============================================================

	if *outputFlag != "" {
		// Determine which extensions list to show in report
		var reportExtensions []string
		if scanMode == "whitelist" {
			reportExtensions = whitelistExts
		} else {
			reportExtensions = []string{"all except blacklist"}
		}

		// Create report
		rep := report.NewReport(
			Version,
			*pathFlag,
			scanMode,
			reportExtensions,
			result,
		)

		// Export report
		err = rep.Export(*outputFlag)
		if err != nil {
			fmt.Printf("\n  Error saving report: %v\n", err)
			os.Exit(1)
		}

		ui.ShowExportSuccess(*outputFlag)
	}

	// ============================================================
	// STEP 15: Exit with appropriate code
	// ============================================================

	// Exit with code 0 (success)
	// Even if cards were found, the scan itself succeeded
	os.Exit(0)
}
