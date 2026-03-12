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
    Shows grouped and alphabetically sorted output:
      - "Metadata to be removed" (sensitive tags)
      - "Metadata to be preserved" (technical tags)
    Includes summary with tag counts

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

## Metadata Handling

IMR uses **selective preservation** - it removes only sensitive metadata while keeping technical image information intact.

### Removed (Sensitive Data)
- **GPS data**: All location information (latitude, longitude, altitude, timestamps)
- **Camera/Lens info**: Make, model, serial numbers
- **Software**: Editing software and version information
- **Creator info**: Artist, copyright, comments, descriptions
- **Timestamps**: Creation, modification, digitization dates
- **Serial numbers**: Body and lens serial numbers

### Preserved (Technical Data)
- **Image properties**: Dimensions, resolution, color space, orientation
- **Exposure settings**: ISO, aperture, shutter speed, focal length
- **Technical tags**: Flash, metering mode, white balance, exposure mode
- **Quality settings**: Compression, sharpness, saturation, contrast
- **Format metadata**: EXIF version, component configuration

**Example**: A typical JPEG with 59 EXIF tags will have 18 sensitive tags removed and 41 technical tags preserved.

### How It Works

1. **Parse EXIF structure**: Read all existing EXIF tags from the image
2. **Filter by category**: Identify which tags are sensitive (40+ tag IDs)
3. **Remove GPS IFD**: Completely drop the GPS sub-directory
4. **Rebuild EXIF**: Create new EXIF block with only non-sensitive tags
5. **Write output**: Save image with filtered EXIF metadata

The tool uses `github.com/dsoprea/go-exif/v3` for precise EXIF manipulation without re-encoding the image data, preserving maximum quality.

## Supported Formats

**Currently Supported**:
- ✅ JPEG (.jpg, .jpeg) - Selective EXIF preservation (removes sensitive tags, keeps technical ones)

**Currently Supported Formats**:
- ✅ JPEG/JPG (.jpg, .jpeg) - All case variations supported
  - Selective EXIF preservation
  - Removes: GPS, device info, lens info, serial numbers, timestamps, creator info
  - Preserves: Exposure settings, resolution, orientation, color space, etc.
- ✅ PNG (.png) - All case variations supported
  - Selective chunk preservation
  - Removes: tEXt/zTXt/iTXt chunks with sensitive keys (Author, Copyright, Comment, Software, Title, etc.)
  - Removes: eXIf chunks with sensitive EXIF data (GPS, device info, timestamps)
  - Removes: tIME chunks (modification timestamps)
  - Preserves: Technical chunks (IHDR, IDAT, PLTE, gAMA, cHRM, sRGB, iCCP, pHYs, etc.)

**Planned Support** (Phase 3):
- 🔜 WebP (.webp) - RIFF container with EXIF/XMP metadata handling
- 🔜 HEIC/HEIF (.heic, .heif) - Modern smartphone format (complex ISOBMFF container)
- 🔜 TIFF (.tiff, .tif) - EXIF handling (similar to JPEG)
- 🔜 GIF (.gif) - Comment and extension block handling

## Development Status

**Current Phase**: Phase 3 - PNG Support Complete ✓

**Completed Features**:
- ✅ JPEG metadata removal with selective preservation (Phase 1 & 2 Enhancement)
- ✅ PNG metadata removal with selective preservation (Phase 3)
- ✅ Dry-run mode with grouped, alphabetically sorted metadata preview (Phase 1)
- ✅ Multiple file processing (Phase 1)
- ✅ Custom output suffix (Phase 1)
- ✅ Recursive directory processing (Phase 2)
- ✅ Smart path handling - mix files and directories (Phase 2)
- ✅ Selective EXIF preservation - keeps technical tags, removes sensitive ones (Phase 2 Enhancement)
- ✅ Tag classification fixes - MakerNote, GPSInfoIFDPointer, SubSec timestamps, thumbnails (Phase 2 Enhancement)
- ✅ Alphabetically sorted tag display in dry-run mode (Phase 2 Enhancement)
- ✅ PNG chunk-based metadata handling (Phase 3)
- ✅ Text chunk filtering with pattern matching (Phase 3)
- ✅ PNG EXIF metadata filtering (Phase 3)

**Roadmap**:
- Phase 0: Project scaffolding and dummy CLI ✅
- Phase 1: JPEG support with EXIF removal and dry-run mode ✅
- Phase 2: Recursive directory processing ✅
- Phase 2 Enhancement: Selective EXIF preservation ✅
  - Selective preservation implementation ✅
  - MakerNote and GPSInfoIFDPointer fixes ✅
  - SubSec timestamp fixes ✅
  - Thumbnail tag corruption fixes ✅
  - Alphabetical sorting ✅
- Phase 3: Extended format support 🔄
  - PNG support ✅
  - WebP support (next)
  - HEIC/HEIF support (future)
  - TIFF support (future)
  - GIF support (future)
- Phase 4: Progress indicators and configuration file support
- Phase 5: Comprehensive testing, benchmarks, and documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See LICENSE file for details

## Privacy & Security

This tool is designed to help protect your privacy by removing potentially sensitive metadata from your images. Always verify the output before sharing images publicly.
