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

// Searching 16 Length digits for possible card data
func findCardNumber(text string) string {
	consecutiveDigits := ""

	for i := 0; i < len(text); i++ {
		char := text[i]

		if char >= '0' && char <= '9' {
			// Add digit to our string
			consecutiveDigits = consecutiveDigits + string(char)

			// Check if we found 16 digits
			if len(consecutiveDigits) == 16 {
				return consecutiveDigits
			}

		} else if char == ' ' || char == '-' {
			// Space or dash - keep going if we have digits
			if len(consecutiveDigits) == 0 {
				// No digits yet, ignore this space/dash
				consecutiveDigits = ""
			}
			// If we have digits, don't reset! Just skip the space/dash

		} else {
			// Any other character - reset
			consecutiveDigits = ""
		}
	}

	return "" // No card found
}

func scanFile(filepath string) {
	fmt.Printf("Scanning file: %s\n", filepath)

	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return
	}
	defer file.Close() // Make sure file gets closed

	// Create a scanner to read line by line
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	foundCount := 0

	// Read each line
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check this line for cards
		cardNumber := findCardNumber(line)
		if cardNumber != "" {
			foundCount++
			fmt.Printf("  Line %d: FOUND CARD: %s\n", lineNumber, cardNumber)
		}
	}

	fmt.Printf("Scan complete. Found %d potential cards.\n\n", foundCount)
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
