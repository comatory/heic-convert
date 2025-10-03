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

### Options

- `-o <path>` - Output directory (default: current directory)
- `-q <quality>` - JPEG quality 1-100 (default: 100)
- `-v` - Verbose logging
- `-h` - Display help
