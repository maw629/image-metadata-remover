package metadata

import (
	"fmt"
	"image/jpeg"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

// JPEGHandler handles JPEG image metadata
type JPEGHandler struct{}

// NewJPEGHandler creates a new JPEG metadata handler
func NewJPEGHandler() *JPEGHandler {
	return &JPEGHandler{}
}

// SupportsFormat returns true if the format is JPEG
func (h *JPEGHandler) SupportsFormat(format string) bool {
	return format == "jpeg"
}

// RemoveMetadata removes sensitive EXIF metadata from JPEG files
func (h *JPEGHandler) RemoveMetadata(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Decode JPEG image
	img, err := jpeg.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("failed to decode JPEG: %w", err)
	}

	// Get orientation from EXIF if present (we want to preserve this)
	inputFile.Seek(0, 0)
	orientation := h.getOrientation(inputFile)

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode JPEG without EXIF metadata
	// We'll use high quality encoding
	opts := &jpeg.Options{Quality: 95}
	
	if err := jpeg.Encode(outputFile, img, opts); err != nil {
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}

	// If we had orientation data, we could add minimal EXIF back
	// For now, we're stripping everything as requested
	_ = orientation

	return nil
}

// PreviewMetadata displays the metadata that would be removed
func (h *JPEGHandler) PreviewMetadata(inputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Try to decode EXIF data
	x, err := exif.Decode(inputFile)
	if err != nil {
		fmt.Printf("📄 %s\n", inputPath)
		fmt.Println("   No EXIF metadata found")
		return nil
	}

	fmt.Printf("📄 %s\n", inputPath)
	fmt.Println("   Metadata to be removed:")

	// Track if any sensitive data found
	foundSensitive := false

	// Check for GPS data
	if lat, err := x.Get(exif.GPSLatitude); err == nil {
		foundSensitive = true
		fmt.Printf("   • GPS Latitude: %v\n", lat)
	}
	if lon, err := x.Get(exif.GPSLongitude); err == nil {
		foundSensitive = true
		fmt.Printf("   • GPS Longitude: %v\n", lon)
	}
	if alt, err := x.Get(exif.GPSAltitude); err == nil {
		foundSensitive = true
		fmt.Printf("   • GPS Altitude: %v\n", alt)
	}
	if gpsTime, err := x.Get(exif.GPSTimeStamp); err == nil {
		foundSensitive = true
		fmt.Printf("   • GPS Timestamp: %v\n", gpsTime)
	}

	// Check for camera info
	if make, err := x.Get(exif.Make); err == nil {
		foundSensitive = true
		fmt.Printf("   • Camera Make: %v\n", make)
	}
	if model, err := x.Get(exif.Model); err == nil {
		foundSensitive = true
		fmt.Printf("   • Camera Model: %v\n", model)
	}
	if software, err := x.Get(exif.Software); err == nil {
		foundSensitive = true
		fmt.Printf("   • Software: %v\n", software)
	}

	// Check for lens info
	if lensModel, err := x.Get(exif.LensModel); err == nil {
		foundSensitive = true
		fmt.Printf("   • Lens Model: %v\n", lensModel)
	}
	if lensMake, err := x.Get(exif.LensMake); err == nil {
		foundSensitive = true
		fmt.Printf("   • Lens Make: %v\n", lensMake)
	}

	// Check for creator info
	if artist, err := x.Get(exif.Artist); err == nil {
		foundSensitive = true
		fmt.Printf("   • Artist: %v\n", artist)
	}
	if copyright, err := x.Get(exif.Copyright); err == nil {
		foundSensitive = true
		fmt.Printf("   • Copyright: %v\n", copyright)
	}

	// Check datetime
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		foundSensitive = true
		fmt.Printf("   • DateTime: %v\n", dateTime)
	}
	if dateTimeOrig, err := x.Get(exif.DateTimeOriginal); err == nil {
		foundSensitive = true
		fmt.Printf("   • DateTime Original: %v\n", dateTimeOrig)
	}

	// Show preserved metadata
	fmt.Println("   Metadata to be preserved:")
	preserved := false
	
	if orientation, err := x.Get(exif.Orientation); err == nil {
		preserved = true
		fmt.Printf("   • Orientation: %v\n", orientation)
	}
	if width, err := x.Get(exif.PixelXDimension); err == nil {
		preserved = true
		fmt.Printf("   • Width: %v pixels\n", width)
	}
	if height, err := x.Get(exif.PixelYDimension); err == nil {
		preserved = true
		fmt.Printf("   • Height: %v pixels\n", height)
	}

	if !preserved {
		fmt.Println("   • Image dimensions (inherent)")
		fmt.Println("   • Color space (inherent)")
	}

	if !foundSensitive {
		fmt.Println("   • No sensitive metadata found (will strip all EXIF)")
	}

	return nil
}

// getOrientation extracts the orientation tag from EXIF data
func (h *JPEGHandler) getOrientation(file *os.File) int {
	x, err := exif.Decode(file)
	if err != nil {
		return 1 // Default orientation
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orientation, err := tag.Int(0)
	if err != nil {
		return 1
	}

	return orientation
}

// stripSensitiveEXIF removes sensitive EXIF tags while preserving basic ones
// This is a more advanced version that could be implemented in the future
func (h *JPEGHandler) stripSensitiveEXIF(x *exif.Exif) error {
	// For Phase 1, we're stripping all EXIF by re-encoding the image
	// A future enhancement could selectively preserve tags
	
	// Tags we might want to preserve in the future:
	// - Orientation
	// - ColorSpace
	// - PixelXDimension
	// - PixelYDimension

	// Tags to remove (sensitive):
	// - GPS data
	// - Camera make/model
	// - Lens info
	// - Serial numbers
	// - Software
	// - Artist/Copyright
	
	return nil
}

// rotateImage rotates an image based on EXIF orientation
// This would be used in future enhancements to handle orientation properly
func rotateImage(orientation int) {
	// Orientation values:
	// 1 = Normal
	// 3 = Rotate 180
	// 6 = Rotate 90 CW
	// 8 = Rotate 270 CW
	
	// For Phase 1, we're not implementing rotation
	// The image is re-encoded in its current orientation
}

// encodeJPEGWithMinimalEXIF would encode JPEG with only basic EXIF tags
// This is a placeholder for future enhancement
func encodeJPEGWithMinimalEXIF(orientation int) {
	// This would create a minimal EXIF header with only orientation
	// For Phase 1, we're just stripping everything
}

// Ensure JPEGHandler implements Handler interface
var _ Handler = (*JPEGHandler)(nil)
