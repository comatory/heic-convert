# heic-convert

CLI tool to convert HEIC images to JPEG format.

## Build

```bash
just build
```

## Usage

```bash
./bin/heic-to-jpeg [options] <input files>
```

It is recommended to copy the binary to `/usr/local/bin` for easier access.

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

Import the provided `HEIC to JPG.workflow` into Automator to create a quick action for converting HEIC images directly from Finder.
The workflow assumes that the binary is in its default location (`/usr/local/bin/heic-to-jpeg`).
