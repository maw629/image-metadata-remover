package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatJPEG    ImageFormat = "jpeg"
	FormatPNG     ImageFormat = "png"
	FormatWebP    ImageFormat = "webp"
	FormatHEIC    ImageFormat = "heic"
	FormatTIFF    ImageFormat = "tiff"
	FormatGIF     ImageFormat = "gif"
	FormatUnknown ImageFormat = "unknown"
)

// DetectFormat detects the image format by reading file magic bytes
func DetectFormat(filePath string) (ImageFormat, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FormatUnknown, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 12 bytes for magic number detection
	header := make([]byte, 12)
	n, err := file.Read(header)
	if err != nil {
		return FormatUnknown, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < 4 {
		return FormatUnknown, fmt.Errorf("file too small to determine format")
	}

	// Check JPEG (FF D8 FF)
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return FormatJPEG, nil
	}

	// Check PNG (89 50 4E 47 0D 0A 1A 0A)
	if n >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E &&
		header[3] == 0x47 && header[4] == 0x0D && header[5] == 0x0A &&
		header[6] == 0x1A && header[7] == 0x0A {
		return FormatPNG, nil
	}

	// Check WebP (RIFF ... WEBP)
	if n >= 12 && header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 &&
		header[3] == 0x46 && header[8] == 0x57 && header[9] == 0x45 &&
		header[10] == 0x42 && header[11] == 0x50 {
		return FormatWebP, nil
	}

	// Check HEIC/HEIF (....ftyp + compatible brand)
	if n >= 12 && header[4] == 0x66 && header[5] == 0x74 && header[6] == 0x79 && header[7] == 0x70 {
		brand := string(header[8:12])
		switch brand {
		case "heic", "heix", "hevc", "hevx", "heim", "heis", "mif1", "msf1":
			return FormatHEIC, nil
		}
	}

	// Check TIFF (49 49 2A 00 or 4D 4D 00 2A)
	if n >= 4 && ((header[0] == 0x49 && header[1] == 0x49 && header[2] == 0x2A && header[3] == 0x00) ||
		(header[0] == 0x4D && header[1] == 0x4D && header[2] == 0x00 && header[3] == 0x2A)) {
		return FormatTIFF, nil
	}

	// Check GIF (GIF87a or GIF89a)
	if n >= 6 && header[0] == 0x47 && header[1] == 0x49 && header[2] == 0x46 &&
		header[3] == 0x38 && (header[4] == 0x37 || header[4] == 0x39) && header[5] == 0x61 {
		return FormatGIF, nil
	}

	// Try extension as fallback
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg":
		return FormatJPEG, nil
	case ".png":
		return FormatPNG, nil
	case ".webp":
		return FormatWebP, nil
	case ".heic", ".heif":
		return FormatHEIC, nil
	case ".tiff", ".tif":
		return FormatTIFF, nil
	case ".gif":
		return FormatGIF, nil
	}

	return FormatUnknown, nil
}

// String returns the string representation of ImageFormat
func (f ImageFormat) String() string {
	return string(f)
}
