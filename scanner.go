package main

import (
	"bufio" // buffered I/O for reading input
	"fmt"   // for printing
	"os"    // operating system stuff (Stdin)
	"strings"
)

func main() {
	// Declare a variable to hold version number
	var version string = "0.01"
	author := "@keraattin"

	fmt.Println("PCI Scanner Starting...")
	fmt.Println("Version:", version)
	fmt.Println("Author:", author)

	// Create a scanner to read input
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("PCI Scanner Setup")
	fmt.Println("-----------------")

	// Ask for input
	fmt.Print("Enter path to scan: ")

	// Read the input
	path, _ := reader.ReadString('\n')

	// Clean up the input (remove newline and spaces)
	path = strings.TrimSpace(path)

	if path == "" {
		fmt.Println("❌ Error: No path provided!")
	} else if path == "/" {
		fmt.Println("⚠️ Warning: Scanning root directory can take a long time!")
	} else {
		fmt.Println("✓ Path to scan:", path)
	}
}
