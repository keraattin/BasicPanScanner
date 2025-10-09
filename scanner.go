package main

import "fmt"

func main() {
	// Declare a variable to hold version number
	var version string = "0.01"

	author := "@keraattin"
	scanCount := 0
	isReady := true

	// Scanner statistics
	filesFound := 1250
	filesScanned := 300
	suspiciousFiles := 5

	filesRemaining := filesFound - filesScanned
	scanPercentage := (filesScanned * 100) / filesFound

	fmt.Println("PCI Scanner Starting...")
	fmt.Println("Version:", version)
	fmt.Println("Author:", author)
	fmt.Println("Scans completed:", scanCount)
	fmt.Println("Ready to scan:", isReady)

	fmt.Println("PCI Scanner Statistics")
	fmt.Println("----------------------")
	fmt.Println("Files found:", filesFound)
	fmt.Println("Files scanned:", filesScanned)
	fmt.Println("Files remaining:", filesRemaining)
	fmt.Println("Progress:", scanPercentage, "%")
	fmt.Println("Suspicious files:", suspiciousFiles)
}
