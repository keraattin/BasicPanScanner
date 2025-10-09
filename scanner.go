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

			// Check if we found 16 digits!
			if len(consecutiveDigits) == 16 {
				return consecutiveDigits // Found a card!
			}
		} else {
			// Not a digit - reset
			consecutiveDigits = ""
		}
	}

	return "" // No card found
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
	}
}
