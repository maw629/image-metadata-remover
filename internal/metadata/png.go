package metadata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// PNGHandler handles PNG image metadata
type PNGHandler struct{}

// NewPNGHandler creates a new PNG metadata handler
func NewPNGHandler() *PNGHandler {
	return &PNGHandler{}
}

// SupportsFormat returns true if the format is PNG
func (h *PNGHandler) SupportsFormat(format string) bool {
	return format == "png"
}

// PNG chunk structure
type pngChunk struct {
	Length uint32
	Type   [4]byte
	Data   []byte
	CRC    uint32
}

// PNG file signature
var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// Sensitive text chunk keys to remove
var sensitiveTextKeys = map[string]bool{
	"Author":      true,
	"Comment":     true,
	"Copyright":   true,
	"Description": true,
	"Software":    true,
	"Source":      true,
	"Disclaimer":  true,
	"Warning":     true,
	"Title":       true,
	"Creation Time": true,
	// Common Adobe XMP keys
	"XML:com.adobe.xmp": true,
	// Common GIMP keys
	"gimp::manifest": true,
	"gimp::thumb-uri": true,
}

// PreviewMetadata displays metadata that would be removed
func (h *PNGHandler) PreviewMetadata(inputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Verify PNG signature
	sig := make([]byte, 8)
	if _, err := io.ReadFull(file, sig); err != nil {
		return fmt.Errorf("failed to read PNG signature: %w", err)
	}
	if !bytes.Equal(sig, pngSignature) {
		return fmt.Errorf("invalid PNG signature")
	}

	toRemove := []string{}
	toPreserve := []string{}
	var exifData *exif.Exif

	// Read all chunks
	for {
		chunk, err := h.readChunk(file)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		chunkType := string(chunk.Type[:])

		switch chunkType {
		case "tEXt", "zTXt", "iTXt":
			// Parse text chunk
			key, value := h.parseTextChunk(chunk.Data, chunkType)
			if h.isSensitiveTextKey(key) {
				toRemove = append(toRemove, fmt.Sprintf("%s.%s = %s", chunkType, key, value))
			} else {
				toPreserve = append(toPreserve, fmt.Sprintf("%s.%s = %s", chunkType, key, value))
			}

		case "eXIf":
			// Parse EXIF data
			exifData, err = exif.Decode(bytes.NewReader(chunk.Data))
			if err == nil {
				// We'll handle EXIF separately below
			}

		case "tIME":
			// Timestamp chunk - always remove
			if len(chunk.Data) >= 7 {
				year := binary.BigEndian.Uint16(chunk.Data[0:2])
				month := chunk.Data[2]
				day := chunk.Data[3]
				hour := chunk.Data[4]
				minute := chunk.Data[5]
				second := chunk.Data[6]
				toRemove = append(toRemove, fmt.Sprintf("tIME = %04d-%02d-%02d %02d:%02d:%02d",
					year, month, day, hour, minute, second))
			}

		case "IHDR":
			toPreserve = append(toPreserve, "IHDR (Image Header)")
		case "PLTE":
			toPreserve = append(toPreserve, "PLTE (Palette)")
		case "IDAT":
			// Don't list every IDAT chunk
			if len(toPreserve) == 0 || toPreserve[len(toPreserve)-1] != "IDAT (Image Data)" {
				toPreserve = append(toPreserve, "IDAT (Image Data)")
			}
		case "gAMA":
			toPreserve = append(toPreserve, "gAMA (Gamma)")
		case "cHRM":
			toPreserve = append(toPreserve, "cHRM (Chromaticity)")
		case "sRGB":
			toPreserve = append(toPreserve, "sRGB (Color Space)")
		case "iCCP":
			toPreserve = append(toPreserve, "iCCP (Color Profile)")
		case "pHYs":
			toPreserve = append(toPreserve, "pHYs (Physical Dimensions)")
		case "sBIT":
			toPreserve = append(toPreserve, "sBIT (Significant Bits)")
		case "bKGD":
			toPreserve = append(toPreserve, "bKGD (Background Color)")
		case "tRNS":
			toPreserve = append(toPreserve, "tRNS (Transparency)")
		case "IEND":
			// End chunk, don't list it
		default:
			// Unknown chunk type
			if h.isAncillary(chunkType) {
				toPreserve = append(toPreserve, fmt.Sprintf("%s (Unknown Ancillary)", chunkType))
			}
		}
	}

	// Handle EXIF data if present
	if exifData != nil {
		sensitiveTags := getSensitiveTagNames()
		
		walker := &pngExifCollector{
			sensitiveTags: sensitiveTags,
			toRemove:      &toRemove,
			toPreserve:    &toPreserve,
		}
		
		exifData.Walk(walker)
	}

	// Sort lists
	sort.Strings(toRemove)
	sort.Strings(toPreserve)

	// Display results
	fmt.Printf("\n=== Metadata Preview for: %s ===\n", inputPath)
	fmt.Printf("Format: PNG\n\n")

	if len(toRemove) > 0 {
		fmt.Println("⚠️  Metadata to be REMOVED (sensitive):")
		for _, item := range toRemove {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(toPreserve) > 0 {
		fmt.Println("✓ Metadata to be PRESERVED:")
		for _, item := range toPreserve {
			fmt.Printf("  - %s\n", item)
		}
		fmt.Println()
	}

	if len(toRemove) == 0 && len(toPreserve) == 0 {
		fmt.Println("ℹ️  No metadata found in this PNG file")
	}

	return nil
}

// RemoveMetadata removes sensitive metadata from PNG files
func (h *PNGHandler) RemoveMetadata(inputPath, outputPath string) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer inputFile.Close()

	// Verify PNG signature
	sig := make([]byte, 8)
	if _, err := io.ReadFull(inputFile, sig); err != nil {
		return fmt.Errorf("failed to read PNG signature: %w", err)
	}
	if !bytes.Equal(sig, pngSignature) {
		return fmt.Errorf("invalid PNG signature")
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Write PNG signature
	if _, err := outputFile.Write(pngSignature); err != nil {
		return fmt.Errorf("failed to write PNG signature: %w", err)
	}

	// Process chunks
	for {
		chunk, err := h.readChunk(inputFile)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		chunkType := string(chunk.Type[:])
		shouldKeep := true

		switch chunkType {
		case "tEXt", "zTXt", "iTXt":
			// Check if text chunk has sensitive key
			key, _ := h.parseTextChunk(chunk.Data, chunkType)
			shouldKeep = !h.isSensitiveTextKey(key)

		case "eXIf":
			// Handle EXIF - filter sensitive tags
			chunk, shouldKeep = h.filterEXIFChunk(chunk)

		case "tIME":
			// Remove timestamp chunk
			shouldKeep = false
		}

		// Write chunk if we're keeping it
		if shouldKeep {
			if err := h.writeChunk(outputFile, chunk); err != nil {
				return fmt.Errorf("failed to write chunk: %w", err)
			}
		}
	}

	return nil
}

// readChunk reads a single PNG chunk
func (h *PNGHandler) readChunk(r io.Reader) (*pngChunk, error) {
	chunk := &pngChunk{}

	// Read length
	if err := binary.Read(r, binary.BigEndian, &chunk.Length); err != nil {
		return nil, err
	}

	// Read type
	if _, err := io.ReadFull(r, chunk.Type[:]); err != nil {
		return nil, err
	}

	// Read data
	chunk.Data = make([]byte, chunk.Length)
	if chunk.Length > 0 {
		if _, err := io.ReadFull(r, chunk.Data); err != nil {
			return nil, err
		}
	}

	// Read CRC
	if err := binary.Read(r, binary.BigEndian, &chunk.CRC); err != nil {
		return nil, err
	}

	// Verify CRC
	crc := crc32.NewIEEE()
	crc.Write(chunk.Type[:])
	crc.Write(chunk.Data)
	if crc.Sum32() != chunk.CRC {
		return nil, fmt.Errorf("CRC mismatch for chunk %s", string(chunk.Type[:]))
	}

	return chunk, nil
}

// writeChunk writes a single PNG chunk
func (h *PNGHandler) writeChunk(w io.Writer, chunk *pngChunk) error {
	// Recalculate CRC
	crc := crc32.NewIEEE()
	crc.Write(chunk.Type[:])
	crc.Write(chunk.Data)
	chunk.CRC = crc.Sum32()

	// Update length
	chunk.Length = uint32(len(chunk.Data))

	// Write length
	if err := binary.Write(w, binary.BigEndian, chunk.Length); err != nil {
		return err
	}

	// Write type
	if _, err := w.Write(chunk.Type[:]); err != nil {
		return err
	}

	// Write data
	if len(chunk.Data) > 0 {
		if _, err := w.Write(chunk.Data); err != nil {
			return err
		}
	}

	// Write CRC
	if err := binary.Write(w, binary.BigEndian, chunk.CRC); err != nil {
		return err
	}

	return nil
}

// parseTextChunk parses tEXt, zTXt, or iTXt chunk data
func (h *PNGHandler) parseTextChunk(data []byte, chunkType string) (key, value string) {
	// All text chunks start with null-terminated keyword
	nullPos := bytes.IndexByte(data, 0)
	if nullPos == -1 {
		return "", ""
	}

	key = string(data[:nullPos])
	
	// For simplicity, we'll just show first 50 chars of value
	// (proper parsing would need to handle compression for zTXt, etc.)
	if nullPos+1 < len(data) {
		valueBytes := data[nullPos+1:]
		if len(valueBytes) > 50 {
			value = fmt.Sprintf("%s... (%d bytes)", string(valueBytes[:50]), len(valueBytes))
		} else {
			value = string(valueBytes)
		}
	}

	return key, value
}

// filterEXIFChunk filters sensitive EXIF tags from eXIf chunk
func (h *PNGHandler) filterEXIFChunk(chunk *pngChunk) (*pngChunk, bool) {
	// Parse EXIF data
	exifData, err := exif.Decode(bytes.NewReader(chunk.Data))
	if err != nil {
		// If we can't parse it, remove it to be safe
		return chunk, false
	}

	sensitiveTags := getSensitiveTagNames()
	hasAnyTag := false
	hasSensitiveTag := false
	
	// Check tags
	walker := &pngExifChecker{
		sensitiveTags:   sensitiveTags,
		hasAnyTag:       &hasAnyTag,
		hasSensitiveTag: &hasSensitiveTag,
	}
	
	exifData.Walk(walker)

	// If no non-sensitive tags, remove the entire chunk
	if !hasAnyTag {
		return chunk, false
	}

	// If any sensitive data found, remove chunk (TODO: implement selective preservation)
	if hasSensitiveTag {
		return chunk, false
	}

	return chunk, true
}

// isSensitiveTextKey checks if a text chunk key contains sensitive data
func (h *PNGHandler) isSensitiveTextKey(key string) bool {
	// Direct match
	if sensitiveTextKeys[key] {
		return true
	}

	// Case-insensitive check for common patterns
	lowerKey := strings.ToLower(key)
	sensitivePatterns := []string{
		"author", "creator", "artist",
		"copyright", "owner",
		"comment", "description",
		"software", "source",
		"date", "time", "timestamp",
		"gps", "location",
		"xmp", "adobe",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}

	return false
}

// isAncillary checks if a chunk type is ancillary (can be safely ignored)
func (h *PNGHandler) isAncillary(chunkType string) bool {
	// Ancillary chunks have bit 5 of first byte set (lowercase first letter)
	if len(chunkType) == 0 {
		return false
	}
	firstChar := chunkType[0]
	return firstChar >= 'a' && firstChar <= 'z'
}

// pngExifCollector implements the exif.Walker interface for metadata preview
type pngExifCollector struct {
	sensitiveTags map[string]bool
	toRemove      *[]string
	toPreserve    *[]string
}

func (w *pngExifCollector) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagName := string(name)
	value := formatTagValue(tag)
	
	if w.sensitiveTags[tagName] {
		*w.toRemove = append(*w.toRemove, fmt.Sprintf("eXIf.%s = %s", tagName, value))
	} else {
		*w.toPreserve = append(*w.toPreserve, fmt.Sprintf("eXIf.%s = %s", tagName, value))
	}
	return nil
}

// pngExifChecker implements the exif.Walker interface for checking sensitive tags
type pngExifChecker struct {
	sensitiveTags   map[string]bool
	hasAnyTag       *bool
	hasSensitiveTag *bool
}

func (w *pngExifChecker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	tagName := string(name)
	*w.hasAnyTag = true
	
	if w.sensitiveTags[tagName] {
		*w.hasSensitiveTag = true
	}
	return nil
}
