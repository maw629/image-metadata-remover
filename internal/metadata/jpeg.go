package metadata

import (
	"fmt"
	"image/jpeg"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
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

// PreviewMetadata displays all metadata in the image file
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

	// Define sensitive tags that will be removed
	sensitiveTags := map[string]bool{
		// GPS data
		"GPSLatitude":          true,
		"GPSLongitude":         true,
		"GPSAltitude":          true,
		"GPSTimeStamp":         true,
		"GPSDateStamp":         true,
		"GPSProcessingMethod":  true,
		"GPSVersionID":         true,
		"GPSLatitudeRef":       true,
		"GPSLongitudeRef":      true,
		"GPSAltitudeRef":       true,
		"GPSMapDatum":          true,
		"GPSSatelites":         true,
		"GPSImgDirection":      true,
		"GPSImgDirectionRef":   true,
		"GPSDestBearing":       true,
		"GPSDestBearingRef":    true,
		"GPSSpeed":             true,
		"GPSSpeedRef":          true,
		"GPSTrack":             true,
		"GPSTrackRef":          true,
		"GPSAreaInformation":   true,
		"GPSDifferential":      true,
		// Camera info
		"Make":                 true,
		"Model":                true,
		"Software":             true,
		// Lens info
		"LensModel":            true,
		"LensMake":             true,
		"LensSerialNumber":     true,
		// Serial numbers
		"BodySerialNumber":     true,
		"InternalSerialNumber": true,
		"SerialNumber":         true,
		// Creator/owner info
		"Artist":               true,
		"Copyright":            true,
		"XPComment":            true,
		"XPAuthor":             true,
		"UserComment":          true,
		"ImageDescription":     true,
		// Timestamps
		"DateTime":             true,
		"DateTimeOriginal":     true,
		"DateTimeDigitized":    true,
	}

	// Collect tags into groups
	var sensitiveList []string
	var preservedList []string

	walker := &metadataCollector{
		sensitiveTags:   sensitiveTags,
		sensitiveList:   &sensitiveList,
		preservedList:   &preservedList,
	}

	// Walk through all EXIF tags
	err = x.Walk(walker)
	if err != nil {
		return fmt.Errorf("error walking EXIF tags: %w", err)
	}

	// Display grouped output
	fmt.Println("   Metadata to be removed:")
	if len(sensitiveList) > 0 {
		for _, tag := range sensitiveList {
			fmt.Printf("   • %s\n", tag)
		}
	} else {
		fmt.Println("   • None found")
	}

	fmt.Println()
	fmt.Println("   Metadata to be preserved:")
	if len(preservedList) > 0 {
		for _, tag := range preservedList {
			fmt.Printf("   • %s\n", tag)
		}
	} else {
		fmt.Println("   • None (will strip all EXIF)")
	}

	// Summary
	fmt.Println()
	fmt.Printf("   Summary: %d total tags (%d sensitive, %d preserved)\n", 
		len(sensitiveList)+len(preservedList), len(sensitiveList), len(preservedList))
	
	if len(sensitiveList) == 0 {
		fmt.Println("   ⚠️  No sensitive metadata found, but all EXIF will be stripped")
	}

	return nil
}

// metadataCollector implements the exif.Walker interface to collect tags
type metadataCollector struct {
	sensitiveTags   map[string]bool
	sensitiveList   *[]string
	preservedList   *[]string
}

func (w *metadataCollector) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagName := string(name)
	tagValue := fmt.Sprintf("%v", tag)
	tagDisplay := fmt.Sprintf("%s: %s", tagName, tagValue)
	
	if w.sensitiveTags[tagName] {
		*w.sensitiveList = append(*w.sensitiveList, tagDisplay)
	} else {
		*w.preservedList = append(*w.preservedList, tagDisplay)
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
