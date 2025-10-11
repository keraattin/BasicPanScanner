package main

import (
	"bufio" // buffered I/O for reading input
	"flag"
	"fmt" // for printing
	"os"  // operating system stuff (Stdin)
	"path/filepath"
	"strings"
	"time"
)

// validateDirectory checks if the path exists and is a directory
func validateDirectory(dirPath string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dirPath)
		}
		return fmt.Errorf("error accessing directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	return nil
}

// findCardNumber searches for credit card patterns (13-19 consecutive digits)
// Different card types have different lengths:
// - Visa: 13, 16, or 19 digits
// - MasterCard: 16 digits
// - Amex: 15 digits
// - Discover: 16 digits
// - Diners: 14 digits
// It handles common formats like spaces and dashes between digit groups
func findCardNumber(text string) string {
	consecutiveDigits := ""

	// Iterate through each character in the text
	for i := 0; i < len(text); i++ {
		char := text[i]

		if char >= '0' && char <= '9' {
			// Found a digit, add it to our collection
			consecutiveDigits = consecutiveDigits + string(char)

			// Check if we have a valid card length (13-19 digits)
			length := len(consecutiveDigits)

			// If we have between 13-19 digits, check what comes next
			if length >= 13 && length <= 19 {
				// Look ahead to see if there are more digits
				if i+1 < len(text) && text[i+1] >= '0' && text[i+1] <= '9' {
					// More digits coming, keep collecting
					continue
				}

				// No more digits, return what we have if it's valid length
				return consecutiveDigits
			}

		} else if char == ' ' || char == '-' {
			// Space or dash - could be formatting in a card number
			// Only reset if we haven't started collecting digits
			if len(consecutiveDigits) == 0 {
				consecutiveDigits = ""
			}
			// Otherwise, continue collecting (skip the separator)

		} else {
			// Any other character breaks the sequence
			// Check if we had a valid length before resetting
			length := len(consecutiveDigits)
			if length >= 13 && length <= 19 {
				return consecutiveDigits
			}

			// Reset for next potential card
			consecutiveDigits = ""
		}
	}

	// Check final collection at end of string
	length := len(consecutiveDigits)
	if length >= 13 && length <= 19 {
		return consecutiveDigits
	}

	// No valid card number found
	return ""
}

// scanDirectory
func scanDirectoryWithOptions(dirPath string, outputFile string, silent bool, extensions []string) error {

	if !silent {
		fmt.Printf("\nScanning directory: %s\n", dirPath)
		fmt.Println(strings.Repeat("=", 50))
	}

	fmt.Printf("\nScanning directory: %s\n", dirPath)
	fmt.Println(strings.Repeat("=", 50))

	// Single pass - no pre-counting
	startTime := time.Now()
	totalFiles := 0
	scannedFiles := 0
	foundCards := 0

	// Progress indicator without knowing total
	lastUpdate := time.Now()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		totalFiles++

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))

		// Only scan text-like files
		if ext == ".txt" || ext == ".log" || ext == ".csv" {
			scannedFiles++

			// Update progress every 100ms (not every file)
			if time.Since(lastUpdate) > 100*time.Millisecond {
				fmt.Printf("\r[Scanning... Files checked: %d, Scanned: %d, Cards found: %d]",
					totalFiles, scannedFiles, foundCards)
				lastUpdate = time.Now()
			}

			// Scan the file
			cardsFound := scanFileWithCount(path)
			if cardsFound > 0 {
				foundCards += cardsFound
				// Show files with findings immediately
				fmt.Printf("\n✓ Found %d cards in: %s\n", cardsFound, filepath.Base(path))
			}
		}

		return nil
	})

	// Clear the progress line
	fmt.Print("\r" + strings.Repeat(" ", 70) + "\r")

	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	elapsed := time.Since(startTime)

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("✓ Directory scan complete!\n")
	fmt.Printf("  Time taken: %s\n", elapsed.Round(time.Second))
	fmt.Printf("  Total files: %d\n", totalFiles)
	fmt.Printf("  Files scanned: %d\n", scannedFiles)
	fmt.Printf("  Cards found: %d\n", foundCards)

	if elapsed.Seconds() > 0 {
		rate := float64(scannedFiles) / elapsed.Seconds()
		fmt.Printf("  Scan rate: %.1f files/second\n", rate)
	}

	return nil
}

