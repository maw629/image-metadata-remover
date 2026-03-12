package metadata

import (
	"fmt"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	
	exifv3 "github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
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
// while preserving non-sensitive tags
func (h *JPEGHandler) RemoveMetadata(inputPath, outputPath string) error {
	// Read the input file
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse JPEG structure
	jmp := jis.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(inputData)
	if err != nil {
		return fmt.Errorf("failed to parse JPEG: %w", err)
	}

	sl := intfc.(*jis.SegmentList)

	// Try to get existing EXIF data
	rootIfd, _, err := sl.Exif()
	if err != nil {
		if err == exifv3.ErrNoExif {
			// No EXIF data - just copy the file
			return os.WriteFile(outputPath, inputData, 0644)
		}
		return fmt.Errorf("failed to read EXIF: %w", err)
	}

	// Build filtered EXIF with only non-sensitive tags
	filteredIb, hasPreserved, err := h.buildFilteredEXIF(rootIfd, nil)
	if err != nil {
		return fmt.Errorf("failed to filter EXIF: %w", err)
	}

	// If no tags preserved, just strip all EXIF
	if !hasPreserved {
		// Remove all EXIF and write output
		_, err := sl.DropExif()
		if err != nil {
			return fmt.Errorf("failed to drop EXIF: %w", err)
		}

		outputFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()

		if err := sl.Write(outputFile); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}

		return nil
	}

	// Set the filtered EXIF
	if err := sl.SetExif(filteredIb); err != nil {
		return fmt.Errorf("failed to set filtered EXIF: %w", err)
	}

	// Write the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if err := sl.Write(outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

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
		fmt.Println("   ✅ No sensitive metadata found - file is clean!")
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

// buildFilteredEXIF creates a new EXIF IFD builder with only non-sensitive tags
func (h *JPEGHandler) buildFilteredEXIF(rootIfd *exifv3.Ifd, rawExif []byte) (*exifv3.IfdBuilder, bool, error) {
	// Define sensitive tag IDs
	sensitiveTags := h.getSensitiveTagIDs()

	// Create a new IFD builder from existing chain
	rootIb := exifv3.NewIfdBuilderFromExistingChain(rootIfd)

	// Track if we preserved any tags
	preservedCount := 0

	// Remove sensitive tags from all IFDs
	if err := h.removeSensitiveTags(rootIb, sensitiveTags, &preservedCount); err != nil {
		return nil, false, err
	}

	return rootIb, preservedCount > 0, nil
}

// removeSensitiveTags removes sensitive tags from IFD builder recursively
func (h *JPEGHandler) removeSensitiveTags(ib *exifv3.IfdBuilder, sensitiveTags map[uint16]bool, preservedCount *int) error {
	// Delete sensitive tags by ID
	for tagID := range sensitiveTags {
		n, err := ib.DeleteAll(tagID)
		if err != nil {
			// Continue on error (tag might not exist)
			continue
		}
		// Don't count deleted tags
		_ = n
	}

	// Count remaining (preserved) tags
	tags := ib.Tags()
	*preservedCount += len(tags)

	// Handle child IFDs - we need to remove the GPS IFD entirely
	// Try to get GPS IFD (tag 0x8825)
	_, err := ib.ChildWithTagId(0x8825)
	if err == nil {
		// GPS IFD exists, delete it
		_, err := ib.DeleteAll(0x8825)
		if err != nil {
			// Continue on error
		}
	}

	// Process EXIF SubIFD if present (tag 0x8769)
	exifIb, err := ib.ChildWithTagId(0x8769)
	if err == nil {
		// Recursively remove sensitive tags from EXIF SubIFD
		if err := h.removeSensitiveTags(exifIb, sensitiveTags, preservedCount); err != nil {
			// Continue on error
		}
	}

	// Process next IFD in chain (IFD1, typically thumbnail)
	nextIb, err := ib.NextIb()
	if err == nil && nextIb != nil {
		if err := h.removeSensitiveTags(nextIb, sensitiveTags, preservedCount); err != nil {
			// Continue on error
		}
	}

	return nil
}

// getSensitiveTagIDs returns a map of sensitive EXIF tag IDs
func (h *JPEGHandler) getSensitiveTagIDs() map[uint16]bool {
	return map[uint16]bool{
		// GPS tags (all GPS tags are sensitive)
		// These are in the GPS IFD which we skip entirely
		0x0000: true, // GPSVersionID
		0x0001: true, // GPSLatitudeRef
		0x0002: true, // GPSLatitude
		0x0003: true, // GPSLongitudeRef
		0x0004: true, // GPSLongitude
		0x0005: true, // GPSAltitudeRef
		0x0006: true, // GPSAltitude
		0x0007: true, // GPSTimeStamp
		0x0008: true, // GPSSatellites
		0x0009: true, // GPSStatus
		0x000A: true, // GPSMeasureMode
		0x000B: true, // GPSDOP
		0x000C: true, // GPSSpeedRef
		0x000D: true, // GPSSpeed
		0x000E: true, // GPSTrackRef
		0x000F: true, // GPSTrack
		0x0010: true, // GPSImgDirectionRef
		0x0011: true, // GPSImgDirection
		0x0012: true, // GPSMapDatum
		0x0013: true, // GPSDestLatitudeRef
		0x0014: true, // GPSDestLatitude
		0x0015: true, // GPSDestLongitudeRef
		0x0016: true, // GPSDestLongitude
		0x0017: true, // GPSDestBearingRef
		0x0018: true, // GPSDestBearing
		0x0019: true, // GPSDestDistanceRef
		0x001A: true, // GPSDestDistance
		0x001B: true, // GPSProcessingMethod
		0x001C: true, // GPSAreaInformation
		0x001D: true, // GPSDateStamp
		0x001E: true, // GPSDifferential

		// Camera/Device information
		0x010F: true, // Make
		0x0110: true, // Model
		0x0131: true, // Software
		0xA433: true, // LensMake
		0xA434: true, // LensModel
		0xA435: true, // LensSerialNumber

		// Serial numbers
		0xA431: true, // BodySerialNumber
		0x927C: true, // MakerNote (contains serial numbers)

		// Creator/Copyright information
		0x013B: true, // Artist
		0x8298: true, // Copyright
		0x9286: true, // UserComment
		0x010E: true, // ImageDescription

		// Windows-specific metadata
		0x9C9B: true, // XPTitle
		0x9C9C: true, // XPComment
		0x9C9D: true, // XPAuthor
		0x9C9E: true, // XPKeywords
		0x9C9F: true, // XPSubject

		// Timestamps (making these sensitive too)
		0x0132: true, // DateTime (Modify Date)
		0x9003: true, // DateTimeOriginal
		0x9004: true, // DateTimeDigitized
		0x9290: true, // SubSecTime
		0x9291: true, // SubSecTimeOriginal
		0x9292: true, // SubSecTimeDigitized
	}
}

// Ensure JPEGHandler implements Handler interface
var _ Handler = (*JPEGHandler)(nil)
