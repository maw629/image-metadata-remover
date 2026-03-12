# Image Metadata Remover (IMR)

A lean and efficient CLI tool written in Go that removes sensitive metadata from images while preserving image quality.

## Features

- 🔒 Remove sensitive metadata (GPS, camera info, personal data)
- 🖼️ Support for multiple formats: JPEG, PNG, WebP, TIFF, GIF
- ⚡ Fast and efficient processing
- 📦 Zero-dependency binary
- 🔧 Simple CLI interface

## Installation

### From Source

```bash
git clone https://github.com/yourusername/image-metadata-remover.git
cd image-metadata-remover
go build -o imr ./cmd/imr
```

### Binary

Download the latest release from the [releases page](https://github.com/yourusername/image-metadata-remover/releases).

## Usage

### Basic Usage

```bash
# Process a single image
imr photo.jpg

# Preview metadata without removing (dry-run)
imr --dry-run photo.jpg

# Process directory recursively
imr -r photos/

# Process multiple files and directories
imr photo1.jpg photo2.png images/

# Use custom suffix
imr -suffix _clean vacation/*.jpg

# Verbose recursive processing
imr -r -verbose photos/
```

### Options

```
--dry-run
    Preview metadata without removing (no output file created)

-r, --recursive
    Process directories recursively
    
-suffix string
    Suffix to append to output filenames (default "_imr")
    
-verbose
    Enable verbose output
    
-version, -v
    Show version information
    
-help, -h
    Show help message
```

### Output

Output files are saved with the specified suffix (default: `_imr`):
- `photo.jpg` → `photo_imr.jpg`
- `image.png` → `image_imr.png`

## Metadata Removal

### Removed (Sensitive Data)
- GPS coordinates and location data
- Camera make, model, and serial numbers
- Lens information
- Software and editing history
- Copyright and creator information
- Timestamps (optional)

### Preserved (Basic Data)
- Image dimensions
- Color space
- Orientation
- Basic quality settings

## Supported Formats

**Currently Supported**:
- ✅ JPEG (.jpg, .jpeg) - Full EXIF metadata removal

**Planned Support** (Phase 3):
- 🔜 PNG (.png) - tEXt, iTXt, zTXt chunks
- 🔜 WebP (.webp)
- 🔜 TIFF (.tiff, .tif)
- 🔜 GIF (.gif)

## Development Status

**Current Phase**: Phase 2 - Enhanced Features ✓

**Completed Features**:
- ✅ JPEG metadata removal (Phase 1)
- ✅ Dry-run mode to preview all metadata (Phase 1)
- ✅ Multiple file processing (Phase 1)
- ✅ Custom output suffix (Phase 1)
- ✅ Recursive directory processing (Phase 2)
- ✅ Smart path handling - mix files and directories (Phase 2)

**Roadmap**:
- Phase 0: Project scaffolding and dummy CLI ✅
- Phase 1: JPEG support with EXIF removal and dry-run mode ✅
- Phase 2: Recursive directory processing ✅
- Phase 3: Extended format support (PNG, WebP, TIFF, GIF)
- Phase 4: Progress indicators and configuration file support
- Phase 5: Comprehensive testing, benchmarks, and documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details

## Privacy & Security

This tool is designed to help protect your privacy by removing potentially sensitive metadata from your images. Always verify the output before sharing images publicly.
