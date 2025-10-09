package main

import (
	"bufio" // buffered I/O for reading input
	"fmt"   // for printing
	"os"    // operating system stuff (Stdin)
	"strings"
)

// Get Path From User
func getPathFromUser() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter path to scan: ")
	path, _ := reader.ReadString('\n')
	path = strings.TrimSpace(path)

	return path
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

	// Process each line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check current line for card patterns
		cardNumber := findCardNumber(line)
		if cardNumber != "" {
			foundCount++
			// In production, we'd hash the card number instead of printing it
			fmt.Printf("  Line %d: FOUND CARD: %s\n", lineNumber, cardNumber)
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Print summary
	fmt.Printf("Scan complete. Found %d potential cards.\n\n", foundCount)
	return nil
}

func main() {
	// Declare a variable to hold version number
	var version string = "0.01"
	author := "@keraattin"

	fmt.Println("PCI Scanner Starting...")
	fmt.Println("Version:", version)
	fmt.Println("Author:", author)

	// Get the Scan Path
	scanPath := getPathFromUser()

	if scanPath == "" {
		fmt.Println("❌ Error: No path provided!")
	} else if scanPath == "/" {
		fmt.Println("⚠️ Warning: Scanning root directory can take a long time!")
	} else {
		fmt.Println("✓ Path to scan:", scanPath)
		scanFile(scanPath) // Do the scan
	}
}
