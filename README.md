# heic-convert

CLI tool to convert HEIC images to JPEG format.

## Installation

### Homebrew (macOS)

```bash
brew install comatory/tap/comatory-heic-convert
```

The binary will be available as `heic-to-jpeg`. See man page with `man heic-to-jpeg`.

### Build from source

```bash
just build
```

## Usage

```bash
heic-to-jpeg [options] <input files>
```

### Options

- `-o <path>` - Output directory (default: current directory)
- `-q <quality>` - JPEG quality 1-100 (default: 100)
- `-v` - Verbose logging
- `-h` - Display help

### Example

Convert a single HEIC file to JPEG:

```bash
heic-to-jpeg image.heic
```

Convert multiple HEIC files to a specified output directory with quality 80:

```bash
heic-to-jpeg -o /path/to/output -q 80 *.heic
```

## Automator workflow (macOS)

For Finder right-click integration, install the Quick Action workflow:

**Via Homebrew:**
After installing with brew, run:
```bash
open "$(brew --prefix)/share/comatory-heic-convert/HEIC to JPG.workflow"
```

**Manual:**
Double-click `HEIC to JPG.workflow` to install. The workflow uses the `heic-to-jpeg` binary from your PATH.
