package main

import (
	"bufio" // buffered I/O for reading input
	"fmt"   // for printing
	"os"    // operating system stuff (Stdin)
	"path/filepath"
	"strings"
)

// getDirectoryFromUser prompts user for a directory path
func getDirectoryFromUser() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter directory to scan (or press Enter for current): ")
	dirPath, _ := reader.ReadString('\n')
	dirPath = strings.TrimSpace(dirPath)

	// Default to current directory if empty
	if dirPath == "" {
		dirPath = "."
		fmt.Println("Using current directory")
	}

	return dirPath
}

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

// findCardNumber searches for 16 consecutive digits in a string
// It handles common formats like spaces and dashes between digit groups
func findCardNumber(text string) string {
	consecutiveDigits := ""

	// Iterate through each character in the text
	for i := 0; i < len(text); i++ {
		char := text[i]

		if char >= '0' && char <= '9' {
			// Found a digit, add it to our collection
			consecutiveDigits = consecutiveDigits + string(char)

			// Check if we've found exactly 16 digits
			if len(consecutiveDigits) == 16 {
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
			consecutiveDigits = ""
		}
	}

	// No 16-digit sequence found
	return ""
}

// scanDirectory walks through a directory and scans all .txt and .log files
func scanDirectory(dirPath string) error {
	fmt.Printf("Scanning directory: %s\n", dirPath)
	fmt.Println("=" + strings.Repeat("=", 40))

	totalFiles := 0
	scannedFiles := 0

	// Walk through the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		// Check for walk errors
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Continue walking despite error
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		totalFiles++

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))

		// Only scan text-like files
		if ext == ".txt" || ext == ".log" || ext == ".csv" {
			scannedFiles++

			// Scan this file
			err := scanFile(path)
			if err != nil {
				fmt.Printf("Error scanning %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	// Print summary
	fmt.Println("=" + strings.Repeat("=", 40))
	fmt.Printf("Directory scan complete\n")
	fmt.Printf("Total files found: %d\n", totalFiles)
	fmt.Printf("Files scanned: %d\n", scannedFiles)

	return nil
}

// scanFile reads a file line by line and checks for credit card patterns
func scanFile(filepath string) error {
	fmt.Printf("Scanning file: %s\n", filepath)

	// Attempt to open the file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	// Ensure file is closed when function exits
	defer file.Close()

	// Create a scanner for line-by-line reading
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	foundCount := 0
	validCount := 0 // Track valid cards

	// Process each line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check current line for card patterns
		cardNumber := findCardNumber(line)
		if cardNumber != "" {
			foundCount++

			// Validate with Luhn algorithm
			if validateLuhn(cardNumber) {
				validCount++

				// Get and display card type
				cardType := getCardType(cardNumber)

				// Mask the card number for safe display
				maskedCard := maskCardNumber(cardNumber)

				fmt.Printf("  Line %d: %s card: %s ✓\n", lineNumber, cardType, maskedCard)
			} else {
				fmt.Printf("  Line %d: Invalid pattern: %s (failed Luhn check)\n", lineNumber, cardNumber)
			}
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Print summary with validation info
	fmt.Printf("Scan complete. Found %d patterns, %d valid cards.\n\n", foundCount, validCount)
	return nil
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
	// Ensure we have enough digits to check
	if len(cardNumber) < 13 {
		return "Unknown"
	}

	// Check first digits for card type
	firstDigit := cardNumber[0]
	firstTwo := cardNumber[0:2]
	firstFour := ""
	if len(cardNumber) >= 4 {
		firstFour = cardNumber[0:4]
	}

	// Visa: Starts with 4
	if firstDigit == '4' {
		return "Visa"
	}

	// MasterCard: Starts with 51-55 or 2221-2720
	if (firstTwo >= "51" && firstTwo <= "55") ||
		(firstFour >= "2221" && firstFour <= "2720") {
		return "MasterCard"
	}

	// American Express: Starts with 34 or 37
	if firstTwo == "34" || firstTwo == "37" {
		return "Amex"
	}

	// Discover: Starts with 6011, 622126-622925, 644-649, 65
	if firstFour == "6011" ||
		(firstTwo >= "64" && firstTwo <= "65") {
		return "Discover"
	}

	// JCB: Starts with 3528-3589
	if firstFour >= "3528" && firstFour <= "3589" {
		return "JCB"
	}

	// Diners Club: Starts with 36, 38, or 300-305
	if firstTwo == "36" || firstTwo == "38" ||
		(firstFour >= "3000" && firstFour <= "3059") {
		return "Diners"
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

func main() {
	// Declare a variable to hold version number
	var version string = "0.01"
	author := "@keraattin"

	fmt.Println("PCI Scanner Starting...")
	fmt.Println("Version:", version)
	fmt.Println("Author:", author)

	// Get directory from user
	dirPath := getDirectoryFromUser()

	// Validate the directory
	err := validateDirectory(dirPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Scan the directory
	err = scanDirectory(dirPath)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
	}
}
