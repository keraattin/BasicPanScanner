# BasicPanScanner

![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg)

A lightweight Go tool for scanning files and directories to detect **credit card numbers** for **PCI DSS** compliance.

## 📋 Requirements

- **Go**: Version 1.19 or higher
- **Operating System**: Linux, macOS, or Windows
- **Permissions**: Read access to target files/directories

## ⚠️ Security Notice

This tool is for **authorized security auditing only**. Use only on systems you own or have explicit permission to scan.

## 🎯 Purpose

BasicPanScanner helps organizations maintain PCI DSS compliance by identifying potential credit card data (PANs - Primary Account Numbers) stored in files across their systems. This tool is essential for:

- **Security Audits**: Discover where sensitive payment card data might be stored
- **PCI Compliance**: Meet PCI DSS requirement 3.2 (data discovery)
- **Data Breach Prevention**: Identify and secure exposed cardholder data
- **Incident Response**: Quickly scan systems after security incidents

## ✨ Features (v1.1.0)
 
- **🔍 Smart Pattern Detection**: Identifies 13-19 digit card numbers
- **✅ Luhn Algorithm Validation**: Eliminates false positives with checksum verification
- **💳 Card Type Identification**: Detects Visa, Mastercard, Amex, Discover, JCB, Diners Club, and UnionPay
- **🔒 PCI-Compliant Masking**: Displays only first 6 and last 4 digits (BIN + last4)
- **📁 Directory Scanning**: Recursively scan entire directory trees
- **⚡ Concurrent File Processing**: Parallel scanning with goroutines for improved performance
- **📊 Real-time Progress Indicators**: Visual feedback during long-running scans
- **🎨 Professional Banner**: Clear visual identity with version info
- **🎛️ CLI Arguments**: Flexible command-line options for custom scans
- **📝 Detailed Reporting**: JSON and text output formats

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

### Scan Current Directory
```
./scanner
# Press Enter when prompted for directory
```


## 📊 Output Examples

### Text Output

```
============================================================
              BasicPanScanner v1.1.0
      PCI-Compliant Credit Card Number Scanner
============================================================

Scanning: /var/log/application

[FOUND] /var/log/app.log:42
  Card: 453201******0366
  Type: Visa
  Valid: ✓

[FOUND] /var/log/transactions.txt:108
  Card: 378282******1005
  Type: American Express
  Valid: ✓

Summary:
--------
Files Scanned: 234
Cards Found: 2
Scan Duration: 1.23s
```

### JSON Output

```json
{
  "version": "1.1.0",
  "scan_date": "2025-10-11T10:30:00Z",
  "path": "/var/log",
  "summary": {
    "files_scanned": 234,
    "cards_found": 2,
    "duration_seconds": 1.23
  },
  "findings": [
    {
      "file": "/var/log/app.log",
      "line": 42,
      "masked_card": "453201******0366",
      "card_type": "Visa",
      "valid": true
    },
    {
      "file": "/var/log/transactions.txt",
      "line": 108,
      "masked_card": "378282******1005",
      "card_type": "American Express",
      "valid": true
    }
  ]
}
```

## How It Works
- Pattern Detection: Identifies sequences of 13-19 digits
- Format Handling: Recognizes cards with spaces/dashes
- Luhn Validation: Verifies mathematical validity
- Type Detection: Identifies card issuer by BIN
- Secure Display: Masks middle digits for security

## 📝 Changelog

### v1.1.0 (2025-10-11)

**Added:**
- ⚡ Concurrent file processing with goroutines
- 📊 Real-time progress indicators
- 🎛️ Comprehensive CLI arguments
- 📝 JSON output format support
- 🔧 Configurable worker pool size
- 📈 Enhanced performance for large-scale scans

## Disclaimer
This tool is for educational and authorized security auditing purposes only. Users are responsible for complying with all applicable laws and regulations.