# BasicPanScanner

![Version](https://img.shields.io/badge/version-2.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg)
![PCI DSS](https://img.shields.io/badge/PCI%20DSS-Compliant-success.svg)

A lightweight, high-performance Go tool for scanning files and directories to detect **credit card numbers (PANs)** for **PCI DSS compliance**. Features regex-based pattern matching, Luhn validation, and comprehensive reporting in multiple formats.

## 📋 Table of Contents

- [Features](#-features)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [Configuration](#-configuration)
- [Export Formats](#-export-formats)
- [Supported Card Issuers](#-supported-card-issuers)
- [Examples](#-examples)
- [Performance](#-performance)
- [Security Notice](#-security-notice)
- [Changelog](#-changelog)

---

## ✨ Features

### Core Functionality
- 🔍 **Smart Pattern Detection**: Regex-based detection for 11 international card issuers
- ✅ **Luhn Algorithm Validation**: Eliminates false positives with checksum verification
- 💳 **Card Type Identification**: Detects Visa, Mastercard, Amex, Discover, JCB, Diners Club, UnionPay, Maestro, RuPay, Troy, and Mir
- 🔒 **PCI-Compliant Masking**: Displays only first 6 (BIN) and last 4 digits
- 📁 **Recursive Directory Scanning**: Scan entire directory trees with smart exclusions
- ⚡ **Concurrent Processing**: Multi-threaded scanning with configurable worker pools
- 🎯 **120+ File Types**: Comprehensive coverage of text, data, office, and code files

### Reporting & Analytics
- 📊 **Comprehensive Statistics**: Card type distribution, risk assessment, file analysis
- 📈 **Risk Assessment**: Automatic classification (High/Medium/Low risk)
- 🎨 **5 Export Formats**: JSON, CSV, XML, HTML, TXT
- 🗂️ **Grouped Findings**: Results organized by file for easy remediation
- 📉 **Executive Summaries**: High-level overview for management
- 🎭 **Interactive HTML Reports**: Accordion UI with charts and visual analytics

### Performance & Reliability
- 🚀 **High Performance**: Scans 100,000+ files efficiently with concurrent processing
- 🛡️ **Robust Configuration**: Validates settings and provides helpful warnings
- 💾 **Smart File Filtering**: Excludes 100+ common build/cache directories
- 📏 **Size Limits**: Configurable maximum file size (default: 50MB)
- ⚙️ **Flexible Configuration**: JSON-based config with CLI overrides

---

## 📋 Requirements

- **Go**: Version 1.19 or higher
- **Operating System**: Linux, macOS, or Windows
- **Permissions**: Read access to target files/directories
- **Memory**: Minimum 512MB RAM (recommended: 1GB+)

---

## 🔧 Installation

### Method 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/keraattin/BasicPanScanner.git
cd BasicPanScanner

# Build the binary
go build scanner.go

# Make executable (Linux/macOS)
chmod +x scanner
```

### Method 2: Direct Build

```bash
# Download source
wget https://raw.githubusercontent.com/keraattin/BasicPanScanner/main/scanner.go
wget https://raw.githubusercontent.com/keraattin/BasicPanScanner/main/config.json

# Build
go build scanner.go
```

---

## 🚀 Quick Start

```bash
# Basic scan
./scanner -path /var/log

# Scan with HTML report
./scanner -path /var/log -output report.html

# Fast scan with 4 workers
./scanner -path /data -workers 4 -output report.json
```

---

## 💻 Usage

### Command Line Options

```
BasicPanScanner v2.0.0 - PCI Compliance Scanner

Required:
    -path <directory>      Directory to scan

Options:
    -output <file>         Save results (.json, .csv, .html, .txt, .xml)
    -ext <list>           Extensions to scan (default: from config)
    -exclude <list>       Directories to skip (default: from config)
    -workers <n>          Number of concurrent workers (default: CPU/2)
    -help                 Show help
```

### Examples

#### Basic Scanning
```bash
# Scan current directory
./scanner -path .

# Scan specific directory
./scanner -path /var/log

# Scan with progress output
./scanner -path /data -workers 4
```

#### Custom File Types
```bash
# Scan only text and log files
./scanner -path /var/log -ext txt,log

# Scan specific extensions
./scanner -path /data -ext csv,sql,json
```

#### Output Formats
```bash
# JSON output (machine-readable)
./scanner -path /var/log -output report.json

# CSV output (Excel-ready)
./scanner -path /var/log -output report.csv

# HTML output (interactive, with charts)
./scanner -path /var/log -output report.html

# XML output (enterprise systems)
./scanner -path /var/log -output report.xml

# Plain text (human-readable)
./scanner -path /var/log -output report.txt
```

#### Performance Tuning
```bash
# Single-threaded (lower CPU usage)
./scanner -path /var/log -workers 1

# Maximum performance (use all CPU cores)
./scanner -path /var/log -workers 8

# Balanced (default: half of CPU cores)
./scanner -path /var/log
```

---

## ⚙️ Configuration

### config.json Structure

```json
{
  "extensions": [
    "txt", "log", "csv", "json", "xml", "sql", ...
  ],
  "exclude_dirs": [
    ".git", "node_modules", "vendor", "dist", ...
  ],
  "max_file_size": "50MB"
}
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `extensions` | File types to scan | 120+ types |
| `exclude_dirs` | Directories to skip | 100+ common dirs |
| `max_file_size` | Maximum file size | 50MB |

### File Extensions Scanned (120+ types)

**Data Files**: txt, log, csv, tsv, json, xml, yaml, sql, db, sqlite  
**Office Documents**: doc, docx, xls, xlsx, pdf, rtf, odt, ods  
**Email Files**: eml, msg, pst, mbox, emlx  
**Code Files**: js, py, java, go, php, rb, c, cpp, cs, swift  
**Config Files**: conf, cfg, ini, properties, env, toml  
**Backup Files**: bak, backup, old, tmp, dump, swp  
**Web Files**: html, htm, php, asp, jsp

### Excluded Directories (100+ patterns)

**Version Control**: .git, .svn, .hg  
**Dependencies**: node_modules, vendor, bower_components  
**Build Output**: dist, build, target, bin, out  
**IDE Files**: .idea, .vscode, .vs, .eclipse  
**Caches**: __pycache__, .cache, .gradle, .m2  
**System**: proc, sys, dev, Library, System

---

## 📊 Export Formats

### 1. HTML (Interactive)
- Executive summary with risk assessment
- Interactive accordion UI (click to expand/collapse)
- Animated bar charts for distribution
- Card type icons and visual indicators
- Mobile-responsive design
- Risk-based color coding

### 2. JSON (Machine-Readable)
```json
{
  "version": "2.0.0",
  "scan_info": {
    "scan_date": "2025-01-15T10:30:00Z",
    "directory": "/var/log",
    "duration": "1m23s"
  },
  "summary": {
    "total_cards": 12,
    "high_risk_files": 0
  },
  "statistics": {
    "cards_by_type": {"Visa": 8, "Mastercard": 3}
  },
  "findings": {
    "/var/log/app.log": [...]
  }
}
```

### 3. CSV (Spreadsheet-Ready)
- Summary header with scan information
- Card type distribution with percentages
- Top files ranked by card count
- Grouped findings by file
- Risk level indicators

### 4. XML (Enterprise Systems)
```xml
<ScanReport version="2.0.0">
  <ScanInfo>...</ScanInfo>
  <Statistics>...</Statistics>
  <Findings>
    <FileGroup path="/var/log/app.log" count="4">
      <Finding>...</Finding>
    </FileGroup>
  </Findings>
</ScanReport>
```

### 5. TXT (Human-Readable)
- ASCII art headers and separators
- Box-drawing characters for structure
- Emoji indicators (🔴🟡🟢) for risk levels
- ASCII bar charts for distributions
- Tree-structured findings (├─ └─)

---

## 💳 Supported Card Issuers

| Card Type | Length | Validation |
|-----------|--------|------------|
| **Visa** | 13, 16, 19 | Regex + Luhn |
| **Mastercard** | 16 | Regex + Luhn |
| **American Express** | 15 | Regex + Luhn |
| **Discover** | 16 | Regex + Luhn |
| **Diners Club** | 14 | Regex + Luhn |
| **JCB** | 16 | Regex + Luhn |
| **UnionPay** | 16-19 | Regex + Luhn |
| **Maestro** | 12-19 | Regex + Luhn |
| **RuPay** | 16 | Regex + Luhn |
| **Troy** | 16 | Regex + Luhn |
| **Mir** | 16 | Regex + Luhn |

### Pattern Support
- Unformatted: `XXXXXXXXXXXXXXXX`
- With dashes: `XXXX-XXXX-XXXX-XXXX`
- With spaces: `XXXX XXXX XXXX XXXX`

---

## 📖 Examples

### Example 1: Basic Security Audit

```bash
# Scan production logs
./scanner -path /var/log/production -output security_audit.html

# Output:
# ✓ Loaded 11 card issuer patterns
# 
# Scanning directory: /var/log/production
# Workers: 2 (concurrent scanning enabled)
# Max file size: 50.00 MB
# ============================================================
# 
# ✓ Found 3 cards in: transactions.log
# [Scanned: 156/234 | Cards: 3]
# 
# ============================================================
# ✓ Scan complete!
#   Time: 1m23s
#   Total files: 234
#   Scanned: 156
#   Cards found: 3
#   Scan rate: 1.9 files/second
# 
#   ✓ Saved: security_audit.html
```

### Example 2: Development Environment Scan

```bash
# Scan codebase for hardcoded cards
./scanner -path /home/user/projects -ext py,js,java,go -output code_audit.json

# Excludes build directories automatically
```

### Example 3: Compliance Reporting

```bash
# Monthly compliance scan with detailed report
./scanner -path /data/exports -workers 4 -output monthly_report_2025_01.html

# Share HTML report with compliance team
```

### Example 4: High-Performance Scan

```bash
# Scan large dataset with maximum workers
./scanner -path /mnt/archive -workers 8 -output large_scan.csv

# Example performance: ~50,000 files in 5 minutes
```

---

## ⚡ Performance

### Benchmarks

| Files | Workers | Time | Files/sec |
|-------|---------|------|-----------|
| 1,000 | 1 | 12s | ~83 |
| 1,000 | 2 | 7s | ~143 |
| 1,000 | 4 | 4s | ~250 |
| 10,000 | 4 | 45s | ~222 |
| 100,000 | 8 | 8m | ~208 |

### Performance Tips

1. **Use Multiple Workers**: Default is CPU/2, increase for faster scanning
2. **Optimize Extensions**: Only scan necessary file types
3. **Exclude Build Dirs**: Config already excludes 100+ common directories
4. **Adjust File Size Limit**: Reduce if scanning large binaries
5. **Use SSD Storage**: Significantly faster than HDD for large scans

### Optimization

```bash
# Fast scan (fewer file types)
./scanner -path /var/log -ext txt,log -workers 4

# Thorough scan (all file types)
./scanner -path /data -workers 8

# Memory-efficient scan
./scanner -path /large/dataset -workers 1
```

---

## ⚠️ Security Notice

### Important

This tool is for **authorized security auditing only**. Use only on systems you own or have explicit permission to scan.

### Best Practices

1. **Authorized Use Only**: Obtain proper authorization before scanning
2. **Secure Reports**: Treat output files as sensitive data
3. **Encrypt Reports**: Use encryption for report storage and transmission
4. **Access Control**: Limit access to scan results
5. **Audit Trail**: Log all scans for compliance purposes
6. **Data Retention**: Follow your organization's data retention policies

### PCI DSS Compliance

BasicPanScanner helps meet:
- **Requirement 3.2**: Data discovery and inventory
- **Requirement 12.5**: Security awareness and scanning

---

## 📝 Changelog

### Version 2.0.0 (13.10.2025)

#### Major Features
- **Enhanced Export Formats**: All formats now include statistics and grouped findings
- **Interactive HTML Reports**: Accordion UI, executive summaries, visual charts
- **Comprehensive Statistics**: Risk assessment, card type distribution, top files
- **Expanded Coverage**: 120+ file types, 100+ excluded directories

#### Improvements
- **Regex-Based Detection**: Issuer-specific patterns reduce false positives
- **Better Configuration**: Enhanced validation with helpful warnings
- **Grouped Findings**: Results organized by file for easier remediation
- **Performance**: 20-30% faster with optimized exclusions

#### Bug Fixes
- Fixed file size parsing bug (map iteration order)
- Fixed XML encoding errors (map to slice conversion)
- Removed duplicate directories from config
- Fixed progress indicators

#### Breaking Changes
- Report structure now uses only GroupedFindings
- Removed duplicate flat Findings array
- All exports show grouped format

### Version 1.1.0 (11.10.2025)
- Added regex-based card detection
- Support for 11 card issuers
- Multiple export formats (JSON, CSV, XML, HTML, TXT)

### Version 1.0.0 (09.10.2025)
- Initial release
- Basic card detection with Luhn validation
- Text and JSON output

---


### Contribution Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Reporting Issues

Found a bug? Please open an issue with:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior


## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Uses Go's standard library for reliability and performance
- Inspired by PCI DSS compliance requirements

---