# BasicPanScanner

A lightweight Go tool for scanning files and directories to detect **credit card numbers** for **PCI DSS** compliance.

## ⚠️ Security Notice

This tool is for **authorized security auditing only**. Use only on systems you own or have explicit permission to scan.

## Features (v1.0.0)

✅ **Card Detection**
- Finds 13-19 digit credit card patterns
- Handles common formats (spaces, dashes)
- Supports variable card lengths

✅ **Validation**
- Luhn algorithm validation
- Reduces false positives
- Validates card number checksums

✅ **Card Type Identification**
- Visa (13, 16, 19 digits)
- MasterCard (16 digits)
- American Express (15 digits)
- Discover (16 digits)
- Diners Club (14 digits)
- JCB (16 digits)

✅ **Security Features**
- PCI-compliant card masking (shows only first 6 and last 4 digits)
- Never stores or logs full card numbers
- Safe for production environments

✅ **File Scanning**
- Scans .txt, .log, and .csv files
- Recursive directory scanning
- Line-by-line processing for large files

## Installation
```bash
# Clone the repository
git clone https://github.com/keraattin/BasicPanScanner.git
cd BasicPanScanner

# Build the binary
go build scanner.go

# Or run directly
go run scanner.go
```

## Usage
### Interactive Mode
```
./scanner
# Enter directory when prompted
```
### Scan Current Directory
```
./scanner
# Press Enter when prompted for directory
```


## Example Output
```
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

Enter directory to scan (or press Enter for current): /var/log

Scanning directory: /var/log
==================================================
Scanning file: /var/log/app.log
  Line 42: Visa card: 453201******0366 ✓
  Line 156: MasterCard card: 510510******5100 ✓
Scan complete. Found 2 patterns, 2 valid cards.

==================================================
Directory scan complete
Total files found: 45
Files scanned: 12
```

## How It Works
- Pattern Detection: Identifies sequences of 13-19 digits
- Format Handling: Recognizes cards with spaces/dashes
- Luhn Validation: Verifies mathematical validity
- Type Detection: Identifies card issuer by BIN
- Secure Display: Masks middle digits for security

## Next Features (v1.1.0)
- [ ] Progress indicators for large directories
- [ ] Export results to CSV/JSON
- [ ] Configuration file support
- [ ] Exclude patterns and directories

## Future Plans (v2.0.0)

- [ ] Database scanning (MySQL, PostgreSQL)
- [ ] Memory scanning
- [ ] Network traffic monitoring
- [ ] Cloud storage scanning (S3, Azure)
- [ ] Regular expression patterns
- [ ] Multi-threaded scanning
- [ ] Real-time monitoring mode

## Disclaimer
This tool is for educational and authorized security auditing purposes only. Users are responsible for complying with all applicable laws and regulations.