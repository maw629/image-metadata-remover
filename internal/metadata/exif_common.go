package metadata

import (
	"fmt"

	"github.com/rwcarlsen/goexif/tiff"
)

// getSensitiveTagIDs returns a map of sensitive EXIF tag IDs for removal
// This is used during actual metadata processing to identify which tags to delete
func getSensitiveTagIDs() map[uint16]bool {
	return map[uint16]bool{
		// GPS tags (all GPS tags are sensitive)
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
		0x001F: true, // GPSHPositioningError

		// Device information
		0x010F: true, // Make
		0x0110: true, // Model
		0x0131: true, // Software
		0xA433: true, // LensMake
		0xA434: true, // LensModel
		0xA435: true, // LensSerialNumber

		// Serial numbers and identifiers
		0xA431: true, // BodySerialNumber
		0xA420: true, // ImageUniqueID
		0x927C: true, // MakerNote (contains serial numbers and proprietary camera data)

		// GPS IFD pointer (must remove when GPS IFD is deleted)
		0x8825: true, // GPSInfoIFDPointer

		// Thumbnail pointers (can become corrupted during EXIF rebuild)
		0x0201: true, // JPEGInterchangeFormat (thumbnail offset)
		0x0202: true, // JPEGInterchangeFormatLength (thumbnail size)

		// Creator/Copyright information
		0x013B: true, // Artist
		0x8298: true, // Copyright
		0x9286: true, // UserComment
		0x010E: true, // ImageDescription

		// Windows XP tags
		0x9C9B: true, // XPTitle
		0x9C9C: true, // XPComment
		0x9C9D: true, // XPAuthor
		0x9C9E: true, // XPKeywords
		0x9C9F: true, // XPSubject

		// Timestamps
		0x0132: true, // DateTime
		0x9003: true, // DateTimeOriginal
		0x9004: true, // DateTimeDigitized
		0x9290: true, // SubSecTime
		0x9291: true, // SubSecTimeOriginal
		0x9292: true, // SubSecTimeDigitized
	}
}

// getSensitiveTagNames returns a map of sensitive EXIF tag names for display
// This is used during dry-run mode to show which tags will be removed
func getSensitiveTagNames() map[string]bool {
	return map[string]bool{
		// GPS data
		"GPSVersionID":          true,
		"GPSLatitudeRef":        true,
		"GPSLatitude":           true,
		"GPSLongitudeRef":       true,
		"GPSLongitude":          true,
		"GPSAltitudeRef":        true,
		"GPSAltitude":           true,
		"GPSTimeStamp":          true,
		"GPSSatellites":         true,
		"GPSStatus":             true,
		"GPSMeasureMode":        true,
		"GPSDOP":                true,
		"GPSSpeedRef":           true,
		"GPSSpeed":              true,
		"GPSTrackRef":           true,
		"GPSTrack":              true,
		"GPSImgDirectionRef":    true,
		"GPSImgDirection":       true,
		"GPSMapDatum":           true,
		"GPSDestLatitudeRef":    true,
		"GPSDestLatitude":       true,
		"GPSDestLongitudeRef":   true,
		"GPSDestLongitude":      true,
		"GPSDestBearingRef":     true,
		"GPSDestBearing":        true,
		"GPSDestDistanceRef":    true,
		"GPSDestDistance":       true,
		"GPSProcessingMethod":   true,
		"GPSAreaInformation":    true,
		"GPSDateStamp":          true,
		"GPSDifferential":       true,
		"GPSHPositioningError":  true,

		// Device information
		"Make":                true,
		"Model":               true,
		"Software":            true,
		"LensMake":            true,
		"LensModel":           true,
		"LensSerialNumber":    true,
		"SubSecTime":          true,
		"SubSecTimeOriginal":  true,
		"SubSecTimeDigitized": true,

		// Timestamps
		"DateTime":          true,
		"DateTimeOriginal":  true,
		"DateTimeDigitized": true,

		// Creator/Copyright
		"Artist":           true,
		"Copyright":        true,
		"UserComment":      true,
		"ImageDescription": true,

		// Windows XP tags
		"XPTitle":    true,
		"XPComment":  true,
		"XPAuthor":   true,
		"XPKeywords": true,
		"XPSubject":  true,

		// Serial numbers
		"BodySerialNumber": true,
		"ImageUniqueID":    true,

		// Proprietary data
		"MakerNote": true,

		// Pointers that need removal
		"GPSInfoIFDPointer":              true,
		"ThumbJPEGInterchangeFormat":     true,
		"ThumbJPEGInterchangeFormatLength": true,
	}
}

// formatTagValue formats a tag value for display in dry-run mode
func formatTagValue(tag *tiff.Tag) string {
	val, err := tag.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("%v", tag)
	}
	return string(val)
}
