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

# Process multiple images
imr photo1.jpg photo2.png photo3.webp

# Use custom suffix
imr -suffix _clean vacation/*.jpg

# Verbose output
imr -verbose photo.jpg
```

### Options

```
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

- ✅ JPEG (.jpg, .jpeg)
- ✅ PNG (.png)
- ✅ WebP (.webp)
- ✅ TIFF (.tiff, .tif)
- ✅ GIF (.gif)

## Development Status

**Current Phase**: Phase 0 - Project Scaffolding ✓

**Roadmap**:
- Phase 0: Project scaffolding and dummy CLI ✅
- Phase 1: JPEG support with EXIF removal (In Progress)
- Phase 2: Extended format support (PNG, WebP, TIFF, GIF)
- Phase 3: Advanced features (recursive processing, dry-run mode)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details

## Privacy & Security

This tool is designed to help protect your privacy by removing potentially sensitive metadata from your images. Always verify the output before sharing images publicly.