// scanFileWithCount scans a file and returns number of valid cards found
func scanFileWithCount(filepath string) int {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
		return 0
	}
	defer file.Close()

	// Create a scanner for line-by-line reading
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	validCount := 0

	// Process each line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check current line for card patterns
		cardNumber := findCardNumber(line)
		if cardNumber != "" {
			// Validate with Luhn algorithm
			if validateLuhn(cardNumber) {
				validCount++

				// Get card type
				cardType := getCardType(cardNumber)

				// Mask the card number for safe display
				maskedCard := maskCardNumber(cardNumber)

				fmt.Printf("    Line %d: %s card: %s ✓\n", lineNumber, cardType, maskedCard)
			}
		}
	}

	// Only show summary if cards were found
	if validCount > 0 {
		fmt.Printf("    → Found %d valid cards\n", validCount)
	}

	return validCount
}

//How Luhn Algorithm Works:
// - Start from the rightmost digit
// - Double every second digit (from right)
// - If doubled value > 9, subtract 9
// - Sum all digits
// - Valid if sum is divisible by 10

// validate Luhn checks if a card number passes the Luhn algorithm
// This is a checksum formula used to validate credit card numbers
func validateLuhn(cardNumber string) bool {
	// Remove any spaces or dashes that might still be in the number
	cleaned := ""
	for i := 0; i < len(cardNumber); i++ {
		if cardNumber[i] >= '0' && cardNumber[i] <= '9' {
			cleaned += string(cardNumber[i])
		}
	}

	// Need at least 13 digits for a valid card (some cards are 13-19 digits)
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	sum := 0
	isEven := false

	// Process digits from right to left
	for i := len(cleaned) - 1; i >= 0; i-- {
		// Convert character to integer
		digit := int(cleaned[i] - '0')

		// Every second digit (from right) is doubled
		if isEven {
			digit *= 2
			// If doubled digit is > 9, subtract 9
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	// Valid if sum is divisible by 10
	return sum%10 == 0
}

// getCardType identifies the card issuer based on the card number
func getCardType(cardNumber string) string {
	length := len(cardNumber)

	// Check length is valid for cards
	if length < 13 || length > 19 {
		return "Unknown"
	}

	// Check first digits for card type
	firstDigit := cardNumber[0]
	firstTwo := cardNumber[0:2]
	firstFour := ""
	if len(cardNumber) >= 4 {
		firstFour = cardNumber[0:4]
	}

	// Visa: Starts with 4 (13, 16, or 19 digits)
	if firstDigit == '4' {
		if length == 13 || length == 16 || length == 19 {
			return "Visa"
		}
		return "Unknown" // Wrong length for Visa
	}

	// MasterCard: Starts with 51-55 or 2221-2720 (16 digits only)
	if (firstTwo >= "51" && firstTwo <= "55") ||
		(firstFour >= "2221" && firstFour <= "2720") {
		if length == 16 {
			return "MasterCard"
		}
		return "Unknown" // Wrong length for MasterCard
	}

	// American Express: Starts with 34 or 37 (15 digits only)
	if firstTwo == "34" || firstTwo == "37" {
		if length == 15 {
			return "Amex"
		}
		return "Unknown" // Wrong length for Amex
	}

	// Discover: Starts with 6011, 644-649, 65 (16 digits only)
	if firstFour == "6011" ||
		(firstTwo >= "64" && firstTwo <= "65") {
		if length == 16 {
			return "Discover"
		}
		return "Unknown"
	}

	// Diners Club: Starts with 36, 38, or 300-305 (14 digits)
	if firstTwo == "36" || firstTwo == "38" ||
		(firstFour >= "3000" && firstFour <= "3059") {
		if length == 14 {
			return "Diners"
		}
		return "Unknown"
	}

	// JCB: Starts with 3528-3589 (16 digits)
	if firstFour >= "3528" && firstFour <= "3589" {
		if length == 16 {
			return "JCB"
		}
		return "Unknown"
	}

	return "Unknown"
}

// maskCardNumber returns a masked version for safe display
// PCI compliance: Shows only first 6 (BIN) and last 4 digits
func maskCardNumber(cardNumber string) string {
	length := len(cardNumber)

	// If card number is too short, just return it
	if length <= 10 {
		return cardNumber
	}

	// Build masked number
	masked := ""

	// Add first 6 digits (BIN - Bank Identification Number)
	masked += cardNumber[0:6]

	// Add asterisks for middle digits
	middleDigits := length - 10 // Total minus first 6 and last 4
	for i := 0; i < middleDigits; i++ {
		masked += "*"
	}

	// Add last 4 digits
	masked += cardNumber[length-4:]

	return masked
}

// showHelp displays usage information
func showHelp() {
	fmt.Println(`
BasicPanScanner v1.0.0 - PCI Compliance Scanner
Usage: ./scanner -path <directory> [options]

Required:
    -path <directory>   Directory to scan

Options:
    -output <file>      Export results to CSV file
    -silent            Silent mode (minimal output)
    -ext <list>        File extensions to scan (default: txt,log,csv)
    -help              Show this help message

Examples:
    ./scanner -path /var/log
    ./scanner -path /home/data -output results.csv
    ./scanner -path /var/log -silent
    ./scanner -path . -ext "txt,log,csv,xml"
    
Exit Codes:
    0 - Success
    1 - Error (invalid path, scan failure, etc.)
`)
}

func displayBanner() {
	fmt.Println(`
    ╔══════════════════════════════════════════════════════════╗
    ║                                                          ║
    ║     BasicPanScanner - PCI Compliance Tool                ║
    ║     ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                  ║
    ║     Version: 1.0.0                                       ║
    ║     Author:  @keraattin                                  ║
    ║     Purpose: Detect credit card data in files            ║
    ║                                                          ║
    ║     [████ ████ ████ ████] Card Detection Active          ║
    ║                                                          ║
    ╚══════════════════════════════════════════════════════════╝
    `)
}

func main() {
	// Define command line flags
	pathFlag := flag.String("path", "", "Directory path to scan (required)")
	outputFlag := flag.String("output", "", "Output file for results (CSV format)")
	helpFlag := flag.Bool("help", false, "Show help message")
	silentFlag := flag.Bool("silent", false, "Silent mode - only show summary")
	extensionsFlag := flag.String("ext", "txt,log,csv", "File extensions to scan (comma-separated)")

	// Parse the flags
	flag.Parse()

	// Show help if requested or if no arguments provided
	if *helpFlag || len(os.Args) == 1 {
		showHelp()
		return
	}

	// Path is required
	if *pathFlag == "" {
		fmt.Println("Error: -path flag is required")
		fmt.Println("Use -help for usage information")
		os.Exit(1)
	}

	// Display banner (unless silent mode)
	if !*silentFlag {
		displayBanner()
	}

	// Validate the directory
	err := validateDirectory(*pathFlag)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Parse extensions
	extensions := strings.Split(*extensionsFlag, ",")
	for i := range extensions {
		extensions[i] = "." + strings.TrimSpace(extensions[i])
	}

	// Start scanning
	if !*silentFlag {
		fmt.Printf("Scanning: %s\n", *pathFlag)
		fmt.Printf("Extensions: %s\n", *extensionsFlag)
	}

	// Run the scan
	err = scanDirectoryWithOptions(*pathFlag, *outputFlag, *silentFlag, extensions)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}
}
