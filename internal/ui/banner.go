// Package ui handles user interface elements
// This file provides the application banner display
package ui

import "fmt"

// ShowBanner displays the application banner
// This shows the app name, version, and purpose
//
// Parameters:
//   - version: Application version (e.g., "3.0.0")
//
// Example:
//   ui.ShowBanner("3.0.0")
func ShowBanner(version string) {
	fmt.Println(`
    ╔══════════════════════════════════════════════════════════╗
    ║                                                          ║
    ║     BasicPanScanner - PCI Compliance Tool                ║
    ║     ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀                  ║
    ║     Version: ` + version + `                             ║
    ║     Author:  @keraattin                                  ║
    ║     Purpose: Detect credit card data in files            ║
    ║                                                          ║
    ║     [████ ████ ████ ████] Card Detection Active          ║
    ║                                                          ║
    ╚══════════════════════════════════════════════════════════╝
    `)
}

// ShowHelp displays usage help text
// This is shown when user runs with -help flag or no arguments
func ShowHelp() {
	fmt.Println(`
BasicPanScanner v3.0.0 - PCI Compliance Scanner

Usage: ./scanner -path <directory> [options]

Required:
    -path <directory>      Directory to scan

Options:
    -output <file>         Save results (.json, .csv, .html, .txt, .xml)
    -mode <mode>          Scan mode: 'whitelist' or 'blacklist' (overrides config)
    -ext <list>           Extensions (applies to active mode)
    -exclude <list>       Directories to skip (default: from config)
    -workers <n>          Number of concurrent workers (default: CPU/2)
    -help                 Show this help

Scan Modes:
    whitelist             Scan ONLY specified extensions
    blacklist             Scan everything EXCEPT specified extensions (default)

Examples:
    # Use config.json settings (blacklist mode by default)
    ./scanner -path /var/log

    # Force whitelist mode (scan only .txt and .log)
    ./scanner -path /var/log -mode whitelist -ext txt,log

    # Force blacklist mode (scan everything except .jpg and .png)
    ./scanner -path /data -mode blacklist -ext jpg,png

    # Fast scan with 4 workers
    ./scanner -path /var/log -workers 4 -output report.html

Configuration:
    Edit config.json to set default mode and extension lists.
    CLI flags always override config values.

Performance:
    Default workers: CPU cores / 2 (safe for production)
    More workers = faster scanning (2-4x speed improvement)

Supported Card Types:
    Visa, Mastercard, Amex, Discover, Diners Club, JCB,
    UnionPay, Maestro, RuPay, Troy, Mir
`)
}
